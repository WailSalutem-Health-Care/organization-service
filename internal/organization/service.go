package organization

import (
	"context"
	"fmt"

	"github.com/WailSalutem-Health-Care/organization-service/internal/auth"
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

func (s *Service) ListOrganizations(ctx context.Context, principal *auth.Principal) ([]OrganizationResponse, error) {
	// Check if user is SUPER_ADMIN
	isSuperAdmin := false
	for _, role := range principal.Roles {
		if role == "SUPER_ADMIN" {
			isSuperAdmin = true
			break
		}
	}

	// SUPER_ADMIN can see all organizations
	if isSuperAdmin {
		orgs, err := s.repo.ListOrganizations(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list organizations: %w", err)
		}
		return orgs, nil
	}

	// ORG_ADMIN can only see their own organization
	if principal.OrgID == "" {
		return nil, fmt.Errorf("no organization associated with this user")
	}

	org, err := s.repo.GetOrganization(ctx, principal.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Return as array with single organization
	return []OrganizationResponse{*org}, nil
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
