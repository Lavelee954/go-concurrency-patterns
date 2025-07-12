# Go Concurrency Patterns Makefile

.PHONY: help test test-race bench clean run-all run-example lint fmt vet

# Default target
help:
	@echo "Available targets:"
	@echo "  test       - Run all tests"
	@echo "  test-race  - Run tests with race detection"
	@echo "  bench      - Run benchmarks"
	@echo "  run-all    - Run all examples"
	@echo "  run-example - Run specific example (use EXAMPLE=folder-name)"
	@echo "  lint       - Run linter (requires golangci-lint)"
	@echo "  fmt        - Format code"
	@echo "  vet        - Run go vet"
	@echo "  clean      - Clean build artifacts"

# Test targets
test:
	@echo "Running tests..."
	go test -v ./...

test-race:
	@echo "Running tests with race detection..."
	go test -race -v ./...

test-short:
	@echo "Running short tests..."
	go test -short -v ./...

# Benchmark targets
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

bench-cpu:
	@echo "Running CPU benchmarks..."
	go test -bench=. -benchmem -cpuprofile=cpu.prof ./...

bench-mem:
	@echo "Running memory benchmarks..."
	go test -bench=. -benchmem -memprofile=mem.prof ./...

# Code quality targets
fmt:
	@echo "Formatting code..."
	go fmt ./...

vet:
	@echo "Running go vet..."
	go vet ./...

lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install it from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run

# Run targets
run-all:
	@echo "Running all examples..."
	@for dir in */; do \
		if [ -f "$$dir/main.go" ]; then \
			echo "Running $$dir..."; \
			timeout 5s go run "./$$dir" || echo "$$dir finished or timed out"; \
			echo ""; \
		fi \
	done

run-example:
	@if [ -z "$(EXAMPLE)" ]; then \
		echo "Usage: make run-example EXAMPLE=folder-name"; \
		echo "Available examples:"; \
		ls -d */ | grep -E '^[0-9]' | head -20; \
	else \
		if [ -f "$(EXAMPLE)/main.go" ]; then \
			echo "Running $(EXAMPLE)..."; \
			cd "$(EXAMPLE)" && timeout 10s go run . || echo "Example finished or timed out"; \
		else \
			echo "Example $(EXAMPLE) not found or no main.go file"; \
		fi \
	fi

# Specific example shortcuts
run-boring:
	@echo "Running boring example..."
	cd 1-boring && timeout 3s go run . || echo "Boring example finished"

run-chan:
	@echo "Running channel example..."
	cd 2-chan && go run .

run-generator:
	@echo "Running generator example..."
	cd 3-generator && go run .

run-fanin:
	@echo "Running fan-in example..."
	cd 4-fanin && go run .

run-timeout:
	@echo "Running timeout example..."
	cd 6-select-timeout && timeout 10s go run . || echo "Timeout example finished"

run-worker-pool:
	@echo "Running worker pool example..."
	cd 18-worker-pool && go run .

run-context:
	@echo "Running context example..."
	cd 16-context && timeout 3s go run . || echo "Context example finished"

# Performance analysis
profile-cpu:
	@echo "Running CPU profile..."
	go test -bench=. -cpuprofile=cpu.prof ./...
	@echo "View profile with: go tool pprof cpu.prof"

profile-mem:
	@echo "Running memory profile..."
	go test -bench=. -memprofile=mem.prof ./...
	@echo "View profile with: go tool pprof mem.prof"

profile-trace:
	@echo "Running trace..."
	go test -bench=BenchmarkWorkerPool -trace=trace.out ./...
	@echo "View trace with: go tool trace trace.out"

# Build targets
build-all:
	@echo "Building all examples..."
	@for dir in */; do \
		if [ -f "$$dir/main.go" ]; then \
			echo "Building $$dir..."; \
			cd "$$dir" && go build -o "$${dir%/}" . && cd ..; \
		fi \
	done

# Clean targets
clean:
	@echo "Cleaning build artifacts..."
	find . -name "*.prof" -delete
	find . -name "*.out" -delete
	find . -name "*.test" -delete
	@for dir in */; do \
		if [ -f "$$dir/$${dir%/}" ]; then \
			rm "$$dir/$${dir%/}"; \
		fi \
	done

# Development targets
dev-setup:
	@echo "Setting up development environment..."
	go mod tidy
	@echo "Installing development tools..."
	@which golangci-lint > /dev/null || echo "Consider installing golangci-lint: https://golangci-lint.run/usage/install/"
	@echo "Development setup complete!"

check: fmt vet test-race
	@echo "All checks passed!"

# Documentation targets
docs:
	@echo "Available documentation:"
	@echo "  README.md - Project overview"
	@echo "  PATTERNS.md - Detailed pattern documentation"
	@echo "  CLAUDE.md - Claude Code integration guide"

# Coverage targets
coverage:
	@echo "Running test coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

coverage-func:
	@echo "Running test coverage (function view)..."
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# Stress testing
stress:
	@echo "Running stress tests..."
	go test -count=100 -short ./...

stress-race:
	@echo "Running stress tests with race detection..."
	go test -race -count=10 ./...

# Performance comparison
compare-patterns:
	@echo "Comparing pattern performance..."
	go test -bench=BenchmarkChannelTypes -benchmem ./...
	go test -bench=BenchmarkFanInPattern -benchmem ./...
	go test -bench=BenchmarkWorkerPool -benchmem ./...

# Example-specific tests
test-examples:
	@echo "Testing individual examples..."
	go test -run=TestBoringPattern -v ./...
	go test -run=TestGeneratorPattern -v ./...
	go test -run=TestFanInPattern -v ./...
	go test -run=TestWorkerPoolPattern -v ./...

# All-in-one targets
full-test: fmt vet test-race bench coverage
	@echo "Full test suite completed!"

ci: fmt vet test-race
	@echo "CI checks completed!"

all: clean fmt vet test-race bench
	@echo "All targets completed!"