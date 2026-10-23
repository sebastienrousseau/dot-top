// SPDX-License-Identifier: Apache-2.0 OR MIT
// Copyright (c) 2026 Sebastien Rousseau

package graph

import (
	"math"
	"strings"
	"testing"
)

func repeat(v float64, n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = v
	}
	return out
}

func ramp(n int) []float64 { // 0 .. 100 in n evenly spaced steps
	out := make([]float64, n)
	for i := range out {
		out[i] = 100 * float64(i) / float64(n-1)
	}
	return out
}

func mustNew(t testing.TB, w, h int, s Style) *Graph {
	t.Helper()
	g, err := New(w, h, s)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestGolden(t *testing.T) {
	blank4, full4 := strings.Repeat("⠀", 4), strings.Repeat("⣿", 4)
	blank10, full10 := strings.Repeat("⠀", 10), strings.Repeat("⣿", 10)
	tests := []struct {
		name   string
		w, h   int
		style  Style
		values []float64
		want   string
	}{
		{"4x1 zero", 4, 1, Braille, repeat(0, 5), blank4},
		{"4x1 full", 4, 1, Braille, repeat(100, 5), full4},
		{"4x1 ramp", 4, 1, Braille, ramp(5), "⢀⣠⣴⣾"},
		{"4x1 full, no previous sample", 4, 1, Braille, repeat(100, 4), "⢸⣿⣿⣿"},
		{"4x1 partial history", 4, 1, Braille, []float64{100, 100}, "⠀⠀⢸⣿"},
		{"10x2 zero", 10, 2, Braille, repeat(0, 11), blank10 + "\n" + blank10},
		{"10x2 full", 10, 2, Braille, repeat(100, 11), full10 + "\n" + full10},
		{"10x2 ramp", 10, 2, Braille, ramp(11), "⠀⠀⠀⠀⠀⢀⣠⣤⣴⣾\n⢀⣠⣤⣴⣾⣿⣿⣿⣿⣿"},
		{"clamped and NaN", 4, 1, Braille, []float64{-5, 150, math.NaN(), 250, 0}, "⢸⡇⢸⡇"},
		{"block 4x1 zero", 4, 1, Block, repeat(0, 4), "    "},
		{"block 4x1 full", 4, 1, Block, repeat(100, 4), "████"},
		{"block 4x1 ramp", 4, 1, Block, []float64{25, 50, 75, 100}, "▂▄▆█"},
		{"block 10x1 ramp", 10, 1, Block, ramp(11), "▁▂▂▃▄▅▆▆▇█"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := mustNew(t, tc.w, tc.h, tc.style)
			g.Fill(tc.values)
			if got := g.String(); got != tc.want {
				t.Fatalf("got\n%s\nwant\n%s", got, tc.want)
			}
		})
	}
}

// Pushing one value at a time must draw exactly what Fill draws.
func TestPushMatchesFill(t *testing.T) {
	values := []float64{3, 97, 40, 12, 66, 81, 0, 55, 23, 100, 7, 49, 72, 15}
	for _, size := range [][2]int{{4, 1}, {10, 2}, {7, 3}} {
		inc := mustNew(t, size[0], size[1], Braille)
		full := mustNew(t, size[0], size[1], Braille)
		for i, v := range values {
			inc.Push(v)
			full.Fill(values[:i+1])
			if inc.String() != full.String() {
				t.Fatalf("%dx%d after %d pushes:\n%s\nwant\n%s", size[0], size[1], i+1, inc, full)
			}
		}
	}
}

func TestPushRecomputesOneColumn(t *testing.T) {
	g := mustNew(t, 10, 2, Braille)
	g.Fill(ramp(11))
	before := strings.Split(g.String(), "\n")
	n := g.recomputed
	g.Push(50)
	if g.recomputed-n != 1 {
		t.Fatalf("Push recomputed %d columns, want 1", g.recomputed-n)
	}
	after := strings.Split(g.String(), "\n")
	for r := range before {
		b, a := []rune(before[r]), []rune(after[r])
		if string(a[:9]) != string(b[1:]) {
			t.Fatalf("row %d did not scroll by one cell: %q -> %q", r, before[r], after[r])
		}
	}
}

func TestIncrementalFrameAllocatesNothing(t *testing.T) {
	for _, style := range []Style{Braille, Block} {
		g := mustNew(t, 80, 1, style)
		buf := make([]byte, 0, g.BufferSize())
		v := 0.0
		allocs := testing.AllocsPerRun(200, func() {
			v = math.Mod(v+7, 101)
			g.Push(v)
			buf = g.AppendRows(buf[:0])
		})
		if allocs != 0 {
			t.Fatalf("style %d: %v allocs per frame, want 0", style, allocs)
		}
	}
	g := mustNew(t, 80, 6, Braille)
	buf := make([]byte, 0, g.BufferSize())
	allocs := testing.AllocsPerRun(200, func() {
		g.Push(42)
		buf = g.AppendRows(buf[:0])
	})
	if allocs != 0 {
		t.Fatalf("80x6 braille: %v allocs per frame, want 0", allocs)
	}
}

func TestNewRejectsBadSizes(t *testing.T) {
	bad := []struct {
		w, h  int
		style Style
	}{{0, 1, Braille}, {1, 0, Braille}, {4, 2, Block}, {4, 1, Style(9)}}
	for _, b := range bad {
		if _, err := New(b.w, b.h, b.style); err == nil {
			t.Fatalf("New(%d, %d, %d) accepted", b.w, b.h, b.style)
		}
	}
	g := mustNew(t, 3, 2, Braille)
	if g.Width() != 3 || g.Height() != 2 {
		t.Fatalf("size = %dx%d", g.Width(), g.Height())
	}
}

func BenchmarkIncrementalFrame(b *testing.B) {
	g := mustNew(b, 120, 8, Braille)
	buf := make([]byte, 0, g.BufferSize())
	b.ReportAllocs()
	for i := 0; b.Loop(); i++ {
		g.Push(float64(i % 101))
		buf = g.AppendRows(buf[:0])
	}
}
