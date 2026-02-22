# Organization Service API Documentation

**Base URL**: `https://organization-service-wailsalutem-suite.apps.inholland-minor.openshift.eu`

## Authentication

All endpoints (except `/health`) require JWT authentication via Bearer token.

**Required Header:**
```
Authorization: Bearer <jwt_token>
```

**Optional Header** (for SUPER_ADMIN cross-org access):
```
X-Organization-ID: <organization-uuid>
```

---

## Role-Based Permissions

| Role | Permissions |
|------|------------|
| **SUPER_ADMIN** | All permissions, can manage all organizations |
| **ORG_ADMIN** | Can only manage their own organization (view org, full user/patient management) |
| **CAREGIVER** | Can view patients, manage care sessions, NFC check-in/out |
| **PATIENT** | Can view own profile and care sessions |
| **MUNICIPALITY** | Can view care session reports |
| **INSURER** | Can view care session reports |

---

## 📋 Organizations API

### 1. Create Organization
**POST** `/organizations`

**Permission**: `organization:create` (SUPER_ADMIN only)

**Request Body:**
```json
{
  "name": "LifeCare Healthcare",
  "contact_email": "admin@lifecare.nl",
  "contact_phone": "+31 20 123 4567",
  "address": "Amsterdam, Netherlands"
}
```

**Response:** `201 Created`
```json
{
  "id": "315298bf-0069-4b81-9469-c598670a2af2",
  "name": "LifeCare Healthcare",
  "schema_name": "org_lifecare_healthcare_315298bf",
  "contact_email": "admin@lifecare.nl",
  "contact_phone": "+31 20 123 4567",
  "address": "Amsterdam, Netherlands",
  "status": "active",
  "created_at": "2026-01-10T10:30:00Z"
}
```

---

### 2. List Organizations (with Pagination)
**GET** `/organizations?page=1&size=10`

**Permission**: `organization:view` (SUPER_ADMIN, ORG_ADMIN)

**Behavior:**
- **SUPER_ADMIN**: Returns all organizations
- **ORG_ADMIN**: Returns only their own organization

**Response:** `200 OK`
```json
{
  "success": true,
  "organizations": [
    {
      "id": "315298bf-0069-4b81-9469-c598670a2af2",
      "name": "LifeCare Healthcare",
      "schema_name": "org_lifecare_healthcare_315298bf",
      "contact_email": "admin@lifecare.nl",
      "contact_phone": "+31 20 123 4567",
      "address": "Amsterdam, Netherlands",
      "status": "active",
      "created_at": "2026-01-10T10:30:00Z"
    }
  ],
  "pagination": {
    "current_page": 1,
    "page_size": 10,
    "total_items": 1,
    "total_pages": 1
  }
}
```

---

### 3. Get Organization by ID
**GET** `/organizations/{id}`

**Permission**: `organization:view` (SUPER_ADMIN, ORG_ADMIN)

**Response:** `200 OK`
```json
{
  "id": "315298bf-0069-4b81-9469-c598670a2af2",
  "name": "LifeCare Healthcare",
  "schema_name": "org_lifecare_healthcare_315298bf",
  "contact_email": "admin@lifecare.nl",
  "contact_phone": "+31 20 123 4567",
  "address": "Amsterdam, Netherlands",
  "status": "active",
  "created_at": "2026-01-10T10:30:00Z"
}
```

---

### 4. Update Organization
**PUT/PATCH** `/organizations/{id}`

**Permission**: `organization:update` (SUPER_ADMIN only)

**Request Body:** (all fields optional)
```json
{
  "name": "LifeCare Healthcare B.V.",
  "contact_email": "info@lifecare.nl",
  "contact_phone": "+31 20 999 8888",
  "address": "Rotterdam, Netherlands"
}
```

**Response:** `200 OK`
```json
{
  "id": "315298bf-0069-4b81-9469-c598670a2af2",
  "name": "LifeCare Healthcare B.V.",
  "schema_name": "org_lifecare_healthcare_315298bf",
  "contact_email": "info@lifecare.nl",
  "contact_phone": "+31 20 999 8888",
  "address": "Rotterdam, Netherlands",
  "status": "active",
  "created_at": "2026-01-10T10:30:00Z"
}
```

---

### 5. Delete Organization
**DELETE** `/organizations/{id}`

**Permission**: `organization:delete` (SUPER_ADMIN only)

