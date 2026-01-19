package users

import (
	"errors"
	"testing"

	"github.com/WailSalutem-Health-Care/organization-service/internal/auth"
	"github.com/WailSalutem-Health-Care/organization-service/internal/pagination"
)

// TestCreateUser_Success tests successful user creation by SUPER_ADMIN
func TestCreateUser_Success(t *testing.T) {
	mockRepo := &mockRepository{
		getSchemaNameFunc: func(orgID string) (string, error) {
			return "org_test_12345678", nil
		},
		validateSchemaFunc: func(schemaName string) error {
			return nil
		},
		createFunc: func(user *User) error {
			user.ID = "user-123"
			return nil
		},
	}

	mockKeycloak := &mockKeycloakAdmin{
		createUserFunc: func(user auth.KeycloakUser) (string, error) {
			return "keycloak-user-123", nil
		},
		setPasswordFunc: func(userID, password string, temporary bool) error {
			return nil
		},
		getRoleFunc: func(roleName string) (*auth.KeycloakRole, error) {
			return &auth.KeycloakRole{ID: "role-id", Name: roleName}, nil
		},
		assignRoleFunc: func(userID string, role auth.KeycloakRole) error {
			return nil
		},
	}

	service := NewService(mockRepo, mockKeycloak)

	req := CreateUserRequest{
		Username:          "testuser",
		Email:             "test@example.com",
		FirstName:         "Test",
		LastName:          "User",
		Role:              "CAREGIVER",
		TemporaryPassword: "temp123",
	}

	principal := &auth.Principal{
		UserID: "admin-1",
		Roles:  []string{"SUPER_ADMIN"},
		OrgID:  "",
	}

	user, err := service.CreateUser(req, principal, "org-target-123")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if user == nil {
		t.Fatal("Expected user, got nil")
	}
	if user.KeycloakUserID != "keycloak-user-123" {
		t.Errorf("Expected KeycloakUserID 'keycloak-user-123', got '%s'", user.KeycloakUserID)
	}
	if user.Role != "CAREGIVER" {
		t.Errorf("Expected role 'CAREGIVER', got '%s'", user.Role)
	}
}

// TestCreateUser_OrgAdminSuccess tests ORG_ADMIN creating allowed role
func TestCreateUser_OrgAdminSuccess(t *testing.T) {
	mockRepo := &mockRepository{
		validateSchemaFunc: func(schemaName string) error {
			return nil
		},
		createFunc: func(user *User) error {
			user.ID = "user-456"
			return nil
		},
	}

	mockKeycloak := &mockKeycloakAdmin{
		createUserFunc: func(user auth.KeycloakUser) (string, error) {
			return "keycloak-user-456", nil
		},
		setPasswordFunc: func(userID, password string, temporary bool) error {
			return nil
		},
		getRoleFunc: func(roleName string) (*auth.KeycloakRole, error) {
			return &auth.KeycloakRole{ID: "role-id", Name: roleName}, nil
		},
		assignRoleFunc: func(userID string, role auth.KeycloakRole) error {
			return nil
		},
	}

	service := NewService(mockRepo, mockKeycloak)

	req := CreateUserRequest{
		Username:          "caregiver1",
		Email:             "caregiver@example.com",
		FirstName:         "Care",
		LastName:          "Giver",
		Role:              "CAREGIVER",
		TemporaryPassword: "temp123",
	}

	principal := &auth.Principal{
		UserID:        "admin-2",
		Roles:         []string{"ORG_ADMIN"},
		OrgID:         "org-2",
		OrgSchemaName: "org_myorg_87654321",
	}

	user, err := service.CreateUser(req, principal, "")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if user == nil {
		t.Fatal("Expected user, got nil")
	}
	if user.OrgID != "org-2" {
		t.Errorf("Expected OrgID 'org-2', got '%s'", user.OrgID)
	}
}

