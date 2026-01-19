.PHONY: test test-auth test-coverage test-verbose clean help test-integration test-e2e test-all setup-test-db

# Default target
help:
	@echo "Available targets:"
	@echo "  make test              - Run unit tests only"
	@echo "  make test-auth         - Run authentication tests only"
	@echo "  make test-integration  - Run integration tests (requires PostgreSQL)"
	@echo "  make test-e2e          - Run E2E tests (requires PostgreSQL)"
	@echo "  make test-all          - Run all tests (unit + integration + E2E)"
	@echo "  make test-coverage     - Run tests with coverage report"
	@echo "  make test-verbose      - Run tests with verbose output"
	@echo "  make setup-test-db     - Setup test database (run once)"
	@echo "  make clean             - Clean test cache and coverage files"

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

# Setup test database (run once)
setup-test-db:
	@echo "Setting up test database..."
	@./scripts/setup-test-db.sh

# Run integration tests only (repository layer)
test-integration:
	@echo "Running integration tests (repository layer)..."
	@echo "Make sure PostgreSQL is running: docker ps | grep postgres"
	@CGO_ENABLED=0 go test -tags=integration -v ./internal/organization/... ./internal/users/... ./internal/patient/...

# Run E2E tests only (full stack)
test-e2e:
	@echo "Running E2E tests (full HTTP stack)..."
	@echo "Make sure PostgreSQL is running: docker ps | grep postgres"
	@CGO_ENABLED=0 go test -tags=integration -v ./internal/e2e/...

# Run all tests (unit + integration + E2E)
test-all:
	@echo "Running all tests (unit + integration + E2E)..."
	@make test
	@echo ""
	@make test-integration
	@echo ""
	@make test-e2e
	@echo ""
	@echo " All tests completed!"
