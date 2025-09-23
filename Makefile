# Spotify API CLI Makefile

# Variables
BINARY_NAME=spotify-cli
BUILD_DIR=bin
GO_FILES=$(shell find . -name "*.go" -type f)
VERSION?=dev
GIT_COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/bambithedeer/spotify-api/internal/version.Version=$(VERSION) -X github.com/bambithedeer/spotify-api/internal/version.GitCommit=$(GIT_COMMIT) -X github.com/bambithedeer/spotify-api/internal/version.BuildTime=$(BUILD_TIME)"

# Default target
.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
.PHONY: build
build: ## Build the CLI binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/spotify-cli

.PHONY: build-all
build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/spotify-cli
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/spotify-cli
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/spotify-cli
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/spotify-cli

.PHONY: install
install: ## Install the binary to $GOPATH/bin
	go install $(LDFLAGS) ./cmd/spotify-cli

# Development targets
.PHONY: run
run: ## Run the CLI with help command
	go run ./cmd/spotify-cli --help

.PHONY: run-version
run-version: ## Run the CLI version command
	go run ./cmd/spotify-cli version

.PHONY: dev
dev: build ## Build and run the CLI
	./$(BUILD_DIR)/$(BINARY_NAME) --help

# Testing targets
.PHONY: test
test: ## Run all tests
	go test -v ./...

.PHONY: test-short
test-short: ## Run tests with short flag (skip slow tests)
	go test -short -v ./...

.PHONY: test-cli
test-cli: ## Run only CLI tests
	go test -v ./internal/cli/...

.PHONY: test-auth
test-auth: ## Run authentication command tests
	go test -v ./internal/cli -run "TestGenerateRandomString|TestMaskString|TestFormatDuration|TestAuthCommands"

.PHONY: test-cover
test-cover: ## Run tests with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-race
test-race: ## Run tests with race detection
	go test -race ./...

.PHONY: bench
bench: ## Run benchmarks
	go test -bench=. ./...

# Code quality targets
.PHONY: fmt
fmt: ## Format Go code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint (requires golangci-lint to be installed)
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run

.PHONY: check
check: fmt vet test ## Run format, vet, and tests

# Dependency management
.PHONY: deps
deps: ## Download dependencies
	go mod download

.PHONY: deps-update
deps-update: ## Update dependencies
	go get -u ./...
	go mod tidy

.PHONY: deps-verify
deps-verify: ## Verify dependencies
	go mod verify

.PHONY: tidy
tidy: ## Tidy go modules
	go mod tidy

# Cleanup targets
.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

.PHONY: clean-all
clean-all: clean ## Remove all generated files including vendor
	go clean -cache
	go clean -modcache

# Development setup
.PHONY: setup
setup: deps ## Setup development environment
	@echo "Setting up development environment..."
	@if [ ! -f .env ]; then \
		echo "Creating .env file from .env.example..."; \
		cp .env.example .env 2>/dev/null || echo "No .env.example found. Please create .env manually."; \
	fi
	@echo "Development environment ready!"

# Release targets
.PHONY: release
release: ## Create a release build with version tag
	@if [ -z "$(VERSION)" ] || [ "$(VERSION)" = "dev" ]; then \
		echo "Error: VERSION must be set for release (e.g., make release VERSION=v1.0.0)"; \
		exit 1; \
	fi
	@echo "Creating release $(VERSION)..."
	$(MAKE) clean
	$(MAKE) build-all VERSION=$(VERSION)
	@echo "Release $(VERSION) built successfully!"

# Documentation targets
.PHONY: docs
docs: ## Generate CLI documentation
	@mkdir -p docs
	go run ./cmd/spotify-cli docs --dir ./docs
	@echo "Documentation generated in docs/"

# Quick development commands
.PHONY: auth-status
auth-status: build ## Check authentication status
	./$(BUILD_DIR)/$(BINARY_NAME) auth status

.PHONY: auth-setup
auth-setup: build ## Run auth setup
	./$(BUILD_DIR)/$(BINARY_NAME) auth setup

.PHONY: quick-test
quick-test: ## Quick test of CLI build and basic commands
	@echo "Building CLI..."
	@$(MAKE) build > /dev/null
	@echo "Testing version command..."
	@./$(BUILD_DIR)/$(BINARY_NAME) version
	@echo "Testing help command..."
	@./$(BUILD_DIR)/$(BINARY_NAME) --help > /dev/null
	@echo "Testing auth help..."
	@./$(BUILD_DIR)/$(BINARY_NAME) auth --help > /dev/null
	@echo "✓ Quick test passed!"

# CI/CD targets
.PHONY: ci
ci: deps check test-race ## Run CI pipeline
	@echo "CI pipeline completed successfully!"

.PHONY: pre-commit
pre-commit: fmt vet test-short ## Run pre-commit checks
	@echo "Pre-commit checks completed!"

# Debug targets
.PHONY: debug
debug: ## Build with debug flags
	go build -gcflags="all=-N -l" -o $(BUILD_DIR)/$(BINARY_NAME)-debug ./cmd/spotify-cli

.PHONY: env
env: ## Show environment information
	@echo "Go version: $(shell go version)"
	@echo "Go path: $(shell go env GOPATH)"
	@echo "Go root: $(shell go env GOROOT)"
	@echo "Current directory: $(shell pwd)"
	@echo "Git commit: $(GIT_COMMIT)"
	@echo "Build time: $(BUILD_TIME)"