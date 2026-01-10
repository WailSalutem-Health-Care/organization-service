package users

import (
	"time"

	"github.com/WailSalutem-Health-Care/organization-service/internal/pagination"
)

// User represents a user in the system
type User struct {
	ID             string    `json:"id"`
	KeycloakUserID string    `json:"keycloakUserId"`
	EmployeeID     string    `json:"employeeId,omitempty"`
	Email          string    `json:"email"`
	FirstName      string    `json:"firstName"`
	LastName       string    `json:"lastName"`
	PhoneNumber    string    `json:"phoneNumber,omitempty"`
	Role           string    `json:"role"`
	IsActive       bool      `json:"isActive"`
	OrgID          string    `json:"orgId"`
	OrgSchemaName  string    `json:"orgSchemaName"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt,omitempty"`
}

// CreateUserRequest represents the request to create a new user (non-PATIENT roles only)
type CreateUserRequest struct {
	Username          string `json:"username"`
	Email             string `json:"email"`
	FirstName         string `json:"firstName"`
	LastName          string `json:"lastName"`
	PhoneNumber       string `json:"phoneNumber,omitempty"`
	Role              string `json:"role"`
	TemporaryPassword string `json:"temporaryPassword"`
	SendResetEmail    bool   `json:"sendResetEmail"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Email       string `json:"email,omitempty"`
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
}

// ResetPasswordRequest represents the request to reset a user's password
type ResetPasswordRequest struct {
	TemporaryPassword string `json:"temporaryPassword"`
	SendEmail         bool   `json:"sendEmail"`
}

// AllowedRoles defines which roles an ORG_ADMIN can create
var AllowedRolesForOrgAdmin = map[string]bool{
	"CAREGIVER":    true,
	"PATIENT":      true,
	"MUNICIPALITY": true,
	"INSURER":      true,
}

// IsRoleAllowedForOrgAdmin checks if a role can be created by an org admin
func IsRoleAllowedForOrgAdmin(role string) bool {
	return AllowedRolesForOrgAdmin[role]
}

// Validate validates the create user request
func (r *CreateUserRequest) Validate() error {
	if r.Username == "" {
		return ErrMissingUsername
	}
	if r.Email == "" {
		return ErrMissingEmail
	}
	if r.FirstName == "" {
		return ErrMissingFirstName
	}
	if r.LastName == "" {
		return ErrMissingLastName
	}
	if r.Role == "" {
		return ErrMissingRole
	}
	if r.TemporaryPassword == "" && !r.SendResetEmail {
		return ErrMissingPassword
	}

	return nil
}

// PaginatedUserListResponse represents a paginated list of users
type PaginatedUserListResponse struct {
	Users      []User          `json:"users"`
	Pagination pagination.Meta `json:"pagination"`
}
