BINARY := gofind
BIN_DIR := bin
PKG     := ./cmd/gofind

.PHONY: all build run test vet fmt tidy clean help

all: build

## build: compile the gofind binary into bin/
build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) $(PKG)

## run: build, then run with ARGS (e.g. make run ARGS="index .")
run: build
	./$(BIN_DIR)/$(BINARY) $(ARGS)

## test: run all tests
test:
	go test ./...

## vet: static analysis
vet:
	go vet ./...

## fmt: format every Go file in place
fmt:
	gofmt -w .

## tidy: sync go.mod / go.sum with imports
tidy:
	go mod tidy

## clean: remove the binary and any local indexes
clean:
	rm -rf $(BIN_DIR) .gofind.db .gofind.db-wal .gofind.db-shm

## help: list available targets
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //'
