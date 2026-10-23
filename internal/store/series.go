// SPDX-License-Identifier: Apache-2.0 OR MIT
// Copyright (c) 2026 Sebastien Rousseau

// Package store holds measured samples in fixed-capacity ring buffers.
//
// A Series never grows: once full, each Push overwrites the oldest sample.
// Push allocates nothing, so collectors can record on every tick without
// pressuring the garbage collector.
package store

import (
	"fmt"
	"time"
)

// Point is one timestamped sample.
type Point struct {
	T time.Time
	V float64
}

// Series is a fixed-capacity ring buffer of timestamped float64 samples.
// It is not safe for concurrent use; callers serialise access.
type Series struct {
	times []time.Time
	vals  []float64
	head  int // index of the next write
	n     int // number of stored samples, <= cap
}

// NewSeries returns an empty Series that keeps the last capacity samples.
// It panics if capacity is not positive.
func NewSeries(capacity int) *Series {
	if capacity <= 0 {
		panic(fmt.Sprintf("store: NewSeries capacity %d must be positive", capacity))
	}
	return &Series{
		times: make([]time.Time, capacity),
		vals:  make([]float64, capacity),
	}
}

// Cap returns the fixed capacity.
func (s *Series) Cap() int { return len(s.vals) }

// Len returns the number of stored samples.
func (s *Series) Len() int { return s.n }

// Push appends a sample, overwriting the oldest one when the Series is full.
// It allocates nothing.
func (s *Series) Push(t time.Time, v float64) {
	s.times[s.head] = t
	s.vals[s.head] = v
	s.head++
	if s.head == len(s.vals) {
		s.head = 0
	}
	if s.n < len(s.vals) {
		s.n++
	}
}

// index maps a logical position (0 = oldest) to a physical slot.
func (s *Series) index(i int) int {
	start := s.head - s.n
	if start < 0 {
		start += len(s.vals)
	}
	j := start + i
	if j >= len(s.vals) {
		j -= len(s.vals)
	}
	return j
}

// At returns the i-th sample, ordered oldest (0) to newest (Len()-1).
// It panics if i is out of range.
func (s *Series) At(i int) Point {
	if i < 0 || i >= s.n {
		panic(fmt.Sprintf("store: At(%d) out of range [0,%d)", i, s.n))
	}
	j := s.index(i)
	return Point{T: s.times[j], V: s.vals[j]}
}

// Last returns the newest sample and true, or the zero Point and false when
// the Series is empty.
func (s *Series) Last() (Point, bool) {
	if s.n == 0 {
		return Point{}, false
	}
	return s.At(s.n - 1), true
}

// Values writes the stored values, oldest to newest, into dst[:0] and returns
// the result. It allocates only when cap(dst) < Len().
func (s *Series) Values(dst []float64) []float64 {
	dst = dst[:0]
	for i := 0; i < s.n; i++ {
		dst = append(dst, s.vals[s.index(i)])
	}
	return dst
}
