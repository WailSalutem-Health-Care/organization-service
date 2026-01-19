# End-to-End (E2E) Testing Guide

## Overview

E2E tests validate the entire application stack from HTTP request to database:

```
HTTP Request → Auth → Handler → Service → Repository → Database → Events
```

### Test Types Comparison

| Type | Scope | Dependencies | Speed |
|------|-------|--------------|-------|
| Unit | Single component | Mocked | Fast (ms) |
| Integration | Repository + DB | Real DB | Medium (1-2s) |
| E2E | Full HTTP stack | Real DB + HTTP Server | Slower (2-5s) |

## Current E2E Test Coverage

**Total: 30 tests (29 passing, 1 skipped)**

### Organization Tests (13 tests - ALL PASSING)

**CRUD Operations:**
- Create organization with schema creation
- Get organization by ID
- List organizations with pagination
- Update organization (full and partial)
- Delete organization (soft delete)
- Get non-existent organization (404)
- List empty results

**Security:**
- Authorization: ORG_ADMIN cannot create/update orgs
- Authentication: Missing token returns 401
- Validation: Required field errors

**Events:**
- RabbitMQ event published on deletion

### Users Tests (6 tests - 5 PASSING, 1 SKIPPED)

**CRUD Operations:**
- Create user (SUPER_ADMIN creates ORG_ADMIN)
- List users with pagination
- Update user
- Delete user (SKIPPED - permission debugging needed)

**Security:**
- Authorization: ORG_ADMIN cannot create SUPER_ADMIN

**Events:**
- RabbitMQ event published on creation

### Patient Tests (6 tests - ALL PASSING)

**CRUD Operations:**
- Create patient
- List patients with pagination
- List active patients only
- Get patient by ID
- Update patient
- Delete patient (soft delete)

**Events:**
- RabbitMQ event published on creation and deletion

### Multi-Tenant Tests (5 tests - ALL PASSING)

**Isolation:**
- ORG_ADMIN can only see their own organization
- SUPER_ADMIN can see all organizations
- Tenant schemas are properly isolated
- Users from org A cannot access org B data
- Patients from org A cannot access org B data

## Test Infrastructure

### Mock Components

**Mock Keycloak** (`testutil/keycloak_mock.go`)
- In-memory user storage
- No real HTTP calls to Keycloak
- Implements full KeycloakAdminInterface
- Thread-safe

**Mock RabbitMQ** (`testutil/rabbitmq_mock.go`)
- In-memory event storage
- No real RabbitMQ connections
- Event verification helpers
- Thread-safe

### Test Utilities

**JWT Tokens** (`testutil/jwt.go`)
```go
token := ts.GenerateSuperAdminToken(t)
token := ts.GenerateOrgAdminToken(t, orgID, schemaName)
```

**HTTP Client** (`testutil/http.go`)
```go
client := ts.NewClient(token)
resp := client.POST(t, "/endpoint", body)
resp := client.GET(t, "/endpoint")
testutil.AssertStatusCode(t, resp, http.StatusOK)
```

**Test Setup** (`e2e/setup.go`)
```go
ts := SetupE2ETest(t)
defer ts.Cleanup(t)
```

## Running E2E Tests

### Prerequisites

**PostgreSQL must be running:**
```bash
docker ps | grep postgres
```

**Test database must be set up:**
```bash
make setup-test-db
```

### Commands

```bash
# Run all E2E tests
make test-e2e

# Run specific test
CGO_ENABLED=0 go test -tags=integration -v ./internal/e2e/... -run TestE2E_CreateOrganization

# Run all tests (unit + integration + E2E)
make test-all
```

### Expected Output

```
Running E2E tests (full HTTP stack)...
Make sure PostgreSQL is running: docker ps | grep postgres
=== RUN   TestE2E_CreateOrganization_FullFlow
    organization_e2e_test.go:118: E2E Test Passed: Created organization ...
--- PASS: TestE2E_CreateOrganization_FullFlow (0.23s)
...
PASS
ok      github.com/WailSalutem-Health-Care/organization-service/internal/e2e    4.518s
```

## Writing New E2E Tests

### Basic Template

