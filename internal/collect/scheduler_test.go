// SPDX-License-Identifier: Apache-2.0 OR MIT
// Copyright (c) 2026 Sebastien Rousseau

package collect

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// manualTicker mimics time.Ticker: a one-slot buffer that drops ticks when
// full.
type manualTicker struct{ ch chan time.Time }

func newManualTicker() *manualTicker        { return &manualTicker{ch: make(chan time.Time, 1)} }
func (m *manualTicker) C() <-chan time.Time { return m.ch }
func (m *manualTicker) Stop()               {}

func (m *manualTicker) tick() bool {
	select {
	case m.ch <- time.Now():
		return true
	default:
		return false
	}
}

// fakeCollector reports each call on started, then blocks until release
// (when gated) or ctx is done.
type fakeCollector struct {
	name     string
	interval time.Duration
	gated    bool
	err      error
	started  chan struct{}
	release  chan struct{}
	calls    atomic.Int64
}

func newFake(gated bool) *fakeCollector {
	return &fakeCollector{
		name: "fake", interval: 20 * time.Millisecond, gated: gated,
		started: make(chan struct{}, 16), release: make(chan struct{}),
	}
}

func (f *fakeCollector) Name() string            { return f.name }
func (f *fakeCollector) Interval() time.Duration { return f.interval }
func (f *fakeCollector) Caps() Capabilities      { return Capabilities{{Metric: "n", Supported: true}} }

func (f *fakeCollector) Collect(ctx context.Context, dst *Sample) error {
	n := f.calls.Add(1)
	f.started <- struct{}{}
	if f.gated {
		select {
		case <-f.release:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	dst.Add("n", "", float64(n))
	return f.err
}

type harness struct {
	t      *testing.T
	sched  *Scheduler
	ticker *manualTicker
	out    chan Result
	cancel context.CancelFunc
	done   chan struct{}
}

func start(t *testing.T, c Collector, outCap int) *harness {
	t.Helper()
	out := make(chan Result, outCap)
	s, err := NewScheduler(out, c)
	if err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, sched: s, ticker: newManualTicker(), out: out, done: make(chan struct{})}
	s.newTicker = func(time.Duration) Ticker { return h.ticker }
	ctx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel
	go func() { s.Run(ctx); close(h.done) }()
	t.Cleanup(h.stop)
	return h
}

func (h *harness) stop() {
	h.cancel()
	select {
	case <-h.done:
	case <-time.After(2 * time.Second):
		h.t.Fatal("Run did not return after cancel")
	}
}

func (h *harness) stats() Stats {
	st, _ := h.sched.Stats("fake")
	return st
}

func wait[T any](t *testing.T, ch <-chan T, what string) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for %s", what)
	}
	var zero T
	return zero
}

func TestSchedulerSkipsTicksDuringOverrun(t *testing.T) {
	f := newFake(true)
	h := start(t, f, 0)

	h.ticker.tick()
	wait(t, f.started, "first collect")
	if !h.ticker.tick() {
		t.Fatal("tick during overrun was not buffered")
	}
	if h.ticker.tick() {
		t.Fatal("manual ticker should drop a tick when its slot is full")
	}
	f.release <- struct{}{}
	wait(t, h.out, "first result")
	if st := h.stats(); st.Runs != 1 || st.Skipped != 1 {
		t.Fatalf("after overrun: %+v, want Runs=1 Skipped=1", st)
	}

	// No queued tick may start another collection.
	select {
	case <-f.started:
		t.Fatal("a queued tick started a second collection")
	case <-time.After(50 * time.Millisecond):
	}

	h.ticker.tick()
	wait(t, f.started, "second collect")
	f.release <- struct{}{}
	wait(t, h.out, "second result")
	if st := h.stats(); st.Runs != 2 || st.Skipped != 1 || f.calls.Load() != 2 {
		t.Fatalf("after second tick: %+v calls=%d, want Runs=2 Skipped=1 calls=2", st, f.calls.Load())
	}
}

func TestSchedulerPublishesOwnedCopies(t *testing.T) {
	f := newFake(false)
	h := start(t, f, 0)
	h.ticker.tick()
	first := wait(t, h.out, "first result")
	h.ticker.tick()
	second := wait(t, h.out, "second result")
	if first.Collector != "fake" || first.Sample.Metrics[0].Value != 1 {
		t.Fatalf("first result = %+v, want fake n=1", first)
	}
	if second.Sample.Metrics[0].Value != 2 {
		t.Fatalf("second result = %+v, want n=2", second)
	}
}

func TestSchedulerDropsWhenReceiverStalls(t *testing.T) {
	f := newFake(false)
	h := start(t, f, 0) // unbuffered and nobody reading
	h.ticker.tick()
	deadline := time.Now().Add(2 * time.Second)
	for h.stats().Dropped == 0 {
		if time.Now().After(deadline) {
			t.Fatalf("result never dropped: %+v", h.stats())
		}
		time.Sleep(5 * time.Millisecond)
	}
	// The collector is not wedged: the next tick still produces a result.
	h.ticker.tick()
	if r := wait(t, h.out, "result after drop"); r.Sample.Metrics[0].Value != 2 {
		t.Fatalf("result after drop = %+v, want n=2", r)
	}
}

func TestSchedulerCountsErrors(t *testing.T) {
	f := newFake(false)
	f.err = errors.New("boom")
	h := start(t, f, 1)
	h.ticker.tick()
	r := wait(t, h.out, "result")
	if !errors.Is(r.Err, f.err) || h.stats().Errors != 1 {
		t.Fatalf("err=%v stats=%+v, want boom and Errors=1", r.Err, h.stats())
	}
}

func TestSchedulerCancelDuringCollect(t *testing.T) {
	f := newFake(true)
	h := start(t, f, 0)
	h.ticker.tick()
	wait(t, f.started, "collect")
	h.stop() // fails the test if Run does not return
}

func TestSchedulerCancelDuringPublish(t *testing.T) {
	f := newFake(false)
	f.interval = time.Hour // publish would wait an hour without cancellation
	h := start(t, f, 0)
	h.ticker.tick()
	wait(t, f.started, "collect")
	h.stop()
}

func TestSchedulerRealTicker(t *testing.T) {
	f := newFake(false)
	f.interval = 5 * time.Millisecond
	out := make(chan Result, 4)
	s, err := NewScheduler(out, f)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { s.Run(ctx); close(done) }()
	wait(t, out, "first real tick")
	wait(t, out, "second real tick")
	cancel()
	wait(t, done, "Run to return")
}

func TestNewSchedulerValidates(t *testing.T) {
	out := make(chan Result)
	dup := newFake(false)
	zero := newFake(false)
	zero.interval = 0
	cases := map[string]struct {
		out chan Result
		cs  []Collector
	}{
		"nil out":       {nil, []Collector{newFake(false)}},
		"nil collector": {out, []Collector{nil}},
		"duplicate":     {out, []Collector{dup, dup}},
		"zero interval": {out, []Collector{zero}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := NewScheduler(tc.out, tc.cs...); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
	if _, ok := (&Scheduler{}).Stats("missing"); ok {
		t.Fatal("Stats found an unknown collector")
	}
}
