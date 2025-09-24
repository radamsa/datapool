# Makefile for datapool Go library

# Configuration variables
GO = go
GOFMT = gofmt
GOLINT = golint

# Default target
.PHONY: all
all: test lint

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test

# Run tests with coverage
.PHONY: coverage
coverage:
	@echo "Running tests with coverage..."
	$(GO) test -coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOFMT) -w -s *.go

# Lint code
.PHONY: lint
lint:
	@echo "Running go vet..."
	$(GO) vet
	@echo "Running golint..."
	@if command -v $(GOLINT) > /dev/null; then \
		$(GOLINT) *.go; \
	else \
		echo "golint not installed. Run: go install golang.org/x/lint/golint@latest"; \
	fi

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f coverage.out coverage.html

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all       - Test, and lint (default)"
	@echo "  test      - Run tests"
	@echo "  coverage  - Run tests with coverage"
	@echo "  fmt       - Format code"
	@echo "  lint      - Lint code"
	@echo "  clean     - Clean build artifacts"
	@echo "  help      - Show this help message"
