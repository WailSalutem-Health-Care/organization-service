package organization

import (
	"time"

	"github.com/WailSalutem-Health-Care/organization-service/internal/pagination"
)

// CreateOrganizationRequest represents the request to create a new organization
type CreateOrganizationRequest struct {
	Name         string                 `json:"name"`
	ContactEmail string                 `json:"contact_email"`
	ContactPhone string                 `json:"contact_phone"`
	Address      string                 `json:"address"`
	Settings     map[string]interface{} `json:"settings,omitempty"`
}

// UpdateOrganizationRequest represents the request to update an organization
type UpdateOrganizationRequest struct {
	Name         *string                 `json:"name,omitempty"`
	ContactEmail *string                 `json:"contact_email,omitempty"`
	ContactPhone *string                 `json:"contact_phone,omitempty"`
	Address      *string                 `json:"address,omitempty"`
	Settings     *map[string]interface{} `json:"settings,omitempty"`
}

// OrganizationResponse represents the organization data returned to clients
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

// PaginatedListResponse represents a paginated list of organizations
type PaginatedListResponse struct {
	Success       bool                   `json:"success"`
	Organizations []OrganizationResponse `json:"organizations"`
	Pagination    pagination.Meta        `json:"pagination"`
}
