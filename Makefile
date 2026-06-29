.PHONY: start stop test build

CONFIG ?= configs/config.yaml

start:
	go run cmd/intake-agent/main.go --config $(CONFIG)

build:
	go build -o bin/intake-agent cmd/intake-agent/main.go

test:
	go test ./... -v

stop:
	pkill -f intake-agent || true
