# Go shared libraries Makefile - DRY and KISS principles
GO_VERSION := $(shell go version | cut -d' ' -f3)
BIN_DIR := bin
CMD_DIR := cmd

# Get all cmd directories dynamically
CMD_TARGETS := $(shell find $(CMD_DIR) -name "main.go" -exec dirname {} \; | sed 's|$(CMD_DIR)/||')
BUILD_TARGETS := $(addprefix build-, $(CMD_TARGETS))

.PHONY: all build test clean install help tidy mod-verify $(BUILD_TARGETS)

all: build test ## Build and test everything

build: $(BUILD_TARGETS) ## Build all CLI tools

test: ## Run all tests with coverage
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Tests completed. Coverage report: coverage.html"

# Generic build rule for all cmd targets
$(BUILD_TARGETS): build-%:
	@echo "Building $*..."
	@mkdir -p $(BIN_DIR)
	@go build -ldflags="-s -w" -o $(BIN_DIR)/$* ./$(CMD_DIR)/$*/
	@echo "✓ $* built successfully"

# Special build for gflag-demo (different path)
build-gflag-demo: ## Build the gflag-demo example
	@echo "Building gflag-demo..."
	@mkdir -p $(BIN_DIR)
	@go build -ldflags="-s -w" -o $(BIN_DIR)/gflag-demo ./pkg/gflag/examples/demo/
	@echo "✓ gflag-demo built successfully"

install: build-envinfo ## Install envinfo to ~/go/bin
	@echo "Installing envinfo to ~/go/bin..."
	@mkdir -p ~/go/bin
	@cp $(BIN_DIR)/envinfo ~/go/bin/envinfo
	@echo "✓ envinfo installed to ~/go/bin/envinfo"

clean: ## Clean all build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)/ coverage.out coverage.html
	@echo "✓ Clean completed"

tidy: ## Tidy go modules and format code
	@echo "Tidying modules and formatting..."
	@go mod tidy
	@go fmt ./...
	@echo "✓ Tidy completed"

mod-verify: ## Verify go modules
	@echo "Verifying modules..."
	@go mod verify
	@go mod download
	@echo "✓ Modules verified"

help: ## Show available make targets
	@echo "Shared Go Libraries Makefile"
	@echo "Go version: $(GO_VERSION)"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Available build targets:"
	@echo "  $(BUILD_TARGETS)"
