# Testing Guide

This document describes the testing strategy and how to run tests for the organization-service.

## Quick Summary

**Total Test Cases:** 140 (131 unit + 9 integration)
**All Tests:** PASSING
**Test Files:** 14
**Commits:** 22

**Coverage by Layer:**
- Authentication: 100% (critical paths)
- Authorization: 100% (critical paths)
- Service Layer: 40-100%
- Handler Layer: 47-100% (all modules completed)
- Repository Layer: Organization tested with real database

## Test Suites

### Phase 1: Authentication & Authorization Tests (COMPLETED)

**Location:** `internal/auth/*_test.go`

**Coverage:** 22.6% overall, **100% coverage for critical components:**
- Middleware (100%)
- JWT Verification (91.4%)
- Permission System (100%)
- Permissions Loading (100%)

**Test Files:**
1. `middleware_test.go` - 16 test cases
2. `jwt_verify_test.go` - 8 test cases
3. `permissions_test.go` - 6 test cases

**Total:** 30 test cases, all passing

### Phase 2: Organization Service Tests (COMPLETED)

**Location:** `internal/organization/*_test.go`

**Coverage:** 32.3% overall
- **Service layer:** 85-100% coverage
- **Handler layer:** 63-100% coverage

**Test Files:**
1. `service_test.go` - 17 test cases (service layer)
2. `handler_test.go` - 14 test cases (HTTP layer)
3. `repository_interface.go` - interface for testability
4. `service_interface.go` - interface for handler testing

**Total:** 31 test cases, all passing

### Phase 3: Users Service Tests (COMPLETED)

**Location:** `internal/users/*_test.go`

**Coverage:** 24.7% overall
- **Service layer:** 40-75% coverage

**Test Files:**
1. `service_test.go` - 20 test cases
2. `repository_interface.go` - interface for testability
3. `keycloak_interface.go` - interface for Keycloak mocking

**Total:** 20 test cases, all passing

### Phase 3: Patient Service Tests (COMPLETED)

**Location:** `internal/patient/*_test.go`

**Coverage:** 28.4% overall
- **Service layer:** 73-100% coverage
- **Handler layer:** 47-78% coverage

**Test Files:**
1. `service_test.go` - 13 test cases (service layer)
2. `handler_test.go` - 15 test cases (HTTP layer)
3. `repository_interface.go` - interface for testability
4. `keycloak_interface.go` - interface for Keycloak mocking
5. `service_interface.go` - interface for handler testing
6. `schema_lookup_adapter.go` - adapter for DB schema lookups

**Total:** 28 test cases, all passing

### Phase 4: Users Handler Tests (COMPLETED)

**Location:** `internal/users/*_test.go`

**Coverage:** 38.0% overall (handler layer)
- **Handler layer:** 47-90% coverage

**Test Files:**
1. `handler_test.go` - 22 test cases (HTTP layer)
2. `service_interface.go` - interface for handler testing

**Total:** 22 test cases, all passing

## Running Tests

### Run All Tests
```bash
CGO_ENABLED=0 go test -v ./...
```

### Run Specific Test Suites
```bash
# Authentication tests
CGO_ENABLED=0 go test -v ./internal/auth/...

# Organization service tests
CGO_ENABLED=0 go test -v ./internal/organization/...

# Middleware tests only
CGO_ENABLED=0 go test -v ./internal/auth/... -run TestMiddleware

# JWT verification tests only
CGO_ENABLED=0 go test -v ./internal/auth/... -run TestVerifier

# Organization service tests only
CGO_ENABLED=0 go test -v ./internal/organization/... -run TestCreateOrganization
```

### Generate Coverage Report
```bash
CGO_ENABLED=0 go test -coverprofile=coverage.out ./internal/auth/...
go tool cover -func=coverage.out
```

### View Coverage in Browser
```bash
go tool cover -html=coverage.out
```

## Test Coverage by Component

### Authentication & Middleware
- ✅ Valid JWT token authentication
- ✅ Missing Authorization header (401)
- ✅ Malformed Authorization header (401)
- ✅ Invalid token signature (401)
- ✅ Expired token (401)
- ✅ Token with missing required claims (401)
- ✅ Principal extraction from context
- ✅ Permission enforcement middleware

### JWT Verification
- ✅ Successful token parsing with all claims
- ✅ Empty token handling
- ✅ Invalid issuer rejection
- ✅ Expired token detection
- ✅ Missing subject claim
- ✅ Missing key ID in header
- ✅ Tokens without roles
- ✅ Tokens without organization claims