**Response:** `204 No Content`

---

## 👥 Users API

### 6. Create User
**POST** `/organization/users`

**Permission**: `user:create` (SUPER_ADMIN, ORG_ADMIN)

**Rules:**
- **SUPER_ADMIN**: Can create any role including ORG_ADMIN
- **ORG_ADMIN**: Can only create: CAREGIVER, PATIENT, MUNICIPALITY, INSURER

**Request Body:**
```json
{
  "username": "john.doe",
  "email": "john@lifecare.nl",
  "firstName": "John",
  "lastName": "Doe",
  "phoneNumber": "+31 6 1234 5678",
  "role": "CAREGIVER",
  "temporaryPassword": "TempPass123!",
  "sendResetEmail": false
}
```

**Response:** `201 Created`
```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "keycloakUserId": "8ed51553-2dd2-41e1-9812-45705d538d8e",
  "email": "john@lifecare.nl",
  "firstName": "John",
  "lastName": "Doe",
  "phoneNumber": "+31 6 1234 5678",
  "role": "CAREGIVER",
  "isActive": true,
  "orgId": "315298bf-0069-4b81-9469-c598670a2af2",
  "orgSchemaName": "org_lifecare_healthcare_315298bf",
  "createdAt": "2026-01-11T14:20:00Z"
}
```

---

### 7. List Users (with Pagination)
**GET** `/organization/users?page=1&size=20`

**Permission**: `user:view` (SUPER_ADMIN, ORG_ADMIN)

**Response:** `200 OK`
```json
{
  "users": [
    {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "keycloakUserId": "8ed51553-2dd2-41e1-9812-45705d538d8e",
      "email": "john@lifecare.nl",
      "firstName": "John",
      "lastName": "Doe",
      "phoneNumber": "+31 6 1234 5678",
      "role": "CAREGIVER",
      "isActive": true,
      "orgId": "315298bf-0069-4b81-9469-c598670a2af2",
      "orgSchemaName": "org_lifecare_healthcare_315298bf",
      "createdAt": "2026-01-11T14:20:00Z"
    }
  ],
  "pagination": {
    "current_page": 1,
    "page_size": 20,
    "total_items": 1,
    "total_pages": 1
  }
}
```

---

### 8. Get User by ID
**GET** `/organization/users/{id}`

**Permission**: `user:view` (SUPER_ADMIN, ORG_ADMIN)

**Response:** `200 OK`
```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "keycloakUserId": "8ed51553-2dd2-41e1-9812-45705d538d8e",
  "email": "john@lifecare.nl",
  "firstName": "John",
  "lastName": "Doe",
  "phoneNumber": "+31 6 1234 5678",
  "role": "CAREGIVER",
  "isActive": true,
  "orgId": "315298bf-0069-4b81-9469-c598670a2af2",
  "orgSchemaName": "org_lifecare_healthcare_315298bf",
  "createdAt": "2026-01-11T14:20:00Z"
}
```

---

### 9. Update User
**PATCH** `/organization/users/{id}`

**Permission**: `user:update` (SUPER_ADMIN, ORG_ADMIN)

**Request Body:** (all fields optional)
```json
{
  "email": "john.doe@lifecare.nl",
  "firstName": "John",
  "lastName": "Doe",
  "phoneNumber": "+31 6 9876 5432"
}
```

**Response:** `200 OK`
```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "keycloakUserId": "8ed51553-2dd2-41e1-9812-45705d538d8e",
  "email": "john.doe@lifecare.nl",
  "firstName": "John",
  "lastName": "Doe",
  "phoneNumber": "+31 6 9876 5432",
  "role": "CAREGIVER",
  "isActive": true,
  "orgId": "315298bf-0069-4b81-9469-c598670a2af2",
  "orgSchemaName": "org_lifecare_healthcare_315298bf",
  "createdAt": "2026-01-11T14:20:00Z",
  "updatedAt": "2026-01-11T15:30:00Z"
}
```

---

### 10. Update My Profile
**PATCH** `/organization/users/me`

**Permission**: Any authenticated user (no specific permission required)

**Request Body:** (all fields optional)
```json
{
  "email": "mynewemail@lifecare.nl",
  "firstName": "John",
  "lastName": "Doe",
  "phoneNumber": "+31 6 1111 2222"
}
```

