package patient

import (
	"context"
	"fmt"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type CreatePatientRequest struct {
	FullName              string `json:"full_name" validate:"required"`
	Email                 string `json:"email"`
	PhoneNumber           string `json:"phone_number"`
	DateOfBirth           string `json:"date_of_birth"` // Format: YYYY-MM-DD
	Address               string `json:"address"`
	EmergencyContactName  string `json:"emergency_contact_name"`
	EmergencyContactPhone string `json:"emergency_contact_phone"`
	MedicalNotes          string `json:"medical_notes"`
}

type UpdatePatientRequest struct {
	FullName              *string `json:"full_name,omitempty"`
	Email                 *string `json:"email,omitempty"`
	PhoneNumber           *string `json:"phone_number,omitempty"`
	DateOfBirth           *string `json:"date_of_birth,omitempty"`
	Address               *string `json:"address,omitempty"`
	EmergencyContactName  *string `json:"emergency_contact_name,omitempty"`
	EmergencyContactPhone *string `json:"emergency_contact_phone,omitempty"`
	MedicalNotes          *string `json:"medical_notes,omitempty"`
	IsActive              *bool   `json:"is_active,omitempty"`
}

type PatientResponse struct {
	ID                    string     `json:"id"`
	FullName              string     `json:"full_name"`
	Email                 string     `json:"email"`
	PhoneNumber           string     `json:"phone_number"`
	DateOfBirth           *string    `json:"date_of_birth,omitempty"`
	Address               string     `json:"address"`
	EmergencyContactName  string     `json:"emergency_contact_name"`
	EmergencyContactPhone string     `json:"emergency_contact_phone"`
	MedicalNotes          string     `json:"medical_notes"`
	IsActive              bool       `json:"is_active"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             *time.Time `json:"updated_at,omitempty"`
}

func (s *Service) CreatePatient(ctx context.Context, schemaName string, req CreatePatientRequest) (*PatientResponse, error) {
	if req.FullName == "" {
		return nil, fmt.Errorf("full name is required")
	}

	patient, err := s.repo.CreatePatient(ctx, schemaName, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create patient: %w", err)
	}

	return patient, nil
}

func (s *Service) ListPatients(ctx context.Context, schemaName string) ([]PatientResponse, error) {
	patients, err := s.repo.ListPatients(ctx, schemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to list patients: %w", err)
	}
	return patients, nil
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

func (s *Service) DeletePatient(ctx context.Context, schemaName string, id string) error {
	err := s.repo.DeletePatient(ctx, schemaName, id)
	if err != nil {
		return fmt.Errorf("failed to delete patient: %w", err)
	}
	return nil
}
