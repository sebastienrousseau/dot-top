<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->
<!-- Copyright (c) 2026 Sebastien Rousseau -->

# Governance

dot-top is one repository in the dot family
([sebastienrousseau/dotfiles](https://github.com/sebastienrousseau/dotfiles)
and its `dot-*` siblings). It has a single maintainer today, and this page
says how decisions are made and who does what.

## Model

dot-top follows a maintainer-led model. The maintainer decides what is
merged and released, after discussion in public issues and pull requests.
Decisions that later readers will question are recorded as architecture
decision records in the dotfiles repository (for example ADR-020, command
names, and ADR-021, the shared family version).

## Roles

| Role | Who | Responsibilities |
|---|---|---|
| Maintainer | [@sebastienrousseau](https://github.com/sebastienrousseau) | Reviews and merges pull requests; triages issues; keeps CI, CodeQL, Scorecard and govulncheck green; handles private vulnerability reports as `SECURITY.md` describes; cuts releases at the dot family version and signs the tags. |
| Contributor | Anyone | Opens issues and pull requests that follow `CONTRIBUTING.md` and `AGENTS.md`, and certifies each commit with a DCO `Signed-off-by`. |

## How changes land

Every change reaches `main` through a pull request. Branch protection
requires signed commits, linear history and the CI, govulncheck and CodeQL
checks, and applies to the maintainer too. The review standard is in
`CONTRIBUTING.md`.

## Changing this document

Changes to governance go through a pull request like any other change.
