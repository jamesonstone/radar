.PHONY: build test lint fmt vet clean install tidy all

BINARY_NAME=rdr
# Expects release tags like v1.2.3; falls back to "dev" when no matching tag exists.
VERSION?=$(shell git describe --tags --abbrev=0 --match 'v[0-9]*.[0-9]*.[0-9]*' 2>/dev/null || echo dev)
LDFLAGS=-ldflags "-X github.com/jamesonstone/radar/internal/app.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/rdr

install:
	go install $(LDFLAGS) ./cmd/rdr

test:
	go test -v ./...

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -rf bin/
	go clean

tidy:
	go mod tidy

all: fmt vet test build