// TestCreateUser_OrgAdminForbiddenRole tests ORG_ADMIN cannot create SUPER_ADMIN
func TestCreateUser_OrgAdminForbiddenRole(t *testing.T) {
	mockRepo := &mockRepository{}
	mockKeycloak := &mockKeycloakAdmin{}

	service := NewService(mockRepo, mockKeycloak)

	req := CreateUserRequest{
		Username:          "baduser",
		Email:             "bad@example.com",
		FirstName:         "Bad",
		LastName:          "User",
		Role:              "SUPER_ADMIN", // ORG_ADMIN cannot create SUPER_ADMIN
		TemporaryPassword: "temp123",
	}

	principal := &auth.Principal{
		UserID:        "admin-3",
		Roles:         []string{"ORG_ADMIN"},
		OrgID:         "org-3",
		OrgSchemaName: "org_test_11111111",
	}

	user, err := service.CreateUser(req, principal, "")

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if user != nil {
		t.Error("Expected nil user")
	}
	if err != ErrRoleNotAllowed {
		t.Errorf("Expected ErrRoleNotAllowed, got: %v", err)
	}
}

// TestCreateUser_OrgAdminCrossOrgForbidden tests ORG_ADMIN cannot create in different org
func TestCreateUser_OrgAdminCrossOrgForbidden(t *testing.T) {
	mockRepo := &mockRepository{}
	mockKeycloak := &mockKeycloakAdmin{}

	service := NewService(mockRepo, mockKeycloak)

	req := CreateUserRequest{
		Username:          "user",
		Email:             "user@example.com",
		FirstName:         "Test",
		LastName:          "User",
		Role:              "CAREGIVER",
		TemporaryPassword: "temp123",
	}

	principal := &auth.Principal{
		UserID:        "admin-4",
		Roles:         []string{"ORG_ADMIN"},
		OrgID:         "org-4",
		OrgSchemaName: "org_test_22222222",
	}

	// Trying to create in different org
	user, err := service.CreateUser(req, principal, "org-5")

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if user != nil {
		t.Error("Expected nil user")
	}
	if err != ErrForbidden {
		t.Errorf("Expected ErrForbidden, got: %v", err)
	}
}

// TestCreateUser_SuperAdminMissingOrgID tests SUPER_ADMIN must provide X-Organization-ID
func TestCreateUser_SuperAdminMissingOrgID(t *testing.T) {
	mockRepo := &mockRepository{}
	mockKeycloak := &mockKeycloakAdmin{}

	service := NewService(mockRepo, mockKeycloak)

	req := CreateUserRequest{
		Username:          "user",
		Email:             "user@example.com",
		FirstName:         "Test",
		LastName:          "User",
		Role:              "CAREGIVER",
		TemporaryPassword: "temp123",
	}

	principal := &auth.Principal{
		UserID: "admin-5",
		Roles:  []string{"SUPER_ADMIN"},
		OrgID:  "",
	}

	// SUPER_ADMIN without target org ID
	user, err := service.CreateUser(req, principal, "")

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if user != nil {
		t.Error("Expected nil user")
	}
}

// TestCreateUser_ValidationError tests validation of required fields
func TestCreateUser_ValidationError(t *testing.T) {
	mockRepo := &mockRepository{}
	mockKeycloak := &mockKeycloakAdmin{}

	service := NewService(mockRepo, mockKeycloak)

	testCases := []struct {
		name string
		req  CreateUserRequest
	}{
		{
			name: "Missing username",
			req: CreateUserRequest{
				Email:             "test@example.com",
				FirstName:         "Test",
				LastName:          "User",
				Role:              "CAREGIVER",
				TemporaryPassword: "temp123",
			},
		},
		{
			name: "Missing email",
			req: CreateUserRequest{
				Username:          "testuser",
				FirstName:         "Test",
				LastName:          "User",
				Role:              "CAREGIVER",
				TemporaryPassword: "temp123",
			},
		},
		{
			name: "Missing first name",
			req: CreateUserRequest{
				Username:          "testuser",
				Email:             "test@example.com",
				LastName:          "User",
				Role:              "CAREGIVER",
				TemporaryPassword: "temp123",
			},
		},
		{
			name: "Missing password and email flag",
			req: CreateUserRequest{
				Username:  "testuser",
				Email:     "test@example.com",
				FirstName: "Test",
				LastName:  "User",
				Role:      "CAREGIVER",
			},
		},
	}

	principal := &auth.Principal{
		UserID: "admin-6",
		Roles:  []string{"SUPER_ADMIN"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user, err := service.CreateUser(tc.req, principal, "org-123")

			if err == nil {
				t.Error("Expected validation error, got nil")
			}
			if user != nil {
				t.Error("Expected nil user")
			}
		})
	}
}

