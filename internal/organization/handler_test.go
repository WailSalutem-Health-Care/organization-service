package organization

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

// TestHandlerCreateOrganization_Success tests successful organization creation
func TestHandlerCreateOrganization_Success(t *testing.T) {
	mockService := &mockService{
		createOrgFunc: func(ctx context.Context, req CreateOrganizationRequest) (*OrganizationResponse, error) {
			return &OrganizationResponse{
				ID:        "org-123",
				Name:      req.Name,
				Status:    "active",
				CreatedAt: time.Now(),
			}, nil
		},
	}

	handler := NewHandler(mockService)

	reqBody := CreateOrganizationRequest{
		Name:         "Test Org",
		ContactEmail: "test@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	
	// Add principal to context
	principal := &auth.Principal{UserID: "user-1", Roles: []string{"SUPER_ADMIN"}}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.CreateOrganization(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rec.Code)
	}

	var response SuccessResponse
	json.NewDecoder(rec.Body).Decode(&response)

	if !response.Success {
		t.Error("Expected success to be true")
	}
	if response.Organization == nil {
		t.Error("Expected organization in response")
	}
	if response.Organization.Name != "Test Org" {
		t.Errorf("Expected name 'Test Org', got '%s'", response.Organization.Name)
	}
}

// TestHandlerCreateOrganization_Unauthenticated tests missing authentication
func TestHandlerCreateOrganization_Unauthenticated(t *testing.T) {
	mockService := &mockService{}
	handler := NewHandler(mockService)

	reqBody := CreateOrganizationRequest{Name: "Test Org"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.CreateOrganization(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}

	var response ErrorResponse
	json.NewDecoder(rec.Body).Decode(&response)

	if response.Error != "unauthenticated" {
		t.Errorf("Expected error 'unauthenticated', got '%s'", response.Error)
	}
}

// TestHandlerCreateOrganization_InvalidJSON tests malformed JSON
func TestHandlerCreateOrganization_InvalidJSON(t *testing.T) {
	mockService := &mockService{}
	handler := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewReader([]byte("invalid json")))
	principal := &auth.Principal{UserID: "user-1"}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.CreateOrganization(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}

	var response ErrorResponse
	json.NewDecoder(rec.Body).Decode(&response)

	if response.Error != "invalid_request" {
		t.Errorf("Expected error 'invalid_request', got '%s'", response.Error)
	}
}

// TestHandlerCreateOrganization_EmptyName tests validation
func TestHandlerCreateOrganization_EmptyName(t *testing.T) {
	mockService := &mockService{}
	handler := NewHandler(mockService)

	reqBody := CreateOrganizationRequest{Name: ""}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewReader(body))
	principal := &auth.Principal{UserID: "user-1"}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.CreateOrganization(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}

	var response ErrorResponse
	json.NewDecoder(rec.Body).Decode(&response)

	if response.Error != "validation_error" {
		t.Errorf("Expected error 'validation_error', got '%s'", response.Error)
	}
}

// TestHandlerCreateOrganization_ServiceError tests service layer error
func TestHandlerCreateOrganization_ServiceError(t *testing.T) {
	mockService := &mockService{
		createOrgFunc: func(ctx context.Context, req CreateOrganizationRequest) (*OrganizationResponse, error) {
			return nil, errors.New("database error")
		},
	}

	handler := NewHandler(mockService)

	reqBody := CreateOrganizationRequest{Name: "Test Org"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewReader(body))
	principal := &auth.Principal{UserID: "user-1"}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.CreateOrganization(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rec.Code)
	}
}

