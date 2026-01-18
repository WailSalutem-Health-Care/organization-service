# User Integration Tests

This document describes the integration tests for the users module.

## Overview

The user integration tests verify the repository layer functionality with a real PostgreSQL database. All tests use the multi-tenant schema architecture where each organization has its own schema.

## Test Coverage

### Repository Methods Tested

1. **Create** - Creating users in tenant schemas
2. **GetByID** - Retrieving users by database ID
3. **GetByKeycloakID** - Retrieving users by Keycloak user ID
4. **List** - Listing all users in a schema
5. **ListWithPagination** - Paginated user listing with search
6. **ListActiveUsersByRoleWithPagination** - Role-based filtering with pagination and search
7. **Update** - Updating user information
8. **Delete** - Soft deleting users
9. **GetSchemaNameByOrgID** - Schema name lookup
10. **ValidateOrgSchema** - Schema validation

### Test Cases (20 tests)

#### Core CRUD Operations
- `TestRepositoryCreate_Integration` - Basic user creation
- `TestRepositoryGetByID_Integration` - Retrieve user by ID
- `TestRepositoryGetByKeycloakID_Integration` - Retrieve by Keycloak ID
- `TestRepositoryUpdate_Integration` - Update user details
- `TestRepositoryDelete_Integration` - Soft delete user

#### List Operations
- `TestRepositoryList_Integration` - List all users
- `TestRepositoryListWithPagination_Integration` - Paginated listing
- `TestRepositoryListActiveUsersByRole_Integration` - Filter by role

#### Search Functionality
- `TestRepositoryListWithPagination_Search_Integration` - Search by name/email
- `TestRepositoryListActiveUsersByRole_Search_Integration` - Role-based search

#### Multi-Tenancy
- `TestRepositoryGetByID_CrossTenantIsolation_Integration` - Verify tenant isolation
- Users in one organization cannot access users from another

#### Employee ID Generation
- `TestRepositoryEmployeeIDGeneration_Integration` - Sequential ID generation (EMP-0001, EMP-0002, etc.)

#### Soft Delete Behavior
- `TestRepositorySoftDelete_ExcludesFromActiveList_Integration` - Deleted users excluded from active lists
- `TestRepositoryDelete_AlreadyDeleted_Integration` - Cannot delete twice

#### Schema Operations
- `TestRepositoryGetSchemaNameByOrgID_Integration` - Get schema by org ID
- `TestRepositoryValidateOrgSchema_Integration` - Validate schema exists

#### Error Handling
- `TestRepositoryGetByID_NotFound_Integration` - User not found error
- `TestRepositoryGetSchemaNameByOrgID_NotFound_Integration` - Org not found error
- `TestRepositoryUpdate_NotFound_Integration` - Update non-existent user
- `TestRepositoryDelete_NotFound_Integration` - Delete non-existent user

## Running the Tests

### Prerequisites

1. PostgreSQL must be running (via Docker):
   ```bash
   cd ../WailSalutem-infra && docker-compose up -d postgres
   ```

2. Test database must be set up:
   ```bash
   make setup-test-db
   ```

### Run Integration Tests

```bash
# Run all integration tests
make test-integration

# Run only user integration tests
CGO_ENABLED=0 go test -tags=integration -v ./internal/users/... -run Integration

# Run specific test
CGO_ENABLED=0 go test -tags=integration -v ./internal/users/... -run TestRepositoryCreate_Integration
```

## Test Database

- **Database**: `wailsalutem_test`
- **User**: `localadmin`
- **Connection**: `localhost:5432`

Each test:
1. Creates a test organization with a unique schema (e.g., `org_test_hospital_a`)
2. Runs the test operations
3. Cleans up all test data

## Key Features Tested

### Multi-Tenant Architecture
- Each organization has its own PostgreSQL schema
- Users are completely isolated by schema
- Cross-tenant access is prevented

### Soft Delete
- Users are not physically deleted
- `deleted_at` timestamp is set
- Deleted users are excluded from active user lists
- Deleted users cannot be deleted again

