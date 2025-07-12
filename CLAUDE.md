# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This repository contains 18 standalone Go programs demonstrating various concurrency patterns. Each example is self-contained in its own directory with a `main.go` file (except `16-context` which has additional `client.go` and `server.go` files).

## Common Commands

### Running Examples
```bash
# Using Make (recommended)
make run-example EXAMPLE=1-boring
make run-example EXAMPLE=16-context

# Quick shortcuts for common examples
make run-boring
make run-worker-pool
make run-context
make run-fanin

# Run all examples (with timeout protection)
make run-all

# Traditional Go commands
cd 1-boring && go run main.go
cd 16-context && go run main.go  # This example has multiple files
go run ./1-boring/main.go
go run ./16-context/*.go
```

### Testing and Quality
```bash
# Run comprehensive tests with race detection
make test-race

# Run all tests
make test

# Run benchmarks
make bench

# Code quality checks
make check  # Runs fmt, vet, and test-race
make fmt    # Format code
make vet    # Run go vet
make lint   # Run golangci-lint (if installed)

# Coverage analysis
make coverage
```

### Building
```bash
# Build all examples
make build-all

# Traditional builds
cd 1-boring && go build
go build ./1-boring

# Build all examples manually
for dir in */; do (cd "$dir" && go build); done
```

### Performance Analysis
```bash
# Run benchmarks
make bench

# Compare different patterns
make compare-patterns

# CPU profiling
make profile-cpu

# Memory profiling
make profile-mem

# Trace analysis
make profile-trace
```

### Development Tools
```bash
# Setup development environment
make dev-setup

# Clean build artifacts
make clean

# Full test suite
make full-test

# Stress testing
make stress
make stress-race
```

## Architecture and Structure

### Pattern Categories
- **Basic Patterns (1-8)**: Core concurrency concepts including goroutines, channels, generators, fan-in, timeouts, and quit signals
- **Google Search Examples (9-12)**: Progressive evolution of a concurrent search implementation showing realistic patterns
- **Advanced Patterns (13-18)**: Complex patterns including ping-pong, subscriptions, bounded parallelism, context usage, ring buffers, and worker pools

### Key Architectural Concepts
- Each example demonstrates a specific concurrency pattern in isolation
- Examples progress from simple goroutine usage to complex coordination patterns
- The `16-context` example is the only multi-file example, showing client-server interaction with context cancellation
- All examples use standard library only (no external dependencies)

### Running Environment
- Go version: 1.24+ (currently using 1.24.5)
- Module: `github.com/lotusirous/gochan`
- Platform: All examples should run on any Go-supported platform

## Development Notes

When working with these examples:
- Each directory represents an independent program
- Examples are designed to run indefinitely or for a short duration to demonstrate patterns
- Some examples (like `1-boring`) have commented code showing alternative behaviors
- The `16-context` example demonstrates real-world HTTP client/server patterns with proper context handling