# Project variables
BINARY_NAME := extproctor
MODULE := zntr.io/extproctor
GO := go
GOFLAGS := -trimpath
LDFLAGS := -s -w -buildid=$(shell git rev-parse HEAD)

# Tools
GOFMT := gofmt
GOLANGCI_LINT := golangci-lint
NILAWAY := nilaway
BUF := buf

# Directories
BUILD_DIR := .
COVERAGE_FILE := coverage.out

.PHONY: all build clean test test-coverage lint lint-fix nilaway vet fmt proto install help

## Default target
all: lint test build

## Build the binary
build:
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/extproctor

## Install the binary
install:
	$(GO) install $(GOFLAGS) -ldflags "$(LDFLAGS)" ./cmd/extproctor

## Clean build artifacts
clean:
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -f $(COVERAGE_FILE)

## Run tests
test:
	$(GO) test -race ./...

## Run tests with coverage
test-coverage:
	$(GO) test -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	$(GO) tool cover -func=$(COVERAGE_FILE)

## Run tests with coverage and open HTML report
test-coverage-html: test-coverage
	$(GO) tool cover -html=$(COVERAGE_FILE)

## Run golangci-lint
lint:
	$(GOLANGCI_LINT) run ./...

## Run golangci-lint with auto-fix
lint-fix:
	$(GOLANGCI_LINT) run --fix ./...

## Run nilaway for nil safety analysis
nilaway:
	$(NILAWAY) ./...

## Run go vet
vet:
	$(GO) vet ./...

## Format code
fmt:
	$(GOFMT) -l -s -w .

## Run all checks (lint + nilaway + vet + test)
check: lint nilaway vet test

## Generate protobuf code
proto:
	$(BUF) generate

## Update dependencies
deps:
	$(GO) mod tidy
	$(GO) mod verify

## Install development tools
tools:
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install go.uber.org/nilaway/cmd/nilaway@latest
	$(GO) install mvdan.cc/gofumpt@latest

## Build sample extproc server
build-sample:
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/extproc-server ./sample/extproc

## Run sample extproc server
run-sample: build-sample
	./extproc-server --addr :50051

## Help
help:
	@echo "Available targets:"
	@echo "  all              - Run lint, test, and build (default)"
	@echo "  build            - Build the extproctor binary"
	@echo "  install          - Install the extproctor binary"
	@echo "  clean            - Remove build artifacts"
	@echo "  test             - Run tests"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  test-coverage-html - Run tests and open HTML coverage report"
	@echo "  lint             - Run golangci-lint"
	@echo "  lint-fix         - Run golangci-lint with auto-fix"
	@echo "  nilaway          - Run nilaway nil safety analysis"
	@echo "  vet              - Run go vet"
	@echo "  fmt              - Format code"
	@echo "  check            - Run all checks (lint, nilaway, vet, test)"
	@echo "  proto            - Generate protobuf code"
	@echo "  deps             - Update and verify dependencies"
	@echo "  tools            - Install development tools"
	@echo "  build-sample     - Build sample extproc server"
	@echo "  run-sample       - Run sample extproc server"
	@echo "  help             - Show this help message"

