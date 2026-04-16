.PHONY: help build clean test install dev lint
.DEFAULT_GOAL := build

# Variables
BINARY_NAME := dhcli
GO := go
GOFLAGS := -v
BUILD_DIR := bin
OUTPUT := $(BUILD_DIR)/$(BINARY_NAME)

help:
	@echo "Available targets:"
	@echo "  build      - Build the application"
	@echo "  clean      - Remove build artifacts"
	@echo "  test       - Run tests"
	@echo "  install    - Install the binary to \$$GOPATH/bin"
	@echo "  dev        - Build with debug info for development"
	@echo "  lint       - Run linter (if golangci-lint installed)"
	@echo "  help       - Show this help message"

build: clean
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) -o $(OUTPUT) .

clean:
	@echo "Cleaning build artifacts..."
	$(GO) clean
	rm -rf $(BUILD_DIR)

install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(OUTPUT) $(shell $(GO) env GOPATH)/bin/

test:
	@echo "Running tests..."
	$(GO) test -v ./...

dev:
	@echo "Building $(BINARY_NAME) for development..."
	$(GO) build -gcflags="all=-N -l" -o $(OUTPUT) .

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; exit 1; }
	golangci-lint run ./...
