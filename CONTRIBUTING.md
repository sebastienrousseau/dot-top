<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Contributing

Thank you for helping. dot-top is pre-release, so open an issue before
starting anything large.

1. Read [`AGENTS.md`](AGENTS.md) for the rules every change keeps and
   [`DEVELOPMENT.md`](DEVELOPMENT.md) for the local gates.
2. Run `make lint test build` before you push; CI runs the same gates.
3. Use [Conventional Commits](https://www.conventionalcommits.org/)
   (`feat:`, `fix:`, `docs:` ...), sign your commits, and add a
   `Signed-off-by:` line (`git commit -s`) to certify the
   [Developer Certificate of Origin](https://developercertificate.org/).
4. A behaviour change needs a test that runs the code and asserts the
   outcome, and that you have seen fail without the change.
5. Note user-visible changes under `## Unreleased` in `CHANGELOG.md`.
6. Open a pull request against `main`. It merges once the required checks
   pass and the review below is done.

## Code review

Every pull request is reviewed before it merges. The reviewer checks:

- **Correctness:** the change does what it says, and its test fails
  without it (point 4).
- **The rules in `AGENTS.md`:** CGO off, no I/O in `internal/ui`, no
  invented numbers, complexity limits.
- **Security:** new inputs are validated and fuzzed, and nothing adds
  network access or telemetry (see [`docs/ASSURANCE.md`](docs/ASSURANCE.md)).
- **Docs and changelog** match the behaviour.

A change is acceptable when every required check is green and the
reviewer has no unresolved comments. See [`GOVERNANCE.md`](GOVERNANCE.md)
for who reviews.

Contributions are dual-licensed under Apache-2.0 OR MIT, like the project.
This project follows the [Code of Conduct](CODE_OF_CONDUCT.md).
