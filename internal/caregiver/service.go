package caregiver

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

type CreateCaregiverRequest struct {
	KeycloakUserID string `json:"keycloak_user_id" validate:"required"`
	FullName       string `json:"full_name" validate:"required"`
	Email          string `json:"email" validate:"required"`
	PhoneNumber    string `json:"phone_number"`
	Role           string `json:"role" validate:"required"`
}

type UpdateCaregiverRequest struct {
	FullName    *string `json:"full_name,omitempty"`
	Email       *string `json:"email,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty"`
	Role        *string `json:"role,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

type CaregiverResponse struct {
	ID             string     `json:"id"`
	KeycloakUserID string     `json:"keycloak_user_id"`
	FullName       string     `json:"full_name"`
	Email          string     `json:"email"`
	PhoneNumber    string     `json:"phone_number"`
	Role           string     `json:"role"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty"`
}

func (s *Service) CreateCaregiver(ctx context.Context, schemaName string, req CreateCaregiverRequest) (*CaregiverResponse, error) {
	if req.FullName == "" {
		return nil, fmt.Errorf("full name is required")
	}
	if req.Email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if req.KeycloakUserID == "" {
		return nil, fmt.Errorf("keycloak user ID is required")
	}
	if req.Role == "" {
		return nil, fmt.Errorf("role is required")
	}

	caregiver, err := s.repo.CreateCaregiver(ctx, schemaName, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create caregiver: %w", err)
	}

	return caregiver, nil
}

func (s *Service) ListCaregivers(ctx context.Context, schemaName string) ([]CaregiverResponse, error) {
	caregivers, err := s.repo.ListCaregivers(ctx, schemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to list caregivers: %w", err)
	}
	return caregivers, nil
}

func (s *Service) GetCaregiver(ctx context.Context, schemaName string, id string) (*CaregiverResponse, error) {
	caregiver, err := s.repo.GetCaregiver(ctx, schemaName, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get caregiver: %w", err)
	}
	return caregiver, nil
}

func (s *Service) UpdateCaregiver(ctx context.Context, schemaName string, id string, req UpdateCaregiverRequest) (*CaregiverResponse, error) {
	caregiver, err := s.repo.UpdateCaregiver(ctx, schemaName, id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update caregiver: %w", err)
	}
	return caregiver, nil
}

func (s *Service) DeleteCaregiver(ctx context.Context, schemaName string, id string) error {
	err := s.repo.DeleteCaregiver(ctx, schemaName, id)
	if err != nil {
		return fmt.Errorf("failed to delete caregiver: %w", err)
	}
	return nil
}
