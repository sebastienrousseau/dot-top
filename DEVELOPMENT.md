<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Development

## Toolchain

- Go: `go.mod` pins `go 1.26.0` and `toolchain go1.26.10`; with
  `GOTOOLCHAIN=auto` (the default) the right toolchain is fetched.
- golangci-lint v2 (CI pins v2.13.2), for example `mise use golangci-lint@2`.
- goreleaser v2, only to test release packaging.

## Gates

Run each of these before pushing; CI (`.github/workflows/ci.yml`) runs
the same ones on Ubuntu and macOS.

| Gate | Command |
| :--- | :--- |
| Formatting | `gofmt -l .` (must print nothing) |
| Vet | `go vet ./...` |
| Tests with the race detector | `go test -race ./...` |
| Pure-Go build | `CGO_ENABLED=0 go build ./...` |
| Lint and complexity ceilings | `golangci-lint run ./...` |

`make lint test build` runs them all. `make bench` runs the benchmarks.
`go test -short ./...` skips the test that builds the binary.

The race detector needs cgo, so `make test` enables it for tests only;
shipped binaries are always built with `CGO_ENABLED=0`.

## Test layout

- Unit tests sit beside the code (`*_test.go`).
- `internal/archtest` checks import rules across the module.
- `cmd/dot-top` tests drive `run()` with captured output and also build
  and run the binary once.

## Release model

Releases follow the dot family's lockstep version and are cut by the
maintainer with goreleaser (`.goreleaser.yaml`). None has been made yet.
To check packaging locally: `goreleaser release --snapshot --clean`.