### Employee ID Generation
- Sequential IDs per organization
- Format: `EMP-0001`, `EMP-0002`, etc.
- Starts at 0001 for each organization

### Search Functionality
- Case-insensitive search (ILIKE)
- Searches across: first name, last name, email
- Works with pagination
- Works with role filtering

### Pagination
- Configurable limit and offset
- Returns total count
- Maintains sort order (newest first)

## Bug Fixes

During test development, we discovered and fixed a bug in `ListWithPagination`:
- **Issue**: Count query used wrong parameter placeholder ($3 instead of $1)
- **Fix**: Separate WHERE clauses for count and data queries with correct parameter numbering
- **File**: `internal/users/repository.go` lines 358-384

## Test Results

```
=== RUN   TestRepositoryCreate_Integration
--- PASS: TestRepositoryCreate_Integration (0.06s)
=== RUN   TestRepositoryGetByID_Integration
--- PASS: TestRepositoryGetByID_Integration (0.04s)
=== RUN   TestRepositoryGetByID_CrossTenantIsolation_Integration
--- PASS: TestRepositoryGetByID_CrossTenantIsolation_Integration (0.06s)
=== RUN   TestRepositoryList_Integration
--- PASS: TestRepositoryList_Integration (0.04s)
=== RUN   TestRepositoryListWithPagination_Integration
--- PASS: TestRepositoryListWithPagination_Integration (0.04s)
=== RUN   TestRepositoryListActiveUsersByRole_Integration
--- PASS: TestRepositoryListActiveUsersByRole_Integration (0.04s)
=== RUN   TestRepositoryUpdate_Integration
--- PASS: TestRepositoryUpdate_Integration (0.04s)
=== RUN   TestRepositoryDelete_Integration
--- PASS: TestRepositoryDelete_Integration (0.04s)
=== RUN   TestRepositoryGetByKeycloakID_Integration
--- PASS: TestRepositoryGetByKeycloakID_Integration (0.04s)
=== RUN   TestRepositoryEmployeeIDGeneration_Integration
--- PASS: TestRepositoryEmployeeIDGeneration_Integration (0.04s)
=== RUN   TestRepositoryGetSchemaNameByOrgID_Integration
--- PASS: TestRepositoryGetSchemaNameByOrgID_Integration (0.04s)
=== RUN   TestRepositoryGetSchemaNameByOrgID_NotFound_Integration
--- PASS: TestRepositoryGetSchemaNameByOrgID_NotFound_Integration (0.01s)
=== RUN   TestRepositoryValidateOrgSchema_Integration
--- PASS: TestRepositoryValidateOrgSchema_Integration (0.03s)
=== RUN   TestRepositoryListWithPagination_Search_Integration
--- PASS: TestRepositoryListWithPagination_Search_Integration (0.04s)
=== RUN   TestRepositoryListActiveUsersByRole_Search_Integration
--- PASS: TestRepositoryListActiveUsersByRole_Search_Integration (0.04s)
=== RUN   TestRepositorySoftDelete_ExcludesFromActiveList_Integration
--- PASS: TestRepositorySoftDelete_ExcludesFromActiveList_Integration (0.04s)
=== RUN   TestRepositoryGetByID_NotFound_Integration
--- PASS: TestRepositoryGetByID_NotFound_Integration (0.04s)
=== RUN   TestRepositoryUpdate_NotFound_Integration
--- PASS: TestRepositoryUpdate_NotFound_Integration (0.04s)
=== RUN   TestRepositoryDelete_NotFound_Integration
--- PASS: TestRepositoryDelete_NotFound_Integration (0.04s)
=== RUN   TestRepositoryDelete_AlreadyDeleted_Integration
--- PASS: TestRepositoryDelete_AlreadyDeleted_Integration (0.03s)

PASS - 20/20 tests passing
```

## Next Steps

Consider adding:
1. Service layer integration tests (with mocked Keycloak)
2. Handler/API integration tests (with test HTTP server)
3. Performance tests for large datasets
4. Concurrent access tests
5. Transaction rollback tests
