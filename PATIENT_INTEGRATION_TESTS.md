# Patient Integration Tests

This document describes the integration tests for the patient module.

## Overview

The patient integration tests verify the repository layer functionality with a real PostgreSQL database. All tests use the multi-tenant schema architecture where each organization has its own schema.

## Test Coverage

### Repository Methods Tested

1. **CreatePatient** - Creating patients in tenant schemas
2. **GetPatient** - Retrieving patients by database ID
3. **ListPatients** - Listing all patients (excluding soft-deleted)
4. **ListPatientsWithPagination** - Paginated patient listing with search
5. **ListActivePatientsWithPagination** - Active patients only (is_active=true)
6. **UpdatePatient** - Updating patient information
7. **DeletePatient** - Soft deleting patients

### Test Cases (18 tests)

#### Core CRUD Operations
- `TestRepositoryCreatePatient_Integration` - Basic patient creation
- `TestRepositoryGetPatient_Integration` - Retrieve patient by ID
- `TestRepositoryUpdate_Integration` - Update patient details
- `TestRepositoryDelete_Integration` - Soft delete patient

#### List Operations
- `TestRepositoryListPatients_Integration` - List all patients
- `TestRepositoryListPatientsWithPagination_Integration` - Paginated listing
- `TestRepositoryListActivePatientsWithPagination_Integration` - Filter active patients

#### Search Functionality
- `TestRepositoryListPatientsWithPagination_Search_Integration` - Search by name/email

#### Multi-Tenancy
- `TestRepositoryGetPatient_CrossTenantIsolation_Integration` - Verify tenant isolation
- Patients in one organization cannot access patients from another

#### Patient ID Generation
- `TestRepositoryPatientIDGeneration_Integration` - Sequential ID generation (PT-0001, PT-0002, etc.)

#### Soft Delete Behavior
- `TestRepositorySoftDelete_ExcludesFromList_Integration` - Deleted patients excluded from lists
- `TestRepositoryDelete_AlreadyDeleted_Integration` - Cannot delete twice

#### Patient-Specific Features
- `TestRepositoryPatient_CareplanFields_Integration` - Care plan type and frequency
- `TestRepositoryPatient_EmergencyContact_Integration` - Emergency contact information
- `TestRepositoryPatient_IsActiveFlag_Integration` - Active/inactive status management

#### Error Handling
- `TestRepositoryGetPatient_NotFound_Integration` - Patient not found error
- `TestRepositoryUpdate_NotFound_Integration` - Update non-existent patient
- `TestRepositoryDelete_NotFound_Integration` - Delete non-existent patient

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

# Run only patient integration tests
CGO_ENABLED=0 go test -tags=integration -v ./internal/patient/... -run Integration

# Run specific test
CGO_ENABLED=0 go test -tags=integration -v ./internal/patient/... -run TestRepositoryCreatePatient_Integration
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
- Patients are completely isolated by schema
- Cross-tenant access is prevented

### Soft Delete
- Patients are not physically deleted
- `deleted_at` timestamp is set
- Deleted patients are excluded from all patient lists
- Deleted patients cannot be retrieved or deleted again

### Patient ID Generation
- Sequential IDs per organization
- Format: `PT-0001`, `PT-0002`, etc.
- Starts at 0001 for each organization

### Active/Inactive Status
- Patients can be marked as active or inactive via `is_active` flag
- Inactive patients are excluded from active patient lists
- Inactive patients still appear in regular lists (different from soft delete)
- Patients can be reactivated

### Care Plan Management
- Care plan type (e.g., "basic", "intensive", "palliative")
- Care plan frequency (e.g., "daily", "weekly", "monthly")
- Both fields are optional and updateable

### Emergency Contact Information
- Emergency contact name
- Emergency contact phone number
- Both fields are optional and updateable

### Search Functionality
- Case-insensitive search (ILIKE)
- Searches across: first name, last name, email
- Works with pagination
- Works with active patient filtering

### Pagination
- Configurable limit and offset
- Returns total count
- Maintains sort order (newest first)

## Patient-Specific Data Model

Patients have additional fields compared to regular users:

