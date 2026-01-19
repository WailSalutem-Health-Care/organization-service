package users

import (
	"bytes"
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
	createUserFunc                           func(req CreateUserRequest, principal *auth.Principal, targetOrgID string) (*User, error)
	getUserFunc                              func(userID string, principal *auth.Principal, targetOrgID string) (*User, error)
	listUsersFunc                            func(principal *auth.Principal, targetOrgID string) ([]User, error)
	listUsersWithPaginationFunc              func(principal *auth.Principal, targetOrgID string, params pagination.Params) (*PaginatedUserListResponse, error)
	listActiveUsersByRoleWithPaginationFunc  func(principal *auth.Principal, targetOrgID string, role string, params pagination.Params) (*PaginatedUserListResponse, error)
	updateUserFunc                           func(userID string, req UpdateUserRequest, principal *auth.Principal, targetOrgID string) (*User, error)
	getMyProfileFunc                         func(principal *auth.Principal) (*User, error)
	updateMyProfileFunc                      func(req UpdateUserRequest, principal *auth.Principal) (*User, error)
	resetPasswordFunc                        func(userID string, req ResetPasswordRequest, principal *auth.Principal, targetOrgID string) error
	deleteUserFunc                           func(userID string, principal *auth.Principal) error
}

func (m *mockService) CreateUser(req CreateUserRequest, principal *auth.Principal, targetOrgID string) (*User, error) {
	if m.createUserFunc != nil {
		return m.createUserFunc(req, principal, targetOrgID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) GetUser(userID string, principal *auth.Principal, targetOrgID string) (*User, error) {
	if m.getUserFunc != nil {
		return m.getUserFunc(userID, principal, targetOrgID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) ListUsers(principal *auth.Principal, targetOrgID string) ([]User, error) {
	if m.listUsersFunc != nil {
		return m.listUsersFunc(principal, targetOrgID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) ListUsersWithPagination(principal *auth.Principal, targetOrgID string, params pagination.Params) (*PaginatedUserListResponse, error) {
	if m.listUsersWithPaginationFunc != nil {
		return m.listUsersWithPaginationFunc(principal, targetOrgID, params)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) ListActiveUsersByRoleWithPagination(principal *auth.Principal, targetOrgID string, role string, params pagination.Params) (*PaginatedUserListResponse, error) {
	if m.listActiveUsersByRoleWithPaginationFunc != nil {
		return m.listActiveUsersByRoleWithPaginationFunc(principal, targetOrgID, role, params)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) UpdateUser(userID string, req UpdateUserRequest, principal *auth.Principal, targetOrgID string) (*User, error) {
	if m.updateUserFunc != nil {
		return m.updateUserFunc(userID, req, principal, targetOrgID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) GetMyProfile(principal *auth.Principal) (*User, error) {
	if m.getMyProfileFunc != nil {
		return m.getMyProfileFunc(principal)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) UpdateMyProfile(req UpdateUserRequest, principal *auth.Principal) (*User, error) {
	if m.updateMyProfileFunc != nil {
		return m.updateMyProfileFunc(req, principal)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) ResetPassword(userID string, req ResetPasswordRequest, principal *auth.Principal, targetOrgID string) error {
	if m.resetPasswordFunc != nil {
		return m.resetPasswordFunc(userID, req, principal, targetOrgID)
	}
	return errors.New("not implemented")
}

func (m *mockService) DeleteUser(userID string, principal *auth.Principal) error {
	if m.deleteUserFunc != nil {
		return m.deleteUserFunc(userID, principal)
	}
	return errors.New("not implemented")
}

// Test CreateUser Handler

func TestHandlerCreateUser_Success(t *testing.T) {
	mockSvc := &mockService{
		createUserFunc: func(req CreateUserRequest, principal *auth.Principal, targetOrgID string) (*User, error) {
			return &User{
				ID:             "user-123",
				KeycloakUserID: "kc-123",
				Email:          req.Email,
				FirstName:      req.FirstName,
				LastName:       req.LastName,
				Role:           req.Role,
				IsActive:       true,
				OrgID:          "org-123",
				OrgSchemaName:  "org_123",
				CreatedAt:      time.Now(),
			}, nil
		},
	}

	handler := NewHandler(mockSvc)

	reqBody := CreateUserRequest{
		Username:          "testuser",
		Email:             "test@example.com",
		FirstName:         "Test",
		LastName:          "User",
		Role:              "CAREGIVER",
		TemporaryPassword: "temp123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Organization-ID", "org-123")

	principal := &auth.Principal{
		UserID: "admin-123",
		Roles:  []string{"SUPER_ADMIN"},
		OrgID:  "org-123",
	}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.CreateUser(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rr.Code)
	}

	var user User
	if err := json.NewDecoder(rr.Body).Decode(&user); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", user.Email)
	}
}

func TestHandlerCreateUser_Unauthenticated(t *testing.T) {
	handler := NewHandler(&mockService{})

	reqBody := CreateUserRequest{
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Role:      "CAREGIVER",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.CreateUser(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}
}

func TestHandlerCreateUser_InvalidJSON(t *testing.T) {
	handler := NewHandler(&mockService{})

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte("invalid json")))
	principal := &auth.Principal{UserID: "admin-123", Roles: []string{"SUPER_ADMIN"}}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.CreateUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandlerCreateUser_ValidationError(t *testing.T) {
	mockSvc := &mockService{
		createUserFunc: func(req CreateUserRequest, principal *auth.Principal, targetOrgID string) (*User, error) {
			return nil, ErrMissingEmail
		},
	}

	handler := NewHandler(mockSvc)

	reqBody := CreateUserRequest{
		Username:  "testuser",
		FirstName: "Test",
		LastName:  "User",
		Role:      "CAREGIVER",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	principal := &auth.Principal{UserID: "admin-123", Roles: []string{"SUPER_ADMIN"}}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.CreateUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandlerCreateUser_ForbiddenRole(t *testing.T) {
	mockSvc := &mockService{
		createUserFunc: func(req CreateUserRequest, principal *auth.Principal, targetOrgID string) (*User, error) {
			return nil, ErrRoleNotAllowed
		},
	}

	handler := NewHandler(mockSvc)

	reqBody := CreateUserRequest{
		Username:          "testuser",
		Email:             "test@example.com",
		FirstName:         "Test",
		LastName:          "User",
		Role:              "SUPER_ADMIN",
		TemporaryPassword: "temp123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	principal := &auth.Principal{UserID: "admin-123", Roles: []string{"ORG_ADMIN"}, OrgID: "org-123"}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.CreateUser(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rr.Code)
	}
}

// Test ListUsers Handler

func TestHandlerListUsers_Success(t *testing.T) {
	mockSvc := &mockService{
		listUsersWithPaginationFunc: func(principal *auth.Principal, targetOrgID string, params pagination.Params) (*PaginatedUserListResponse, error) {
			return &PaginatedUserListResponse{
				Users: []User{
					{
						ID:        "user-1",
						Email:     "user1@example.com",
						FirstName: "User",
						LastName:  "One",
						Role:      "CAREGIVER",
						IsActive:  true,
					},
					{
						ID:        "user-2",
						Email:     "user2@example.com",
						FirstName: "User",
						LastName:  "Two",
						Role:      "MUNICIPALITY",
						IsActive:  true,
					},
				},
				Pagination: pagination.Meta{
					CurrentPage:  1,
					PerPage:      10,
					TotalPages:   1,
					TotalRecords: 2,
					HasNext:      false,
					HasPrevious:  false,
				},
			}, nil
		},
	}

	handler := NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	principal := &auth.Principal{UserID: "admin-123", Roles: []string{"ORG_ADMIN"}, OrgID: "org-123"}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ListUsers(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response PaginatedUserListResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(response.Users))
	}
}

func TestHandlerListUsers_Unauthenticated(t *testing.T) {
	handler := NewHandler(&mockService{})

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rr := httptest.NewRecorder()

	handler.ListUsers(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}
}

func TestHandlerListUsers_Forbidden(t *testing.T) {
	mockSvc := &mockService{
		listUsersWithPaginationFunc: func(principal *auth.Principal, targetOrgID string, params pagination.Params) (*PaginatedUserListResponse, error) {
			return nil, ErrForbidden
		},
	}

	handler := NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req.Header.Set("X-Organization-ID", "other-org")
	principal := &auth.Principal{UserID: "admin-123", Roles: []string{"ORG_ADMIN"}, OrgID: "org-123"}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ListUsers(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rr.Code)
	}
}

// Test ListActiveCaregivers Handler

func TestHandlerListActiveCaregivers_Success(t *testing.T) {
	mockSvc := &mockService{
		listActiveUsersByRoleWithPaginationFunc: func(principal *auth.Principal, targetOrgID string, role string, params pagination.Params) (*PaginatedUserListResponse, error) {
			if role != "CAREGIVER" {
				t.Errorf("Expected role CAREGIVER, got %s", role)
			}
			return &PaginatedUserListResponse{
				Users: []User{
					{
						ID:        "user-1",
						Email:     "caregiver@example.com",
						FirstName: "Care",
						LastName:  "Giver",
						Role:      "CAREGIVER",
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

	handler := NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/users/caregivers/active", nil)
	principal := &auth.Principal{UserID: "admin-123", Roles: []string{"ORG_ADMIN"}, OrgID: "org-123"}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ListActiveCaregivers(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

// Test GetUser Handler

func TestHandlerGetUser_Success(t *testing.T) {
	mockSvc := &mockService{
		getUserFunc: func(userID string, principal *auth.Principal, targetOrgID string) (*User, error) {
			if userID != "user-123" {
				t.Errorf("Expected userID user-123, got %s", userID)
			}
			return &User{
				ID:        "user-123",
				Email:     "test@example.com",
				FirstName: "Test",
				LastName:  "User",
				Role:      "CAREGIVER",
				IsActive:  true,
				OrgID:     "org-123",
			}, nil
		},
	}

	handler := NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/users/user-123", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "user-123"})
	principal := &auth.Principal{UserID: "admin-123", Roles: []string{"ORG_ADMIN"}, OrgID: "org-123"}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.GetUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var user User
	if err := json.NewDecoder(rr.Body).Decode(&user); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if user.ID != "user-123" {
		t.Errorf("Expected user ID user-123, got %s", user.ID)
	}
}

func TestHandlerGetUser_NotFound(t *testing.T) {
	mockSvc := &mockService{
		getUserFunc: func(userID string, principal *auth.Principal, targetOrgID string) (*User, error) {
			return nil, ErrUserNotFound
		},
	}

	handler := NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/users/user-999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "user-999"})
	principal := &auth.Principal{UserID: "admin-123", Roles: []string{"ORG_ADMIN"}, OrgID: "org-123"}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.GetUser(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

func TestHandlerGetUser_Forbidden(t *testing.T) {
	mockSvc := &mockService{
		getUserFunc: func(userID string, principal *auth.Principal, targetOrgID string) (*User, error) {
			return nil, ErrForbidden
		},
	}

	handler := NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/users/user-123", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "user-123"})
	req.Header.Set("X-Organization-ID", "other-org")
	principal := &auth.Principal{UserID: "admin-123", Roles: []string{"ORG_ADMIN"}, OrgID: "org-123"}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.GetUser(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rr.Code)
	}
}

// Test UpdateUser Handler

func TestHandlerUpdateUser_Success(t *testing.T) {
	mockSvc := &mockService{
		updateUserFunc: func(userID string, req UpdateUserRequest, principal *auth.Principal, targetOrgID string) (*User, error) {
			return &User{
				ID:        userID,
				Email:     req.Email,
				FirstName: req.FirstName,
				LastName:  req.LastName,
				Role:      "CAREGIVER",
				IsActive:  true,
				OrgID:     "org-123",
			}, nil
		},
	}

	handler := NewHandler(mockSvc)

	reqBody := UpdateUserRequest{
		Email:     "updated@example.com",
		FirstName: "Updated",
		LastName:  "User",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/users/user-123", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "user-123"})
	principal := &auth.Principal{UserID: "admin-123", Roles: []string{"ORG_ADMIN"}, OrgID: "org-123"}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.UpdateUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var user User
	if err := json.NewDecoder(rr.Body).Decode(&user); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if user.Email != "updated@example.com" {
		t.Errorf("Expected email updated@example.com, got %s", user.Email)
	}
}

func TestHandlerUpdateUser_InvalidJSON(t *testing.T) {
	handler := NewHandler(&mockService{})

	req := httptest.NewRequest(http.MethodPut, "/users/user-123", bytes.NewReader([]byte("invalid")))
	req = mux.SetURLVars(req, map[string]string{"id": "user-123"})
	principal := &auth.Principal{UserID: "admin-123", Roles: []string{"ORG_ADMIN"}, OrgID: "org-123"}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.UpdateUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandlerUpdateUser_Forbidden(t *testing.T) {
	mockSvc := &mockService{
		updateUserFunc: func(userID string, req UpdateUserRequest, principal *auth.Principal, targetOrgID string) (*User, error) {
			return nil, ErrForbidden
		},
	}

	handler := NewHandler(mockSvc)

	reqBody := UpdateUserRequest{Email: "test@example.com"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/users/user-123", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "user-123"})
	principal := &auth.Principal{UserID: "admin-123", Roles: []string{"ORG_ADMIN"}, OrgID: "org-123"}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.UpdateUser(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rr.Code)
	}
}

// Test UpdateMyProfile Handler

func TestHandlerUpdateMyProfile_Success(t *testing.T) {
	mockSvc := &mockService{
		updateMyProfileFunc: func(req UpdateUserRequest, principal *auth.Principal) (*User, error) {
			return &User{
				ID:        principal.UserID,
				Email:     req.Email,
				FirstName: req.FirstName,
				LastName:  req.LastName,
				Role:      "CAREGIVER",
				IsActive:  true,
			}, nil
		},
	}

	handler := NewHandler(mockSvc)

	reqBody := UpdateUserRequest{
		Email:     "myemail@example.com",
		FirstName: "My",
		LastName:  "Name",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/users/me", bytes.NewReader(body))
	principal := &auth.Principal{UserID: "user-123", Roles: []string{"CAREGIVER"}, OrgID: "org-123"}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.UpdateMyProfile(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var user User
	if err := json.NewDecoder(rr.Body).Decode(&user); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if user.Email != "myemail@example.com" {
		t.Errorf("Expected email myemail@example.com, got %s", user.Email)
	}
}

func TestHandlerUpdateMyProfile_Unauthenticated(t *testing.T) {
	handler := NewHandler(&mockService{})

	reqBody := UpdateUserRequest{Email: "test@example.com"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/users/me", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.UpdateMyProfile(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}
}

// Test ResetPassword Handler

func TestHandlerResetPassword_Success(t *testing.T) {
	mockSvc := &mockService{
		resetPasswordFunc: func(userID string, req ResetPasswordRequest, principal *auth.Principal, targetOrgID string) error {
			return nil
		},
	}

	handler := NewHandler(mockSvc)

	reqBody := ResetPasswordRequest{
		TemporaryPassword: "newtemp123",
		SendEmail:         true,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/users/user-123/reset-password", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "user-123"})
	principal := &auth.Principal{UserID: "admin-123", Roles: []string{"ORG_ADMIN"}, OrgID: "org-123"}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ResetPassword(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["message"] != "password reset successfully" {
		t.Errorf("Expected success message, got %s", response["message"])
	}
}

func TestHandlerResetPassword_NotFound(t *testing.T) {
	mockSvc := &mockService{
		resetPasswordFunc: func(userID string, req ResetPasswordRequest, principal *auth.Principal, targetOrgID string) error {
			return ErrUserNotFound
		},
	}

	handler := NewHandler(mockSvc)

	reqBody := ResetPasswordRequest{TemporaryPassword: "newtemp123"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/users/user-999/reset-password", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "user-999"})
	principal := &auth.Principal{UserID: "admin-123", Roles: []string{"ORG_ADMIN"}, OrgID: "org-123"}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ResetPassword(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

// Test DeleteUser Handler

func TestHandlerDeleteUser_Success(t *testing.T) {
	mockSvc := &mockService{
		deleteUserFunc: func(userID string, principal *auth.Principal) error {
			return nil
		},
	}

	handler := NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodDelete, "/users/user-123", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "user-123"})
	principal := &auth.Principal{UserID: "admin-123", Roles: []string{"SUPER_ADMIN"}}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.DeleteUser(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", rr.Code)
	}
}

func TestHandlerDeleteUser_NotFound(t *testing.T) {
	mockSvc := &mockService{
		deleteUserFunc: func(userID string, principal *auth.Principal) error {
			return ErrUserNotFound
		},
	}

	handler := NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodDelete, "/users/user-999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "user-999"})
	principal := &auth.Principal{UserID: "admin-123", Roles: []string{"SUPER_ADMIN"}}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.DeleteUser(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

func TestHandlerDeleteUser_Forbidden(t *testing.T) {
	mockSvc := &mockService{
		deleteUserFunc: func(userID string, principal *auth.Principal) error {
			return ErrForbidden
		},
	}

	handler := NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodDelete, "/users/user-123", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "user-123"})
	principal := &auth.Principal{UserID: "admin-123", Roles: []string{"ORG_ADMIN"}, OrgID: "org-123"}
	ctx := auth.ContextWithPrincipal(req.Context(), principal)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.DeleteUser(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rr.Code)
	}
}
