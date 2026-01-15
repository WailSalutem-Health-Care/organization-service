package patient

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/WailSalutem-Health-Care/organization-service/internal/auth"
	"github.com/WailSalutem-Health-Care/organization-service/internal/pagination"
)

// TestCreatePatient_Success tests successful patient creation
func TestCreatePatient_Success(t *testing.T) {
	mockRepo := &mockRepository{
		createPatientFunc: func(ctx context.Context, schemaName, orgID, keycloakUserID string, req CreatePatientRequest) (*PatientResponse, error) {
			dob := "1980-01-01"
			return &PatientResponse{
				ID:             "patient-123",
				PatientID:      "PT-0001",
				KeycloakUserID: keycloakUserID,
				FirstName:      req.FirstName,
				LastName:       req.LastName,
				Email:          req.Email,
				PhoneNumber:    req.PhoneNumber,
				DateOfBirth:    &dob,
				Address:        req.Address,
				IsActive:       true,
				CreatedAt:      time.Now(),
			}, nil
		},
	}

	mockKeycloak := &mockKeycloakAdmin{
		createUserFunc: func(user auth.KeycloakUser) (string, error) {
			return "keycloak-patient-123", nil
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

	req := CreatePatientRequest{
		Username:          "patient1",
		Email:             "patient@example.com",
		FirstName:         "John",
		LastName:          "Doe",
		DateOfBirth:       "1980-01-01",
		Address:           "123 Main St",
		TemporaryPassword: "temp123",
	}

	patient, err := service.CreatePatient(context.Background(), "org_test_12345678", "org-123", req)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if patient == nil {
		t.Fatal("Expected patient, got nil")
	}
	if patient.PatientID != "PT-0001" {
		t.Errorf("Expected PatientID 'PT-0001', got '%s'", patient.PatientID)
	}
	if patient.KeycloakUserID != "keycloak-patient-123" {
		t.Errorf("Expected KeycloakUserID 'keycloak-patient-123', got '%s'", patient.KeycloakUserID)
	}
}

// TestCreatePatient_ValidationError tests validation of required fields
func TestCreatePatient_ValidationError(t *testing.T) {
	mockRepo := &mockRepository{}
	mockKeycloak := &mockKeycloakAdmin{}

	service := NewService(mockRepo, mockKeycloak)

	testCases := []struct {
		name string
		req  CreatePatientRequest
	}{
		{
			name: "Missing username",
			req: CreatePatientRequest{
				Email:             "patient@example.com",
				FirstName:         "John",
				LastName:          "Doe",
				DateOfBirth:       "1980-01-01",
				Address:           "123 Main St",
				TemporaryPassword: "temp123",
			},
		},
		{
			name: "Missing email",
			req: CreatePatientRequest{
				Username:          "patient1",
				FirstName:         "John",
				LastName:          "Doe",
				DateOfBirth:       "1980-01-01",
				Address:           "123 Main St",
				TemporaryPassword: "temp123",
			},
		},
		{
			name: "Missing first name",
			req: CreatePatientRequest{
				Username:          "patient1",
				Email:             "patient@example.com",
				LastName:          "Doe",
				DateOfBirth:       "1980-01-01",
				Address:           "123 Main St",
				TemporaryPassword: "temp123",
			},
		},
		{
			name: "Missing last name",
			req: CreatePatientRequest{
				Username:          "patient1",
				Email:             "patient@example.com",
				FirstName:         "John",
				DateOfBirth:       "1980-01-01",
				Address:           "123 Main St",
				TemporaryPassword: "temp123",
			},
		},
		{
			name: "Missing date of birth",
			req: CreatePatientRequest{
				Username:          "patient1",
				Email:             "patient@example.com",
				FirstName:         "John",
				LastName:          "Doe",
				Address:           "123 Main St",
				TemporaryPassword: "temp123",
			},
		},
		{
			name: "Missing address",
			req: CreatePatientRequest{
				Username:          "patient1",
				Email:             "patient@example.com",
				FirstName:         "John",
				LastName:          "Doe",
				DateOfBirth:       "1980-01-01",
				TemporaryPassword: "temp123",
			},
		},
		{
			name: "Missing password and email flag",
			req: CreatePatientRequest{
				Username:    "patient1",
				Email:       "patient@example.com",
				FirstName:   "John",
				LastName:    "Doe",
				DateOfBirth: "1980-01-01",
				Address:     "123 Main St",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			patient, err := service.CreatePatient(context.Background(), "org_test_12345678", "org-123", tc.req)

			if err == nil {
				t.Error("Expected validation error, got nil")
			}
			if patient != nil {
				t.Error("Expected nil patient")
			}
		})
	}
}

// TestCreatePatient_SendResetEmail tests email flow instead of temporary password
func TestCreatePatient_SendResetEmail(t *testing.T) {
	emailSent := false

	mockRepo := &mockRepository{
		createPatientFunc: func(ctx context.Context, schemaName, orgID, keycloakUserID string, req CreatePatientRequest) (*PatientResponse, error) {
			return &PatientResponse{
				ID:             "patient-456",
				KeycloakUserID: keycloakUserID,
				FirstName:      req.FirstName,
				IsActive:       true,
			}, nil
		},
	}

	mockKeycloak := &mockKeycloakAdmin{
		createUserFunc: func(user auth.KeycloakUser) (string, error) {
			return "keycloak-456", nil
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

	req := CreatePatientRequest{
		Username:       "patient2",
		Email:          "patient2@example.com",
		FirstName:      "Jane",
		LastName:       "Doe",
		DateOfBirth:    "1985-05-05",
		Address:        "456 Oak Ave",
		SendResetEmail: true,
	}

	patient, err := service.CreatePatient(context.Background(), "org_test_12345678", "org-123", req)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if patient == nil {
		t.Fatal("Expected patient, got nil")
	}
	if !emailSent {
		t.Error("Expected email to be sent")
	}
}

// TestCreatePatient_KeycloakRollback tests rollback when database creation fails
func TestCreatePatient_KeycloakRollback(t *testing.T) {
	keycloakDeleteCalled := false

	mockRepo := &mockRepository{
		createPatientFunc: func(ctx context.Context, schemaName, orgID, keycloakUserID string, req CreatePatientRequest) (*PatientResponse, error) {
			return nil, errors.New("database error")
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

	req := CreatePatientRequest{
		Username:          "patient3",
		Email:             "patient3@example.com",
		FirstName:         "Test",
		LastName:          "Patient",
		DateOfBirth:       "1990-01-01",
		Address:           "789 Test Rd",
		TemporaryPassword: "temp123",
	}

	patient, err := service.CreatePatient(context.Background(), "org_test_12345678", "org-123", req)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if patient != nil {
		t.Error("Expected nil patient")
	}
	if !keycloakDeleteCalled {
		t.Error("Expected Keycloak user to be deleted on rollback")
	}
}

// TestListPatients_Success tests listing all patients
func TestListPatients_Success(t *testing.T) {
	mockRepo := &mockRepository{
		listPatientsFunc: func(ctx context.Context, schemaName string) ([]PatientResponse, error) {
			return []PatientResponse{
				{ID: "patient-1", FirstName: "John", LastName: "Doe"},
				{ID: "patient-2", FirstName: "Jane", LastName: "Smith"},
			}, nil
		},
	}

	mockKeycloak := &mockKeycloakAdmin{}
	service := NewService(mockRepo, mockKeycloak)

	patients, err := service.ListPatients(context.Background(), "org_test_12345678")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(patients) != 2 {
		t.Errorf("Expected 2 patients, got %d", len(patients))
	}
}

// TestListPatientsWithPagination_Success tests pagination
func TestListPatientsWithPagination_Success(t *testing.T) {
	mockRepo := &mockRepository{
		listPatientsWithPaginationFunc: func(ctx context.Context, schemaName string, limit, offset int, search string) ([]PatientResponse, int, error) {
			patients := make([]PatientResponse, limit)
			for i := 0; i < limit; i++ {
				patients[i] = PatientResponse{
					ID:        string(rune('0' + i)),
					FirstName: "Patient",
				}
			}
			return patients, 25, nil
		},
	}

	mockKeycloak := &mockKeycloakAdmin{}
	service := NewService(mockRepo, mockKeycloak)

	params := pagination.Params{
		Page:  1,
		Limit: 10,
	}

	response, err := service.ListPatientsWithPagination(context.Background(), "org_test_12345678", params)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(response.Patients) != 10 {
		t.Errorf("Expected 10 patients, got %d", len(response.Patients))
	}
	if response.Pagination.TotalRecords != 25 {
		t.Errorf("Expected total 25, got %d", response.Pagination.TotalRecords)
	}
}

// TestListActivePatientsWithPagination_Success tests active patient filtering
func TestListActivePatientsWithPagination_Success(t *testing.T) {
	mockRepo := &mockRepository{
		listActivePatientsFunc: func(ctx context.Context, schemaName string, limit, offset int, search string) ([]PatientResponse, int, error) {
			return []PatientResponse{
				{ID: "patient-1", FirstName: "Active", IsActive: true},
				{ID: "patient-2", FirstName: "Patient", IsActive: true},
			}, 2, nil
		},
	}

	mockKeycloak := &mockKeycloakAdmin{}
	service := NewService(mockRepo, mockKeycloak)

	params := pagination.Params{
		Page:  1,
		Limit: 10,
	}

	response, err := service.ListActivePatientsWithPagination(context.Background(), "org_test_12345678", params)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(response.Patients) != 2 {
		t.Errorf("Expected 2 active patients, got %d", len(response.Patients))
	}
	for _, p := range response.Patients {
		if !p.IsActive {
			t.Error("Expected all patients to be active")
		}
	}
}

// TestGetPatient_Success tests retrieving a single patient
func TestGetPatient_Success(t *testing.T) {
	mockRepo := &mockRepository{
		getPatientFunc: func(ctx context.Context, schemaName, id string) (*PatientResponse, error) {
			return &PatientResponse{
				ID:        id,
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				IsActive:  true,
			}, nil
		},
	}

	mockKeycloak := &mockKeycloakAdmin{}
	service := NewService(mockRepo, mockKeycloak)

	patient, err := service.GetPatient(context.Background(), "org_test_12345678", "patient-123")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if patient == nil {
		t.Fatal("Expected patient, got nil")
	}
	if patient.ID != "patient-123" {
		t.Errorf("Expected patient ID 'patient-123', got '%s'", patient.ID)
	}
}

// TestUpdatePatient_Success tests successful patient update
func TestUpdatePatient_Success(t *testing.T) {
	newEmail := "newemail@example.com"
	newAddress := "789 New St"

	mockRepo := &mockRepository{
		updatePatientFunc: func(ctx context.Context, schemaName, id string, req UpdatePatientRequest) (*PatientResponse, error) {
			return &PatientResponse{
				ID:      id,
				Email:   *req.Email,
				Address: *req.Address,
			}, nil
		},
	}

	mockKeycloak := &mockKeycloakAdmin{}
	service := NewService(mockRepo, mockKeycloak)

	req := UpdatePatientRequest{
		Email:   &newEmail,
		Address: &newAddress,
	}

	patient, err := service.UpdatePatient(context.Background(), "org_test_12345678", "patient-123", req)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if patient.Email != newEmail {
		t.Errorf("Expected email '%s', got '%s'", newEmail, patient.Email)
	}
	if patient.Address != newAddress {
		t.Errorf("Expected address '%s', got '%s'", newAddress, patient.Address)
	}
}

// TestDeletePatient_Success tests successful patient deletion
func TestDeletePatient_Success(t *testing.T) {
	mockRepo := &mockRepository{
		deletePatientFunc: func(ctx context.Context, schemaName, orgID, id string) error {
			return nil
		},
	}

	mockKeycloak := &mockKeycloakAdmin{}
	service := NewService(mockRepo, mockKeycloak)

	err := service.DeletePatient(context.Background(), "org_test_12345678", "org-123", "patient-123")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// TestDeletePatient_NotFound tests deleting non-existent patient
func TestDeletePatient_NotFound(t *testing.T) {
	mockRepo := &mockRepository{
		deletePatientFunc: func(ctx context.Context, schemaName, orgID, id string) error {
			return errors.New("patient not found")
		},
	}

	mockKeycloak := &mockKeycloakAdmin{}
	service := NewService(mockRepo, mockKeycloak)

	err := service.DeletePatient(context.Background(), "org_test_12345678", "org-123", "nonexistent")

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

// Mock implementations

type mockRepository struct {
	createPatientFunc              func(ctx context.Context, schemaName, orgID, keycloakUserID string, req CreatePatientRequest) (*PatientResponse, error)
	listPatientsFunc               func(ctx context.Context, schemaName string) ([]PatientResponse, error)
	listPatientsWithPaginationFunc func(ctx context.Context, schemaName string, limit, offset int, search string) ([]PatientResponse, int, error)
	listActivePatientsFunc         func(ctx context.Context, schemaName string, limit, offset int, search string) ([]PatientResponse, int, error)
	getPatientFunc                 func(ctx context.Context, schemaName, id string) (*PatientResponse, error)
	updatePatientFunc              func(ctx context.Context, schemaName, id string, req UpdatePatientRequest) (*PatientResponse, error)
	deletePatientFunc              func(ctx context.Context, schemaName, orgID, id string) error
}

func (m *mockRepository) CreatePatient(ctx context.Context, schemaName, orgID, keycloakUserID string, req CreatePatientRequest) (*PatientResponse, error) {
	if m.createPatientFunc != nil {
		return m.createPatientFunc(ctx, schemaName, orgID, keycloakUserID, req)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRepository) ListPatients(ctx context.Context, schemaName string) ([]PatientResponse, error) {
	if m.listPatientsFunc != nil {
		return m.listPatientsFunc(ctx, schemaName)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRepository) ListPatientsWithPagination(ctx context.Context, schemaName string, limit, offset int, search string) ([]PatientResponse, int, error) {
	if m.listPatientsWithPaginationFunc != nil {
		return m.listPatientsWithPaginationFunc(ctx, schemaName, limit, offset, search)
	}
	return nil, 0, errors.New("not implemented")
}

func (m *mockRepository) ListActivePatientsWithPagination(ctx context.Context, schemaName string, limit, offset int, search string) ([]PatientResponse, int, error) {
	if m.listActivePatientsFunc != nil {
		return m.listActivePatientsFunc(ctx, schemaName, limit, offset, search)
	}
	return nil, 0, errors.New("not implemented")
}

func (m *mockRepository) GetPatient(ctx context.Context, schemaName, id string) (*PatientResponse, error) {
	if m.getPatientFunc != nil {
		return m.getPatientFunc(ctx, schemaName, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRepository) UpdatePatient(ctx context.Context, schemaName, id string, req UpdatePatientRequest) (*PatientResponse, error) {
	if m.updatePatientFunc != nil {
		return m.updatePatientFunc(ctx, schemaName, id, req)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRepository) DeletePatient(ctx context.Context, schemaName, orgID, id string) error {
	if m.deletePatientFunc != nil {
		return m.deletePatientFunc(ctx, schemaName, orgID, id)
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
