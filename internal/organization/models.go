package organization

import "time"

// CreateOrganizationRequest represents the request to create a new organization
type CreateOrganizationRequest struct {
	Name         string                 `json:"name"`
	ContactEmail string                 `json:"contact_email"`
	ContactPhone string                 `json:"contact_phone"`
	Address      string                 `json:"address"`
	Settings     map[string]interface{} `json:"settings,omitempty"`
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
