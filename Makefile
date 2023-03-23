all: lint test build
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

build: mem-viz dsc-viz
.PHONY: build

mem-viz:
	go build ./cmd/mem-viz
.PHONY: dsc-viz

dsc-viz:
	go build ./cmd/dsc-viz
.PHONY: dsc-viz

debug-mem:
	DEBUG=y go run -- ./cmd/mem-viz $(ARGS)
.PHONY: debug

debug-dsc:
	DEBUG=y go run -- ./cmd/dsc-viz $(ARGS)
.PHONY: debug
