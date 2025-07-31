# Makefile for ntgrrc - Netgear Remote Control CLI

# Project information
BINARY_NAME := ntgrrc
PACKAGE := ntgrrc
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOCLEAN := $(GOCMD) clean
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt
GOVET := $(GOCMD) vet

# Build flags
LDFLAGS := -ldflags "-X main.VERSION=$(VERSION) -X main.BUILD_TIME=$(BUILD_TIME) -X main.GIT_COMMIT=$(GIT_COMMIT)"
BUILD_FLAGS := $(LDFLAGS) -trimpath

# Directories
BUILD_DIR := build
DIST_DIR := dist
COVERAGE_DIR := coverage
BIN_DIR := bin

# Coverage settings
COVERAGE_OUT := $(COVERAGE_DIR)/coverage.out
COVERAGE_HTML := $(COVERAGE_DIR)/coverage.html
COVERAGE_THRESHOLD := 80

# Cross-compilation targets
PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	linux/386 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64 \
	windows/386 \
	freebsd/amd64

# Default target
.DEFAULT_GOAL := help

## help: Show this help message
.PHONY: help
help:
	@echo "ntgrrc - Netgear Remote Control CLI"
	@echo ""
	@echo "Available targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## clean: Remove build artifacts and temporary files
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -rf $(COVERAGE_DIR)
	rm -rf $(BIN_DIR)
	rm -f $(BINARY_NAME)
	@echo "Clean complete"

## deps: Download and verify dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) verify
	@echo "Dependencies ready"

## tidy: Clean up go.mod and go.sum
.PHONY: tidy
tidy:
	@echo "Tidying go modules..."
	$(GOMOD) tidy
	@echo "Modules tidied"

## fmt: Format Go source code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...
	@echo "Code formatted"

## vet: Run go vet
.PHONY: vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...
	@echo "Vet complete"

## lint: Run basic linting (fmt + vet)
.PHONY: lint
lint: fmt vet
	@echo "Basic linting complete"

## build: Build the binary for current platform
.PHONY: build
build: deps
	@echo "Building $(BINARY_NAME) for current platform..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## build-dev: Build with development flags (no optimizations)
.PHONY: build-dev
build-dev: deps
	@echo "Building $(BINARY_NAME) for development..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -race -o $(BUILD_DIR)/$(BINARY_NAME)-dev .
	@echo "Development build complete: $(BUILD_DIR)/$(BINARY_NAME)-dev"

## build-examples: Build example programs
.PHONY: build-examples
build-examples: deps
	@echo "Building example programs..."
	@mkdir -p $(BUILD_DIR)/examples
	CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/examples/poe_status ./examples/poe_status.go
	CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/examples/poe_status_simple ./examples/poe_status_simple.go
	CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/examples/poe_management ./examples/poe_management.go
	@echo "Example programs built in $(BUILD_DIR)/examples/"
	@echo "Use --debug or -d flag for debug output"

## install: Install the binary to GOPATH/bin
.PHONY: install
install: deps
	@echo "Installing $(BINARY_NAME)..."
	$(GOBUILD) $(BUILD_FLAGS) -o $(GOPATH)/bin/$(BINARY_NAME) .
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

## run: Build and run the application with arguments
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

## run-dev: Run in development mode with race detection
.PHONY: run-dev
run-dev:
	@echo "Running $(BINARY_NAME) in development mode..."
	$(GOCMD) run -race . $(ARGS)

## test: Run all tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...
	@echo "Tests complete"

## test-short: Run tests with short flag
.PHONY: test-short
test-short:
	@echo "Running short tests..."
	$(GOTEST) -short -v ./...
	@echo "Short tests complete"

## test-race: Run tests with race detection
.PHONY: test-race
test-race:
	@echo "Running tests with race detection..."
	$(GOTEST) -race -v ./...
	@echo "Race tests complete"

## test-verbose: Run tests with maximum verbosity
.PHONY: test-verbose
test-verbose:
	@echo "Running verbose tests..."
	$(GOTEST) -v -x ./...
	@echo "Verbose tests complete"

## test-examples: Run tests for example programs
.PHONY: test-examples
test-examples:
	@echo "Running example tests..."
	CGO_ENABLED=0 $(GOTEST) -v ./examples/lib
	@echo "Example tests complete"

## test-coverage: Run tests with coverage analysis
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -coverprofile=$(COVERAGE_OUT) ./...
	@echo "Coverage report generated: $(COVERAGE_OUT)"

## coverage-html: Generate HTML coverage report
.PHONY: coverage-html
coverage-html: test-coverage
	@echo "Generating HTML coverage report..."
	$(GOCMD) tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@echo "HTML coverage report: $(COVERAGE_HTML)"
	@echo "Open in browser: file://$(PWD)/$(COVERAGE_HTML)"

## coverage-func: Show function-level coverage
.PHONY: coverage-func
coverage-func: test-coverage
	@echo "Function coverage report:"
	$(GOCMD) tool cover -func=$(COVERAGE_OUT)

