# backend/Makefile
.PHONY: help build run test test-unit test-integration test-coverage test-race test-bench test-fast test-one clean-test clean docker-build docker-run lint fmt vet mod-tidy mod-download install-tools deps all test-handlers test-middleware test-k8s test-logging test-server test-watch check coverage-badge

# Variables
BINARY_NAME=api
DOCKER_IMAGE=rbac-generator-api
DOCKER_TAG=latest
GO_FILES=$(shell find . -name '*.go' -type f)
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Default target
.DEFAULT_GOAL := help

help: ## Show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# ============================================================================
# Build Targets
# ============================================================================

build: ## Build the binary
	@echo "Building..."
	@go build -o $(BINARY_NAME) ./cmd/api
	@echo "Build complete: $(BINARY_NAME)"

run: ## Run the application
	@go run ./cmd/api

# ============================================================================
# Test Targets
# ============================================================================

test: ## Run all tests
	@echo "Running all tests..."
	go test -v ./...

test-unit: ## Run unit tests only (fast tests)
	@echo "Running unit tests..."
	go test -v -short ./...

test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	go test -v -run Integration ./...

test-server: ## Run server startup/shutdown tests
	@echo "Running server tests..."
	go test -v ./test/...

test-handlers: ## Run handler tests only
	@echo "Running handler tests..."
	go test -v ./internal/handlers/...

test-middleware: ## Run middleware tests only
	@echo "Running middleware tests..."
	go test -v ./internal/middleware/...

test-k8s: ## Run Kubernetes client tests only
	@echo "Running k8s tests..."
	go test -v ./internal/k8s/...

test-logging: ## Run logging tests only
	@echo "Running logging tests..."
	go test -v ./internal/logging/...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -v -coverprofile=$(COVERAGE_FILE) ./...
	go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"
	@echo ""
	@echo "Coverage Summary:"
	@go tool cover -func=$(COVERAGE_FILE) | grep total

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	go test -v -race ./...

test-bench: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

test-fast: ## Run tests with verbose output and fail fast
	@echo "Running tests (fail fast mode)..."
	go test -v -failfast ./...

test-one: ## Run a specific test (will prompt for test name)
	@read -p "Enter test name: " test; \
	echo "Running test: $$test"; \
	go test -v -run $$test ./...

test-watch: ## Watch for changes and run tests (requires entr: brew install entr)
	@echo "Watching for changes..."
	@echo "Press Ctrl+C to stop"
	@find . -name '*.go' | entr -c go test -v ./...

check: test-unit test-race lint ## Run all quality checks (unit tests, race detector, linter)
	@echo ""
	@echo "✅ All checks passed!"

coverage-badge: test-coverage ## Generate coverage percentage for badge
	@echo -n "Coverage: "
	@go tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print $$3}'

# ============================================================================
# Cleanup Targets
# ============================================================================

clean-test: ## Clean test artifacts
	@echo "Cleaning test artifacts..."
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	rm -f /tmp/test-api-server
	go clean -testcache

clean: clean-test ## Clean build artifacts and test files
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@go clean
	@echo "Clean complete"

# ============================================================================
# Docker Targets
# ============================================================================

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p 8080:8080 \
		-v ~/.kube/config:/root/.kube/config:ro \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# ============================================================================
# Code Quality Targets
# ============================================================================

lint: ## Run linter (requires golangci-lint)
	@echo "Running linter..."
	@golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted"

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...
	@echo "Vet complete"

# ============================================================================
# Dependency Management
# ============================================================================

mod-tidy: ## Tidy go modules
	@echo "Tidying modules..."
	@go mod tidy
	@echo "Modules tidied"

mod-download: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@echo "Dependencies downloaded"

deps: mod-download ## Alias for mod-download

# ============================================================================
# Development Tools
# ============================================================================

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Installing goimports..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "Installing staticcheck..."
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@echo "Tools installed"

# ============================================================================
# Composite Targets
# ============================================================================

all: clean fmt vet lint test build ## Run all checks and build
	@echo ""
	@echo "✅ All tasks completed successfully!"

