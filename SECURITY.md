<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Security policy

## Reporting a vulnerability

Report privately through GitHub Security Advisories:
**Security → Report a vulnerability** on the
[dot-top repository](https://github.com/sebastienrousseau/dot-top/security/advisories/new).
Do not open a public issue.

Include the version (`dot-top --version`), platform, and steps to
reproduce. You will get an acknowledgement; fixes ship in the next family
release, or sooner for a serious issue.

## Supported versions

dot-top is pre-release. Only the `main` branch is supported until the
first release; after that, the latest family release.

## Security posture

- Pure Go, `CGO_ENABLED=0`.
- Local only: no network access and no telemetry.
- Signals only processes owned by the same user (planned; no signalling
  code exists yet).
