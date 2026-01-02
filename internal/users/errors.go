package users

import "errors"

var (
	ErrMissingUsername  = errors.New("username is required")
	ErrMissingEmail     = errors.New("email is required")
	ErrMissingFirstName = errors.New("first name is required")
	ErrMissingLastName  = errors.New("last name is required")
	ErrMissingRole      = errors.New("role is required")
	ErrMissingPassword  = errors.New("temporary password is required when not sending reset email")
	ErrInvalidRole      = errors.New("invalid role")
	ErrRoleNotAllowed   = errors.New("org_admin cannot create this role")
	ErrUserNotFound     = errors.New("user not found")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden - insufficient permissions")
	ErrInvalidOrgSchema = errors.New("invalid organization schema name")
)