pre-commit: fmt vet lint test-unit ## Run pre-commit checks (fast)
	@echo ""
	@echo "✅ Pre-commit checks passed!"

ci: lint test-race test-coverage ## Run CI pipeline checks
	@echo ""
	@echo "✅ CI checks passed!"

# ============================================================================
# Development Workflow Targets
# ============================================================================

dev: ## Start development mode (build and run)
	@echo "Starting development mode..."
	@$(MAKE) build
	@$(MAKE) run

quick-test: ## Quick test (unit tests only, no race detector)
	@echo "Running quick tests..."
	@go test -short ./...

full-test: clean-test test-coverage test-race ## Run full test suite with coverage and race detection
	@echo ""
	@echo "✅ Full test suite completed!"

# ============================================================================
# Benchmarking and Profiling
# ============================================================================

bench-compare: ## Run benchmarks and save for comparison
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./... | tee bench-new.txt
	@echo ""
	@echo "Benchmark results saved to bench-new.txt"
	@echo "To compare with previous run: benchcmp bench-old.txt bench-new.txt"

profile-cpu: ## Run tests with CPU profiling
	@echo "Running tests with CPU profiling..."
	@go test -cpuprofile=cpu.prof -bench=. ./...
	@echo "CPU profile saved to cpu.prof"
	@echo "View with: go tool pprof cpu.prof"

profile-mem: ## Run tests with memory profiling
	@echo "Running tests with memory profiling..."
	@go test -memprofile=mem.prof -bench=. ./...
	@echo "Memory profile saved to mem.prof"
	@echo "View with: go tool pprof mem.prof"

# ============================================================================
# Documentation
# ============================================================================

docs: ## Generate and serve documentation
	@echo "Generating documentation..."
	@godoc -http=:6060 &
	@echo "Documentation server started at http://localhost:6060"
	@echo "Press Ctrl+C to stop"

# ============================================================================
# Statistics and Reporting
# ============================================================================

stats: ## Show code statistics
	@echo "Code Statistics:"
	@echo "================"
	@echo "Total Go files: $$(find . -name '*.go' | wc -l)"
	@echo "Total lines of code: $$(find . -name '*.go' -exec wc -l {} + | tail -1 | awk '{print $$1}')"
	@echo "Total test files: $$(find . -name '*_test.go' | wc -l)"
	@echo ""
	@echo "Package breakdown:"
	@find . -name '*.go' -not -name '*_test.go' | xargs wc -l | tail -1
	@echo ""
	@echo "Test file breakdown:"
	@find . -name '*_test.go' | xargs wc -l | tail -1

test-summary: test-coverage ## Show test coverage summary
	@echo ""
	@echo "Test Coverage Summary:"
	@echo "======================"
	@go tool cover -func=$(COVERAGE_FILE) | grep -v "100.0%" | head -20
	@echo ""
	@echo "Overall Coverage:"
	@go tool cover -func=$(COVERAGE_FILE) | grep total

# ============================================================================
# Verification Targets
# ============================================================================

verify-deps: ## Verify all dependencies are present
	@echo "Verifying dependencies..."
	@go mod verify
	@echo "✅ Dependencies verified"

verify-build: ## Verify the project builds successfully
	@echo "Verifying build..."
	@go build -v ./...
	@echo "✅ Build verification complete"

verify-all: verify-deps verify-build test-unit ## Verify dependencies, build, and run unit tests
	@echo ""
	@echo "✅ All verifications passed!"

# ============================================================================
# Documentation Targets
# ============================================================================

swagger-validate: ## Validate OpenAPI specification
	@echo "Validating OpenAPI specification..."
	@docker run --rm -v $(PWD):/workspace openapitools/openapi-generator-cli validate -i /workspace/api/swagger.json

swagger-serve: ## Serve Swagger UI locally
	@echo "Starting Swagger UI server..."
	@echo "Open http://localhost:8080/swagger in your browser"
	@$(MAKE) run

docs-gen: ## Generate API client from OpenAPI spec
	@echo "Generating API clients..."
	@docker run --rm -v $(PWD):/workspace openapitools/openapi-generator-cli generate \
		-i /workspace/api/swagger.json \
		-g go \
		-o /workspace/generated/client