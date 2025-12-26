package organization

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

type CreateOrganizationRequest struct {
	Name         string                 `json:"name"`
	ContactEmail string                 `json:"contact_email"`
	ContactPhone string                 `json:"contact_phone"`
	Address      string                 `json:"address"`
	Settings     map[string]interface{} `json:"settings,omitempty"`
}

type OrganizationResponse struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	SchemaName   string                 `json:"schema_name"`
	ContactEmail string                 `json:"contact_email"`
	ContactPhone string                 `json:"contact_phone"`
	Address      string                 `json:"address"`
	Status       string                 `json:"status"`
	Settings     map[string]interface{} `json:"settings,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

func (s *Service) CreateOrganization(ctx context.Context, req CreateOrganizationRequest) (*OrganizationResponse, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("organization name is required")
	}

	org, err := s.repo.CreateOrganization(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	return org, nil
}

func (s *Service) ListOrganizations(ctx context.Context) ([]OrganizationResponse, error) {
	orgs, err := s.repo.ListOrganizations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}
	return orgs, nil
}

func (s *Service) GetOrganization(ctx context.Context, id string) (*OrganizationResponse, error) {
	org, err := s.repo.GetOrganization(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	return org, nil
}
