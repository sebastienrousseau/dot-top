// SPDX-License-Identifier: Apache-2.0 OR MIT
// Copyright (c) 2026 Sebastien Rousseau

package collect

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Result is one published measurement. Sample is a copy the receiver owns.
type Result struct {
	Collector string
	Sample    Sample
	Err       error
	Took      time.Duration
}

// Stats counts what happened to one collector's ticks.
type Stats struct {
	Runs    uint64 // Collect calls that returned
	Errors  uint64 // Collect calls that returned an error
	Skipped uint64 // ticks discarded because the previous Collect overran
	Dropped uint64 // results discarded because the receiver did not take them in time
}

// Ticker is the tick source a Scheduler uses; time.Ticker satisfies it
// through the default factory. Tests inject a manual one.
type Ticker interface {
	C() <-chan time.Time
	Stop()
}

type realTicker struct{ t *time.Ticker }

func (r realTicker) C() <-chan time.Time { return r.t.C }
func (r realTicker) Stop()               { r.t.Stop() }

func newRealTicker(d time.Duration) Ticker { return realTicker{time.NewTicker(d)} }

// Scheduler runs one goroutine per Collector.
//
// Scheduling policy:
//   - Collect runs on each tick, not at start; the first result arrives one
//     interval after Run starts.
//   - Overrun: a tick that arrives while Collect is still running is
//     discarded and counted in Stats.Skipped; ticks never queue, so a slow
//     collector runs at most once per interval and never catches up in a
//     burst. (The runtime ticker also coalesces ticks beyond its one-slot
//     buffer; those are not counted, so Skipped is a lower bound.)
//   - Publishing: each result is sent on the output channel. If the
//     receiver does not take it within one interval, it is discarded and
//     counted in Stats.Dropped, so a stalled receiver never blocks a
//     collector for longer than one interval.
//   - Cancellation: Run returns after ctx is done and every collector
//     goroutine has exited. The ctx is passed to Collect, which should
//     return promptly when it is cancelled.
type Scheduler struct {
	collectors []Collector
	out        chan<- Result
	newTicker  func(time.Duration) Ticker
	stats      map[string]*counters
}

type counters struct{ runs, errors, skipped, dropped atomic.Uint64 }

// NewScheduler validates the collectors and returns a Scheduler that
// publishes to out. Names must be unique and intervals positive.
func NewScheduler(out chan<- Result, cs ...Collector) (*Scheduler, error) {
	if out == nil {
		return nil, errors.New("collect: nil output channel")
	}
	s := &Scheduler{out: out, newTicker: newRealTicker, stats: make(map[string]*counters, len(cs))}
	for _, c := range cs {
		if err := s.add(c); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *Scheduler) add(c Collector) error {
	if c == nil {
		return errors.New("collect: nil collector")
	}
	name := c.Name()
	if _, dup := s.stats[name]; dup {
		return fmt.Errorf("collect: duplicate collector name %q", name)
	}
	if c.Interval() <= 0 {
		return fmt.Errorf("collect: collector %q has non-positive interval %v", name, c.Interval())
	}
	s.stats[name] = &counters{}
	s.collectors = append(s.collectors, c)
	return nil
}

// Stats returns a snapshot of one collector's counters.
func (s *Scheduler) Stats(name string) (Stats, bool) {
	c, ok := s.stats[name]
	if !ok {
		return Stats{}, false
	}
	return Stats{
		Runs:    c.runs.Load(),
		Errors:  c.errors.Load(),
		Skipped: c.skipped.Load(),
		Dropped: c.dropped.Load(),
	}, true
}

// Run starts every collector and blocks until ctx is done and all of them
// have stopped.
func (s *Scheduler) Run(ctx context.Context) {
	var wg sync.WaitGroup
	for _, c := range s.collectors {
		wg.Go(func() { s.loop(ctx, c) })
	}
	wg.Wait()
}

func (s *Scheduler) loop(ctx context.Context, c Collector) {
	interval := c.Interval()
	tick := s.newTicker(interval)
	defer tick.Stop()
	timer := time.NewTimer(interval)
	timer.Stop()
	st := s.stats[c.Name()]
	var dst Sample
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C():
		}
		r := s.collectOnce(ctx, c, &dst, st)
		skipPending(tick.C(), st)
		if !publish(ctx, s.out, r, timer, interval, st) {
			return
		}
	}
}

func (s *Scheduler) collectOnce(ctx context.Context, c Collector, dst *Sample, st *counters) Result {
	dst.Reset()
	start := time.Now()
	err := c.Collect(ctx, dst)
	took := time.Since(start)
	st.runs.Add(1)
	if err != nil {
		st.errors.Add(1)
	}
	return Result{Collector: c.Name(), Sample: dst.Clone(), Err: err, Took: took}
}

// skipPending discards ticks that arrived while Collect was running.
func skipPending(ch <-chan time.Time, st *counters) {
	for {
		select {
		case <-ch:
			st.skipped.Add(1)
		default:
			return
		}
	}
}

// publish sends r, waiting at most one interval. It returns false when ctx
// is done.
func publish(ctx context.Context, out chan<- Result, r Result, timer *time.Timer, wait time.Duration, st *counters) bool {
	timer.Reset(wait)
	defer timer.Stop()
	select {
	case out <- r:
		return true
	case <-ctx.Done():
		return false
	case <-timer.C:
		st.dropped.Add(1)
		return true
	}
}
