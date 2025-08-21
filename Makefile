all: lint test build
.PHONY: all

setup:
	go mod download
.PHONY: setup

lint:
	gofmt -d -e -s .
	go vet ./...
	go tool staticcheck ./...
.PHONY: lint

test:
	go test -v ./...
.PHONY: test

build: mem-viz dsc-viz macho-viz
.PHONY: build

mem-viz:
	go build ./cmd/mem-viz
.PHONY: mem-viz

dsc-viz:
	go build ./cmd/dsc-viz
.PHONY: dsc-viz

macho-viz:
	go build ./cmd/macho-viz
.PHONY: macho-viz

debug-mem:
	DEBUG=y go run -- ./cmd/mem-viz $(ARGS)
.PHONY: debug-mem

debug-dsc:
	DEBUG=y go run -- ./cmd/dsc-viz $(ARGS)
.PHONY: debug-dsc

debug-macho:
	DEBUG=y go run -- ./cmd/macho-viz $(ARGS)
.PHONY: debug-macho
