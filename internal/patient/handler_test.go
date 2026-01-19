package patient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/WailSalutem-Health-Care/organization-service/internal/auth"
	"github.com/WailSalutem-Health-Care/organization-service/internal/pagination"
	"github.com/gorilla/mux"
)

// mockService implements ServiceInterface for testing
type mockService struct {
	createPatientFunc                    func(ctx context.Context, schemaName, orgID string, req CreatePatientRequest) (*PatientResponse, error)
	getPatientFunc                       func(ctx context.Context, schemaName, id string) (*PatientResponse, error)
	listPatientsFunc                     func(ctx context.Context, schemaName string) ([]PatientResponse, error)
	listPatientsWithPaginationFunc       func(ctx context.Context, schemaName string, params pagination.Params) (*PaginatedPatientListResponse, error)
	listActivePatientsWithPaginationFunc func(ctx context.Context, schemaName string, params pagination.Params) (*PaginatedPatientListResponse, error)
	updatePatientFunc                    func(ctx context.Context, schemaName, id string, req UpdatePatientRequest) (*PatientResponse, error)
	deletePatientFunc                    func(ctx context.Context, schemaName, orgID, id string) error
}

func (m *mockService) CreatePatient(ctx context.Context, schemaName, orgID string, req CreatePatientRequest) (*PatientResponse, error) {
	if m.createPatientFunc != nil {
		return m.createPatientFunc(ctx, schemaName, orgID, req)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) GetPatient(ctx context.Context, schemaName, id string) (*PatientResponse, error) {
	if m.getPatientFunc != nil {
		return m.getPatientFunc(ctx, schemaName, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) ListPatients(ctx context.Context, schemaName string) ([]PatientResponse, error) {
	if m.listPatientsFunc != nil {
		return m.listPatientsFunc(ctx, schemaName)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) ListPatientsWithPagination(ctx context.Context, schemaName string, params pagination.Params) (*PaginatedPatientListResponse, error) {
	if m.listPatientsWithPaginationFunc != nil {
		return m.listPatientsWithPaginationFunc(ctx, schemaName, params)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) ListActivePatientsWithPagination(ctx context.Context, schemaName string, params pagination.Params) (*PaginatedPatientListResponse, error) {
	if m.listActivePatientsWithPaginationFunc != nil {
		return m.listActivePatientsWithPaginationFunc(ctx, schemaName, params)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) UpdatePatient(ctx context.Context, schemaName, id string, req UpdatePatientRequest) (*PatientResponse, error) {
	if m.updatePatientFunc != nil {
		return m.updatePatientFunc(ctx, schemaName, id, req)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) DeletePatient(ctx context.Context, schemaName, orgID, id string) error {
	if m.deletePatientFunc != nil {
		return m.deletePatientFunc(ctx, schemaName, orgID, id)
	}
	return errors.New("not implemented")
}

// mockSchemaLookup implements SchemaLookup for testing
type mockSchemaLookup struct {
	getSchemaNameByOrgIDFunc func(ctx context.Context, orgID string) (string, error)
}

func (m *mockSchemaLookup) GetSchemaNameByOrgID(ctx context.Context, orgID string) (string, error) {
	if m.getSchemaNameByOrgIDFunc != nil {
		return m.getSchemaNameByOrgIDFunc(ctx, orgID)
	}
	return "", errors.New("not implemented")
}

// Test CreatePatient Handler

func TestHandlerCreatePatient_Success(t *testing.T) {
	mockSvc := &mockService{
		createPatientFunc: func(ctx context.Context, schemaName, orgID string, req CreatePatientRequest) (*PatientResponse, error) {
			return &PatientResponse{
				ID:             "patient-123",
				PatientID:      "P-001",
				KeycloakUserID: "kc-123",
				FirstName:      req.FirstName,
				LastName:       req.LastName,
				Email:          req.Email,
				Address:        req.Address,
				IsActive:       true,
				CreatedAt:      time.Now(),
			}, nil
		},
	}

	mockLookup := &mockSchemaLookup{
		getSchemaNameByOrgIDFunc: func(ctx context.Context, orgID string) (string, error) {
			return "org_123", nil
		},
	}

	handler := NewHandler(mockSvc, mockLookup)

	reqBody := CreatePatientRequest{
		Username:    "patient1",
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "john@example.com",
		DateOfBirth: "1990-01-01",
		Address:     "123 Main St",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/patients", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Organization-ID", "org-123")

	principal := &auth.Principal{
		UserID: "admin-123",
		Roles:  []string{"SUPER_ADMIN"},
	}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.CreatePatient(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rr.Code)
	}

	var response PatientSuccessResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}

	if response.Patient.FirstName != "John" {
		t.Errorf("Expected first name John, got %s", response.Patient.FirstName)
	}
}

func TestHandlerCreatePatient_Unauthenticated(t *testing.T) {
	handler := NewHandler(&mockService{}, &mockSchemaLookup{})

	reqBody := CreatePatientRequest{
		Username:    "patient1",
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "john@example.com",
		DateOfBirth: "1990-01-01",
		Address:     "123 Main St",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/patients", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.CreatePatient(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}
}

func TestHandlerCreatePatient_SuperAdminMissingOrgHeader(t *testing.T) {
	handler := NewHandler(&mockService{}, &mockSchemaLookup{})

	reqBody := CreatePatientRequest{
		Username:    "patient1",
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "john@example.com",
		DateOfBirth: "1990-01-01",
		Address:     "123 Main St",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/patients", bytes.NewReader(body))
	principal := &auth.Principal{
		UserID: "admin-123",
		Roles:  []string{"SUPER_ADMIN"},
	}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.CreatePatient(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandlerCreatePatient_InvalidJSON(t *testing.T) {
	handler := NewHandler(&mockService{}, &mockSchemaLookup{})

	req := httptest.NewRequest(http.MethodPost, "/patients", bytes.NewReader([]byte("invalid json")))
	principal := &auth.Principal{
		UserID:        "admin-123",
		Roles:         []string{"ORG_ADMIN"},
		OrgID:         "org-123",
		OrgSchemaName: "org_123",
	}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.CreatePatient(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandlerCreatePatient_MissingFirstName(t *testing.T) {
	handler := NewHandler(&mockService{}, &mockSchemaLookup{})

	reqBody := CreatePatientRequest{
		Username:    "patient1",
		LastName:    "Doe",
		Email:       "john@example.com",
		DateOfBirth: "1990-01-01",
		Address:     "123 Main St",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/patients", bytes.NewReader(body))
	principal := &auth.Principal{
		UserID:        "admin-123",
		Roles:         []string{"ORG_ADMIN"},
		OrgID:         "org-123",
		OrgSchemaName: "org_123",
	}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.CreatePatient(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandlerCreatePatient_OrgAdminSuccess(t *testing.T) {
	mockSvc := &mockService{
		createPatientFunc: func(ctx context.Context, schemaName, orgID string, req CreatePatientRequest) (*PatientResponse, error) {
			return &PatientResponse{
				ID:        "patient-123",
				FirstName: req.FirstName,
				LastName:  req.LastName,
				Email:     req.Email,
				IsActive:  true,
			}, nil
		},
	}

	handler := NewHandler(mockSvc, &mockSchemaLookup{})

	reqBody := CreatePatientRequest{
		Username:    "patient1",
		FirstName:   "Jane",
		LastName:    "Smith",
		Email:       "jane@example.com",
		DateOfBirth: "1985-05-15",
		Address:     "456 Oak Ave",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/patients", bytes.NewReader(body))
	principal := &auth.Principal{
		UserID:        "admin-123",
		Roles:         []string{"ORG_ADMIN"},
		OrgID:         "org-123",
		OrgSchemaName: "org_123",
	}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.CreatePatient(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rr.Code)
	}
}

// Test ListPatients Handler

func TestHandlerListPatients_Success(t *testing.T) {
	mockSvc := &mockService{
		listPatientsWithPaginationFunc: func(ctx context.Context, schemaName string, params pagination.Params) (*PaginatedPatientListResponse, error) {
			return &PaginatedPatientListResponse{
				Success: true,
				Patients: []PatientResponse{
					{
						ID:        "patient-1",
						FirstName: "John",
						LastName:  "Doe",
						Email:     "john@example.com",
						IsActive:  true,
					},
					{
						ID:        "patient-2",
						FirstName: "Jane",
						LastName:  "Smith",
						Email:     "jane@example.com",
						IsActive:  true,
					},
				},
				Pagination: pagination.Meta{
					CurrentPage:  1,
					PerPage:      10,
					TotalPages:   1,
					TotalRecords: 2,
				},
			}, nil
		},
	}

	handler := NewHandler(mockSvc, &mockSchemaLookup{})

	req := httptest.NewRequest(http.MethodGet, "/patients", nil)
	principal := &auth.Principal{
		UserID:        "admin-123",
		Roles:         []string{"ORG_ADMIN"},
		OrgID:         "org-123",
		OrgSchemaName: "org_123",
	}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ListPatients(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response PaginatedPatientListResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Patients) != 2 {
		t.Errorf("Expected 2 patients, got %d", len(response.Patients))
	}
}

func TestHandlerListPatients_Unauthenticated(t *testing.T) {
	handler := NewHandler(&mockService{}, &mockSchemaLookup{})

	req := httptest.NewRequest(http.MethodGet, "/patients", nil)
	rr := httptest.NewRecorder()

	handler.ListPatients(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}
}

// Test ListActivePatients Handler

func TestHandlerListActivePatients_Success(t *testing.T) {
	mockSvc := &mockService{
		listActivePatientsWithPaginationFunc: func(ctx context.Context, schemaName string, params pagination.Params) (*PaginatedPatientListResponse, error) {
			return &PaginatedPatientListResponse{
				Success: true,
				Patients: []PatientResponse{
					{
						ID:        "patient-1",
						FirstName: "Active",
						LastName:  "Patient",
						Email:     "active@example.com",
						IsActive:  true,
					},
				},
				Pagination: pagination.Meta{
					CurrentPage:  1,
					PerPage:      10,
					TotalPages:   1,
					TotalRecords: 1,
				},
			}, nil
		},
	}

	handler := NewHandler(mockSvc, &mockSchemaLookup{})

	req := httptest.NewRequest(http.MethodGet, "/patients/active", nil)
	principal := &auth.Principal{
		UserID:        "admin-123",
		Roles:         []string{"ORG_ADMIN"},
		OrgID:         "org-123",
		OrgSchemaName: "org_123",
	}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ListActivePatients(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

// Test GetPatient Handler

func TestHandlerGetPatient_Success(t *testing.T) {
	mockSvc := &mockService{
		getPatientFunc: func(ctx context.Context, schemaName, id string) (*PatientResponse, error) {
			if id != "patient-123" {
				t.Errorf("Expected id patient-123, got %s", id)
			}
			return &PatientResponse{
				ID:        "patient-123",
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				IsActive:  true,
			}, nil
		},
	}

	handler := NewHandler(mockSvc, &mockSchemaLookup{})

	req := httptest.NewRequest(http.MethodGet, "/patients/patient-123", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "patient-123"})
	principal := &auth.Principal{
		UserID:        "admin-123",
		Roles:         []string{"ORG_ADMIN"},
		OrgID:         "org-123",
		OrgSchemaName: "org_123",
	}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.GetPatient(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response PatientSuccessResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Patient.ID != "patient-123" {
		t.Errorf("Expected patient ID patient-123, got %s", response.Patient.ID)
	}
}

func TestHandlerGetPatient_NotFound(t *testing.T) {
	mockSvc := &mockService{
		getPatientFunc: func(ctx context.Context, schemaName, id string) (*PatientResponse, error) {
			return nil, errors.New("patient not found")
		},
	}

	handler := NewHandler(mockSvc, &mockSchemaLookup{})

	req := httptest.NewRequest(http.MethodGet, "/patients/patient-999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "patient-999"})
	principal := &auth.Principal{
		UserID:        "admin-123",
		Roles:         []string{"ORG_ADMIN"},
		OrgID:         "org-123",
		OrgSchemaName: "org_123",
	}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.GetPatient(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

// Test UpdatePatient Handler

func TestHandlerUpdatePatient_Success(t *testing.T) {
	firstName := "Updated"
	mockSvc := &mockService{
		updatePatientFunc: func(ctx context.Context, schemaName, id string, req UpdatePatientRequest) (*PatientResponse, error) {
			return &PatientResponse{
				ID:        id,
				FirstName: *req.FirstName,
				LastName:  "Doe",
				Email:     "john@example.com",
				IsActive:  true,
			}, nil
		},
	}

	handler := NewHandler(mockSvc, &mockSchemaLookup{})

	reqBody := UpdatePatientRequest{
		FirstName: &firstName,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/patients/patient-123", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "patient-123"})
	principal := &auth.Principal{
		UserID:        "admin-123",
		Roles:         []string{"ORG_ADMIN"},
		OrgID:         "org-123",
		OrgSchemaName: "org_123",
	}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.UpdatePatient(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response PatientSuccessResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Patient.FirstName != "Updated" {
		t.Errorf("Expected first name Updated, got %s", response.Patient.FirstName)
	}
}

func TestHandlerUpdatePatient_InvalidJSON(t *testing.T) {
	handler := NewHandler(&mockService{}, &mockSchemaLookup{})

	req := httptest.NewRequest(http.MethodPut, "/patients/patient-123", bytes.NewReader([]byte("invalid")))
	req = mux.SetURLVars(req, map[string]string{"id": "patient-123"})
	principal := &auth.Principal{
		UserID:        "admin-123",
		Roles:         []string{"ORG_ADMIN"},
		OrgID:         "org-123",
		OrgSchemaName: "org_123",
	}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.UpdatePatient(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

// Test DeletePatient Handler

func TestHandlerDeletePatient_Success(t *testing.T) {
	mockSvc := &mockService{
		deletePatientFunc: func(ctx context.Context, schemaName, orgID, id string) error {
			return nil
		},
	}

	handler := NewHandler(mockSvc, &mockSchemaLookup{})

	req := httptest.NewRequest(http.MethodDelete, "/patients/patient-123", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "patient-123"})
	principal := &auth.Principal{
		UserID:        "admin-123",
		Roles:         []string{"ORG_ADMIN"},
		OrgID:         "org-123",
		OrgSchemaName: "org_123",
	}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.DeletePatient(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Error("Expected success to be true")
	}
}

func TestHandlerDeletePatient_ServiceError(t *testing.T) {
	mockSvc := &mockService{
		deletePatientFunc: func(ctx context.Context, schemaName, orgID, id string) error {
			return errors.New("deletion failed")
		},
	}

	handler := NewHandler(mockSvc, &mockSchemaLookup{})

	req := httptest.NewRequest(http.MethodDelete, "/patients/patient-123", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "patient-123"})
	principal := &auth.Principal{
		UserID:        "admin-123",
		Roles:         []string{"ORG_ADMIN"},
		OrgID:         "org-123",
		OrgSchemaName: "org_123",
	}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.DeletePatient(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rr.Code)
	}
}
