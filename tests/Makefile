# Makefile for Personal Finance Tracker Tests

.PHONY: test test-verbose test-coverage test-clean test-models test-controllers test-auth help

# Default target
help:
	@echo "Available test commands:"
	@echo "  test          - Run all tests"
	@echo "  test-verbose  - Run tests with verbose output"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  test-clean    - Clean test cache and coverage files"
	@echo "  test-models   - Run only model tests"
	@echo "  test-controllers - Run only controller tests"
	@echo "  test-auth     - Run only auth-related tests"
	@echo "  help          - Show this help message"

# Set environment variables for testing
export JWT_SECRET=test_secret_key_for_testing_only
export GIN_MODE=test

# Run all tests
test:
	@echo "🧪 Running all tests..."
	@go test ./...

# Run tests with verbose output
test-verbose:
	@echo "🧪 Running tests with verbose output..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "📄 Coverage report saved to coverage.html"
	@echo "📈 Coverage summary:"
	@go tool cover -func=coverage.out

# Clean test cache and coverage files
test-clean:
	@echo "🧹 Cleaning test cache and coverage files..."
	@go clean -testcache
	@rm -f coverage.out coverage.html

# Run only model tests
test-models:
	@echo "🧪 Running model tests..."
	@go test -v ./backend/models/...

# Run only controller tests
test-controllers:
	@echo "🧪 Running controller tests..."
	@go test -v ./backend/controllers/...

# Run only auth-related tests
test-auth:
	@echo "🧪 Running auth tests..."
	@go test -v ./backend/controllers/... -run "Auth"

# Run with database skip
test-no-db:
	@echo "🧪 Running tests without database..."
	@TEST_SKIP_DB=true go test -v ./...

# Watch mode (requires entr)
test-watch:
	@echo "👀 Running tests in watch mode (Ctrl+C to stop)..."
	@find . -name "*.go" | entr -c make test