// TestHandlerListOrganizations_Success tests successful listing
func TestHandlerListOrganizations_Success(t *testing.T) {
	mockService := &mockService{
		listOrgsPaginatedFunc: func(ctx context.Context, principal *auth.Principal, params pagination.Params) (*PaginatedListResponse, error) {
			return &PaginatedListResponse{
				Success: true,
				Organizations: []OrganizationResponse{
					{ID: "org-1", Name: "Org 1"},
					{ID: "org-2", Name: "Org 2"},
				},
				Pagination: pagination.Meta{
					CurrentPage:  1,
					TotalRecords: 2,
				},
			}, nil
		},
	}

	handler := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/organizations?page=1&limit=10", nil)
	principal := &auth.Principal{UserID: "user-1", Roles: []string{"SUPER_ADMIN"}}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.ListOrganizations(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var response PaginatedListResponse
	json.NewDecoder(rec.Body).Decode(&response)

	if len(response.Organizations) != 2 {
		t.Errorf("Expected 2 organizations, got %d", len(response.Organizations))
	}
}

// TestHandlerListOrganizations_Unauthenticated tests missing authentication
func TestHandlerListOrganizations_Unauthenticated(t *testing.T) {
	mockService := &mockService{}
	handler := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/organizations", nil)
	rec := httptest.NewRecorder()

	handler.ListOrganizations(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

// TestHandlerGetOrganization_Success tests successful retrieval
func TestHandlerGetOrganization_Success(t *testing.T) {
	mockService := &mockService{
		getOrgFunc: func(ctx context.Context, id string, principal *auth.Principal) (*OrganizationResponse, error) {
			return &OrganizationResponse{
				ID:     id,
				Name:   "Test Org",
				Status: "active",
			}, nil
		},
	}

	handler := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/organizations/org-123", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "org-123"})
	principal := &auth.Principal{UserID: "user-1", Roles: []string{"SUPER_ADMIN"}}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.GetOrganization(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var response SuccessResponse
	json.NewDecoder(rec.Body).Decode(&response)

	if response.Organization.ID != "org-123" {
		t.Errorf("Expected org ID 'org-123', got '%s'", response.Organization.ID)
	}
}

// TestHandlerGetOrganization_Forbidden tests forbidden access
func TestHandlerGetOrganization_Forbidden(t *testing.T) {
	mockService := &mockService{
		getOrgFunc: func(ctx context.Context, id string, principal *auth.Principal) (*OrganizationResponse, error) {
			return nil, errors.New("forbidden")
		},
	}

	handler := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/organizations/org-123", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "org-123"})
	principal := &auth.Principal{UserID: "user-1", Roles: []string{"ORG_ADMIN"}, OrgID: "org-456"}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.GetOrganization(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rec.Code)
	}

	var response ErrorResponse
	json.NewDecoder(rec.Body).Decode(&response)

	if response.Error != "forbidden" {
		t.Errorf("Expected error 'forbidden', got '%s'", response.Error)
	}
}

// TestHandlerGetOrganization_NotFound tests not found error
func TestHandlerGetOrganization_NotFound(t *testing.T) {
	mockService := &mockService{
		getOrgFunc: func(ctx context.Context, id string, principal *auth.Principal) (*OrganizationResponse, error) {
			return nil, errors.New("organization not found")
		},
	}

	handler := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/organizations/nonexistent", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "nonexistent"})
	principal := &auth.Principal{UserID: "user-1", Roles: []string{"SUPER_ADMIN"}}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.GetOrganization(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rec.Code)
	}
}

// TestHandlerUpdateOrganization_Success tests successful update
func TestHandlerUpdateOrganization_Success(t *testing.T) {
	newName := "Updated Org"
	mockService := &mockService{
		updateOrgFunc: func(ctx context.Context, id string, req UpdateOrganizationRequest, principal *auth.Principal) (*OrganizationResponse, error) {
			return &OrganizationResponse{
				ID:   id,
				Name: *req.Name,
			}, nil
		},
	}

	handler := NewHandler(mockService)

	reqBody := UpdateOrganizationRequest{Name: &newName}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/organizations/org-123", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "org-123"})
	principal := &auth.Principal{UserID: "user-1", Roles: []string{"SUPER_ADMIN"}}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.UpdateOrganization(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var response SuccessResponse
	json.NewDecoder(rec.Body).Decode(&response)

	if response.Organization.Name != "Updated Org" {
		t.Errorf("Expected name 'Updated Org', got '%s'", response.Organization.Name)
	}
}

// TestHandlerUpdateOrganization_Forbidden tests forbidden update
func TestHandlerUpdateOrganization_Forbidden(t *testing.T) {
	mockService := &mockService{
		updateOrgFunc: func(ctx context.Context, id string, req UpdateOrganizationRequest, principal *auth.Principal) (*OrganizationResponse, error) {
			return nil, errors.New("forbidden")
		},
	}

	handler := NewHandler(mockService)

	newName := "Updated Org"
	reqBody := UpdateOrganizationRequest{Name: &newName}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/organizations/org-123", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "org-123"})
	principal := &auth.Principal{UserID: "user-1", Roles: []string{"ORG_ADMIN"}}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.UpdateOrganization(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rec.Code)
	}
}

// TestHandlerDeleteOrganization_Success tests successful deletion
func TestHandlerDeleteOrganization_Success(t *testing.T) {
	mockService := &mockService{
		deleteOrgFunc: func(ctx context.Context, id string) error {
			return nil
		},
	}

	handler := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodDelete, "/organizations/org-123", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "org-123"})
	principal := &auth.Principal{UserID: "user-1", Roles: []string{"SUPER_ADMIN"}}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.DeleteOrganization(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", rec.Code)
	}

	if rec.Body.Len() != 0 {
		t.Error("Expected empty body for 204 response")
	}
}

