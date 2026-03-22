VERSION ?= dev

.PHONY: build test test-integration lint install clean

build:
	go build -ldflags "-X main.version=$(VERSION)" -o bin/ci-debugger ./cmd/ci-debugger

test:
	go test ./...

test-integration:
	go test -tags integration -v ./...

lint:
	golangci-lint run

install:
	go install -ldflags "-X main.version=$(VERSION)" ./cmd/ci-debugger

clean:
	rm -rf bin/

# Quick sanity check — list workflows in testdata
demo-list:
	./bin/ci-debugger list --workflow testdata/simple.yml 2>/dev/null || \
		./bin/ci-debugger --help

.DEFAULT_GOAL := build
