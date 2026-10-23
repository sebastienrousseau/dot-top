# SPDX-License-Identifier: Apache-2.0 OR MIT
# Copyright (c) 2026 Sebastien Rousseau

VERSION ?= dev
# DEBUG=1 keeps the symbol table and DWARF data (no -s -w).
DEBUG ?= 0
ifeq ($(DEBUG),1)
STRIP_FLAGS :=
else
STRIP_FLAGS := -s -w
endif
LDFLAGS := $(STRIP_FLAGS) -X main.version=$(VERSION)
export CGO_ENABLED := 0

# Install locations follow the GNU conventions: PREFIX and DESTDIR.
PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
DESTDIR ?=

.PHONY: all test lint build bench fmt-check vet install uninstall repro

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

install: build
	install -d '$(DESTDIR)$(BINDIR)'
	install -m 0755 bin/dot-top '$(DESTDIR)$(BINDIR)/dot-top'

uninstall:
	rm -f '$(DESTDIR)$(BINDIR)/dot-top'

# Build twice from clean caches and require identical binaries.
repro:
	go build -trimpath -ldflags '$(LDFLAGS)' -o bin/repro-a ./cmd/dot-top
	go clean -cache
	go build -trimpath -ldflags '$(LDFLAGS)' -o bin/repro-b ./cmd/dot-top
	cmp bin/repro-a bin/repro-b
	@echo "reproducible: $$(shasum -a 256 bin/repro-a | cut -d' ' -f1)"

bench:
	go test -run '^$$' -bench . -benchmem ./...