// TestCreateUser_PatientRoleRejected tests PATIENT role cannot be created via users endpoint
func TestCreateUser_PatientRoleRejected(t *testing.T) {
	mockRepo := &mockRepository{
		getSchemaNameFunc: func(orgID string) (string, error) {
			return "org_test_12345678", nil
		},
		validateSchemaFunc: func(schemaName string) error {
			return nil
		},
	}

	mockKeycloak := &mockKeycloakAdmin{
		createUserFunc: func(user auth.KeycloakUser) (string, error) {
			return "keycloak-123", nil
		},
		setPasswordFunc: func(userID, password string, temporary bool) error {
			return nil
		},
		getRoleFunc: func(roleName string) (*auth.KeycloakRole, error) {
			return &auth.KeycloakRole{ID: "role-id", Name: roleName}, nil
		},
		assignRoleFunc: func(userID string, role auth.KeycloakRole) error {
			return nil
		},
		deleteUserFunc: func(userID string) error {
			return nil
		},
	}

	service := NewService(mockRepo, mockKeycloak)

	req := CreateUserRequest{
		Username:          "patient1",
		Email:             "patient@example.com",
		FirstName:         "Patient",
		LastName:          "User",
		Role:              "PATIENT",
		TemporaryPassword: "temp123",
	}

	principal := &auth.Principal{
		UserID: "admin-7",
		Roles:  []string{"SUPER_ADMIN"},
	}

	user, err := service.CreateUser(req, principal, "org-123")

	if err == nil {
		t.Error("Expected error for PATIENT role, got nil")
	}
	if user != nil {
		t.Error("Expected nil user")
	}
}

// TestCreateUser_KeycloakRollback tests rollback when database creation fails
func TestCreateUser_KeycloakRollback(t *testing.T) {
	keycloakDeleteCalled := false

	mockRepo := &mockRepository{
		getSchemaNameFunc: func(orgID string) (string, error) {
			return "org_test_12345678", nil
		},
		validateSchemaFunc: func(schemaName string) error {
			return nil
		},
		createFunc: func(user *User) error {
			return errors.New("database error")
		},
	}

	mockKeycloak := &mockKeycloakAdmin{
		createUserFunc: func(user auth.KeycloakUser) (string, error) {
			return "keycloak-789", nil
		},
		setPasswordFunc: func(userID, password string, temporary bool) error {
			return nil
		},
		getRoleFunc: func(roleName string) (*auth.KeycloakRole, error) {
			return &auth.KeycloakRole{ID: "role-id", Name: roleName}, nil
		},
		assignRoleFunc: func(userID string, role auth.KeycloakRole) error {
			return nil
		},
		deleteUserFunc: func(userID string) error {
			keycloakDeleteCalled = true
			return nil
		},
	}

	service := NewService(mockRepo, mockKeycloak)

	req := CreateUserRequest{
		Username:          "testuser",
		Email:             "test@example.com",
		FirstName:         "Test",
		LastName:          "User",
		Role:              "CAREGIVER",
		TemporaryPassword: "temp123",
	}

	principal := &auth.Principal{
		UserID: "admin-8",
		Roles:  []string{"SUPER_ADMIN"},
	}

	user, err := service.CreateUser(req, principal, "org-123")

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if user != nil {
		t.Error("Expected nil user")
	}
	if !keycloakDeleteCalled {
		t.Error("Expected Keycloak user to be deleted on rollback")
	}
}

