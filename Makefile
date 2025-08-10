BINARY_NAME=pihole-analyzer
TEST_BINARY=pihole-analyzer-test
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

.PHONY: build build-test build-all run clean install-deps help setup-pihole analyze-pihole pre-push ci-test test-mode cache-info cache-clean docker-build docker-dev docker-prod version release-setup commit release-dry-run release-status

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Semantic Release Commands
version: ## Show current version information
	@echo "📦 Version Information:"
	@echo "Current Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Branch: $$(git branch --show-current 2>/dev/null || echo 'unknown')"
	@echo "Git Commit: $$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "Git Status: $$(git status --porcelain | wc -l | xargs echo) uncommitted changes"

release-setup: ## Install semantic-release dependencies (requires Node.js)
	@echo "🔧 Setting up semantic release..."
	@if command -v npm >/dev/null 2>&1; then \
		echo "📦 Installing semantic-release dependencies..."; \
		npm install; \
		echo "🔗 Setting up Git hooks..."; \
		npx husky install; \
		echo ""; \
		echo "✅ Semantic release setup complete!"; \
		echo ""; \
		echo "💡 Next steps:"; \
		echo "  • Use 'make commit' for conventional commits"; \
		echo "  • Use 'make release-dry-run' to test releases"; \
		echo "  • Push to main/release branches for automated releases"; \
	else \
		echo "❌ Node.js not found. Install options:"; \
		echo ""; \
		echo "🍺 Via Homebrew (recommended):"; \
		echo "   brew install node"; \
		echo ""; \
		echo "📦 Via Node Version Manager:"; \
		echo "   curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash"; \
		echo "   nvm install node"; \
		echo ""; \
		echo "🌐 Direct download:"; \
		echo "   https://nodejs.org/"; \
		echo ""; \
		echo "After installing Node.js, run 'make release-setup' again"; \
		exit 1; \
	fi

commit: ## Interactive commit with conventional format
	@echo "🚀 Creating conventional commit..."
	@if command -v npx >/dev/null 2>&1 && [ -d "node_modules" ]; then \
		npx git-cz; \
	else \
		echo "⚠️  Commitizen not available. Using manual commit format..."; \
		echo "Format: <type>[scope]: <description>"; \
		echo "Types: feat, fix, docs, style, refactor, test, chore, perf, ci, build, revert"; \
		echo ""; \
		read -p "Enter commit message: " msg; \
		git commit -m "$$msg"; \
	fi

release-dry-run: ## Run semantic-release in dry-run mode
	@echo "🧪 Running release dry-run..."
	@if command -v npx >/dev/null 2>&1 && [ -d "node_modules" ]; then \
		npx semantic-release --dry-run; \
	else \
		echo "❌ semantic-release not available. Run 'make release-setup' first"; \
		exit 1; \
	fi

release-status: ## Check semantic-release setup status
	@echo "🔍 Semantic Release Status:"
	@echo ""
	@if command -v node >/dev/null 2>&1; then \
		echo "✅ Node.js: $$(node --version)"; \
	else \
		echo "❌ Node.js: Not installed"; \
	fi
	@if command -v npm >/dev/null 2>&1; then \
		echo "✅ npm: $$(npm --version)"; \
	else \
		echo "❌ npm: Not available"; \
	fi
	@if [ -f "package.json" ]; then \
		echo "✅ package.json: Present"; \
	else \
		echo "❌ package.json: Missing"; \
	fi
	@if [ -d "node_modules" ]; then \
		echo "✅ Dependencies: Installed"; \
	else \
		echo "❌ Dependencies: Not installed (run 'make release-setup')"; \
	fi
	@if [ -f ".husky/commit-msg" ]; then \
		echo "✅ Git hooks: Configured"; \
	else \
		echo "⚠️  Git hooks: Not configured"; \
	fi

protect-release-branch: ## Protect a release branch (usage: make protect-release-branch VERSION=v1.1)
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ VERSION is required. Usage: make protect-release-branch VERSION=v1.1"; \
		exit 1; \
	fi
	@./scripts/protect-release-branch.sh $(VERSION)

# Container Build Commands

# Enhanced Build Commands
build-all: build build-test ## Build all binaries
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
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg BUILDKIT_INLINE_CACHE=1 \
		-t pihole-analyzer:latest \
		-t pihole-analyzer:$(VERSION) .; \
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
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--target production \
		-t pihole-analyzer:latest \
		-t pihole-analyzer:$(VERSION) .

docker-test: ## Run tests in Docker container
	@echo "🧪 Running tests in Docker container..."
	docker-compose run --rm pihole-analyzer-test

docker-test-quick: docker-build ## Quick test of Docker container
	@echo "🧪 Quick testing Docker container..."
	@docker run --rm pihole-analyzer:latest --help
	@echo "✅ Docker container test completed"

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

.PHONY: docker-build docker-build-dev docker-build-prod docker-build-multi docker-test docker-test-quick docker-dev docker-prod docker-clean docker-logs docker-shell docker-push benchmark analyze-size watch cache-info cache-warm cache-clean fast-build dev-setup
