<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->
<!-- Copyright (c) 2026 Sebastien Rousseau -->

# Security assurance case

Why dot-top's security requirements are met, and how each claim is
checked. It describes the code on `main`; planned behaviour is marked
**planned** and becomes a requirement when it is built.

## Security requirements

1. dot-top never sends data off the machine: no network access and no
   telemetry.
2. It needs no privileges beyond the user who runs it.
3. It only reads: process tables, system counters and agent transcripts.
   It never writes outside its own state directory (**planned**: the
   recorder).
4. It may signal only processes owned by the same user, and only after
   checking the process is still the one the user chose (**planned**: T2).
5. Malformed input cannot crash it or make it use unbounded memory.

## Threat model

| Threat | Example | Requirement |
|---|---|---|
| Data leaves the machine | A dependency phones home | 1 |
| Privilege misuse | dot-top run as root kills another user's process | 2, 4 |
| Wrong process signalled | A PID is reused between sampling and the kill | 4 |
| Hostile local input | A crafted transcript or `/proc` entry, or arguments | 5 |
| Supply chain | A compromised dependency or build step | 1, 5 |

Out of scope: an attacker who already runs code as the same user (they
can do anything dot-top can), and kernel compromise.

## Trust boundaries

- **User to dot-top:** command-line arguments and the terminal. Only the
  documented options and commands are accepted.
- **Operating system to dot-top:** `/proc`, sysctl and libproc data
  (**planned**: T1 collectors) is parsed as untrusted.
- **Agent transcripts to dot-top:** files under `~/.claude` and `~/.codex`
  (**planned**: T2), parsed as untrusted and read for metadata only.
- **dot-top to other processes:** signals (**planned**: T2), same user only.

## Secure design principles applied

- **Least privilege:** no network code at all (requirement 1). An
  architecture test (`internal/archtest`) fails the build if the UI
  packages import `os`, `net` or `syscall`.
- **Economy of mechanism:** pure Go with `CGO_ENABLED=0`; macOS libraries
  will be called through purego rather than cgo.
- **Fail-safe defaults:** a metric that cannot be measured shows "n/a"
  with the reason instead of a guessed number.
- **Complete mediation (planned):** every signal re-checks the PID's start
  time and program before it is sent.

## Common weaknesses countered

| Weakness | Countermeasure | Evidence |
|---|---|---|
| Memory corruption (CWE-119 family) | Go is memory-safe; no cgo, no `unsafe` | `CGO_ENABLED=0` in CI and releases |
| Unbounded resource use (CWE-400) | Fixed-capacity ring buffers; the scheduler drops ticks instead of queuing | `TestSeriesPushAllocatesNothing`, `TestSchedulerSkipsTicksDuringOverrun` |
| Improper input validation (CWE-20) | Arguments are an allowlist; others exit 64 or 127 | `FuzzRun`, `TestRun` |
| Command injection (CWE-78) | dot-top runs no shell and builds no command lines | no `os/exec` import outside `_test.go` files |
| Vulnerable dependencies (CWE-1395) | govulncheck in CI, Dependabot weekly | `.github/workflows/ci.yml`, `.github/dependabot.yml` |

## Hardening

- Statically linked, pure Go binaries (`CGO_ENABLED=0`) built with
  `-trimpath`.
- CodeQL (Go and workflows), gosec and the race detector on every pull
  request; native fuzz targets for every parser added.
- Branch protection: signed commits, required checks, no force pushes.
