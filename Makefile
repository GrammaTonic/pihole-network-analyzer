BINARY_NAME=pihole-analyzer
TEST_BINARY=pihole-analyzer-test
PHASE5_BINARY=pihole-analyzer-phase5
PIHOLE_CONFIG=pihole-config.json

# Build optimization variables
GOCACHE_DIR=$(shell go env GOCACHE)
GOMOD_CACHE=$(shell go env GOMODCACHE)
BUILD_FLAGS=-trimpath
LDFLAGS=-s -w
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Enable Go build cache and optimization
export GOCACHE
export GOMODCACHE

.PHONY: build build-test build-phase5 build-all run clean install-deps help setup-pihole analyze-pihole pre-push ci-test test-mode cache-info cache-clean phase5-build phase5-test phase5-deploy docker-api-only

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Phase 5 Commands
phase5-build: ## Build Phase 5 analyzer with API-first architecture
	@echo "🚀 Building Phase 5 analyzer..."
	@start_time=$$(date +%s); \
	CGO_ENABLED=0 go build $(BUILD_FLAGS) -ldflags="$(LDFLAGS) -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)" \
		-o $(PHASE5_BINARY) ./cmd/pihole-analyzer-phase5/; \
	end_time=$$(date +%s); \
	duration=$$((end_time - start_time)); \
	echo "✅ Phase 5 build completed in $${duration}s"

phase5-test: ## Test Phase 5 analyzer
	@echo "🧪 Testing Phase 5 analyzer..."
	@go test -v ./internal/analyzer/phase5_*.go
	@echo "✅ Phase 5 tests completed"

phase5-integration: ## Run Phase 5 integration tests with API scenarios
	@echo "🔗 Running Phase 5 integration tests..."
	@./scripts/integration-test.sh --phase5
	@echo "✅ Phase 5 integration tests completed"

phase5-deploy: build-phase5 ## Deploy Phase 5 analyzer (build + install)
	@echo "📦 Deploying Phase 5 analyzer..."
	@sudo cp $(PHASE5_BINARY) /usr/local/bin/
	@echo "✅ Phase 5 analyzer deployed to /usr/local/bin/$(PHASE5_BINARY)"

phase5-validate: ## Validate Phase 5 configuration
	@echo "✅ Validating Phase 5 configuration..."
	@if [ -f $(PHASE5_BINARY) ]; then \
		./$(PHASE5_BINARY) --help; \
	else \
		echo "❌ Phase 5 binary not found. Run 'make phase5-build' first"; \
		exit 1; \
	fi

# API-Only Container Commands
docker-api-only: ## Build API-only container variant
	@echo "🐳 Building API-only container..."
	@docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--platform linux/amd64,linux/arm64,linux/arm/v7 \
		-t pihole-analyzer:api-only \
		-f Dockerfile.api-only .
	@echo "✅ API-only container built"

docker-api-only-test: docker-api-only ## Test API-only container deployment
	@echo "🧪 Testing API-only container..."
	@docker run --rm pihole-analyzer:api-only --help
	@echo "✅ API-only container test completed"

docker-api-only-deploy: docker-api-only ## Deploy API-only container environment
	@echo "🚀 Deploying API-only container environment..."
	@docker-compose -f docker-compose.api-only.yml up -d
	@echo "✅ API-only environment deployed"

# Enhanced Build Commands
build-all: build build-test build-phase5 ## Build all binaries including Phase 5
	@echo "✅ All binaries built successfully"

# Cache management
cache-info: ## Show build cache information
	@echo "📊 Build Cache Information:"
	@echo "GOCACHE: $$(go env GOCACHE)"
	@echo "GOMODCACHE: $$(go env GOMODCACHE)"
	@echo ""
	@echo "Cache sizes:"
	@if [ -d "$$(go env GOCACHE)" ]; then \
		echo "  Build cache: $$(du -sh "$$(go env GOCACHE)" 2>/dev/null | cut -f1)"; \
	fi
	@if [ -d "$$(go env GOMODCACHE)" ]; then \
		echo "  Module cache: $$(du -sh "$$(go env GOMODCACHE)" 2>/dev/null | cut -f1)"; \
	fi

cache-warm: ## Warm up build caches for faster builds
	@echo "🔥 Warming up build caches..."
	@./scripts/cache-warm.sh

cache-clean: ## Clean Go build and module caches
	@echo "🧹 Cleaning build caches..."
	@go clean -cache -modcache -testcache
	@echo "✅ Caches cleaned"

