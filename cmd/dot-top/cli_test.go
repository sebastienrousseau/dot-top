// SPDX-License-Identifier: Apache-2.0 OR MIT
// Copyright (c) 2026 Sebastien Rousseau

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func runCapture(args ...string) (code int, stdout, stderr string) {
	var out, errb bytes.Buffer
	code = run(args, &out, &errb)
	return code, out.String(), errb.String()
}

func TestRun(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		code   int
		stdout string // substring; "" means stdout must be empty
		stderr string // substring; "" means stderr must be empty
	}{
		{"short help", []string{"-h"}, 0, "Usage: dot-top", ""},
		{"long help", []string{"--help"}, 0, "Usage: dot-top", ""},
		{"help wins over bad command", []string{"nope", "--help"}, 0, "Usage: dot-top", ""},
		{"short version", []string{"-V"}, 0, "dot-top dev\n", ""},
		{"long version", []string{"--version"}, 0, "dot-top dev\n", ""},
		{"no args", nil, 0, "", "interactive mode not built yet"},
		{"verbose alone", []string{"-v"}, 0, "", "interactive mode not built yet"},
		{"once plain", []string{"--once"}, 0, "nothing measured", ""},
		{"once verbose", []string{"--once", "--verbose"}, 0, "nothing measured", "0 collectors registered"},
		{"measure plain", []string{"measure"}, 0, "nothing measured", ""},
		{"json without once", []string{"--json"}, 64, "", "--json needs --once"},
		{"unknown long option", []string{"--frobnicate"}, 64, "", `unknown option "--frobnicate"`},
		{"unknown short option", []string{"-x"}, 64, "", `unknown option "-x"`},
		{"unknown command", []string{"explode"}, 127, "", `unknown command "explode"`},
		{"planned command", []string{"report", "show"}, 127, "", "planned but not built"},
		{"planned object default", []string{"report"}, 127, "", "planned but not built"},
		{"measure extra word", []string{"measure", "now"}, 127, "", `unknown command "measure now"`},
		{"double dash ends options", []string{"--", "-V"}, 127, "", `unknown command "-V"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code, stdout, stderr := runCapture(tc.args...)
			if code != tc.code {
				t.Errorf("exit = %d, want %d (stderr %q)", code, tc.code, stderr)
			}
			checkStream(t, "stdout", stdout, tc.stdout)
			checkStream(t, "stderr", stderr, tc.stderr)
		})
	}
}

func checkStream(t *testing.T, name, got, want string) {
	t.Helper()
	if want == "" && got != "" {
		t.Errorf("%s = %q, want empty", name, got)
	}
	if !strings.Contains(got, want) {
		t.Errorf("%s = %q, want it to contain %q", name, got, want)
	}
}

func TestSnapshotJSON(t *testing.T) {
	for _, args := range [][]string{{"--once", "--json"}, {"measure", "--json"}} {
		code, stdout, stderr := runCapture(args...)
		if code != 0 || stderr != "" {
			t.Fatalf("%v: exit %d stderr %q", args, code, stderr)
		}
		var doc map[string]json.RawMessage
		if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
			t.Fatalf("%v: invalid JSON %q: %v", args, stdout, err)
		}
		want := map[string]string{"schema": `"top/snapshot.v1"`, "version": `"dev"`, "collectors": `[]`}
		if len(doc) != len(want) {
			t.Fatalf("%v: keys %v, want exactly %v", args, doc, want)
		}
		for k, v := range want {
			if string(doc[k]) != v {
				t.Fatalf("%v: %s = %s, want %s", args, k, doc[k], v)
			}
		}
	}
}

// The built binary takes its version from -ldflags and, with no TTY and no
// arguments, exits promptly instead of waiting for input.
func TestBinary(t *testing.T) {
	if testing.Short() {
		t.Skip("builds the binary")
	}
	bin := filepath.Join(t.TempDir(), "dot-top")
	build := exec.Command("go", "build", "-ldflags", "-X main.version=9.9.9-test", "-o", bin, ".")
	build.Env = append(build.Environ(), "CGO_ENABLED=0")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	if out, err := exec.Command(bin, "--version").Output(); err != nil || string(out) != "dot-top 9.9.9-test\n" {
		t.Fatalf("--version = %q, %v", out, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin) // stdin is /dev/null: no TTY
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil || errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("no-arg run: %v (ctx %v)", err, ctx.Err())
	}
	if !strings.Contains(stderr.String(), "interactive mode not built yet") {
		t.Fatalf("stderr = %q", stderr.String())
	}
	code := exec.Command(bin, "explode").Run()
	var exitErr *exec.ExitError
	if !errors.As(code, &exitErr) || exitErr.ExitCode() != 127 {
		t.Fatalf("unknown command exit = %v, want 127", code)
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("closed") }

func TestWriteFailureExitsOne(t *testing.T) {
	for _, args := range [][]string{{"--version"}, {"--help"}, {"--once"}, {"--once", "--json"}} {
		var stderr bytes.Buffer
		if code := run(args, failingWriter{}, &stderr); code != 1 {
			t.Fatalf("%v with failing stdout: exit %d, want 1", args, code)
		}
		if !strings.Contains(stderr.String(), "closed") {
			t.Fatalf("%v: stderr %q does not report the write error", args, stderr.String())
		}
	}
}
