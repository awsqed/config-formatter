# Config Formatter Makefile

# Binary name
BINARY_NAME=config-formatter
VERSION?=dev
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Build directories
BUILD_DIR=./build
DIST_DIR=./dist

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Main package
MAIN_PACKAGE=./main.go

# Color output
COLOR_RESET=\033[0m
COLOR_BOLD=\033[1m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m
COLOR_BLUE=\033[34m

.PHONY: all build clean test run help install \
        build-linux build-darwin build-windows \
        build-linux-amd64 build-linux-arm64 \
        build-darwin-amd64 build-darwin-arm64 \
        build-windows-amd64 build-windows-arm64 \
        build-all

## help: Display this help message
help:
	@echo "$(COLOR_BOLD)Docker Compose Formatter - Makefile Commands$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_GREEN)Development:$(COLOR_RESET)"
	@echo "  make run FILE=<file>    - Run without building (e.g., make run FILE=tailscale.yml)"
	@echo "  make test               - Run tests"
	@echo "  make clean              - Remove build artifacts"
	@echo ""
	@echo "$(COLOR_GREEN)Building:$(COLOR_RESET)"
	@echo "  make build              - Build for current platform"
	@echo "  make build-all          - Build for all platforms"
	@echo "  make install            - Build and install to GOPATH/bin"
	@echo ""
	@echo "$(COLOR_GREEN)Platform-specific builds:$(COLOR_RESET)"
	@echo "  make build-linux        - Build for all Linux platforms"
	@echo "  make build-darwin       - Build for all macOS platforms"
	@echo "  make build-windows      - Build for all Windows platforms"
	@echo ""
	@echo "$(COLOR_YELLOW)Individual platform builds:$(COLOR_RESET)"
	@echo "  make build-linux-amd64"
	@echo "  make build-linux-arm64"
	@echo "  make build-darwin-amd64"
	@echo "  make build-darwin-arm64"
	@echo "  make build-windows-amd64"
	@echo "  make build-windows-arm64"
	@echo ""
	@echo "$(COLOR_BLUE)Examples:$(COLOR_RESET)"
	@echo "  make run FILE=tailscale.yml"
	@echo "  make run FILE=vaultwarden.yml ARGS=\"-check\""
	@echo "  make build-all"

## run: Run without building (requires FILE=<filename>)
run:
	@if [ -z "$(FILE)" ]; then \
		echo "$(COLOR_YELLOW)Error: FILE parameter is required$(COLOR_RESET)"; \
		echo "Usage: make run FILE=<filename> [ARGS=\"additional args\"]"; \
		echo "Example: make run FILE=tailscale.yml"; \
		exit 1; \
	fi
	@echo "$(COLOR_GREEN)Running formatter on $(FILE)...$(COLOR_RESET)"
	$(GORUN) $(MAIN_PACKAGE) -input $(FILE) $(ARGS)

## build: Build for current platform
build:
	@echo "$(COLOR_GREEN)Building $(BINARY_NAME) for current platform...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(COLOR_GREEN)✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)"

## install: Build and install to GOPATH/bin
install:
	@echo "$(COLOR_GREEN)Installing $(BINARY_NAME)...$(COLOR_RESET)"
	$(GOCMD) install $(LDFLAGS) $(MAIN_PACKAGE)
	@echo "$(COLOR_GREEN)✓ Installed to GOPATH/bin$(COLOR_RESET)"

## test: Run tests
test:
	@echo "$(COLOR_GREEN)Running tests...$(COLOR_RESET)"
	$(GOTEST) -v ./...

## clean: Remove build artifacts
clean:
	@echo "$(COLOR_YELLOW)Cleaning build artifacts...$(COLOR_RESET)"
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	@echo "$(COLOR_GREEN)✓ Clean complete$(COLOR_RESET)"

## build-all: Build for all platforms
build-all: build-linux build-darwin build-windows
	@echo "$(COLOR_GREEN)✓ All platform builds complete$(COLOR_RESET)"

## build-linux: Build for all Linux platforms
build-linux: build-linux-amd64 build-linux-arm64

## build-darwin: Build for all macOS platforms
build-darwin: build-darwin-amd64 build-darwin-arm64

## build-windows: Build for all Windows platforms
build-windows: build-windows-amd64 build-windows-arm64

## build-linux-amd64: Build for Linux AMD64
build-linux-amd64:
	@echo "$(COLOR_BLUE)Building for Linux AMD64...$(COLOR_RESET)"
	@mkdir -p $(DIST_DIR)/linux-amd64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/linux-amd64/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(COLOR_GREEN)✓ $(DIST_DIR)/linux-amd64/$(BINARY_NAME)$(COLOR_RESET)"

## build-linux-arm64: Build for Linux ARM64
build-linux-arm64:
	@echo "$(COLOR_BLUE)Building for Linux ARM64...$(COLOR_RESET)"
	@mkdir -p $(DIST_DIR)/linux-arm64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/linux-arm64/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(COLOR_GREEN)✓ $(DIST_DIR)/linux-arm64/$(BINARY_NAME)$(COLOR_RESET)"

## build-darwin-amd64: Build for macOS AMD64 (Intel)
build-darwin-amd64:
	@echo "$(COLOR_BLUE)Building for macOS AMD64 (Intel)...$(COLOR_RESET)"
	@mkdir -p $(DIST_DIR)/darwin-amd64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/darwin-amd64/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(COLOR_GREEN)✓ $(DIST_DIR)/darwin-amd64/$(BINARY_NAME)$(COLOR_RESET)"

## build-darwin-arm64: Build for macOS ARM64 (Apple Silicon)
build-darwin-arm64:
	@echo "$(COLOR_BLUE)Building for macOS ARM64 (Apple Silicon)...$(COLOR_RESET)"
	@mkdir -p $(DIST_DIR)/darwin-arm64
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/darwin-arm64/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(COLOR_GREEN)✓ $(DIST_DIR)/darwin-arm64/$(BINARY_NAME)$(COLOR_RESET)"

## build-windows-amd64: Build for Windows AMD64
build-windows-amd64:
	@echo "$(COLOR_BLUE)Building for Windows AMD64...$(COLOR_RESET)"
	@mkdir -p $(DIST_DIR)/windows-amd64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/windows-amd64/$(BINARY_NAME).exe $(MAIN_PACKAGE)
	@echo "$(COLOR_GREEN)✓ $(DIST_DIR)/windows-amd64/$(BINARY_NAME).exe$(COLOR_RESET)"

## build-windows-arm64: Build for Windows ARM64
build-windows-arm64:
	@echo "$(COLOR_BLUE)Building for Windows ARM64...$(COLOR_RESET)"
	@mkdir -p $(DIST_DIR)/windows-arm64
	GOOS=windows GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/windows-arm64/$(BINARY_NAME).exe $(MAIN_PACKAGE)
	@echo "$(COLOR_GREEN)✓ $(DIST_DIR)/windows-arm64/$(BINARY_NAME).exe$(COLOR_RESET)"

## mod-tidy: Tidy go modules
mod-tidy:
	@echo "$(COLOR_GREEN)Tidying go modules...$(COLOR_RESET)"
	$(GOMOD) tidy
	@echo "$(COLOR_GREEN)✓ Modules tidied$(COLOR_RESET)"

## mod-download: Download go modules
mod-download:
	@echo "$(COLOR_GREEN)Downloading go modules...$(COLOR_RESET)"
	$(GOMOD) download
	@echo "$(COLOR_GREEN)✓ Modules downloaded$(COLOR_RESET)"

# Default target
.DEFAULT_GOAL := help