## coverage-check: Check if coverage meets threshold
.PHONY: coverage-check
coverage-check: test-coverage
	@echo "Checking coverage threshold ($(COVERAGE_THRESHOLD)%)..."
	@COVERAGE=$$($(GOCMD) tool cover -func=$(COVERAGE_OUT) | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ $$(echo "$$COVERAGE < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
		echo "❌ Coverage $$COVERAGE% is below threshold $(COVERAGE_THRESHOLD)%"; \
		exit 1; \
	else \
		echo "✅ Coverage $$COVERAGE% meets threshold $(COVERAGE_THRESHOLD)%"; \
	fi

## test-integration: Run integration tests only
.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -run "TestComplete|TestMulti|TestConcurrent" ./...
	@echo "Integration tests complete"

## test-unit: Run unit tests only (excluding integration)
.PHONY: test-unit
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v -run "^Test" -skip "TestComplete|TestMulti|TestConcurrent" ./...
	@echo "Unit tests complete"

## test-models: Run model-specific tests
.PHONY: test-models
test-models:
	@echo "Running model-specific tests..."
	$(GOTEST) -v -run "Test.*305EP|Test.*316EP" ./...
	@echo "Model tests complete"

## bench: Run benchmark tests
.PHONY: bench
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...
	@echo "Benchmarks complete"

## cross-build: Build for all supported platforms
.PHONY: cross-build
cross-build: deps
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		output_name=$(BINARY_NAME)-$$GOOS-$$GOARCH; \
		if [ $$GOOS = "windows" ]; then output_name=$$output_name.exe; fi; \
		echo "Building for $$GOOS/$$GOARCH..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH $(GOBUILD) $(BUILD_FLAGS) -o $(DIST_DIR)/$$output_name .; \
		if [ $$? -ne 0 ]; then \
			echo "❌ Failed to build for $$GOOS/$$GOARCH"; \
			exit 1; \
		fi; \
	done
	@echo "✅ Cross-compilation complete. Binaries in $(DIST_DIR)/"

## release: Create release archives for all platforms
.PHONY: release
release: cross-build
	@echo "Creating release archives..."
	@cd $(DIST_DIR) && for binary in *; do \
		if [[ $$binary == *.exe ]]; then \
			zip $$binary.zip $$binary README.md LICENSE; \
		else \
			tar -czf $$binary.tar.gz $$binary README.md LICENSE; \
		fi; \
	done
	@echo "✅ Release archives created in $(DIST_DIR)/"

## docker-build: Build Docker image
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .
	docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest
	@echo "Docker image built: $(BINARY_NAME):$(VERSION)"

## docker-run: Run in Docker container
.PHONY: docker-run
docker-run: docker-build
	@echo "Running in Docker..."
	docker run --rm -it $(BINARY_NAME):$(VERSION) $(ARGS)

## ci: Run all CI checks
.PHONY: ci
ci: clean deps lint test-race coverage-check
	@echo "✅ All CI checks passed"

## dev: Set up development environment
.PHONY: dev
dev: deps build-dev
	@echo "✅ Development environment ready"
	@echo "Run 'make run-dev ARGS=\"--help\"' to test"

## version: Show version information
.PHONY: version
version:
	@echo "ntgrrc version information:"
	@echo "  Version: $(VERSION)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  Go Version: $$($(GOCMD) version)"

## size: Show binary sizes
.PHONY: size
size: build
	@echo "Binary sizes:"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME) | awk '{print "  " $$9 ": " $$5}'
	@if [ -d $(DIST_DIR) ]; then \
		echo "Cross-compiled binaries:"; \
		ls -lh $(DIST_DIR)/ | grep -v "^total" | awk '{print "  " $$9 ": " $$5}'; \
	fi

## debug: Build with debug symbols and run with dlv
.PHONY: debug
debug: deps
	@echo "Building debug version..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -gcflags="all=-N -l" -o $(BUILD_DIR)/$(BINARY_NAME)-debug .
	@echo "Debug build complete. Run with: dlv exec ./$(BUILD_DIR)/$(BINARY_NAME)-debug"

## profile: Build with profiling enabled
.PHONY: profile
profile: deps
	@echo "Building with profiling..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -tags profile -o $(BUILD_DIR)/$(BINARY_NAME)-profile .
	@echo "Profile build complete: $(BUILD_DIR)/$(BINARY_NAME)-profile"

## check: Run all quality checks
.PHONY: check
check: lint test-race coverage-check
	@echo "✅ All quality checks passed"

## watch: Watch for changes and run tests
.PHONY: watch
watch:
	@echo "Watching for changes... (requires 'entr' command)"
	@find . -name "*.go" | entr -r make test

## mock-data: Validate test data files
.PHONY: mock-data
mock-data:
	@echo "Validating test data files..."
	@for dir in test-data/*/; do \
		echo "Checking $$dir"; \
		find "$$dir" -name "*.html" -exec echo "  Found: {}" \;; \
	done
	@echo "Test data validation complete"

# Create necessary directories
$(DIST_DIR) $(COVERAGE_DIR) $(BIN_DIR):
	@mkdir -p $@

# Build directory creation (separate to avoid conflict with build target)
create-build-dir:
	@mkdir -p $(BUILD_DIR)

# Phony targets for directory creation
.PHONY: directories create-build-dir
directories: create-build-dir $(DIST_DIR) $(COVERAGE_DIR) $(BIN_DIR)

# Include local customizations if they exist
-include Makefile.local