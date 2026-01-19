package organization

import (
	"context"

	"github.com/WailSalutem-Health-Care/organization-service/internal/auth"
	"github.com/WailSalutem-Health-Care/organization-service/internal/pagination"
)

// ServiceInterface defines the contract for organization business logic
type ServiceInterface interface {
	CreateOrganization(ctx context.Context, req CreateOrganizationRequest) (*OrganizationResponse, error)
	ListOrganizations(ctx context.Context, principal *auth.Principal) ([]OrganizationResponse, error)
	ListOrganizationsWithPagination(ctx context.Context, principal *auth.Principal, params pagination.Params) (*PaginatedListResponse, error)
	GetOrganization(ctx context.Context, id string, principal *auth.Principal) (*OrganizationResponse, error)
	UpdateOrganization(ctx context.Context, id string, req UpdateOrganizationRequest, principal *auth.Principal) (*OrganizationResponse, error)
	DeleteOrganization(ctx context.Context, id string) error
}

// Ensure Service implements ServiceInterface
var _ ServiceInterface = (*Service)(nil)
