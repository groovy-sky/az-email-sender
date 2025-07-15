# Azure Email Sender CLI - Multi-platform build

# Application name
APP_NAME := azemailsender

# Build directory
BUILD_DIR := build

# Go build flags
LDFLAGS := -s -w
GCFLAGS := 

# Version info (can be overridden)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build targets
TARGETS := \
	linux/amd64 \
	linux/arm64 \
	windows/amd64 \
	darwin/amd64 \
	darwin/arm64

.PHONY: all clean build test help $(TARGETS)

# Default target
all: clean build

# Help target
help:
	@echo "Available targets:"
	@echo "  all       - Clean and build for all platforms"
	@echo "  build     - Build for all platforms"
	@echo "  clean     - Remove build directory"
	@echo "  test      - Run tests"
	@echo "  help      - Show this help"
	@echo ""
	@echo "Platform-specific targets:"
	@echo "  linux/amd64   - Build for Linux x86_64"
	@echo "  linux/arm64   - Build for Linux ARM64"
	@echo "  windows/amd64 - Build for Windows x86_64"
	@echo "  darwin/amd64  - Build for macOS x86_64"
	@echo "  darwin/arm64  - Build for macOS ARM64"

# Build for all platforms
build: $(TARGETS)

# Clean build directory
clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)

# Test target
test:
	@echo "Running tests..."
	@go test -v ./...

# Generic build rule for each platform
$(TARGETS):
	$(eval GOOS := $(word 1,$(subst /, ,$@)))
	$(eval GOARCH := $(word 2,$(subst /, ,$@)))
	$(eval BINARY_NAME := $(APP_NAME)$(if $(filter windows,$(GOOS)),.exe))
	$(eval OUTPUT_PATH := $(BUILD_DIR)/$(GOOS)-$(GOARCH)/$(BINARY_NAME))
	@echo "Building $(APP_NAME) for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(dir $(OUTPUT_PATH))
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-ldflags "$(LDFLAGS) -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)" \
		-gcflags "$(GCFLAGS)" \
		-o $(OUTPUT_PATH) \
		./cmd/$(APP_NAME)
	@echo "Built: $(OUTPUT_PATH)"

# Create release archives (optional)
release: build
	@echo "Creating release archives..."
	@cd $(BUILD_DIR) && \
	for dir in */; do \
		platform=$${dir%/}; \
		echo "Creating archive for $$platform..."; \
		if echo "$$platform" | grep -q windows; then \
			zip -r $(APP_NAME)-$(VERSION)-$$platform.zip $$platform/; \
		else \
			tar -czf $(APP_NAME)-$(VERSION)-$$platform.tar.gz $$platform/; \
		fi; \
	done
	@echo "Release archives created in $(BUILD_DIR)/"

# Install locally (builds for current platform)
install:
	@echo "Installing $(APP_NAME) locally..."
	@go install ./cmd/$(APP_NAME)

# Development build (current platform only)
dev:
	@echo "Building $(APP_NAME) for development..."
	@go build -o $(APP_NAME) ./cmd/$(APP_NAME)

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping lint"; \
	fi

# Vet code
vet:
	@echo "Vetting code..."
	@go vet ./...

# Check dependencies
deps:
	@echo "Checking dependencies..."
	@go mod tidy
	@go mod verify

# Full check (format, vet, lint, test)
check: fmt vet lint test
	@echo "All checks passed!"