### Permission System
- ✅ Single role with permission (allowed)
- ✅ Single role without permission (denied)
- ✅ Multiple roles with permission check
- ✅ Unknown roles handling
- ✅ Permissions loading from YAML
- ✅ Invalid YAML handling
- ✅ Empty permissions file
- ✅ Real permissions.yml validation

## What's Tested

### Security Critical ✅
1. **Authentication** - All JWT validation paths covered
2. **Authorization** - Permission checking logic fully tested
3. **Token Validation** - Signature, expiration, issuer verification
4. **Role-Based Access Control** - Multi-role permission resolution

### Edge Cases ✅
- Missing headers
- Malformed tokens
- Expired credentials
- Invalid configurations
- Empty role assignments
- Missing organization claims

## Next Steps

The following test suites need to be implemented:

### Phase 5: Repository Layer Tests (DATABASE REQUIRED)
- [ ] `internal/organization/repository_test.go`
- [ ] `internal/users/repository_test.go`
- [ ] `internal/patient/repository_test.go`

These tests require a running PostgreSQL database with multi-tenant schema setup.

### Phase 6: Integration Tests (INFRASTRUCTURE REQUIRED)
- [ ] Database integration tests (requires PostgreSQL)
- [ ] Keycloak integration tests (requires Keycloak instance)
- [ ] RabbitMQ integration tests (requires RabbitMQ)
- [ ] End-to-end API tests (requires full stack)

These tests require the full infrastructure stack running.

## Testing Standards

### Code Coverage Goals
- **Minimum:** 60% overall coverage
- **Target:** 75-80% overall coverage
- **Critical paths:** 90%+ coverage (auth, multi-tenancy, authorization)

### Current Status
- **Authentication layer:** 100% coverage (critical)
- **Authorization logic:** 100% coverage (critical)
- **Organization service:** 85-100% coverage (completed)
- **Organization handler:** 63-100% coverage (completed)
- **Users service:** 40-75% coverage (completed)
- **Users handler:** 47-90% coverage (completed)
- **Patient service:** 73-100% coverage (completed)
- **Patient handler:** 47-78% coverage (completed)
- **Repository layer:** 0% coverage (requires database)

### Phase 4 Complete
All handler tests for organization, users, and patient modules are now complete.
Total of 131 unit test cases covering authentication, authorization, service logic, and HTTP handlers.

### Phase 5: Integration Tests (IN PROGRESS)

**Location:** `internal/organization/repository_integration_test.go`

**Coverage:** Organization repository with real PostgreSQL

**Test Files:**
1. `repository_integration_test.go` - 9 integration tests
2. `testutil/database.go` - Test helpers
3. `scripts/setup-test-db.sh` - Test database setup

**Total:** 9 integration tests, all passing

**What's tested:**
- CreateOrganization with real database
- Multi-tenant schema creation
- GetOrganization by ID
- ListOrganizations with pagination
- UpdateOrganization
- DeleteOrganization (soft delete)
- Database constraints and transactions
- Schema isolation

**Running integration tests:**
```bash
# Setup test database (one-time)
make setup-test-db

# Run integration tests
make test-integration

# Run all tests (unit + integration)
make test-all
```

### Test Naming Convention
- Test files: `*_test.go`
- Test functions: `Test<ComponentName>_<Scenario>`
- Example: `TestVerifier_ParseAndVerifyToken_ExpiredToken`

### Test Structure
```go
func TestComponent_Scenario(t *testing.T) {
    // Setup
    // Execute
    // Verify
}
```

## Notes

- **CGO_ENABLED=0** is required due to macOS build issues
- Mock JWKS implementation avoids real Keycloak dependency
- All auth tests run in < 2 seconds
- Tests use real RSA key generation for realistic JWT signing
- Permission tests validate against actual `permissions.yml` file

## CI/CD Integration

Add to your CI pipeline:
```yaml
- name: Run Tests
  run: CGO_ENABLED=0 go test -v -coverprofile=coverage.out ./...
  
- name: Check Coverage
  run: |
    coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    if [ $(echo "$coverage < 60" | bc) -eq 1 ]; then
      echo "Coverage $coverage% is below 60% threshold"
      exit 1
    fi
```

## Troubleshooting

### Issue: `dyld missing LC_UUID` error
**Solution:** Run tests with `CGO_ENABLED=0`

### Issue: Tests fail to compile
**Solution:** Ensure you're in the project root and run `go mod tidy`

### Issue: Coverage not showing
**Solution:** Run coverage command with output file, then view with `go tool cover`