// TestCreateUser_SendResetEmail tests email flow instead of temporary password
func TestCreateUser_SendResetEmail(t *testing.T) {
	emailSent := false

	mockRepo := &mockRepository{
		getSchemaNameFunc: func(orgID string) (string, error) {
			return "org_test_12345678", nil
		},
		validateSchemaFunc: func(schemaName string) error {
			return nil
		},
		createFunc: func(user *User) error {
			user.ID = "user-999"
			return nil
		},
	}

	mockKeycloak := &mockKeycloakAdmin{
		createUserFunc: func(user auth.KeycloakUser) (string, error) {
			return "keycloak-999", nil
		},
		sendEmailActionFunc: func(userID string, actions []string) error {
			emailSent = true
			return nil
		},
		getRoleFunc: func(roleName string) (*auth.KeycloakRole, error) {
			return &auth.KeycloakRole{ID: "role-id", Name: roleName}, nil
		},
		assignRoleFunc: func(userID string, role auth.KeycloakRole) error {
			return nil
		},
	}

	service := NewService(mockRepo, mockKeycloak)

	req := CreateUserRequest{
		Username:       "testuser",
		Email:          "test@example.com",
		FirstName:      "Test",
		LastName:       "User",
		Role:           "CAREGIVER",
		SendResetEmail: true,
	}

	principal := &auth.Principal{
		UserID: "admin-9",
		Roles:  []string{"SUPER_ADMIN"},
	}

	user, err := service.CreateUser(req, principal, "org-123")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if user == nil {
		t.Fatal("Expected user, got nil")
	}
	if !emailSent {
		t.Error("Expected email to be sent")
	}
}

// TestGetUser_Success tests successful user retrieval
func TestGetUser_Success(t *testing.T) {
	mockRepo := &mockRepository{
		getSchemaNameFunc: func(orgID string) (string, error) {
			return "org_test_12345678", nil
		},
		getByIDFunc: func(schemaName, userID string) (*User, error) {
			return &User{
				ID:             userID,
				Email:          "user@example.com",
				FirstName:      "Test",
				LastName:       "User",
				Role:           "CAREGIVER",
				OrgID:          "org-123",
				OrgSchemaName:  schemaName,
				KeycloakUserID: "keycloak-123",
			}, nil
		},
	}

	mockKeycloak := &mockKeycloakAdmin{}
	service := NewService(mockRepo, mockKeycloak)

	principal := &auth.Principal{
		UserID: "admin-10",
		Roles:  []string{"SUPER_ADMIN"},
	}

	user, err := service.GetUser("user-123", principal, "org-123")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if user == nil {
		t.Fatal("Expected user, got nil")
	}
	if user.ID != "user-123" {
		t.Errorf("Expected user ID 'user-123', got '%s'", user.ID)
	}
}

// TestGetUser_OrgAdminCrossOrgForbidden tests ORG_ADMIN cannot view users from other orgs
func TestGetUser_OrgAdminCrossOrgForbidden(t *testing.T) {
	mockRepo := &mockRepository{}
	mockKeycloak := &mockKeycloakAdmin{}

	service := NewService(mockRepo, mockKeycloak)

	principal := &auth.Principal{
		UserID: "admin-11",
		Roles:  []string{"ORG_ADMIN"},
		OrgID:  "org-1",
	}

	// Trying to get user from different org
	user, err := service.GetUser("user-123", principal, "org-2")

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if user != nil {
		t.Error("Expected nil user")
	}
	if err != ErrForbidden {
		t.Errorf("Expected ErrForbidden, got: %v", err)
	}
}

