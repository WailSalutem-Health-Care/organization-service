package patient

import (
	"context"
	"fmt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreatePatient(ctx context.Context, schemaName string, req CreatePatientRequest) (*PatientResponse, error) {
	if req.FirstName == "" {
		return nil, fmt.Errorf("first name is required")
	}
	if req.LastName == "" {
		return nil, fmt.Errorf("last name is required")
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