install-deps: ## Install Go dependencies with caching
	@echo "📦 Installing dependencies with caching..."
	@start_time=$$(date +%s); \
	go mod tidy; \
	go mod download; \
	end_time=$$(date +%s); \
	duration=$$((end_time - start_time)); \
	echo "✅ Dependencies installed in $${duration}s"

build: ## Build the production application with caching
	@echo "🚀 Building $(BINARY_NAME) with optimization..."
	@start_time=$$(date +%s); \
	go build $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/pihole-analyzer; \
	end_time=$$(date +%s); \
	duration=$$((end_time - start_time)); \
	size=$$(ls -lah $(BINARY_NAME) | awk '{print $$5}'); \
	echo "✅ Build completed in $${duration}s (size: $${size})"

build-test: ## Build the test runner binary with caching
	@echo "🧪 Building $(TEST_BINARY) with optimization..."
	@start_time=$$(date +%s); \
	go build $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o $(TEST_BINARY) ./cmd/pihole-analyzer-test; \
	end_time=$$(date +%s); \
	duration=$$((end_time - start_time)); \
	size=$$(ls -lah $(TEST_BINARY) | awk '{print $$5}'); \
	echo "✅ Test build completed in $${duration}s (size: $${size})"

# Cached incremental build - only rebuild if sources changed
build-cached: $(BINARY_NAME) ## Incremental build (only if sources changed)

$(BINARY_NAME): $(shell find . -name "*.go" -not -path "./vendor/*")
	@echo "🔄 Incremental build triggered..."
	@$(MAKE) build

test-mode: build-test ## Run in test mode with mock data
	./$(TEST_BINARY) --test

analyze-pihole: build ## Analyze Pi-hole live data (requires pihole-config.json)
	./$(BINARY_NAME) --pihole $(PIHOLE_CONFIG)

setup-pihole: build ## Setup Pi-hole API configuration
	./$(BINARY_NAME) --pihole-setup

test-pihole: build ## Test Pi-hole connection and analyze data
	@if [ ! -f $(PIHOLE_CONFIG) ]; then \
		echo "Pi-hole config not found. Run 'make setup-pihole' first."; \
		exit 1; \
	fi
	./$(BINARY_NAME) --pihole $(PIHOLE_CONFIG)

clean: ## Clean build artifacts
	rm -f $(BINARY_NAME)
	rm -f $(TEST_BINARY)
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

ci-test: ## Run the same tests as CI locally with caching
	@echo "🧪 Running CI tests locally with caching optimizations..."
	@start_time=$$(date +%s); \
	go mod tidy; \
	go mod verify; \
	go mod download; \
	echo "Building applications..."; \
	go build $(BUILD_FLAGS) -o pihole-analyzer ./cmd/pihole-analyzer; \
	go build $(BUILD_FLAGS) -o pihole-analyzer-test ./cmd/pihole-analyzer-test; \
	echo "Running test mode..."; \
	./pihole-analyzer-test --test; \
	echo "Checking code formatting..."; \
	if [ "$$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "❌ Code formatting issues found:"; \
		gofmt -s -l .; \
		echo "Fix with: make fmt"; \
		exit 1; \
	fi; \
	echo "Running go vet..."; \
	go vet ./...; \
	end_time=$$(date +%s); \
	duration=$$((end_time - start_time)); \
	echo "✅ All CI tests passed in $${duration}s!"

multi-build: build ## Test multi-platform builds (like CI) with parallel compilation
	@echo "🏗️ Testing multi-platform builds with parallel compilation..."
	@start_time=$$(date +%s); \
	echo "Starting parallel cross-platform builds..."; \
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o /tmp/test-linux-amd64 ./cmd/pihole-analyzer & \
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o /tmp/test-windows-amd64.exe ./cmd/pihole-analyzer & \
	GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) -o /tmp/test-darwin-arm64 ./cmd/pihole-analyzer & \
	wait; \
	echo "✅ All platform builds successful!"; \
	rm -f /tmp/test-*; \
	end_time=$$(date +%s); \
	duration=$$((end_time - start_time)); \
	echo "✅ Multi-platform builds completed in $${duration}s"

# Enhanced build targets with timing and optimization
fast-build: ## Fast optimized build with all caching enabled
	@echo "⚡ Fast build with maximum caching..."
	@$(MAKE) cache-info
	@$(MAKE) build-cached
	@$(MAKE) build-test