```go
//go:build integration

package e2e

import (
    "net/http"
    "testing"
    "github.com/WailSalutem-Health-Care/organization-service/internal/testutil"
)

func TestE2E_YourFeature(t *testing.T) {
    // Setup
    ts := SetupE2ETest(t)
    defer ts.Cleanup(t)
    
    // Generate token
    token := ts.GenerateSuperAdminToken(t)
    client := ts.NewClient(token)
    
    // Make HTTP request
    reqBody := map[string]interface{}{
        "field": "value",
    }
    resp := client.POST(t, "/endpoint", reqBody)
    
    // Verify HTTP response
    testutil.AssertStatusCode(t, resp, http.StatusCreated)
    
    var result struct {
        Success bool   `json:"success"`
        Data    string `json:"data"`
    }
    testutil.DecodeJSON(t, resp, &result)
    
    if !result.Success {
        t.Error("Expected success to be true")
    }
    
    // Verify database state
    var dbValue string
    err := ts.DB.QueryRow("SELECT value FROM table WHERE id = $1", id).Scan(&dbValue)
    if err != nil {
        t.Fatalf("Failed to query database: %v", err)
    }
    
    // Verify event was published
    ts.MockPublisher.AssertEventPublished(t, "your.event")
    
    t.Logf("E2E Test Passed: Feature works correctly")
}
```

### Testing Different Roles

```go
// Test with SUPER_ADMIN (should succeed)
superToken := ts.GenerateSuperAdminToken(t)
superClient := ts.NewClient(superToken)
resp := superClient.POST(t, "/endpoint", body)
testutil.AssertStatusCode(t, resp, http.StatusCreated)

// Test with ORG_ADMIN (should fail)
orgToken := ts.GenerateOrgAdminToken(t, orgID, schemaName)
orgClient := ts.NewClient(orgToken)
resp = orgClient.POST(t, "/endpoint", body)
testutil.AssertStatusCode(t, resp, http.StatusForbidden)
```

### Verifying Events

```go
// Check event was published
ts.MockPublisher.AssertEventPublished(t, "user.created")

// Check event count
ts.MockPublisher.AssertEventCount(t, "user.created", 1)

// Get event data
events := ts.MockPublisher.GetEventsByKey("user.created")
lastEvent := ts.MockPublisher.GetLastEventByKey("user.created")
```

## What to Test in E2E

### Do Test
- Complete user workflows (create → read → update → delete)
- Authentication and authorization
- HTTP layer integration (status codes, headers, JSON)
- Database persistence
- Event publishing
- Multi-tenant isolation

### Don't Test
- Business logic details (use unit tests)
- Database query optimization (use integration tests)
- Individual function edge cases (use unit tests)

## Test Structure

```
internal/e2e/
├── setup.go                    # Test environment setup
├── organization_e2e_test.go    # Organization tests (13 tests)
├── users_e2e_test.go          # User tests (6 tests)
├── patient_e2e_test.go        # Patient tests (6 tests)
└── multitenant_e2e_test.go    # Multi-tenant tests (5 tests)

internal/testutil/
├── jwt.go                      # JWT token generation
├── http.go                     # HTTP test client
├── auth.go                     # Test verifier
├── keycloak_mock.go           # Mock Keycloak client
├── rabbitmq_mock.go           # Mock RabbitMQ publisher
└── database.go                 # Database helpers
```

## Troubleshooting

### Database Connection Issues
```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Verify test database exists
psql -h localhost -U localadmin -d wailsalutem_test -c "\dt"

# Re-setup if needed
make setup-test-db
```

### Test Failures
```bash
# Run with verbose output
CGO_ENABLED=0 go test -tags=integration -v ./internal/e2e/...

# Run specific test
CGO_ENABLED=0 go test -tags=integration -v ./internal/e2e/... -run TestE2E_YourTest
```

### Clean Test Data
```bash
# Tests automatically clean up
# Manual cleanup if needed:
psql -h localhost -U localadmin -d wailsalutem_test \
  -c "TRUNCATE TABLE wailsalutem.organizations CASCADE"
```

## Key Features

### No External Dependencies
- Mock Keycloak: No real Keycloak server needed
- Mock RabbitMQ: No real message broker needed
- Real Database: PostgreSQL for accurate testing
- Real HTTP: Full HTTP stack validation

### Test Isolation
- Each test gets fresh database
- Automatic cleanup after tests
- No interdependencies between tests

### Realistic Testing
- Real HTTP requests and responses
- Real JWT token validation
- Real database operations
- Real routing and middleware

## Related Documentation

- [TESTING.md](./TESTING.md) - Unit and integration testing guide
- [Makefile](./Makefile) - Test commands reference

## Summary

- **30 E2E tests** covering all major functionality
- **29 passing** (96.7% pass rate)
- **1 skipped** (user delete - permission debugging needed)
- Tests organization, users, patients, and multi-tenant isolation
- Mock Keycloak and RabbitMQ for no external dependencies
- Production-ready test suite
