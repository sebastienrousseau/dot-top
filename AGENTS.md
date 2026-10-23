<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# AGENTS.md

Rules for humans and AI agents changing this repository. Each is a gate;
CI or a test enforces the ones marked *(enforced)*.

## Invariants

1. **CGO off.** Every build uses `CGO_ENABLED=0` *(enforced: CI build,
   Makefile, `.goreleaser.yaml`)*. macOS libraries, when needed, are
   loaded with `github.com/ebitengine/purego`, added only when used.
2. **No I/O in the UI.** Packages under `internal/ui/` must not import
   `os`, `net`, `syscall` or their subpackages, directly or through another
   package of this module *(enforced: `internal/archtest`)*. Update and View
   stay pure; collectors do the reading.
3. **No invented numbers.** A value that was not measured is shown as
   `n/a` with a reason (`collect.Capabilities`), never as 0 or an estimate.
4. **Lockstep family version.** dot-top ships at the dot family version
   (ADR-021). Do not tag or bump versions independently.
5. **Complexity ceilings.** Per function: cyclomatic ≤ 10, cognitive ≤ 15,
   ≤ 60 lines; ≤ 500 lines per file *(enforced: golangci-lint for the
   function limits; file length by review)*.
6. **Tests run the code.** Every test executes the behaviour and asserts
   the outcome; a new test must be seen failing against a broken
   implementation first.

## Commits

- Conventional Commits.
- Commits are signed. The `Signed-off-by:` (DCO) line is added by the
  human who reviews and certifies the change, never by an agent. Agents
  record their part with an `Assisted-by:` trailer.
- No commit, tag, push or release without the maintainer's review.

## Layout

See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md).