// TestListUsers_Success tests listing users
func TestListUsers_Success(t *testing.T) {
	mockRepo := &mockRepository{
		getSchemaNameFunc: func(orgID string) (string, error) {
			return "org_test_12345678", nil
		},
		listFunc: func(schemaName string) ([]User, error) {
			return []User{
				{ID: "user-1", Email: "user1@example.com", Role: "CAREGIVER"},
				{ID: "user-2", Email: "user2@example.com", Role: "MUNICIPALITY"},
			}, nil
		},
	}

	mockKeycloak := &mockKeycloakAdmin{}
	service := NewService(mockRepo, mockKeycloak)

	principal := &auth.Principal{
		UserID: "admin-12",
		Roles:  []string{"SUPER_ADMIN"},
	}

	users, err := service.ListUsers(principal, "org-123")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

// TestListUsersWithPagination_Success tests pagination
func TestListUsersWithPagination_Success(t *testing.T) {
	mockRepo := &mockRepository{
		getSchemaNameFunc: func(orgID string) (string, error) {
			return "org_test_12345678", nil
		},
		listWithPaginationFunc: func(schemaName string, limit, offset int, search string) ([]User, int, error) {
			users := make([]User, limit)
			for i := 0; i < limit; i++ {
				users[i] = User{
					ID:    string(rune('0' + i)),
					Email: "user@example.com",
				}
			}
			return users, 25, nil
		},
	}

	mockKeycloak := &mockKeycloakAdmin{}
	service := NewService(mockRepo, mockKeycloak)

	principal := &auth.Principal{
		UserID: "admin-13",
		Roles:  []string{"SUPER_ADMIN"},
	}

	params := pagination.Params{
		Page:  1,
		Limit: 10,
	}

	response, err := service.ListUsersWithPagination(principal, "org-123", params)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(response.Users) != 10 {
		t.Errorf("Expected 10 users, got %d", len(response.Users))
	}
	if response.Pagination.TotalRecords != 25 {
		t.Errorf("Expected total 25, got %d", response.Pagination.TotalRecords)
	}
}

// TestListActiveUsersByRoleWithPagination_Success tests role filtering
func TestListActiveUsersByRoleWithPagination_Success(t *testing.T) {
	mockRepo := &mockRepository{
		getSchemaNameFunc: func(orgID string) (string, error) {
			return "org_test_12345678", nil
		},
		listActiveByRoleFunc: func(schemaName string, role string, limit, offset int, search string) ([]User, int, error) {
			return []User{
				{ID: "user-1", Role: role, IsActive: true},
				{ID: "user-2", Role: role, IsActive: true},
			}, 2, nil
		},
	}

	mockKeycloak := &mockKeycloakAdmin{}
	service := NewService(mockRepo, mockKeycloak)

	principal := &auth.Principal{
		UserID: "admin-14",
		Roles:  []string{"SUPER_ADMIN"},
	}

	params := pagination.Params{
		Page:  1,
		Limit: 10,
	}

	response, err := service.ListActiveUsersByRoleWithPagination(principal, "org-123", "CAREGIVER", params)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(response.Users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(response.Users))
	}
	if response.Users[0].Role != "CAREGIVER" {
		t.Errorf("Expected role 'CAREGIVER', got '%s'", response.Users[0].Role)
	}
}

// TestUpdateUser_Success tests successful user update
func TestUpdateUser_Success(t *testing.T) {
	mockRepo := &mockRepository{
		getSchemaNameFunc: func(orgID string) (string, error) {
			return "org_test_12345678", nil
		},
		getByIDFunc: func(schemaName, userID string) (*User, error) {
			return &User{
				ID:             userID,
				Email:          "old@example.com",
				FirstName:      "Old",
				LastName:       "Name",
				KeycloakUserID: "keycloak-123",
				OrgSchemaName:  schemaName,
			}, nil
		},
		updateFunc: func(user *User) error {
			return nil
		},
	}

	mockKeycloak := &mockKeycloakAdmin{
		getUserFunc: func(userID string) (*auth.KeycloakUser, error) {
			return &auth.KeycloakUser{
				ID:       userID,
				Username: "testuser",
			}, nil
		},
		updateUserFunc: func(userID string, user auth.KeycloakUser) error {
			return nil
		},
	}

	service := NewService(mockRepo, mockKeycloak)

	req := UpdateUserRequest{
		Email:     "new@example.com",
		FirstName: "New",
		LastName:  "Name",
	}

	principal := &auth.Principal{
		UserID: "admin-15",
		Roles:  []string{"SUPER_ADMIN"},
	}

	user, err := service.UpdateUser("user-123", req, principal, "org-123")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if user.Email != "new@example.com" {
		t.Errorf("Expected email 'new@example.com', got '%s'", user.Email)
	}
}

