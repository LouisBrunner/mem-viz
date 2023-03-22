all: lint test
.PHONY: all

setup:
	go mod download
.PHONY: setup

lint:
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck ./...
.PHONY: lint

test:
	go test -v ./...
.PHONY: test

build:
	go build ./cmd/dsc-viz
.PHONY: build

debug:
	DEBUG=y go run -- ./cmd/dsc-viz $(ARGS)
.PHONY: debug
