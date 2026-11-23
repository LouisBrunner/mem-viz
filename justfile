all: lint test build

setup:
    go mod download

lint:
    gofmt -d -e -s .
    go vet ./...
    go tool staticcheck ./...

test:
    go test -v ./...

build: mem-viz dsc-viz macho-viz

[private]
_build cmd:
    go build ./cmd/{{ cmd }}

mem-viz: (_build "mem-viz")

dsc-viz: (_build "dsc-viz")

macho-viz: (_build "macho-viz")

debug-mem *args: (_run "y" "mem-viz" args)

debug-dsc *args: (_run "y" "dsc-viz" args)

debug-macho *args: (_run "y" "macho-viz" args)

[private]
_run debug cmd *args:
    DEBUG={{ debug }} go run -- ./cmd/{{ cmd }} {{ args }}
