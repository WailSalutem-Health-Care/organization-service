package organization

import "context"

// RepositoryInterface defines the contract for organization data access
type RepositoryInterface interface {
	CreateOrganization(ctx context.Context, req CreateOrganizationRequest) (*OrganizationResponse, error)
	ListOrganizations(ctx context.Context) ([]OrganizationResponse, error)
	ListOrganizationsWithPagination(ctx context.Context, limit, offset int, search, status string) ([]OrganizationResponse, int, error)
	GetOrganization(ctx context.Context, id string) (*OrganizationResponse, error)
	UpdateOrganization(ctx context.Context, id string, req UpdateOrganizationRequest) (*OrganizationResponse, error)
	DeleteOrganization(ctx context.Context, id string) error
}

// Ensure Repository implements RepositoryInterface
var _ RepositoryInterface = (*Repository)(nil)