// TestUpdateMyProfile_Success tests user updating their own profile
func TestUpdateMyProfile_Success(t *testing.T) {
	mockRepo := &mockRepository{
		getSchemaNameFunc: func(orgID string) (string, error) {
			return "org_test_12345678", nil
		},
		getByKeycloakIDFunc: func(schemaName, keycloakID string) (*User, error) {
			return &User{
				ID:             "user-123",
				Email:          "old@example.com",
				FirstName:      "Old",
				LastName:       "Name",
				KeycloakUserID: keycloakID,
				OrgSchemaName:  schemaName,
			}, nil
		},
		updateFunc: func(user *User) error {
			return nil
		},
	}

	mockKeycloak := &mockKeycloakAdmin{
		getUserFunc: func(userID string) (*auth.KeycloakUser, error) {
			return &auth.KeycloakUser{
				ID:       userID,
				Username: "testuser",
			}, nil
		},
		updateUserFunc: func(userID string, user auth.KeycloakUser) error {
			return nil
		},
	}

	service := NewService(mockRepo, mockKeycloak)

	req := UpdateUserRequest{
		Email:     "mynewemail@example.com",
		FirstName: "Updated",
	}

	principal := &auth.Principal{
		UserID: "keycloak-my-id",
		Roles:  []string{"CAREGIVER"},
		OrgID:  "org-123",
	}

	user, err := service.UpdateMyProfile(req, principal)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if user.Email != "mynewemail@example.com" {
		t.Errorf("Expected email 'mynewemail@example.com', got '%s'", user.Email)
	}
}

