// SPDX-License-Identifier: Apache-2.0 OR MIT
// Copyright (c) 2026 Sebastien Rousseau

package store

import (
	"slices"
	"testing"
	"time"
)

var epoch = time.Unix(1_700_000_000, 0)

func push(s *Series, vals ...float64) {
	for i, v := range vals {
		s.Push(epoch.Add(time.Duration(i)*time.Second), v)
	}
}

func TestSeriesOrder(t *testing.T) {
	tests := []struct {
		name   string
		cap    int
		pushed []float64
		want   []float64
	}{
		{"empty", 3, nil, []float64{}},
		{"partial", 3, []float64{1, 2}, []float64{1, 2}},
		{"exactly full", 3, []float64{1, 2, 3}, []float64{1, 2, 3}},
		{"wrapped once", 3, []float64{1, 2, 3, 4}, []float64{2, 3, 4}},
		{"wrapped twice", 3, []float64{1, 2, 3, 4, 5, 6, 7}, []float64{5, 6, 7}},
		{"capacity one", 1, []float64{1, 2, 3}, []float64{3}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := NewSeries(tc.cap)
			push(s, tc.pushed...)
			if got := s.Values(nil); !slices.Equal(got, tc.want) {
				t.Fatalf("Values = %v, want %v", got, tc.want)
			}
			if s.Len() != len(tc.want) {
				t.Fatalf("Len = %d, want %d", s.Len(), len(tc.want))
			}
			for i, w := range tc.want {
				if p := s.At(i); p.V != w {
					t.Fatalf("At(%d).V = %v, want %v", i, p.V, w)
				}
			}
		})
	}
}

func TestSeriesTimestampsFollowValues(t *testing.T) {
	s := NewSeries(2)
	push(s, 10, 20, 30) // timestamps epoch+0s, +1s, +2s
	if p := s.At(0); p.V != 20 || !p.T.Equal(epoch.Add(time.Second)) {
		t.Fatalf("At(0) = %+v, want 20 at epoch+1s", p)
	}
	last, ok := s.Last()
	if !ok || last.V != 30 || !last.T.Equal(epoch.Add(2*time.Second)) {
		t.Fatalf("Last = %+v, %v; want 30 at epoch+2s, true", last, ok)
	}
}

func TestSeriesLastEmpty(t *testing.T) {
	if p, ok := NewSeries(4).Last(); ok || p != (Point{}) {
		t.Fatalf("Last on empty = %+v, %v; want zero, false", p, ok)
	}
}

func TestSeriesValuesReusesDst(t *testing.T) {
	s := NewSeries(4)
	push(s, 1, 2, 3)
	dst := make([]float64, 0, 8)
	got := s.Values(dst)
	if &got[0] != &dst[:1][0] {
		t.Fatal("Values did not reuse dst's backing array")
	}
	if !slices.Equal(got, []float64{1, 2, 3}) {
		t.Fatalf("Values = %v", got)
	}
}

func TestSeriesPanics(t *testing.T) {
	cases := map[string]func(){
		"zero capacity":    func() { NewSeries(0) },
		"At negative":      func() { NewSeries(2).At(-1) },
		"At past Len":      func() { s := NewSeries(2); push(s, 1); s.At(1) },
		"At on empty slot": func() { NewSeries(2).At(0) },
	}
	for name, fn := range cases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected panic")
				}
			}()
			fn()
		})
	}
}

func TestSeriesPushAllocatesNothing(t *testing.T) {
	s := NewSeries(64)
	i := 0
	allocs := testing.AllocsPerRun(1000, func() {
		s.Push(epoch, float64(i))
		i++
	})
	if allocs != 0 {
		t.Fatalf("Push allocated %v times per run, want 0", allocs)
	}
}

func TestSeriesValuesAllocatesNothingWithCapacity(t *testing.T) {
	s := NewSeries(64)
	push(s, make([]float64, 100)...)
	dst := make([]float64, 0, 64)
	allocs := testing.AllocsPerRun(100, func() { dst = s.Values(dst) })
	if allocs != 0 {
		t.Fatalf("Values allocated %v times per run, want 0", allocs)
	}
}

func BenchmarkSeriesPush(b *testing.B) {
	s := NewSeries(1024)
	b.ReportAllocs()
	for i := 0; b.Loop(); i++ {
		s.Push(epoch, float64(i))
	}
}

func BenchmarkSeriesValues(b *testing.B) {
	s := NewSeries(1024)
	push(s, make([]float64, 1500)...)
	dst := make([]float64, 0, 1024)
	b.ReportAllocs()
	for b.Loop() {
		dst = s.Values(dst)
	}
}
