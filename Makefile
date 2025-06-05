# 🚀 GoIRC Makefile
# Modern IRC Client built with Go and Bubble Tea

# ==================== CONFIGURATION ====================
BINARY_NAME := goirc
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION := $(shell go version | awk '{print $$3}')

# Build flags
LDFLAGS := -ldflags="-s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.buildTime=$(BUILD_TIME) \
	-X main.goVersion=$(GO_VERSION)"

# Colors for output
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
RED := \033[31m
RESET := \033[0m
BOLD := \033[1m

# Platform detection
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

# ==================== MAIN TARGETS ====================

.PHONY: all
all: info deps fmt lint test build ## Run all checks and build

.PHONY: info
info: ## Show build information
	@echo "$(BOLD)$(BLUE)🚀 GoIRC Build Information$(RESET)"
	@echo "$(GREEN)Version:$(RESET)    $(VERSION)"
	@echo "$(GREEN)Commit:$(RESET)     $(COMMIT)"
	@echo "$(GREEN)Build Time:$(RESET) $(BUILD_TIME)"
	@echo "$(GREEN)Go Version:$(RESET) $(GO_VERSION)"
	@echo "$(GREEN)Platform:$(RESET)   $(UNAME_S)/$(UNAME_M)"
	@echo ""

# ==================== BUILD TARGETS ====================

.PHONY: build
build: ## Build for current platform
	@echo "$(BOLD)$(BLUE)🔨 Building GoIRC...$(RESET)"
	@go build $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "$(GREEN)✅ Build complete: $(BINARY_NAME)$(RESET)"

.PHONY: build-fast
build-fast: ## Fast build without optimizations (for development)
	@echo "$(BOLD)$(BLUE)⚡ Fast building GoIRC...$(RESET)"
	@go build -o $(BINARY_NAME) .
	@echo "$(GREEN)✅ Fast build complete: $(BINARY_NAME)$(RESET)"

.PHONY: build-all
build-all: clean ## Cross-compile for all platforms
	@echo "$(BOLD)$(BLUE)🌍 Cross-compiling for all platforms...$(RESET)"
	@mkdir -p dist/
	
	@echo "$(YELLOW)Building Linux AMD64...$(RESET)"
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)_linux_amd64 .
	@tar -czf dist/$(BINARY_NAME)_linux_amd64.tar.gz -C dist $(BINARY_NAME)_linux_amd64
	
	@echo "$(YELLOW)Building Linux ARM64...$(RESET)"
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)_linux_arm64 .
	@tar -czf dist/$(BINARY_NAME)_linux_arm64.tar.gz -C dist $(BINARY_NAME)_linux_arm64
	
	@echo "$(YELLOW)Building macOS AMD64...$(RESET)"
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)_darwin_amd64 .
	@tar -czf dist/$(BINARY_NAME)_darwin_amd64.tar.gz -C dist $(BINARY_NAME)_darwin_amd64
	
	@echo "$(YELLOW)Building macOS ARM64...$(RESET)"
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)_darwin_arm64 .
	@tar -czf dist/$(BINARY_NAME)_darwin_arm64.tar.gz -C dist $(BINARY_NAME)_darwin_arm64
	
	@echo "$(YELLOW)Building Windows AMD64...$(RESET)"
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)_windows_amd64.exe .
	@cd dist && zip -q $(BINARY_NAME)_windows_amd64.zip $(BINARY_NAME)_windows_amd64.exe

	@echo "$(GREEN)✅ Cross-compilation complete! Check dist/ directory$(RESET)"
	@ls -la dist/

# ==================== DEVELOPMENT TARGETS ====================

.PHONY: deps
deps: ## Install dependencies
	@echo "$(BOLD)$(BLUE)📦 Installing dependencies...$(RESET)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✅ Dependencies installed$(RESET)"

.PHONY: dev-deps
dev-deps: deps ## Install development dependencies
	@echo "$(BOLD)$(BLUE)🛠️  Installing development tools...$(RESET)"
	@command -v golangci-lint >/dev/null 2>&1 || \
		(echo "Installing golangci-lint..." && \
		 go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@command -v gosec >/dev/null 2>&1 || \
		(echo "Installing gosec..." && \
		 go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest)
	@command -v gofumpt >/dev/null 2>&1 || \
		(echo "Installing gofumpt..." && \
		 go install mvdan.cc/gofumpt@latest)
	@echo "$(GREEN)✅ Development tools installed$(RESET)"

.PHONY: run
run: build ## Build and run the application
	@echo "$(BOLD)$(BLUE)▶️  Running GoIRC...$(RESET)"
	@./$(BINARY_NAME)

.PHONY: run-dev
run-dev: ## Run without building (uses go run)
	@echo "$(BOLD)$(BLUE)🚀 Running GoIRC in development mode...$(RESET)"
	@go run .

# ==================== QUALITY TARGETS ====================

.PHONY: fmt
fmt: ## Format code
	@echo "$(BOLD)$(BLUE)🎨 Formatting code...$(RESET)"
	@go fmt ./...
	@if command -v gofumpt >/dev/null 2>&1; then \
		gofumpt -l -w .; \
	fi
	@echo "$(GREEN)✅ Code formatted$(RESET)"

.PHONY: lint
lint: ## Lint code
	@echo "$(BOLD)$(BLUE)🔍 Linting code...$(RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
		echo "$(GREEN)✅ Linting complete$(RESET)"; \
	else \
		echo "$(YELLOW)⚠️  golangci-lint not found. Run 'make dev-deps' to install$(RESET)"; \
	fi

.PHONY: sec
sec: ## Security scan
	@echo "$(BOLD)$(BLUE)🛡️  Running security scan...$(RESET)"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
		echo "$(GREEN)✅ Security scan complete$(RESET)"; \
	else \
		echo "$(YELLOW)⚠️  gosec not found. Run 'make dev-deps' to install$(RESET)"; \
	fi

.PHONY: test
test: ## Run tests
	@echo "$(BOLD)$(BLUE)🧪 Running tests...$(RESET)"
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)✅ Tests complete$(RESET)"

