// SPDX-License-Identifier: Apache-2.0 OR MIT
// Copyright (c) 2026 Sebastien Rousseau

// Package graph renders percentage series as terminal graphs.
//
// In braille style each terminal cell is a 2x4 dot matrix (U+2800 block).
// As in btop, a cell's left dot column shows the previous sample and its
// right dot column the current one, so the graph scrolls one cell per
// sample and a new sample changes only the rightmost cell column. Columns
// are cached in a ring; Push recomputes one column and AppendRows copies the
// cache, so an incremental frame allocates nothing once the output buffer is
// large enough.
//
// Block style (h = 1 only) uses the eighth blocks ▁▂▃▄▅▆▇█, one sample per
// cell, for terminals or fonts without braille.
//
// Values are clamped to 0..100. NaN (not measured) draws no dots, the same
// as 0; callers that need to show "n/a" do so outside the graph.
package graph

import (
	"errors"
	"math"
	"unicode/utf8"
)

// Style selects the glyph set.
type Style int

const (
	// Braille draws 2x4 dots per cell.
	Braille Style = iota
	// Block draws one eighth-block per cell; only valid for h = 1.
	Block
)

const brailleBase = 0x2800

// Bits lighting the bottom n dots of the left and right braille columns.
var (
	leftBits  = [5]rune{0, 0x40, 0x44, 0x46, 0x47}
	rightBits = [5]rune{0, 0x80, 0xA0, 0xB0, 0xB8}
	blocks    = [9]rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
)

// Graph is a w x h cell graph with a cached column per cell.
type Graph struct {
	w, h  int
	style Style
	cells []rune // column-major: physical column p occupies cells[p*h : p*h+h], row 0 at top
	head  int    // physical index of the oldest (leftmost) column
	last  int    // level of the previous sample, -1 when there is none
	// recomputed counts columns computed, so tests can prove Push touches one.
	recomputed int
}

// New returns an empty graph of w x h cells.
func New(w, h int, style Style) (*Graph, error) {
	if w <= 0 || h <= 0 {
		return nil, errors.New("graph: width and height must be positive")
	}
	if style == Block && h != 1 {
		return nil, errors.New("graph: block style needs height 1")
	}
	if style != Braille && style != Block {
		return nil, errors.New("graph: unknown style")
	}
	g := &Graph{w: w, h: h, style: style, cells: make([]rune, w*h)}
	g.Reset()
	return g, nil
}

// Width returns the width in cells.
func (g *Graph) Width() int { return g.w }

// Height returns the height in cells.
func (g *Graph) Height() int { return g.h }

// Reset clears every column and forgets the previous sample.
func (g *Graph) Reset() {
	blank := g.blank()
	for i := range g.cells {
		g.cells[i] = blank
	}
	g.head = 0
	g.last = -1
}

func (g *Graph) blank() rune {
	if g.style == Block {
		return ' '
	}
	return brailleBase
}

// levels is the number of vertical steps the graph can show.
func (g *Graph) levels() int {
	if g.style == Block {
		return len(blocks) - 1
	}
	return 4 * g.h
}

func (g *Graph) level(v float64) int {
	if math.IsNaN(v) || v <= 0 {
		return 0
	}
	if v >= 100 {
		return g.levels()
	}
	return int(math.Round(v * float64(g.levels()) / 100))
}

// Push scrolls the graph left by one cell and draws v in the rightmost
// column. It allocates nothing.
func (g *Graph) Push(v float64) {
	cur := g.level(v)
	col := g.cells[g.head*g.h : g.head*g.h+g.h]
	if g.style == Block {
		col[0] = blocks[cur]
	} else {
		g.brailleColumn(col, max(g.last, 0), cur)
	}
	g.recomputed++
	g.last = cur
	g.head++
	if g.head == g.w {
		g.head = 0
	}
}

func (g *Graph) brailleColumn(col []rune, left, right int) {
	for r := range col {
		below := (g.h - 1 - r) * 4
		col[r] = brailleBase | leftBits[clamp4(left-below)] | rightBits[clamp4(right-below)]
	}
}

func clamp4(n int) int { return min(max(n, 0), 4) }

// Fill resets the graph and draws the newest values that fit: w+1 samples
// for braille (the first only supplies the leftmost cell's left column) and
// w for block style.
func (g *Graph) Fill(values []float64) {
	g.Reset()
	keep := g.w
	if g.style == Braille {
		keep++
	}
	if len(values) > keep {
		values = values[len(values)-keep:]
	}
	if g.style == Braille && len(values) == keep {
		g.last = g.level(values[0])
		values = values[1:]
	}
	for _, v := range values {
		g.Push(v)
	}
}

// AppendRows appends the graph as UTF-8, rows top to bottom separated by
// '\n' with no trailing newline. It allocates only when dst lacks capacity;
// BufferSize gives the capacity that always suffices.
func (g *Graph) AppendRows(dst []byte) []byte {
	for r := 0; r < g.h; r++ {
		if r > 0 {
			dst = append(dst, '\n')
		}
		for c := 0; c < g.w; c++ {
			p := g.head + c
			if p >= g.w {
				p -= g.w
			}
			dst = utf8.AppendRune(dst, g.cells[p*g.h+r])
		}
	}
	return dst
}

// BufferSize is the most bytes AppendRows writes.
func (g *Graph) BufferSize() int { return g.w*g.h*utf8.UTFMax + g.h }

// String renders the graph; it allocates, so frame loops use AppendRows.
func (g *Graph) String() string {
	return string(g.AppendRows(make([]byte, 0, g.BufferSize())))
}
