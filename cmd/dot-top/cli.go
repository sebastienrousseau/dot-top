// SPDX-License-Identifier: Apache-2.0 OR MIT
// Copyright (c) 2026 Sebastien Rousseau

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/sebastienrousseau/dot-top/internal/collect"
)

// Exit codes from the dot family contract (ADR-020, N10).
const (
	exitOK      = 0
	exitFailed  = 1
	exitUsage   = 64
	exitUnknown = 127
)

// snapshotSchema identifies the --json document.
const snapshotSchema = "top/snapshot.v1"

const usage = `Usage: dot-top [options] [command]

What the machine and its agents are doing.
Pre-release: no collectors or TUI are built yet.

Commands:
  measure          Sample once and print the result (same as --once)
  report show      Usage report (planned, not built)

Options:
  -h, --help       Show this help and exit
  -V, --version    Print the version and exit
  -v, --verbose    Log diagnostics to stderr
      --once       Sample once, print, and exit
      --json       With --once or measure: print a top/snapshot.v1 document
`

type options struct {
	help, version, verbose, once, json bool
	command                            []string
}

var errUsage = errors.New("usage")

func parseArgs(args []string) (options, error) {
	var o options
	flags := map[string]*bool{
		"-h": &o.help, "--help": &o.help,
		"-V": &o.version, "--version": &o.version,
		"-v": &o.verbose, "--verbose": &o.verbose,
		"--once": &o.once, "--json": &o.json,
	}
	for i, a := range args {
		if a == "--" {
			o.command = append(o.command, args[i+1:]...)
			break
		}
		if p, ok := flags[a]; ok {
			*p = true
			continue
		}
		if strings.HasPrefix(a, "-") && a != "-" {
			return o, fmt.Errorf("%w: unknown option %q", errUsage, a)
		}
		o.command = append(o.command, a)
	}
	return o, nil
}

// snapshot is the --json document. Collectors is empty until collectors
// exist; nothing unmeasured is ever filled in.
type snapshot struct {
	Schema     string              `json:"schema"`
	Version    string              `json:"version"`
	Collectors []collectorSnapshot `json:"collectors"`
}

type collectorSnapshot struct {
	Name    string               `json:"name"`
	Caps    collect.Capabilities `json:"capabilities"`
	Metrics []collect.Metric     `json:"metrics"`
}

func run(args []string, stdout, stderr io.Writer) int {
	o, err := parseArgs(args)
	if err != nil {
		logf(stderr, "dot-top: %v\nTry 'dot-top --help'.\n", err)
		return exitUsage
	}
	switch {
	case o.help:
		return write(stdout, stderr, usage)
	case o.version:
		return write(stdout, stderr, "dot-top "+version+"\n")
	}
	return dispatch(o, stdout, stderr)
}

func dispatch(o options, stdout, stderr io.Writer) int {
	switch cmd := strings.Join(o.command, " "); cmd {
	case "":
		if o.once {
			return measure(o, stdout, stderr)
		}
		if o.json {
			logf(stderr, "dot-top: --json needs --once or the measure command\n")
			return exitUsage
		}
		logf(stderr, "dot-top: interactive mode not built yet; try 'dot-top --once --json'\n")
		return exitOK
	case "measure":
		return measure(o, stdout, stderr)
	case "report", "report show":
		logf(stderr, "dot-top: %q is planned but not built yet\n", cmd)
		return exitUnknown
	default:
		logf(stderr, "dot-top: unknown command %q\nTry 'dot-top --help'.\n", cmd)
		return exitUnknown
	}
}

func measure(o options, stdout, stderr io.Writer) int {
	snap := snapshot{Schema: snapshotSchema, Version: version, Collectors: []collectorSnapshot{}}
	if o.verbose {
		logf(stderr, "dot-top: %d collectors registered\n", len(snap.Collectors))
	}
	if !o.json {
		return write(stdout, stderr, "no collectors built yet; nothing measured\n")
	}
	enc := json.NewEncoder(stdout)
	if err := enc.Encode(snap); err != nil {
		logf(stderr, "dot-top: write snapshot: %v\n", err)
		return exitFailed
	}
	return exitOK
}

// logf writes a diagnostic to stderr. A failed diagnostic write has nowhere
// left to be reported, so its error is deliberately dropped.
func logf(stderr io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(stderr, format, args...)
}

// write prints a result on stdout and turns a failed write into exit 1.
func write(stdout, stderr io.Writer, s string) int {
	if _, err := io.WriteString(stdout, s); err != nil {
		logf(stderr, "dot-top: write: %v\n", err)
		return exitFailed
	}
	return exitOK
}
