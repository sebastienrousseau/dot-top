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

Contributions are dual-licensed under Apache-2.0 OR MIT, like the project.
This project follows the [Code of Conduct](CODE_OF_CONDUCT.md).
