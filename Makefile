GO          := go
GOOS        := $(shell go env GOOS)
GOARCH      := $(shell go env GOARCH)
ORB_MACHINE ?= dev

mods = $(patsubst %/,%,$(wildcard */go.mod) $(wildcard */*/go.mod))

default: help

help:
	@echo "Available targets:"
	@echo "  build         - Build snapshot binary for current OS/ARCH via goreleaser"
	@echo "  snapshot      - Build snapshot release for all platforms"
	@echo "  test          - Run tests across all workspace modules (macOS)"
	@echo "  tidy          - Tidy all Go workspace modules"
	@echo "  lint          - Run golangci-lint"
	@echo "  orb-test      - Run tests in OrbStack Linux VM (with race detector)"
	@echo "  orb-build     - Build Linux binary in OrbStack VM; prints binary size"

build:
	goreleaser build --snapshot --clean --single-target

snapshot:
	goreleaser release --snapshot --clean

test:
	$(GO) test ./...
	@rc=0; for dir in $$(awk '/\t\.\//{sub(/\t\.\//, ""); print}' go.work); do \
	  if find $$dir \( -mindepth 1 -type d -exec test -f "{}/go.mod" \; -prune \) \
	          -o \( -type f -name "*_test.go" -print -quit \) 2>/dev/null | grep -q .; then \
	    (cd $$dir && $(GO) test ./...) || rc=1; \
	  fi; \
	done; exit $$rc

tidy:
	$(GO) mod tidy
	$(foreach m,$(mods),(cd $(m) && $(GO) mod tidy) &&) true

lint:
	golangci-lint run ./...

orb-test:
	orb run -m $(ORB_MACHINE) -- bash -lc "cd $(CURDIR) && GOFLAGS='-race -count=1' make test"

orb-build:
	orb run -m $(ORB_MACHINE) -- bash -lc "cd $(CURDIR) && go build -trimpath -ldflags='-s -w' -o /tmp/shelly ./cmd/shelly && du -h /tmp/shelly"

.PHONY: help build snapshot test tidy lint orb-test orb-build