// TestResetPassword_Success tests password reset
func TestResetPassword_Success(t *testing.T) {
	mockRepo := &mockRepository{
		getSchemaNameFunc: func(orgID string) (string, error) {
			return "org_test_12345678", nil
		},
		getByIDFunc: func(schemaName, userID string) (*User, error) {
			return &User{
				ID:             userID,
				KeycloakUserID: "keycloak-123",
			}, nil
		},
	}

	mockKeycloak := &mockKeycloakAdmin{
		setPasswordFunc: func(userID, password string, temporary bool) error {
			return nil
		},
	}

	service := NewService(mockRepo, mockKeycloak)

	req := ResetPasswordRequest{
		TemporaryPassword: "newtemp123",
	}

	principal := &auth.Principal{
		UserID: "admin-16",
		Roles:  []string{"SUPER_ADMIN"},
	}

	err := service.ResetPassword("user-123", req, principal, "org-123")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// TestDeleteUser_Success tests successful user deletion
func TestDeleteUser_Success(t *testing.T) {
	mockRepo := &mockRepository{
		getSchemaNameFunc: func(orgID string) (string, error) {
			return "org_test_12345678", nil
		},
		getByIDFunc: func(schemaName, userID string) (*User, error) {
			return &User{
				ID:             userID,
				KeycloakUserID: "keycloak-123",
				OrgID:          "org-123",
				Role:           "CAREGIVER",
				OrgSchemaName:  schemaName,
			}, nil
		},
		deleteFunc: func(schemaName, orgID, userID, role string) error {
			return nil
		},
	}

	mockKeycloak := &mockKeycloakAdmin{
		deleteUserFunc: func(userID string) error {
			return nil
		},
	}

	service := NewService(mockRepo, mockKeycloak)

	principal := &auth.Principal{
		UserID:        "admin-17",
		Roles:         []string{"ORG_ADMIN"},
		OrgID:         "org-123",
		OrgSchemaName: "org_test_12345678",
	}

	err := service.DeleteUser("user-123", principal)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// Mock implementations

type mockRepository struct {
	getSchemaNameFunc      func(orgID string) (string, error)
	validateSchemaFunc     func(schemaName string) error
	createFunc             func(user *User) error
	getByIDFunc            func(schemaName, userID string) (*User, error)
	getByKeycloakIDFunc    func(schemaName, keycloakID string) (*User, error)
	listFunc               func(schemaName string) ([]User, error)
	listWithPaginationFunc func(schemaName string, limit, offset int, search string) ([]User, int, error)
	listActiveByRoleFunc   func(schemaName string, role string, limit, offset int, search string) ([]User, int, error)
	updateFunc             func(user *User) error
	deleteFunc             func(schemaName, orgID, userID, role string) error
}

func (m *mockRepository) GetSchemaNameByOrgID(orgID string) (string, error) {
	if m.getSchemaNameFunc != nil {
		return m.getSchemaNameFunc(orgID)
	}
	return "", errors.New("not implemented")
}

func (m *mockRepository) ValidateOrgSchema(schemaName string) error {
	if m.validateSchemaFunc != nil {
		return m.validateSchemaFunc(schemaName)
	}
	return errors.New("not implemented")
}

func (m *mockRepository) Create(user *User) error {
	if m.createFunc != nil {
		return m.createFunc(user)
	}
	return errors.New("not implemented")
}

func (m *mockRepository) GetByID(schemaName, userID string) (*User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(schemaName, userID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRepository) GetByKeycloakID(schemaName, keycloakUserID string) (*User, error) {
	if m.getByKeycloakIDFunc != nil {
		return m.getByKeycloakIDFunc(schemaName, keycloakUserID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRepository) List(schemaName string) ([]User, error) {
	if m.listFunc != nil {
		return m.listFunc(schemaName)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRepository) ListWithPagination(schemaName string, limit, offset int, search string) ([]User, int, error) {
	if m.listWithPaginationFunc != nil {
		return m.listWithPaginationFunc(schemaName, limit, offset, search)
	}
	return nil, 0, errors.New("not implemented")
}

func (m *mockRepository) ListActiveUsersByRoleWithPagination(schemaName string, role string, limit, offset int, search string) ([]User, int, error) {
	if m.listActiveByRoleFunc != nil {
		return m.listActiveByRoleFunc(schemaName, role, limit, offset, search)
	}
	return nil, 0, errors.New("not implemented")
}

func (m *mockRepository) Update(user *User) error {
	if m.updateFunc != nil {
		return m.updateFunc(user)
	}
	return errors.New("not implemented")
}

func (m *mockRepository) Delete(schemaName, orgID, userID string, role string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(schemaName, orgID, userID, role)
	}
	return errors.New("not implemented")
}

type mockKeycloakAdmin struct {
	createUserFunc      func(user auth.KeycloakUser) (string, error)
	setPasswordFunc     func(userID, password string, temporary bool) error
	sendEmailActionFunc func(userID string, actions []string) error
	getRoleFunc         func(roleName string) (*auth.KeycloakRole, error)
	assignRoleFunc      func(userID string, role auth.KeycloakRole) error
	deleteUserFunc      func(userID string) error
	updateUserFunc      func(userID string, user auth.KeycloakUser) error
	getUserFunc         func(userID string) (*auth.KeycloakUser, error)
}

func (m *mockKeycloakAdmin) CreateUser(user auth.KeycloakUser) (string, error) {
	if m.createUserFunc != nil {
		return m.createUserFunc(user)
	}
	return "", errors.New("not implemented")
}

func (m *mockKeycloakAdmin) SetPassword(userID, password string, temporary bool) error {
	if m.setPasswordFunc != nil {
		return m.setPasswordFunc(userID, password, temporary)
	}
	return errors.New("not implemented")
}

func (m *mockKeycloakAdmin) SendEmailAction(userID string, actions []string) error {
	if m.sendEmailActionFunc != nil {
		return m.sendEmailActionFunc(userID, actions)
	}
	return errors.New("not implemented")
}

func (m *mockKeycloakAdmin) GetRole(roleName string) (*auth.KeycloakRole, error) {
	if m.getRoleFunc != nil {
		return m.getRoleFunc(roleName)
	}
	return nil, errors.New("not implemented")
}

func (m *mockKeycloakAdmin) AssignRole(userID string, role auth.KeycloakRole) error {
	if m.assignRoleFunc != nil {
		return m.assignRoleFunc(userID, role)
	}
	return errors.New("not implemented")
}

func (m *mockKeycloakAdmin) DeleteUser(userID string) error {
	if m.deleteUserFunc != nil {
		return m.deleteUserFunc(userID)
	}
	return errors.New("not implemented")
}

func (m *mockKeycloakAdmin) UpdateUser(userID string, user auth.KeycloakUser) error {
	if m.updateUserFunc != nil {
		return m.updateUserFunc(userID, user)
	}
	return errors.New("not implemented")
}

func (m *mockKeycloakAdmin) GetUser(userID string) (*auth.KeycloakUser, error) {
	if m.getUserFunc != nil {
		return m.getUserFunc(userID)
	}
	return nil, errors.New("not implemented")
}
