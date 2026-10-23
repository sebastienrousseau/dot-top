// SPDX-License-Identifier: Apache-2.0 OR MIT
// Copyright (c) 2026 Sebastien Rousseau

// Command dot-top is `dot top`: a system monitor that attributes processes
// to AI coding-agent sessions. This build is the foundation only; see
// README.md for what works.
package main

import "os"

// version is set at build time with -ldflags "-X main.version=<v>".
var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