**Response:** `200 OK`
```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "keycloakUserId": "8ed51553-2dd2-41e1-9812-45705d538d8e",
  "email": "mynewemail@lifecare.nl",
  "firstName": "John",
  "lastName": "Doe",
  "phoneNumber": "+31 6 1111 2222",
  "role": "CAREGIVER",
  "isActive": true,
  "orgId": "315298bf-0069-4b81-9469-c598670a2af2",
  "orgSchemaName": "org_lifecare_healthcare_315298bf",
  "createdAt": "2026-01-11T14:20:00Z",
  "updatedAt": "2026-01-11T16:00:00Z"
}
```

---

### 11. Reset User Password
**POST** `/organization/users/{id}/reset-password`

**Permission**: `user:update` (SUPER_ADMIN, ORG_ADMIN)

**Request Body:**
```json
{
  "temporaryPassword": "NewTempPass123!",
  "sendEmail": true
}
```

**Response:** `200 OK`
```json
{
  "message": "Password reset successful"
}
```

---

### 12. Delete User
**DELETE** `/organization/users/{id}`

**Permission**: `user:delete` (SUPER_ADMIN, ORG_ADMIN)

**Response:** `204 No Content`

---

### 13. List Active Caregivers
**GET** `/organization/users/caregivers/active?page=1&size=20`

**Permission**: `user:view` (SUPER_ADMIN, ORG_ADMIN)

**Response:** `200 OK`
```json
{
  "users": [
    {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "keycloakUserId": "8ed51553-2dd2-41e1-9812-45705d538d8e",
      "email": "john@lifecare.nl",
      "firstName": "John",
      "lastName": "Doe",
      "phoneNumber": "+31 6 1234 5678",
      "role": "CAREGIVER",
      "isActive": true,
      "orgId": "315298bf-0069-4b81-9469-c598670a2af2",
      "orgSchemaName": "org_lifecare_healthcare_315298bf",
      "createdAt": "2026-01-11T14:20:00Z"
    }
  ],
  "pagination": {
    "current_page": 1,
    "page_size": 20,
    "total_items": 1,
    "total_pages": 1
  }
}
```

---

### 14. List Active Municipality Users
**GET** `/organization/users/municipality/active?page=1&size=20`

**Permission**: `user:view` (SUPER_ADMIN, ORG_ADMIN)

**Response:** Same format as List Active Caregivers

---

### 15. List Active Insurers
**GET** `/organization/users/insurers/active?page=1&size=20`

**Permission**: `user:view` (SUPER_ADMIN, ORG_ADMIN)

**Response:** Same format as List Active Caregivers

---

### 16. List Active Org Admins
**GET** `/organization/users/org-admins/active?page=1&size=20`

**Permission**: `user:view` (SUPER_ADMIN, ORG_ADMIN)

**Response:** Same format as List Active Caregivers

---

## 🏥 Patients API

### 17. Create Patient
**POST** `/organization/patients`

**Permission**: `patient:create` (SUPER_ADMIN, ORG_ADMIN, CAREGIVER)

**Request Body:**
```json
{
  "username": "jane.smith",
  "temporaryPassword": "PatientPass123!",
  "sendResetEmail": false,
  "firstName": "Jane",
  "lastName": "Smith",
  "email": "jane.smith@example.com",
  "phoneNumber": "+31 6 8888 9999",
  "dateOfBirth": "1960-05-15",
  "address": "123 Main St, Amsterdam",
  "emergencyContactName": "John Smith",
  "emergencyContactPhone": "+31 6 7777 6666",
  "medicalNotes": "Diabetes, requires insulin",
  "careplanType": "intensive",
  "careplanFrequency": "daily"
}
```

**Response:** `201 Created`
```json
{
  "id": "p1a2b3c4-d5e6-7890-abcd-ef1234567890",
  "patient_id": "PAT-001",
  "keycloak_user_id": "7bd41442-1dc1-40e0-8821-34604d427d7d",
  "first_name": "Jane",
  "last_name": "Smith",
  "email": "jane.smith@example.com",
  "phone_number": "+31 6 8888 9999",
  "date_of_birth": "1960-05-15",
  "address": "123 Main St, Amsterdam",
  "emergency_contact_name": "John Smith",
  "emergency_contact_phone": "+31 6 7777 6666",
  "medical_notes": "Diabetes, requires insulin",
  "careplan_type": "intensive",
  "careplan_frequency": "daily",
  "is_active": true,
  "created_at": "2026-01-11T10:00:00Z"
}
```

