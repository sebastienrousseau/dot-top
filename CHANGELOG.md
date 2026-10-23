<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Changelog

All notable changes are recorded here, following
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/). dot-top ships at
the dot family version.

## Unreleased

### Added

- `make install` and `make uninstall` (honouring `PREFIX` and `DESTDIR`),
  `make build DEBUG=1` to keep symbols, and `make repro`, which CI runs to
  require bit-for-bit identical builds.
- `GOVERNANCE.md`, `ROADMAP.md`, `docs/ASSURANCE.md` (security assurance
  case), a code review section in `CONTRIBUTING.md`, and REUSE licence
  texts under `LICENSES/`.
- `internal/store`: fixed-capacity sample ring buffer with zero-allocation
  push.
- `internal/collect`: `Collector` interface, capabilities with reasons,
  and a scheduler that skips ticks on overrun.
- `internal/ui/graph`: braille graph renderer with a block fallback and
  zero-allocation incremental frames.
- `dot-top` CLI skeleton: `--help`, `--version`, `--verbose`,
  `--once --json` (`top/snapshot.v1`, empty collector list) and family
  exit codes.
- `dot-plugin.toml` manifest for the `top` domain.
