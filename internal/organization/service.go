package organization

import (
	"context"
	"fmt"

	"github.com/WailSalutem-Health-Care/organization-service/internal/pagination"
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

// ListOrganizationsWithPagination retrieves organizations with pagination
func (s *Service) ListOrganizationsWithPagination(ctx context.Context, params pagination.Params) (*PaginatedListResponse, error) {
	// Validate pagination parameters
	params.Validate()

	// Get paginated data from repository
	orgs, totalCount, err := s.repo.ListOrganizationsWithPagination(ctx, params.Limit, params.CalculateOffset())
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	// Calculate pagination metadata
	meta := params.CalculateMeta(totalCount)

	response := &PaginatedListResponse{
		Success:       true,
		Organizations: orgs,
		Pagination:    meta,
	}

	return response, nil
}

func (s *Service) GetOrganization(ctx context.Context, id string) (*OrganizationResponse, error) {
	org, err := s.repo.GetOrganization(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	return org, nil
}

func (s *Service) UpdateOrganization(ctx context.Context, id string, req UpdateOrganizationRequest) (*OrganizationResponse, error) {
	org, err := s.repo.UpdateOrganization(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}
	return org, nil
}

func (s *Service) DeleteOrganization(ctx context.Context, id string) error {
	err := s.repo.DeleteOrganization(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}
	return nil
}