---

### 18. List Patients (with Pagination)
**GET** `/organization/patients?page=1&size=20`

**Permission**: `patient:view` (SUPER_ADMIN, ORG_ADMIN, CAREGIVER, PATIENT)

**Response:** `200 OK`
```json
{
  "success": true,
  "patients": [
    {
      "id": "p1a2b3c4-d5e6-7890-abcd-ef1234567890",
      "patient_id": "PAT-001",
      "keycloak_user_id": "7bd41442-1dc1-40e0-8821-34604d427d7d",
      "first_name": "Jane",
      "last_name": "Smith",
      "email": "jane.smith@example.com",
      "phone_number": "+31 6 8888 9999",
      "date_of_birth": "1960-05-15",
      "address": "123 Main St, Amsterdam",
      "emergency_contact_name": "John Smith",
      "emergency_contact_phone": "+31 6 7777 6666",
      "medical_notes": "Diabetes, requires insulin",
      "careplan_type": "intensive",
      "careplan_frequency": "daily",
      "is_active": true,
      "created_at": "2026-01-11T10:00:00Z"
    }
  ],
  "pagination": {
    "current_page": 1,
    "page_size": 20,
    "total_items": 1,
    "total_pages": 1
  }
}
```

---

### 19. List Active Patients
**GET** `/organization/patients/active?page=1&size=20`

**Permission**: `patient:view` (SUPER_ADMIN, ORG_ADMIN, CAREGIVER, PATIENT)

**Response:** Same format as List Patients, but only includes active patients

---

### 20. Get Patient by ID
**GET** `/organization/patients/{id}`

**Permission**: `patient:view` (SUPER_ADMIN, ORG_ADMIN, CAREGIVER, PATIENT)

**Response:** `200 OK`
```json
{
  "id": "p1a2b3c4-d5e6-7890-abcd-ef1234567890",
  "patient_id": "PAT-001",
  "keycloak_user_id": "7bd41442-1dc1-40e0-8821-34604d427d7d",
  "first_name": "Jane",
  "last_name": "Smith",
  "email": "jane.smith@example.com",
  "phone_number": "+31 6 8888 9999",
  "date_of_birth": "1960-05-15",
  "address": "123 Main St, Amsterdam",
  "emergency_contact_name": "John Smith",
  "emergency_contact_phone": "+31 6 7777 6666",
  "medical_notes": "Diabetes, requires insulin",
  "careplan_type": "intensive",
  "careplan_frequency": "daily",
  "is_active": true,
  "created_at": "2026-01-11T10:00:00Z"
}
```

---

### 21. Update Patient
**PUT/PATCH** `/organization/patients/{id}`

**Permission**: `patient:update` (SUPER_ADMIN, ORG_ADMIN, PATIENT)

**Request Body:** (all fields optional)
```json
{
  "first_name": "Jane",
  "last_name": "Smith-Johnson",
  "email": "jane.johnson@example.com",
  "phone_number": "+31 6 9999 0000",
  "address": "456 New St, Rotterdam",
  "emergency_contact_name": "Mary Johnson",
  "emergency_contact_phone": "+31 6 5555 4444",
  "medical_notes": "Diabetes, requires insulin. Added heart condition.",
  "is_active": true,
  "careplan_type": "intensive",
  "careplan_frequency": "twice-daily"
}
```

**Response:** `200 OK`
```json
{
  "id": "p1a2b3c4-d5e6-7890-abcd-ef1234567890",
  "patient_id": "PAT-001",
  "keycloak_user_id": "7bd41442-1dc1-40e0-8821-34604d427d7d",
  "first_name": "Jane",
  "last_name": "Smith-Johnson",
  "email": "jane.johnson@example.com",
  "phone_number": "+31 6 9999 0000",
  "date_of_birth": "1960-05-15",
  "address": "456 New St, Rotterdam",
  "emergency_contact_name": "Mary Johnson",
  "emergency_contact_phone": "+31 6 5555 4444",
  "medical_notes": "Diabetes, requires insulin. Added heart condition.",
  "careplan_type": "intensive",
  "careplan_frequency": "twice-daily",
  "is_active": true,
  "created_at": "2026-01-11T10:00:00Z",
  "updated_at": "2026-01-11T18:00:00Z"
}
```

---

