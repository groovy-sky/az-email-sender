# Makefile for azemailsender-cli

# Variables
APP_NAME := azemailsender-cli
PKG := github.com/groovy-sky/azemailsender
CMD_PATH := ./cmd/$(APP_NAME)
BUILD_DIR := dist
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Go build flags
GO_BUILD_FLAGS := -trimpath

# Platforms for cross-compilation
PLATFORMS := \
	darwin/amd64 \
	darwin/arm64 \
	linux/amd64 \
	linux/arm64 \
	windows/amd64

.PHONY: all build build-all clean test lint deps help install

# Default target
all: build

# Build for current platform
build:
	@echo "Building $(APP_NAME) for current platform..."
	@mkdir -p $(BUILD_DIR)
	go build $(GO_BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(CMD_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

# Build for all platforms
build-all: clean
	@echo "Building $(APP_NAME) for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		OS=$$(echo $$platform | cut -d'/' -f1); \
		ARCH=$$(echo $$platform | cut -d'/' -f2); \
		OUTPUT_NAME=$(APP_NAME)-$$OS-$$ARCH; \
		if [ "$$OS" = "windows" ]; then \
			OUTPUT_NAME=$$OUTPUT_NAME.exe; \
		fi; \
		echo "Building for $$OS/$$ARCH..."; \
		GOOS=$$OS GOARCH=$$ARCH go build $(GO_BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$$OUTPUT_NAME $(CMD_PATH); \
	done
	@echo "Cross-compilation complete. Binaries in $(BUILD_DIR)/"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run linting
lint:
	@echo "Running linting..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

# Install CLI locally
install: build
	@echo "Installing $(APP_NAME) to GOPATH/bin..."
	go install $(LDFLAGS) $(CMD_PATH)
	@echo "Installation complete. Run '$(APP_NAME) --help' to get started."

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	go clean

# Development build (with debug info)
dev-build:
	@echo "Building $(APP_NAME) for development..."
	@mkdir -p $(BUILD_DIR)
	go build -race $(GO_BUILD_FLAGS) -ldflags "-X main.version=$(VERSION)-dev -X main.commit=$(COMMIT) -X main.date=$(DATE)" -o $(BUILD_DIR)/$(APP_NAME)-dev $(CMD_PATH)
	@echo "Development build complete: $(BUILD_DIR)/$(APP_NAME)-dev"

# Run the CLI (for testing)
run: build
	@echo "Running $(APP_NAME)..."
	./$(BUILD_DIR)/$(APP_NAME) --help

# Create release archives
release: build-all
	@echo "Creating release archives..."
	@cd $(BUILD_DIR) && \
	for platform in $(PLATFORMS); do \
		OS=$$(echo $$platform | cut -d'/' -f1); \
		ARCH=$$(echo $$platform | cut -d'/' -f2); \
		BINARY_NAME=$(APP_NAME)-$$OS-$$ARCH; \
		if [ "$$OS" = "windows" ]; then \
			BINARY_NAME=$$BINARY_NAME.exe; \
		fi; \
		if [ -f "$$BINARY_NAME" ]; then \
			ARCHIVE_NAME=$(APP_NAME)-$(VERSION)-$$OS-$$ARCH; \
			if [ "$$OS" = "windows" ]; then \
				zip $$ARCHIVE_NAME.zip $$BINARY_NAME; \
			else \
				tar -czf $$ARCHIVE_NAME.tar.gz $$BINARY_NAME; \
			fi; \
			echo "Created $$ARCHIVE_NAME archive"; \
		fi; \
	done

# Show available targets
help:
	@echo "Available targets:"
	@echo "  build       - Build for current platform"
	@echo "  build-all   - Build for all platforms"
	@echo "  deps        - Install dependencies"
	@echo "  test        - Run tests"
	@echo "  lint        - Run linting"
	@echo "  install     - Install CLI locally"
	@echo "  clean       - Clean build artifacts"
	@echo "  dev-build   - Build for development"
	@echo "  run         - Build and run CLI"
	@echo "  release     - Create release archives"
	@echo "  help        - Show this help"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION     - Version to build (default: git describe or 'dev')"
	@echo "  BUILD_DIR   - Build output directory (default: dist)"