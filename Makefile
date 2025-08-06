BINARY_NAME=dns-analyzer
SOURCE_FILE=main.go
CSV_FILE=test.csv
PIHOLE_CONFIG=pihole-config.json

.PHONY: build run clean install-deps help setup-pihole analyze-pihole

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install-deps: ## Install Go dependencies
	go mod tidy
	go mod download

build: ## Build the application
	go build -o $(BINARY_NAME) $(SOURCE_FILE)

run: build ## Build and run the application with test.csv
	./$(BINARY_NAME) $(CSV_FILE)

analyze: build ## Alias for run
	./$(BINARY_NAME) $(CSV_FILE)

run-with-file: build ## Run with a specific CSV file (usage: make run-with-file CSV_FILE=yourfile.csv)
	./$(BINARY_NAME) $(CSV_FILE)

setup-pihole: build ## Setup Pi-hole SSH configuration
	./$(BINARY_NAME) --pihole-setup

analyze-pihole: build ## Analyze Pi-hole live data (requires pihole-config.json)
	./$(BINARY_NAME) --pihole $(PIHOLE_CONFIG)

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
