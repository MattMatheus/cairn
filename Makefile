.PHONY: build test smoke pilot-check

GO ?= go
GOCACHE ?= /tmp/cairn-go-build-cache

build:
	$(GO) build -o ./bin/cairn ./cmd/cairn

test:
	$(GO) test ./...

smoke:
	GOCACHE=$(GOCACHE) deployments/local-dev/core-smoke.sh

pilot-check:
	GOCACHE=$(GOCACHE) deployments/local-dev/pilot-check.sh
