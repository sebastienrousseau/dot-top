// SPDX-License-Identifier: Apache-2.0 OR MIT
// Copyright (c) 2026 Sebastien Rousseau

package main

import (
	"bytes"
	"strings"
	"testing"
)

// FuzzRun runs the CLI on arbitrary argument lists (split on NUL) and
// checks it never panics and only exits with a family exit code.
func FuzzRun(f *testing.F) {
	for _, seed := range []string{"", "-V", "--once\x00--json", "measure\x00--json", "--json", "bogus", "--bogus", "report\x00show"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, joined string) {
		var args []string
		if joined != "" {
			args = strings.Split(joined, "\x00")
		}
		var out, errOut bytes.Buffer
		switch rc := run(args, &out, &errOut); rc {
		case exitOK, exitFailed, exitUsage, exitUnknown:
		default:
			t.Fatalf("run(%q) = %d, not a family exit code", args, rc)
		}
	})
}