### 22. Delete Patient
**DELETE** `/organization/patients/{id}`

**Permission**: `patient:delete` (SUPER_ADMIN, ORG_ADMIN)

**Response:** `204 No Content`

---

## 🏥 Health Check

### 23. Health Check (Public)
**GET** `/health`

**Permission**: None (public endpoint)

**Response:** `200 OK`
```json
{
  "status": "ok",
  "service": "organization-service"
}
```

---

## Error Responses

All error responses follow this format:

### 400 Bad Request
```json
{
  "error": "invalid request body"
}
```

### 401 Unauthorized
```json
{
  "error": "unauthorized"
}
```

### 403 Forbidden
```json
{
  "error": "forbidden"
}
```

### 404 Not Found
```json
{
  "error": "organization not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "internal server error"
}
```

---

## Pagination

All list endpoints support pagination with query parameters:
- `page` (default: 1) - Page number
- `size` (default: 20) - Items per page

**Response includes:**
```json
{
  "pagination": {
    "current_page": 1,
    "page_size": 20,
    "total_items": 45,
    "total_pages": 3
  }
}
```

---

## Quick Reference Table

| Method | Endpoint | Permission | Roles |
|--------|----------|------------|-------|
| POST | `/organizations` | `organization:create` | SUPER_ADMIN |
| GET | `/organizations` | `organization:view` | SUPER_ADMIN, ORG_ADMIN |
| GET | `/organizations/{id}` | `organization:view` | SUPER_ADMIN, ORG_ADMIN |
| PUT/PATCH | `/organizations/{id}` | `organization:update` | SUPER_ADMIN |
| DELETE | `/organizations/{id}` | `organization:delete` | SUPER_ADMIN |
| POST | `/organization/users` | `user:create` | SUPER_ADMIN, ORG_ADMIN |
| GET | `/organization/users` | `user:view` | SUPER_ADMIN, ORG_ADMIN |
| GET | `/organization/users/{id}` | `user:view` | SUPER_ADMIN, ORG_ADMIN |
| PATCH | `/organization/users/{id}` | `user:update` | SUPER_ADMIN, ORG_ADMIN |
| PATCH | `/organization/users/me` | None | All authenticated |
| POST | `/organization/users/{id}/reset-password` | `user:update` | SUPER_ADMIN, ORG_ADMIN |
| DELETE | `/organization/users/{id}` | `user:delete` | SUPER_ADMIN, ORG_ADMIN |
| GET | `/organization/users/caregivers/active` | `user:view` | SUPER_ADMIN, ORG_ADMIN |
| GET | `/organization/users/municipality/active` | `user:view` | SUPER_ADMIN, ORG_ADMIN |
| GET | `/organization/users/insurers/active` | `user:view` | SUPER_ADMIN, ORG_ADMIN |
| GET | `/organization/users/org-admins/active` | `user:view` | SUPER_ADMIN, ORG_ADMIN |
| POST | `/organization/patients` | `patient:create` | SUPER_ADMIN, ORG_ADMIN, CAREGIVER |
| GET | `/organization/patients` | `patient:view` | SUPER_ADMIN, ORG_ADMIN, CAREGIVER, PATIENT |
| GET | `/organization/patients/active` | `patient:view` | SUPER_ADMIN, ORG_ADMIN, CAREGIVER, PATIENT |
| GET | `/organization/patients/{id}` | `patient:view` | SUPER_ADMIN, ORG_ADMIN, CAREGIVER, PATIENT |
| PUT/PATCH | `/organization/patients/{id}` | `patient:update` | SUPER_ADMIN, ORG_ADMIN, PATIENT |
| DELETE | `/organization/patients/{id}` | `patient:delete` | SUPER_ADMIN, ORG_ADMIN |
| GET | `/health` | None | Public |

---

## Frontend Implementation Notes

1. **Store JWT token** after Keycloak login
2. **Include token** in Authorization header for all requests
3. **Handle pagination** - most list endpoints support `?page=1&size=20`
4. **Check user role** to show/hide UI elements based on permissions
5. **Handle errors** - all errors return consistent format
6. **SUPER_ADMIN cross-org access** - use `X-Organization-ID` header when needed
7. **Date formats** - Use `YYYY-MM-DD` for date_of_birth
8. **Password requirements** - Ensure temporary passwords meet complexity requirements

---

**Last Updated:** January 12, 2026
