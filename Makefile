BINARY_NAME=pihole-analyzer
PIHOLE_CONFIG=pihole-config.json

.PHONY: build run clean install-deps help setup-pihole analyze-pihole pre-push ci-test

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install-deps: ## Install Go dependencies
	go mod tidy
	go mod download

build: ## Build the application
	go build -o $(BINARY_NAME) ./cmd/pihole-analyzer

analyze-pihole: build ## Analyze Pi-hole live data (requires pihole-config.json)
	./$(BINARY_NAME) --pihole $(PIHOLE_CONFIG)

setup-pihole: build ## Setup Pi-hole SSH configuration
	./$(BINARY_NAME) --pihole-setup

test-pihole: build ## Test Pi-hole connection and analyze data
	@if [ ! -f $(PIHOLE_CONFIG) ]; then \
		echo "Pi-hole config not found. Run 'make setup-pihole' first."; \
		exit 1; \
	fi
	./$(BINARY_NAME) --pihole $(PIHOLE_CONFIG)

clean: ## Clean build artifacts
	rm -f $(BINARY_NAME)
	rm -f dns_usage_report_*.txt
	rm -f pihole-data-*.db

test: ## Run tests (if any)
	go test ./...

fmt: ## Format Go code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

all: install-deps fmt vet build ## Install deps, format, vet, and build

pre-push: ## Run comprehensive pre-push tests
	./scripts/pre-push-test.sh

ci-test: ## Run the same tests as CI locally
	@echo "🧪 Running CI tests locally..."
	go mod tidy
	go mod verify
	go mod download
	go build -o pihole-analyzer ./cmd/pihole-analyzer
	./pihole-analyzer --test
	@if [ "$$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "❌ Code formatting issues found:"; \
		gofmt -s -l .; \
		echo "Fix with: make fmt"; \
		exit 1; \
	fi
	go vet ./...
	@echo "✅ All CI tests passed!"

multi-build: build ## Test multi-platform builds (like CI)
	@echo "🏗️ Testing multi-platform builds..."
	GOOS=linux GOARCH=amd64 go build -o /tmp/test-linux-amd64 ./cmd/pihole-analyzer
	GOOS=windows GOARCH=amd64 go build -o /tmp/test-windows-amd64.exe ./cmd/pihole-analyzer
	GOOS=darwin GOARCH=arm64 go build -o /tmp/test-darwin-arm64 ./cmd/pihole-analyzer
	@echo "✅ All platform builds successful!"
	rm -f /tmp/test-*

feature-branch: ci-test ## Validate feature branch (run before push)
	@echo "🌿 Feature branch validation complete!"
	@echo "Your code is ready to push to a feature branch."

release-build: ci-test multi-build ## Full release build validation
	@echo "🚀 Release build validation complete!"
	@echo "Your code is ready for main/master branch."
