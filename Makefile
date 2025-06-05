# GoIRC Makefile
BINARY_NAME=goirc
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags="-s -w -X main.version=$(VERSION)"

# Default target
.PHONY: all
all: build

# Build for current platform
.PHONY: build
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Clean build artifacts
.PHONY: clean
clean:
	rm -f $(BINARY_NAME)*
	rm -f dist/*
	rm -rf dist/

# Run tests
.PHONY: test
test:
	go test -v ./...

# Run the application
.PHONY: run
run:
	go run . 

# Cross-compile for all platforms
.PHONY: build-all
build-all: clean
	mkdir -p dist/
	
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)_linux_amd64 .
	tar -czf dist/$(BINARY_NAME)_linux_amd64.tar.gz -C dist $(BINARY_NAME)_linux_amd64
	
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)_linux_arm64 .
	tar -czf dist/$(BINARY_NAME)_linux_arm64.tar.gz -C dist $(BINARY_NAME)_linux_arm64
	
	# macOS AMD64 (Intel)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)_darwin_amd64 .
	tar -czf dist/$(BINARY_NAME)_darwin_amd64.tar.gz -C dist $(BINARY_NAME)_darwin_amd64
	
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)_darwin_arm64 .
	tar -czf dist/$(BINARY_NAME)_darwin_arm64.tar.gz -C dist $(BINARY_NAME)_darwin_arm64
	
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)_windows_amd64.exe .
	cd dist && zip $(BINARY_NAME)_windows_amd64.zip $(BINARY_NAME)_windows_amd64.exe

# Install dependencies
.PHONY: deps
deps:
	go mod download
	go mod tidy

# Format code
.PHONY: fmt
fmt:
	go fmt ./...

# Lint code (requires golangci-lint to be installed)
.PHONY: lint
lint:
	golangci-lint run

# Check for security issues (requires gosec to be installed)
.PHONY: sec
sec:
	gosec ./...

# Development workflow
.PHONY: dev
dev: deps fmt test build

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build      Build for current platform"
	@echo "  build-all  Cross-compile for all platforms"
	@echo "  test       Run tests"
	@echo "  run        Run the application"
	@echo "  clean      Clean build artifacts"
	@echo "  deps       Install dependencies"
	@echo "  fmt        Format code"
	@echo "  lint       Lint code (requires golangci-lint)"
	@echo "  sec        Security check (requires gosec)"
	@echo "  dev        Development workflow (deps + fmt + test + build)"
	@echo "  help       Show this help message"