// TestHandlerDeleteOrganization_ServiceError tests deletion error
func TestHandlerDeleteOrganization_ServiceError(t *testing.T) {
	mockService := &mockService{
		deleteOrgFunc: func(ctx context.Context, id string) error {
			return errors.New("deletion failed")
		},
	}

	handler := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodDelete, "/organizations/org-123", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "org-123"})
	principal := &auth.Principal{UserID: "user-1", Roles: []string{"SUPER_ADMIN"}}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.DeleteOrganization(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rec.Code)
	}
}

// Mock service implementation

type mockService struct {
	createOrgFunc         func(ctx context.Context, req CreateOrganizationRequest) (*OrganizationResponse, error)
	listOrgsFunc          func(ctx context.Context, principal *auth.Principal) ([]OrganizationResponse, error)
	listOrgsPaginatedFunc func(ctx context.Context, principal *auth.Principal, params pagination.Params) (*PaginatedListResponse, error)
	getOrgFunc            func(ctx context.Context, id string, principal *auth.Principal) (*OrganizationResponse, error)
	updateOrgFunc         func(ctx context.Context, id string, req UpdateOrganizationRequest, principal *auth.Principal) (*OrganizationResponse, error)
	deleteOrgFunc         func(ctx context.Context, id string) error
}

func (m *mockService) CreateOrganization(ctx context.Context, req CreateOrganizationRequest) (*OrganizationResponse, error) {
	if m.createOrgFunc != nil {
		return m.createOrgFunc(ctx, req)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) ListOrganizations(ctx context.Context, principal *auth.Principal) ([]OrganizationResponse, error) {
	if m.listOrgsFunc != nil {
		return m.listOrgsFunc(ctx, principal)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) ListOrganizationsWithPagination(ctx context.Context, principal *auth.Principal, params pagination.Params) (*PaginatedListResponse, error) {
	if m.listOrgsPaginatedFunc != nil {
		return m.listOrgsPaginatedFunc(ctx, principal, params)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) GetOrganization(ctx context.Context, id string, principal *auth.Principal) (*OrganizationResponse, error) {
	if m.getOrgFunc != nil {
		return m.getOrgFunc(ctx, id, principal)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) UpdateOrganization(ctx context.Context, id string, req UpdateOrganizationRequest, principal *auth.Principal) (*OrganizationResponse, error) {
	if m.updateOrgFunc != nil {
		return m.updateOrgFunc(ctx, id, req, principal)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) DeleteOrganization(ctx context.Context, id string) error {
	if m.deleteOrgFunc != nil {
		return m.deleteOrgFunc(ctx, id)
	}
	return errors.New("not implemented")
}

