# SPDX-License-Identifier: Apache-2.0 OR MIT
# Copyright (c) 2026 Sebastien Rousseau

VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)
export CGO_ENABLED := 0

.PHONY: all test lint build bench fmt-check vet

all: lint test build

test:
	CGO_ENABLED=1 go test -race ./...

lint: fmt-check vet
	golangci-lint run ./...

fmt-check:
	@out="$$(gofmt -l .)"; if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi

vet:
	go vet ./...

build:
	go build -trimpath -ldflags '$(LDFLAGS)' -o bin/dot-top ./cmd/dot-top

bench:
	go test -run '^$$' -bench . -benchmem ./...
