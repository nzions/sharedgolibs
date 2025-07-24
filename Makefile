.PHONY: regen-autoport build-tools build-binarycleaner build-servicemanager clean-bins help

build-tools: build-servicemanager build-binarycleaner ## Build all CLI tools

build-servicemanager: ## Build the servicemanager CLI tool
	@echo "Building servicemanager..."
	@mkdir -p bin
	@go build -o bin/servicemanager ./cmd/servicemanager/
	@echo "✓ servicemanager built successfully"

build-binarycleaner: ## Build the binarycleaner CLI tool
	@echo "Building binarycleaner..."
	@mkdir -p bin
	@go build -o bin/binarycleaner ./cmd/binarycleaner/
	@echo "✓ binarycleaner built successfully"

clean-bins: ## Remove all binary files from bin/ directory
	@echo "Cleaning bin/ directory..."
	@rm -rf bin/
	@echo "✓ bin/ directory cleaned"

regen-autoport: build-servicemanager ## Regenerate autoport configuration from docker-compose.yml
	@echo "Regenerating autoport configuration..."
	@if [ ! -f "docker-compose.yml" ]; then \
		echo "Error: docker-compose.yml not found in current directory"; \
		echo "Please run this command from a directory containing docker-compose.yml"; \
		exit 1; \
	fi
	@./bin/servicemanager -generate=docker-compose.yml
	@echo "✓ Autoport configuration regenerated successfully"

help: ## Show available make targets
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