release-build: ci-test multi-build ## Full release build validation with optimizations
	@echo "🚀 Release build validation with optimizations..."
	@start_time=$$(date +%s); \
	$(MAKE) cache-info; \
	end_time=$$(date +%s); \
	duration=$$((end_time - start_time)); \
	echo "✅ Release build validation complete in $${duration}s!"
	@echo "Your code is ready for main/master branch."

# Development workflow with caching
dev-setup: install-deps cache-info ## Setup development environment with caching info
	@echo "🔧 Development environment setup complete!"
	@echo "Use 'make fast-build' for optimized builds"

# Watch and rebuild (requires entr: brew install entr)
watch: ## Watch for changes and rebuild automatically (requires entr)
	@echo "👀 Watching for changes... (Ctrl+C to stop)"
	@find . -name "*.go" | entr -r make build-cached

# Docker build targets with caching
docker-build: ## Build Docker image with caching
	@echo "🐳 Building Docker image with caching..."
	@start_time=$$(date +%s); \
	docker build --build-arg BUILDKIT_INLINE_CACHE=1 -t pihole-analyzer:latest .; \
	end_time=$$(date +%s); \
	duration=$$((end_time - start_time)); \
	echo "✅ Docker build completed in $${duration}s"

docker-build-dev: ## Build development Docker image
	@echo "🐳 Building development Docker image..."
	docker-compose -f docker-compose.dev.yml build pihole-analyzer-dev

docker-build-prod: ## Build production Docker image
	@echo "🐳 Building production Docker image..."
	docker-compose -f docker-compose.prod.yml build pihole-analyzer

docker-build-multi: ## Build multi-architecture images
	@echo "🐳 Building multi-architecture images..."
	docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 \
		--target production -t pihole-analyzer:latest .

docker-test: ## Run tests in Docker container
	@echo "🧪 Running tests in Docker container..."
	docker-compose run --rm pihole-analyzer-test

docker-dev: ## Start development environment with Docker
	@echo "🔧 Starting development environment..."
	docker-compose -f docker-compose.dev.yml up -d pihole-analyzer-dev
	@echo "✅ Development container started. Use 'docker exec -it pihole-analyzer-dev sh' to access"

docker-prod: ## Start production environment with Docker
	@echo "🚀 Starting production environment..."
	docker-compose -f docker-compose.prod.yml up -d
	@echo "✅ Production containers started"

docker-clean: ## Clean Docker images and containers
	@echo "🧹 Cleaning Docker resources..."
	docker-compose down --rmi all --volumes --remove-orphans
	docker-compose -f docker-compose.dev.yml down --rmi all --volumes --remove-orphans
	docker-compose -f docker-compose.prod.yml down --rmi all --volumes --remove-orphans
	docker system prune -f

docker-logs: ## Show container logs
	@echo "📝 Showing container logs..."
	docker-compose logs -f

docker-shell: ## Access development container shell
	@echo "🐚 Accessing development container..."
	docker exec -it pihole-analyzer-dev sh

docker-push: ## Push images to registry (requires authentication)
	@echo "📤 Pushing images to registry..."
	docker tag pihole-analyzer:latest ghcr.io/grammatonic/pihole-analyzer:latest
	docker push ghcr.io/grammatonic/pihole-analyzer:latest

# Performance benchmarking
benchmark: ## Run performance benchmarks
	@echo "📊 Running performance benchmarks..."
	@start_time=$$(date +%s); \
	go test -bench=. -benchmem -run=Benchmark ./...; \
	end_time=$$(date +%s); \
	duration=$$((end_time - start_time)); \
	echo "✅ Benchmarks completed in $${duration}s"

# Build size analysis
analyze-size: build ## Analyze binary size and dependencies
	@echo "📏 Analyzing binary size..."
	@if command -v objdump >/dev/null 2>&1; then \
		echo "Binary sections:"; \
		objdump -h $(BINARY_NAME) 2>/dev/null || echo "objdump not available or not supported"; \
	fi
	@echo "Binary size: $$(ls -lah $(BINARY_NAME) | awk '{print $$5}')"
	@echo "Dependencies:"
	@go list -m all | head -10

.PHONY: docker-build docker-build-dev docker-build-prod docker-build-multi docker-test docker-dev docker-prod docker-clean docker-logs docker-shell docker-push benchmark analyze-size watch cache-info cache-warm cache-clean fast-build dev-setup
