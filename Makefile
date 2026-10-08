VERSION ?= $(or $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//'),dev)
COMMIT  ?= $(or $(shell git rev-parse --short=12 HEAD 2>/dev/null),unknown)
PKG     := github.com/Nicolas1bhr/agentpool/internal/version
LDFLAGS := -s -w -X $(PKG).Version=$(VERSION) -X $(PKG).Commit=$(COMMIT)

.PHONY: build test check fmt vet clean

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/agentpool ./cmd/agentpool

test:
	go test -race -count=1 ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

check: vet test
	@test -z "$$(gofmt -l .)" || (echo "gofmt needed on:"; gofmt -l .; exit 1)

clean:
	rm -rf bin dist
