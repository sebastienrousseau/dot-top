// SPDX-License-Identifier: Apache-2.0 OR MIT
// Copyright (c) 2026 Sebastien Rousseau

package graph

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// FuzzGraph feeds arbitrary samples into graphs of arbitrary size and checks
// the frame keeps its shape and alphabet, and that pushing sample by sample
// draws the same frame as filling the same history at once.
func FuzzGraph(f *testing.F) {
	f.Add(uint8(10), uint8(2), false, []byte{0, 50, 100, 255})
	f.Add(uint8(4), uint8(1), true, []byte{1, 2, 3})
	f.Fuzz(func(t *testing.T, wb, hb uint8, block bool, data []byte) {
		w, h, style := int(wb%64)+1, int(hb%8)+1, Braille
		if block {
			h, style = 1, Block
		}
		pushed, err := New(w, h, style)
		if err != nil {
			t.Fatal(err)
		}
		values := make([]float64, len(data))
		for i, b := range data {
			values[i] = float64(b) / 2.55 // 0..100, with some samples just past it
			pushed.Push(values[i])
		}
		frame := pushed.String()
		checkFrame(t, frame, w, h, style)
		filled, _ := New(w, h, style)
		filled.Fill(values)
		if filled.String() != frame {
			t.Fatalf("Push and Fill disagree:\n%s\n---\n%s", frame, filled.String())
		}
		if n := len(pushed.AppendRows(nil)); n > pushed.BufferSize() {
			t.Fatalf("AppendRows wrote %d bytes, BufferSize is %d", n, pushed.BufferSize())
		}
	})
}

func checkFrame(t *testing.T, frame string, w, h int, style Style) {
	t.Helper()
	rows := strings.Split(strings.TrimSuffix(frame, "\n"), "\n")
	if len(rows) != h {
		t.Fatalf("%d rows, want %d", len(rows), h)
	}
	for _, row := range rows {
		if utf8.RuneCountInString(row) != w {
			t.Fatalf("row %q has %d cells, want %d", row, utf8.RuneCountInString(row), w)
		}
		for _, r := range row {
			if style == Braille && (r < brailleBase || r > brailleBase+0xFF) {
				t.Fatalf("rune %U is not braille", r)
			}
			if style == Block && !strings.ContainsRune(string(blocks[:]), r) {
				t.Fatalf("rune %U is not a block symbol", r)
			}
		}
	}
}
