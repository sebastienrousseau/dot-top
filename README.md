<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

<h1 align="center">dot-top</h1>

<p align="center">
  What the machine and its AI coding agents are doing: a terminal system
  monitor that attributes processes to Claude Code and Codex sessions.
</p>

<p align="center">
  <a href="LICENSE-APACHE"><img src="https://img.shields.io/badge/license-Apache--2.0%20OR%20MIT-blue.svg?style=for-the-badge" alt="License: Apache-2.0 OR MIT" height="28" /></a>
  <a href="go.mod"><img src="https://img.shields.io/badge/go-1.26-93450a.svg?style=for-the-badge&logo=go" alt="Go 1.26" height="28" /></a>
</p>

> **Pre-release: not usable yet (foundation only).** This repository holds
> the measurement scheduler, sample store, graph renderer and CLI skeleton.
> There are no collectors and no interactive UI yet; `dot-top --once --json`
> prints an empty, honest snapshot. Nothing has been released or tagged.

---

## Contents

- [Install](#install) · [Requirements](#requirements) · [Quick Start](#quick-start)
- [The dot family](#the-dot-family) · [Capabilities at a glance](#capabilities-at-a-glance)
- [When not to use dot-top](#when-not-to-use-dot-top) · [Development](#development)
- [Security](#security) · [Documentation](#documentation)
- [Stability guarantees](#stability-guarantees) · [License](#license)

---

## Install

No release exists yet. Build from source:

```bash
make build        # CGO_ENABLED=0; writes bin/dot-top
```

---

## Requirements

Go 1.26 (the toolchain is pinned in `go.mod`). Targets: Linux, macOS and
Windows on amd64 and arm64. The binary is pure Go (`CGO_ENABLED=0`).

---

## Quick Start

```bash
bin/dot-top --once --json
# {"schema":"top/snapshot.v1","version":"dev","collectors":[]}
```

The collectors list is empty because no collector exists yet. dot-top never
fills in a number it did not measure; unsupported metrics will appear as
`n/a` with a reason.

---

## The dot family

dot-top provides `dot top` and the binary `dot-top` in the dot command
family (ADR-020 in the [dotfiles](https://github.com/sebastienrousseau/dotfiles)
repository). Every family repository ships at the same family version.

---

## Capabilities at a glance

| Area | Capability | Status |
| :--- | :--- | :--- |
| Store | Fixed-capacity ring buffer of samples, zero-allocation push | built |
| Scheduling | One goroutine per collector; overruns skip ticks | built |
| Rendering | Braille graph (2x4 dots per cell) and block fallback, zero allocations per frame | built |
| CLI | `--help`, `--version`, `--verbose`, `--once --json`, family exit codes | built |
| Collectors | CPU, memory, disk, network, processes | not started |
| Agent attribution | Claude Code and Codex sessions | not started |
| Interactive UI | Bubble Tea v2 TUI | not started |

---

## When not to use dot-top

- Today, for anything: it measures nothing yet. Use btop, htop or atop.
- For fleet or remote monitoring: dot-top reads the local host only and
  makes no network connections.
- For alerting or long-term retention: it keeps a short in-memory history.

---

## Development

```bash
make lint test build bench
```

See [`DEVELOPMENT.md`](DEVELOPMENT.md) for every gate and how CI runs it.

---

## Security

Report vulnerabilities privately according to [`SECURITY.md`](SECURITY.md).

dot-top is designed to read process tables and agent transcripts, signal
only processes owned by the same user, and never use the network. It sends
no telemetry. It is pure Go with cgo disabled.

---

## Documentation

- User manual: not written yet (pre-release).
- API reference: [pkg.go.dev](https://pkg.go.dev/github.com/sebastienrousseau/dot-top) once the module is published.
- Developer docs: [`DEVELOPMENT.md`](DEVELOPMENT.md) and [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md).
- Family map: [dotfiles](https://github.com/sebastienrousseau/dotfiles), ADR-020 (command names) and ADR-021 (lockstep version).

---

## Stability guarantees

None yet. Before the first release, flags, output and the
`top/snapshot.v1` document may change without notice. After it, the CLI
follows the family contract: a removed or renamed command works with a
warning for one family release, and a change to a `--json` document's
shape bumps its schema version.

---

## License

Dual-licensed under [Apache-2.0](LICENSE-APACHE) or [MIT](LICENSE-MIT), at
your option. Contributions are accepted under the same terms.