.PHONY: test-coverage
test-coverage: test ## Run tests with coverage report
	@echo "$(BOLD)$(BLUE)📊 Generating coverage report...$(RESET)"
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ Coverage report generated: coverage.html$(RESET)"

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "$(BOLD)$(BLUE)⚡ Running benchmarks...$(RESET)"
	@go test -bench=. -benchmem ./...

# ==================== UTILITY TARGETS ====================

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(BOLD)$(BLUE)🧹 Cleaning build artifacts...$(RESET)"
	@rm -f $(BINARY_NAME)*
	@rm -f coverage.out coverage.html
	@rm -rf dist/
	@echo "$(GREEN)✅ Clean complete$(RESET)"

.PHONY: install
install: build ## Install binary to system
	@echo "$(BOLD)$(BLUE)📥 Installing GoIRC...$(RESET)"
	@install -Dm755 $(BINARY_NAME) ~/.local/bin/$(BINARY_NAME)
	@echo "$(GREEN)✅ GoIRC installed to ~/.local/bin/$(BINARY_NAME)$(RESET)"
	@echo "$(YELLOW)💡 Make sure ~/.local/bin is in your PATH$(RESET)"

.PHONY: uninstall
uninstall: ## Uninstall binary from system
	@echo "$(BOLD)$(BLUE)🗑️  Uninstalling GoIRC...$(RESET)"
	@rm -f ~/.local/bin/$(BINARY_NAME)
	@echo "$(GREEN)✅ GoIRC uninstalled$(RESET)"

# ==================== RELEASE TARGETS ====================

.PHONY: release
release: ## Create a new release (usage: make release VERSION=v1.0.0)
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)❌ VERSION is required. Usage: make release VERSION=v1.0.0$(RESET)"; \
		exit 1; \
	fi
	@echo "$(BOLD)$(BLUE)🚀 Creating release $(VERSION)...$(RESET)"
	@./scripts/release.sh $(VERSION)

.PHONY: changelog
changelog: ## Generate changelog
	@echo "$(BOLD)$(BLUE)📝 Generating changelog...$(RESET)"
	@git log --oneline --pretty=format:"- %s" HEAD...$(shell git describe --tags --abbrev=0 2>/dev/null || echo "HEAD") > CHANGELOG.tmp
	@echo "$(GREEN)✅ Changelog generated: CHANGELOG.tmp$(RESET)"

# ==================== DOCKER TARGETS ====================

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(BOLD)$(BLUE)🐳 Building Docker image...$(RESET)"
	@docker build -t $(BINARY_NAME):$(VERSION) -t $(BINARY_NAME):latest .
	@echo "$(GREEN)✅ Docker image built$(RESET)"

.PHONY: docker-run
docker-run: docker-build ## Run in Docker container
	@echo "$(BOLD)$(BLUE)🐳 Running GoIRC in Docker...$(RESET)"
	@docker run -it --rm $(BINARY_NAME):latest

# ==================== WORKFLOW TARGETS ====================

.PHONY: dev
dev: dev-deps fmt lint test build ## Full development workflow

.PHONY: ci
ci: deps fmt lint sec test ## Continuous integration workflow

.PHONY: check
check: fmt lint test ## Quick check (fmt + lint + test)

# ==================== HELP TARGET ====================

.PHONY: help
help: ## Show this help message
	@echo "$(BOLD)$(BLUE)🚀 GoIRC Makefile$(RESET)"
	@echo ""
	@echo "$(BOLD)Available targets:$(RESET)"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ { \
		printf "  $(GREEN)%-15s$(RESET) %s\n", $$1, $$2 \
	}' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(BOLD)Examples:$(RESET)"
	@echo "  make dev          # Full development workflow"
	@echo "  make build        # Build for current platform"
	@echo "  make build-all    # Cross-compile for all platforms"
	@echo "  make install      # Install to ~/.local/bin"
	@echo "  make release VERSION=v1.0.0  # Create release"

# Default target when no target is specified
.DEFAULT_GOAL := help
