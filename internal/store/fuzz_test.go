// SPDX-License-Identifier: Apache-2.0 OR MIT
// Copyright (c) 2026 Sebastien Rousseau

package store

import (
	"math"
	"testing"
	"time"
)

// FuzzSeries pushes an arbitrary sequence into a ring of arbitrary capacity
// and checks the ring against a plain slice holding the same samples.
func FuzzSeries(f *testing.F) {
	f.Add(uint8(4), []byte{1, 2, 3, 4, 5, 6})
	f.Add(uint8(1), []byte{9})
	f.Add(uint8(64), []byte{})
	f.Fuzz(func(t *testing.T, capByte uint8, data []byte) {
		capacity := int(capByte%64) + 1
		s := NewSeries(capacity)
		all := make([]float64, 0, len(data))
		base := time.Unix(0, 0)
		for i, b := range data {
			v := float64(b) - 128
			s.Push(base.Add(time.Duration(i)*time.Second), v)
			all = append(all, v)
		}
		want := all[max(0, len(all)-capacity):]
		if s.Len() != len(want) {
			t.Fatalf("Len = %d, want %d", s.Len(), len(want))
		}
		got := s.Values(nil)
		for i := range want {
			if got[i] != want[i] || s.At(i).V != want[i] {
				t.Fatalf("sample %d = %v / %v, want %v", i, got[i], s.At(i).V, want[i])
			}
		}
		last, ok := s.Last()
		if ok != (len(want) > 0) || (ok && math.Float64bits(last.V) != math.Float64bits(want[len(want)-1])) {
			t.Fatalf("Last = %v %v, want %v", last, ok, want)
		}
	})
}