- **Patient ID**: Sequential display ID (PT-0001, PT-0002, etc.)
- **Date of Birth**: Patient's birth date
- **Address**: Physical address
- **Emergency Contact**: Name and phone number
- **Medical Notes**: Free-text medical information
- **Care Plan**: Type and frequency of care
- **Is Active**: Active/inactive status flag

## Bug Fixes

During test development, we discovered and fixed the same SQL parameter bug found in users:
- **Issue**: `ListPatientsWithPagination` and `ListActivePatientsWithPagination` search had wrong SQL parameter numbering
- **Fix**: Separate WHERE clauses for count and data queries with correct parameter numbering
- **Files**: `internal/patient/repository.go` lines 298-327 and 429-458

## Test Results

```
=== RUN   TestRepositoryCreatePatient_Integration
--- PASS: TestRepositoryCreatePatient_Integration (0.06s)
=== RUN   TestRepositoryGetPatient_Integration
--- PASS: TestRepositoryGetPatient_Integration (0.05s)
=== RUN   TestRepositoryGetPatient_CrossTenantIsolation_Integration
--- PASS: TestRepositoryGetPatient_CrossTenantIsolation_Integration (0.06s)
=== RUN   TestRepositoryListPatients_Integration
--- PASS: TestRepositoryListPatients_Integration (0.04s)
=== RUN   TestRepositoryListPatientsWithPagination_Integration
--- PASS: TestRepositoryListPatientsWithPagination_Integration (0.05s)
=== RUN   TestRepositoryListPatientsWithPagination_Search_Integration
--- PASS: TestRepositoryListPatientsWithPagination_Search_Integration (0.06s)
=== RUN   TestRepositoryListActivePatientsWithPagination_Integration
--- PASS: TestRepositoryListActivePatientsWithPagination_Integration (0.04s)
=== RUN   TestRepositoryUpdate_Integration
--- PASS: TestRepositoryUpdate_Integration (0.04s)
=== RUN   TestRepositoryDelete_Integration
--- PASS: TestRepositoryDelete_Integration (0.04s)
=== RUN   TestRepositoryPatientIDGeneration_Integration
--- PASS: TestRepositoryPatientIDGeneration_Integration (0.04s)
=== RUN   TestRepositorySoftDelete_ExcludesFromList_Integration
--- PASS: TestRepositorySoftDelete_ExcludesFromList_Integration (0.04s)
=== RUN   TestRepositoryGetPatient_NotFound_Integration
--- PASS: TestRepositoryGetPatient_NotFound_Integration (0.03s)
=== RUN   TestRepositoryUpdate_NotFound_Integration
--- PASS: TestRepositoryUpdate_NotFound_Integration (0.03s)
=== RUN   TestRepositoryDelete_NotFound_Integration
--- PASS: TestRepositoryDelete_NotFound_Integration (0.03s)
=== RUN   TestRepositoryDelete_AlreadyDeleted_Integration
--- PASS: TestRepositoryDelete_AlreadyDeleted_Integration (0.04s)
=== RUN   TestRepositoryPatient_CareplanFields_Integration
--- PASS: TestRepositoryPatient_CareplanFields_Integration (0.04s)
=== RUN   TestRepositoryPatient_EmergencyContact_Integration
--- PASS: TestRepositoryPatient_EmergencyContact_Integration (0.04s)
=== RUN   TestRepositoryPatient_IsActiveFlag_Integration
--- PASS: TestRepositoryPatient_IsActiveFlag_Integration (0.04s)

PASS - 18/18 tests passing
```

## Comparison with Users Tests

| Feature | Users | Patients |
|---------|-------|----------|
| CRUD Operations | ✅ | ✅ |
| Pagination | ✅ | ✅ |
| Search | ✅ | ✅ |
| Multi-tenancy | ✅ | ✅ |
| Soft Delete | ✅ | ✅ |
| Sequential IDs | EMP-0001 | PT-0001 |
| Role Filtering | ✅ | N/A |
| Active/Inactive | N/A | ✅ |
| Care Plan | N/A | ✅ |
| Emergency Contact | N/A | ✅ |

## Next Steps

Consider adding:
1. Service layer integration tests (with mocked Keycloak)
2. Handler/API integration tests (with test HTTP server)
3. Tests for patient-caregiver assignments
4. Tests for care plan updates and history
5. Performance tests for large patient datasets
