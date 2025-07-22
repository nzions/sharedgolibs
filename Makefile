.PHONY: regen-autoport

regen-autoport: ## Regenerate autoport configuration from docker-compose.yml
	@echo "Building servicemanager..."
	@mkdir -p bin
	@go build -o bin/servicemanager ./cmd/servicemanager/
	@echo "Regenerating autoport configuration..."
	@if [ ! -f "docker-compose.yml" ]; then \
		echo "Error: docker-compose.yml not found in current directory"; \
		echo "Please run this command from a directory containing docker-compose.yml"; \
		exit 1; \
	fi
	@./bin/servicemanager -generate=docker-compose.yml
	@echo "✓ Autoport configuration regenerated successfully"
