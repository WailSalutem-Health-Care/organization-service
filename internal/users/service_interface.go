package users

import (
	"github.com/WailSalutem-Health-Care/organization-service/internal/auth"
	"github.com/WailSalutem-Health-Care/organization-service/internal/pagination"
)

// ServiceInterface defines the contract for user business logic operations
type ServiceInterface interface {
	CreateUser(req CreateUserRequest, principal *auth.Principal, targetOrgID string) (*User, error)
	GetUser(userID string, principal *auth.Principal, targetOrgID string) (*User, error)
	ListUsers(principal *auth.Principal, targetOrgID string) ([]User, error)
	ListUsersWithPagination(principal *auth.Principal, targetOrgID string, params pagination.Params) (*PaginatedUserListResponse, error)
	ListActiveUsersByRoleWithPagination(principal *auth.Principal, targetOrgID string, role string, params pagination.Params) (*PaginatedUserListResponse, error)
	UpdateUser(userID string, req UpdateUserRequest, principal *auth.Principal, targetOrgID string) (*User, error)
	UpdateMyProfile(req UpdateUserRequest, principal *auth.Principal) (*User, error)
	ResetPassword(userID string, req ResetPasswordRequest, principal *auth.Principal, targetOrgID string) error
	DeleteUser(userID string, principal *auth.Principal) error
}
