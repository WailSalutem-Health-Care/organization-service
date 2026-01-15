.PHONY: test test-auth test-coverage test-verbose clean help

# Default target
help:
	@echo "Available targets:"
	@echo "  make test          - Run all tests"
	@echo "  make test-auth     - Run authentication tests only"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make test-verbose  - Run tests with verbose output"
	@echo "  make clean         - Clean test cache and coverage files"

# Run all tests
test:
	@echo "Running all tests..."
	CGO_ENABLED=0 go test ./...

# Run authentication tests only
test-auth:
	@echo "Running authentication tests..."
	CGO_ENABLED=0 go test -v ./internal/auth/...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	CGO_ENABLED=0 go test -coverprofile=coverage.out ./...
	@echo "\n=== Coverage Summary ==="
	@go tool cover -func=coverage.out | grep total
	@echo "\nRun 'make coverage-html' to view detailed coverage in browser"

# View coverage in browser
coverage-html:
	@echo "Opening coverage report in browser..."
	go tool cover -html=coverage.out

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	CGO_ENABLED=0 go test -v ./...

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	go test -race ./...

# Clean test cache and coverage files
clean:
	@echo "Cleaning test cache and coverage files..."
	go clean -testcache
	rm -f coverage.out

# Run a specific test
# Usage: make test-specific TEST=TestMiddleware_ValidToken
test-specific:
	@echo "Running test: $(TEST)"
	CGO_ENABLED=0 go test -v ./internal/auth/... -run $(TEST)
