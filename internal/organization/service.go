package organization

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
