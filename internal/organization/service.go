package organization

import (
	"context"
	"fmt"
	"log"

	"github.com/WailSalutem-Health-Care/organization-service/internal/auth"
	"github.com/WailSalutem-Health-Care/organization-service/internal/pagination"
)

type Service struct {
	repo RepositoryInterface
}

func NewService(repo RepositoryInterface) *Service {
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

func (s *Service) ListOrganizations(ctx context.Context, principal *auth.Principal) ([]OrganizationResponse, error) {
	// Debug logging
	log.Printf("[DEBUG] ListOrganizations - UserID: %s, OrgID: %s, Roles: %v", principal.UserID, principal.OrgID, principal.Roles)

	// Check if user is SUPER_ADMIN
	isSuperAdmin := false
	for _, role := range principal.Roles {
		if role == "SUPER_ADMIN" {
			isSuperAdmin = true
			break
		}
	}

	log.Printf("[DEBUG] isSuperAdmin: %v", isSuperAdmin)

	// SUPER_ADMIN can see all organizations
	if isSuperAdmin {
		orgs, err := s.repo.ListOrganizations(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list organizations: %w", err)
		}
		log.Printf("[DEBUG] Returning %d organizations for SUPER_ADMIN", len(orgs))
		return orgs, nil
	}

	// ORG_ADMIN can only see their own organization
	if principal.OrgID == "" {
		log.Printf("[DEBUG] No org_id for non-SUPER_ADMIN user")
		return nil, fmt.Errorf("no organization associated with this user")
	}

	org, err := s.repo.GetOrganization(ctx, principal.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Return as array with single organization
	return []OrganizationResponse{*org}, nil
}

// ListOrganizationsWithPagination retrieves organizations with pagination and authorization
func (s *Service) ListOrganizationsWithPagination(ctx context.Context, principal *auth.Principal, params pagination.Params) (*PaginatedListResponse, error) {
	// Validate pagination parameters
	params.Validate()

	// Check if user is SUPER_ADMIN
	isSuperAdmin := false
	for _, role := range principal.Roles {
		if role == "SUPER_ADMIN" {
			isSuperAdmin = true
			break
		}
	}

	// SUPER_ADMIN can see all organizations with pagination
	if isSuperAdmin {
		// Get paginated data from repository with search and status filters
		orgs, totalCount, err := s.repo.ListOrganizationsWithPagination(ctx, params.Limit, params.CalculateOffset(), params.Search, params.Status)
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

	// ORG_ADMIN can only see their own organization (pagination not really needed, but keep consistent)
	if principal.OrgID == "" {
		return nil, fmt.Errorf("no organization associated with this user")
	}

	org, err := s.repo.GetOrganization(ctx, principal.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Return as array with single organization
	meta := params.CalculateMeta(1)
	response := &PaginatedListResponse{
		Success:       true,
		Organizations: []OrganizationResponse{*org},
		Pagination:    meta,
	}

	return response, nil
}

func (s *Service) GetOrganization(ctx context.Context, id string, principal *auth.Principal) (*OrganizationResponse, error) {
	// Check if user is SUPER_ADMIN
	isSuperAdmin := false
	for _, role := range principal.Roles {
		if role == "SUPER_ADMIN" {
			isSuperAdmin = true
			break
		}
	}

	// ORG_ADMIN can only view their own organization
	if !isSuperAdmin && principal.OrgID != id {
		return nil, fmt.Errorf("forbidden")
	}

	org, err := s.repo.GetOrganization(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	return org, nil
}

func (s *Service) UpdateOrganization(ctx context.Context, id string, req UpdateOrganizationRequest, principal *auth.Principal) (*OrganizationResponse, error) {
	// Check if user is SUPER_ADMIN
	isSuperAdmin := false
	for _, role := range principal.Roles {
		if role == "SUPER_ADMIN" {
			isSuperAdmin = true
			break
		}
	}

	// Only SUPER_ADMIN can update organizations
	if !isSuperAdmin {
		return nil, fmt.Errorf("forbidden")
	}

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
