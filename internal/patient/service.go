package patient

import (
	"context"
	"fmt"
	"log"

	"github.com/WailSalutem-Health-Care/organization-service/internal/auth"
	"github.com/WailSalutem-Health-Care/organization-service/internal/pagination"
)

type Service struct {
	repo          *Repository
	keycloakAdmin *auth.KeycloakAdminClient
}

func NewService(repo *Repository, keycloakAdmin *auth.KeycloakAdminClient) *Service {
	return &Service{
		repo:          repo,
		keycloakAdmin: keycloakAdmin,
	}
}

func (s *Service) CreatePatient(ctx context.Context, schemaName string, orgID string, req CreatePatientRequest) (*PatientResponse, error) {
	// Validate required fields
	if req.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if req.Email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if req.FirstName == "" {
		return nil, fmt.Errorf("first name is required")
	}
	if req.LastName == "" {
		return nil, fmt.Errorf("last name is required")
	}
	if req.DateOfBirth == "" {
		return nil, fmt.Errorf("date of birth is required")
	}
	if req.Address == "" {
		return nil, fmt.Errorf("address is required")
	}
	if req.TemporaryPassword == "" && !req.SendResetEmail {
		return nil, fmt.Errorf("either temporaryPassword or sendResetEmail must be provided")
	}

	// Create user in Keycloak
	keycloakUser := auth.KeycloakUser{
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Enabled:   true,
		Attributes: map[string][]string{
			"organizationID": {orgID},
			"orgSchemaName":  {schemaName},
		},
	}

	log.Printf("Creating patient in Keycloak: %s", req.Username)
	keycloakUserID, err := s.keycloakAdmin.CreateUser(keycloakUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create user in Keycloak: %w", err)
	}
	log.Printf("Created patient in Keycloak with ID: %s", keycloakUserID)

	// Set password or send reset email
	if req.TemporaryPassword != "" {
		err = s.keycloakAdmin.SetPassword(keycloakUserID, req.TemporaryPassword, false)
		if err != nil {
			log.Printf("Failed to set password, rolling back: %s", keycloakUserID)
			_ = s.keycloakAdmin.DeleteUser(keycloakUserID)
			return nil, fmt.Errorf("failed to set password: %w", err)
		}
	} else if req.SendResetEmail {
		err = s.keycloakAdmin.SendEmailAction(keycloakUserID, []string{"UPDATE_PASSWORD"})
		if err != nil {
			log.Printf("Failed to send reset email, rolling back: %s", keycloakUserID)
			_ = s.keycloakAdmin.DeleteUser(keycloakUserID)
			return nil, fmt.Errorf("failed to send reset email: %w", err)
		}
	}

	// Assign PATIENT role
	role, err := s.keycloakAdmin.GetRole("PATIENT")
	if err != nil {
		log.Printf("Failed to get PATIENT role, rolling back: %s", keycloakUserID)
		_ = s.keycloakAdmin.DeleteUser(keycloakUserID)
		return nil, fmt.Errorf("failed to get PATIENT role: %w", err)
	}

	err = s.keycloakAdmin.AssignRole(keycloakUserID, *role)
	if err != nil {
		log.Printf("Failed to assign PATIENT role, rolling back: %s", keycloakUserID)
		_ = s.keycloakAdmin.DeleteUser(keycloakUserID)
		return nil, fmt.Errorf("failed to assign PATIENT role: %w", err)
	}

	// Create patient in database with keycloak_user_id
	patient, err := s.repo.CreatePatient(ctx, schemaName, orgID, keycloakUserID, req)
	if err != nil {
		log.Printf("Failed to create patient in database, rolling back: %s", keycloakUserID)
		_ = s.keycloakAdmin.DeleteUser(keycloakUserID)
		return nil, fmt.Errorf("failed to create patient in database: %w", err)
	}

	log.Printf("Successfully created patient end-to-end: %s (Keycloak ID: %s, DB ID: %s)", req.Username, keycloakUserID, patient.ID)

	return patient, nil
}

func (s *Service) ListPatients(ctx context.Context, schemaName string) ([]PatientResponse, error) {
	patients, err := s.repo.ListPatients(ctx, schemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to list patients: %w", err)
	}
	return patients, nil
}

// ListPatientsWithPagination retrieves patients with pagination
func (s *Service) ListPatientsWithPagination(ctx context.Context, schemaName string, params pagination.Params) (*PaginatedPatientListResponse, error) {
	// Validate pagination parameters
	params.Validate()

	// Get paginated data from repository
	patients, totalCount, err := s.repo.ListPatientsWithPagination(ctx, schemaName, params.Limit, params.CalculateOffset(), params.Search)
	if err != nil {
		return nil, fmt.Errorf("failed to list patients: %w", err)
	}

	// Calculate pagination metadata
	meta := params.CalculateMeta(totalCount)

	response := &PaginatedPatientListResponse{
		Success:    true,
		Patients:   patients,
		Pagination: meta,
	}

	return response, nil
}

// ListActivePatientsWithPagination retrieves active patients (not soft deleted and is_active = true) with pagination
func (s *Service) ListActivePatientsWithPagination(ctx context.Context, schemaName string, params pagination.Params) (*PaginatedPatientListResponse, error) {
	// Validate pagination parameters
	params.Validate()

	// Get paginated data from repository
	patients, totalCount, err := s.repo.ListActivePatientsWithPagination(ctx, schemaName, params.Limit, params.CalculateOffset(), params.Search)
	if err != nil {
		return nil, fmt.Errorf("failed to list active patients: %w", err)
	}

	// Calculate pagination metadata
	meta := params.CalculateMeta(totalCount)

	response := &PaginatedPatientListResponse{
		Success:    true,
		Patients:   patients,
		Pagination: meta,
	}

	return response, nil
}

func (s *Service) GetPatient(ctx context.Context, schemaName string, id string) (*PatientResponse, error) {
	patient, err := s.repo.GetPatient(ctx, schemaName, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get patient: %w", err)
	}
	return patient, nil
}

func (s *Service) UpdatePatient(ctx context.Context, schemaName string, id string, req UpdatePatientRequest) (*PatientResponse, error) {
	patient, err := s.repo.UpdatePatient(ctx, schemaName, id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update patient: %w", err)
	}
	return patient, nil
}

func (s *Service) DeletePatient(ctx context.Context, schemaName string, orgID string, id string) error {
	err := s.repo.DeletePatient(ctx, schemaName, orgID, id)
	if err != nil {
		return fmt.Errorf("failed to delete patient: %w", err)
	}
	return nil
}
