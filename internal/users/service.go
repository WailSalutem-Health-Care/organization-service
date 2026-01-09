package users

import (
	"fmt"
	"log"

	"github.com/WailSalutem-Health-Care/organization-service/internal/auth"
	"github.com/WailSalutem-Health-Care/organization-service/internal/pagination"
)

type Service struct {
	repo          *Repository
	keycloakAdmin *auth.KeycloakAdminClient
}

func NewService(repo *Repository, keycloakAdmin *auth.KeycloakAdminClient) *Service {
	return &Service{
		repo:          repo,
		keycloakAdmin: keycloakAdmin,
	}
}

func (s *Service) CreateUser(req CreateUserRequest, principal *auth.Principal, targetOrgID string) (*User, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var effectiveOrgID string

	if s.hasRole(principal, "SUPER_ADMIN") {
		if targetOrgID == "" {
			log.Printf("SUPER_ADMIN must provide X-Organization-ID header")
			return nil, fmt.Errorf("SUPER_ADMIN must provide X-Organization-ID header to specify target organization")
		}
		effectiveOrgID = targetOrgID
		log.Printf("SUPER_ADMIN creating user in org: %s", effectiveOrgID)
	} else {
		if targetOrgID != "" {
			log.Printf("ORG_ADMIN attempted to create user in different org")
			return nil, ErrForbidden
		}
		effectiveOrgID = principal.OrgID
		if effectiveOrgID == "" {
			log.Printf("No organization ID in token")
			return nil, ErrInvalidOrgSchema
		}
		log.Printf("ORG_ADMIN creating user in own org: %s", effectiveOrgID)
	}

	if !s.hasRole(principal, "SUPER_ADMIN") {
		if !IsRoleAllowedForOrgAdmin(req.Role) {
			log.Printf("ORG_ADMIN attempted to create forbidden role: %s", req.Role)
			return nil, ErrRoleNotAllowed
		}
	}

	orgSchemaName := principal.OrgSchemaName

	if targetOrgID != "" || orgSchemaName == "" {
		var err error
		orgSchemaName, err = s.repo.GetSchemaNameByOrgID(effectiveOrgID)
		if err != nil {
			log.Printf("Failed to get schema name for orgId %s: %v", effectiveOrgID, err)
			return nil, ErrInvalidOrgSchema
		}
		log.Printf("Looked up schema name '%s' for orgId '%s'", orgSchemaName, effectiveOrgID)
	}

	if err := s.repo.ValidateOrgSchema(orgSchemaName); err != nil {
		return nil, err
	}

	keycloakUser := auth.KeycloakUser{
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Enabled:   true,
		Attributes: map[string][]string{
			"organizationID": {effectiveOrgID},
			"orgSchemaName":  {orgSchemaName},
		},
	}

	log.Printf("Creating Keycloak user with attributes: organizationID=%s, orgSchemaName=%s", effectiveOrgID, orgSchemaName)

	keycloakUserID, err := s.keycloakAdmin.CreateUser(keycloakUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create user in Keycloak: %w", err)
	}

	log.Printf("Created user in Keycloak: %s (ID: %s)", req.Username, keycloakUserID)

	if req.TemporaryPassword != "" {
		err = s.keycloakAdmin.SetPassword(keycloakUserID, req.TemporaryPassword, false)
		if err != nil {
			log.Printf("Failed to set password, rolling back user creation: %s", keycloakUserID)
			_ = s.keycloakAdmin.DeleteUser(keycloakUserID)
			return nil, fmt.Errorf("failed to set password: %w", err)
		}
	} else if req.SendResetEmail {
		err = s.keycloakAdmin.SendEmailAction(keycloakUserID, []string{"UPDATE_PASSWORD"})
		if err != nil {
			log.Printf("Failed to send reset email, rolling back user creation: %s", keycloakUserID)
			_ = s.keycloakAdmin.DeleteUser(keycloakUserID)
			return nil, fmt.Errorf("failed to send reset email: %w", err)
		}
	}

	role, err := s.keycloakAdmin.GetRole(req.Role)
	if err != nil {
		log.Printf("Failed to get role, rolling back user creation: %s", keycloakUserID)
		_ = s.keycloakAdmin.DeleteUser(keycloakUserID)
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	err = s.keycloakAdmin.AssignRole(keycloakUserID, *role)
	if err != nil {
		log.Printf("Failed to assign role, rolling back user creation: %s", keycloakUserID)
		_ = s.keycloakAdmin.DeleteUser(keycloakUserID)
		return nil, fmt.Errorf("failed to assign role: %w", err)
	}

	user := &User{
		KeycloakUserID: keycloakUserID,
		Email:          req.Email,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		PhoneNumber:    req.PhoneNumber,
		Role:           req.Role,
		OrgID:          effectiveOrgID,
		OrgSchemaName:  orgSchemaName,
	}

	// Reject PATIENT role - patients should be created via /organization/patients endpoint
	if req.Role == "PATIENT" {
		log.Printf("Attempted to create PATIENT via users endpoint: %s", keycloakUserID)
		_ = s.keycloakAdmin.DeleteUser(keycloakUserID)
		return nil, fmt.Errorf("PATIENT users must be created via /organization/patients endpoint")
	}

	// Create user in users table (for CAREGIVER, MUNICIPALITY, INSURER, etc.)
	err = s.repo.Create(user)
	if err != nil {
		log.Printf("Failed to create user in database, rolling back: %s", keycloakUserID)
		_ = s.keycloakAdmin.DeleteUser(keycloakUserID)
		return nil, fmt.Errorf("failed to create user in database: %w", err)
	}
	log.Printf("Successfully created user record: %s", user.ID)

	log.Printf("Successfully created user end-to-end: %s (Keycloak ID: %s, DB ID: %s)", req.Username, keycloakUserID, user.ID)

	return user, nil
}

func (s *Service) GetUser(userID string, principal *auth.Principal, targetOrgID string) (*User, error) {
	var effectiveOrgID string

	if s.hasRole(principal, "SUPER_ADMIN") {
		if targetOrgID != "" {
			effectiveOrgID = targetOrgID
			log.Printf("SUPER_ADMIN getting user from org: %s", effectiveOrgID)
		} else {
			effectiveOrgID = principal.OrgID
			if effectiveOrgID == "" {
				log.Printf("SUPER_ADMIN token has no orgId and no X-Organization-ID header provided")
				return nil, ErrInvalidOrgSchema
			}
			log.Printf("SUPER_ADMIN getting user from own org: %s", effectiveOrgID)
		}
	} else {
		if targetOrgID != "" {
			log.Printf("ORG_ADMIN attempted to get user from different org")
			return nil, ErrForbidden
		}
		effectiveOrgID = principal.OrgID
		if effectiveOrgID == "" {
			log.Printf("No organization ID in token")
			return nil, ErrInvalidOrgSchema
		}
		log.Printf("ORG_ADMIN getting user from own org: %s", effectiveOrgID)
	}

	orgSchemaName, err := s.repo.GetSchemaNameByOrgID(effectiveOrgID)
	if err != nil {
		log.Printf("Failed to get schema name for orgId %s: %v", effectiveOrgID, err)
		return nil, ErrInvalidOrgSchema
	}
	log.Printf("Looked up schema name '%s' for orgId '%s'", orgSchemaName, effectiveOrgID)

	user, err := s.repo.GetByID(orgSchemaName, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) ListUsers(principal *auth.Principal, targetOrgID string) ([]User, error) {
	var effectiveOrgID string

	if s.hasRole(principal, "SUPER_ADMIN") {
		if targetOrgID != "" {
			effectiveOrgID = targetOrgID
			log.Printf("SUPER_ADMIN listing users from org: %s", effectiveOrgID)
		} else {
			effectiveOrgID = principal.OrgID
			if effectiveOrgID == "" {
				log.Printf("SUPER_ADMIN token has no orgId and no X-Organization-ID header provided")
				return nil, ErrInvalidOrgSchema
			}
			log.Printf("SUPER_ADMIN listing users from own org: %s", effectiveOrgID)
		}
	} else {
		if targetOrgID != "" {
			log.Printf("ORG_ADMIN attempted to list users from different org")
			return nil, ErrForbidden
		}
		effectiveOrgID = principal.OrgID
		if effectiveOrgID == "" {
			log.Printf("No organization ID in token")
			return nil, ErrInvalidOrgSchema
		}
		log.Printf("ORG_ADMIN listing users from own org: %s", effectiveOrgID)
	}

	orgSchemaName, err := s.repo.GetSchemaNameByOrgID(effectiveOrgID)
	if err != nil {
		log.Printf("Failed to get schema name for orgId %s: %v", effectiveOrgID, err)
		return nil, ErrInvalidOrgSchema
	}
	log.Printf("Looked up schema name '%s' for orgId '%s'", orgSchemaName, effectiveOrgID)

	users, err := s.repo.List(orgSchemaName)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// ListUsersWithPagination retrieves users with pagination
func (s *Service) ListUsersWithPagination(principal *auth.Principal, targetOrgID string, params pagination.Params) (*PaginatedUserListResponse, error) {
	var effectiveOrgID string

	if s.hasRole(principal, "SUPER_ADMIN") {
		if targetOrgID != "" {
			effectiveOrgID = targetOrgID
			log.Printf("SUPER_ADMIN listing users from org: %s", effectiveOrgID)
		} else {
			effectiveOrgID = principal.OrgID
			if effectiveOrgID == "" {
				log.Printf("SUPER_ADMIN token has no orgId and no X-Organization-ID header provided")
				return nil, ErrInvalidOrgSchema
			}
			log.Printf("SUPER_ADMIN listing users from own org: %s", effectiveOrgID)
		}
	} else {
		if targetOrgID != "" {
			log.Printf("ORG_ADMIN attempted to list users from different org")
			return nil, ErrForbidden
		}
		effectiveOrgID = principal.OrgID
		if effectiveOrgID == "" {
			log.Printf("No organization ID in token")
			return nil, ErrInvalidOrgSchema
		}
		log.Printf("ORG_ADMIN listing users from own org: %s", effectiveOrgID)
	}

	orgSchemaName, err := s.repo.GetSchemaNameByOrgID(effectiveOrgID)
	if err != nil {
		log.Printf("Failed to get schema name for orgId %s: %v", effectiveOrgID, err)
		return nil, ErrInvalidOrgSchema
	}
	log.Printf("Looked up schema name '%s' for orgId '%s'", orgSchemaName, effectiveOrgID)

	// Validate pagination parameters
	params.Validate()

	// Get paginated data from repository
	users, totalCount, err := s.repo.ListWithPagination(orgSchemaName, params.Limit, params.CalculateOffset(), params.Search)
	if err != nil {
		return nil, err
	}

	// Calculate pagination metadata
	meta := params.CalculateMeta(totalCount)

	response := &PaginatedUserListResponse{
		Users:      users,
		Pagination: meta,
	}

	return response, nil
}

// ListActiveUsersByRoleWithPagination retrieves active users (not soft deleted) by role with pagination
func (s *Service) ListActiveUsersByRoleWithPagination(principal *auth.Principal, targetOrgID string, role string, params pagination.Params) (*PaginatedUserListResponse, error) {
	var effectiveOrgID string

	if s.hasRole(principal, "SUPER_ADMIN") {
		if targetOrgID != "" {
			effectiveOrgID = targetOrgID
			log.Printf("SUPER_ADMIN listing active %s users from org: %s", role, effectiveOrgID)
		} else {
			effectiveOrgID = principal.OrgID
			if effectiveOrgID == "" {
				log.Printf("SUPER_ADMIN token has no orgId and no X-Organization-ID header provided")
				return nil, ErrInvalidOrgSchema
			}
			log.Printf("SUPER_ADMIN listing active %s users from own org: %s", role, effectiveOrgID)
		}
	} else {
		if targetOrgID != "" {
			log.Printf("ORG_ADMIN attempted to list active %s users from different org", role)
			return nil, ErrForbidden
		}
		effectiveOrgID = principal.OrgID
		if effectiveOrgID == "" {
			log.Printf("No organization ID in token")
			return nil, ErrInvalidOrgSchema
		}
		log.Printf("ORG_ADMIN listing active %s users from own org: %s", role, effectiveOrgID)
	}

	orgSchemaName, err := s.repo.GetSchemaNameByOrgID(effectiveOrgID)
	if err != nil {
		log.Printf("Failed to get schema name for orgId %s: %v", effectiveOrgID, err)
		return nil, ErrInvalidOrgSchema
	}
	log.Printf("Looked up schema name '%s' for orgId '%s'", orgSchemaName, effectiveOrgID)

	// Validate pagination parameters
	params.Validate()

	// Get paginated data from repository
	users, totalCount, err := s.repo.ListActiveUsersByRoleWithPagination(orgSchemaName, role, params.Limit, params.CalculateOffset(), params.Search)
	if err != nil {
		return nil, err
	}

	// Calculate pagination metadata
	meta := params.CalculateMeta(totalCount)

	response := &PaginatedUserListResponse{
		Users:      users,
		Pagination: meta,
	}

	return response, nil
}

func (s *Service) UpdateUser(userID string, req UpdateUserRequest, principal *auth.Principal, targetOrgID string) (*User, error) {
	var effectiveOrgID string

	if s.hasRole(principal, "SUPER_ADMIN") {
		if targetOrgID != "" {
			effectiveOrgID = targetOrgID
			log.Printf("SUPER_ADMIN updating user in org: %s", effectiveOrgID)
		} else {
			effectiveOrgID = principal.OrgID
			if effectiveOrgID == "" {
				log.Printf("SUPER_ADMIN token has no orgId and no X-Organization-ID header provided")
				return nil, ErrInvalidOrgSchema
			}
			log.Printf("SUPER_ADMIN updating user in own org: %s", effectiveOrgID)
		}
	} else {
		if targetOrgID != "" {
			log.Printf("ORG_ADMIN attempted to update user in different org")
			return nil, ErrForbidden
		}
		effectiveOrgID = principal.OrgID
		if effectiveOrgID == "" {
			log.Printf("No organization ID in token")
			return nil, ErrInvalidOrgSchema
		}
		log.Printf("ORG_ADMIN updating user in own org: %s", effectiveOrgID)
	}

	orgSchemaName, err := s.repo.GetSchemaNameByOrgID(effectiveOrgID)
	if err != nil {
		log.Printf("Failed to get schema name for orgId %s: %v", effectiveOrgID, err)
		return nil, ErrInvalidOrgSchema
	}
	log.Printf("Looked up schema name '%s' for orgId '%s'", orgSchemaName, effectiveOrgID)

	user, err := s.repo.GetByID(orgSchemaName, userID)
	if err != nil {
		return nil, err
	}

	keycloakUpdateNeeded := false

	if req.Email != "" {
		user.Email = req.Email
		keycloakUpdateNeeded = true
	}
	if req.FirstName != "" {
		user.FirstName = req.FirstName
		keycloakUpdateNeeded = true
	}
	if req.LastName != "" {
		user.LastName = req.LastName
		keycloakUpdateNeeded = true
	}
	if req.PhoneNumber != "" {
		user.PhoneNumber = req.PhoneNumber
	}

	if keycloakUpdateNeeded {
		keycloakUserData, err := s.keycloakAdmin.GetUser(user.KeycloakUserID)
		if err != nil {
			log.Printf("Failed to get user from Keycloak: %v", err)
			return nil, fmt.Errorf("failed to get user from Keycloak: %w", err)
		}

		keycloakUser := auth.KeycloakUser{
			Username:  keycloakUserData.Username,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Enabled:   true,
		}

		err = s.keycloakAdmin.UpdateUser(user.KeycloakUserID, keycloakUser)
		if err != nil {
			log.Printf("Failed to update user in Keycloak: %v", err)
			return nil, fmt.Errorf("failed to update user in Keycloak: %w", err)
		}
		log.Printf("Updated user in Keycloak: %s", user.KeycloakUserID)
	}

	err = s.repo.Update(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) UpdateMyProfile(req UpdateUserRequest, principal *auth.Principal) (*User, error) {
	if principal.OrgID == "" {
		log.Printf("User token missing organizationID claim - cannot update profile")
		return nil, fmt.Errorf("user token must contain organizationID claim")
	}

	orgSchemaName, err := s.repo.GetSchemaNameByOrgID(principal.OrgID)
	if err != nil {
		log.Printf("Failed to get schema name for orgId %s: %v", principal.OrgID, err)
		return nil, ErrInvalidOrgSchema
	}

	user, err := s.repo.GetByKeycloakID(orgSchemaName, principal.UserID)
	if err != nil {
		log.Printf("Failed to get user by Keycloak ID: %v", err)
		return nil, ErrUserNotFound
	}

	keycloakUpdateNeeded := false

	if req.Email != "" {
		user.Email = req.Email
		keycloakUpdateNeeded = true
	}
	if req.FirstName != "" {
		user.FirstName = req.FirstName
		keycloakUpdateNeeded = true
	}
	if req.LastName != "" {
		user.LastName = req.LastName
		keycloakUpdateNeeded = true
	}
	if req.PhoneNumber != "" {
		user.PhoneNumber = req.PhoneNumber
	}

	if keycloakUpdateNeeded {
		keycloakUserData, err := s.keycloakAdmin.GetUser(user.KeycloakUserID)
		if err != nil {
			log.Printf("Failed to get user from Keycloak: %v", err)
			return nil, fmt.Errorf("failed to get user from Keycloak: %w", err)
		}

		keycloakUser := auth.KeycloakUser{
			Username:  keycloakUserData.Username,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Enabled:   true,
		}

		err = s.keycloakAdmin.UpdateUser(user.KeycloakUserID, keycloakUser)
		if err != nil {
			log.Printf("Failed to update user in Keycloak: %v", err)
			return nil, fmt.Errorf("failed to update user in Keycloak: %w", err)
		}
		log.Printf("User updated their own profile in Keycloak: %s", user.KeycloakUserID)
	}

	err = s.repo.Update(user)
	if err != nil {
		return nil, err
	}

	log.Printf("User updated their own profile: %s (Keycloak ID: %s)", user.Email, user.KeycloakUserID)

	return user, nil
}

func (s *Service) ResetPassword(userID string, req ResetPasswordRequest, principal *auth.Principal, targetOrgID string) error {
	var effectiveOrgID string

	if s.hasRole(principal, "SUPER_ADMIN") {
		if targetOrgID != "" {
			effectiveOrgID = targetOrgID
			log.Printf("SUPER_ADMIN resetting password for user in org: %s", effectiveOrgID)
		} else {
			effectiveOrgID = principal.OrgID
			if effectiveOrgID == "" {
				log.Printf("SUPER_ADMIN token has no orgId and no X-Organization-ID header provided")
				return ErrInvalidOrgSchema
			}
			log.Printf("SUPER_ADMIN resetting password for user in own org: %s", effectiveOrgID)
		}
	} else {
		if targetOrgID != "" {
			log.Printf("ORG_ADMIN attempted to reset password in different org")
			return ErrForbidden
		}
		effectiveOrgID = principal.OrgID
		if effectiveOrgID == "" {
			log.Printf("No organization ID in token")
			return ErrInvalidOrgSchema
		}
		log.Printf("ORG_ADMIN resetting password for user in own org: %s", effectiveOrgID)
	}

	orgSchemaName, err := s.repo.GetSchemaNameByOrgID(effectiveOrgID)
	if err != nil {
		log.Printf("Failed to get schema name for orgId %s: %v", effectiveOrgID, err)
		return ErrInvalidOrgSchema
	}
	log.Printf("Looked up schema name '%s' for orgId '%s'", orgSchemaName, effectiveOrgID)

	user, err := s.repo.GetByID(orgSchemaName, userID)
	if err != nil {
		return err
	}

	if req.TemporaryPassword != "" {
		err = s.keycloakAdmin.SetPassword(user.KeycloakUserID, req.TemporaryPassword, true)
		if err != nil {
			return fmt.Errorf("failed to reset password: %w", err)
		}
	} else if req.SendEmail {
		err = s.keycloakAdmin.SendEmailAction(user.KeycloakUserID, []string{"UPDATE_PASSWORD"})
		if err != nil {
			return fmt.Errorf("failed to send reset email: %w", err)
		}
	}

	log.Printf("Reset password for user: %s (Keycloak ID: %s)", user.Email, user.KeycloakUserID)

	return nil
}

func (s *Service) DeleteUser(userID string, principal *auth.Principal) error {
	orgSchemaName := principal.OrgSchemaName
	if orgSchemaName == "" {
		if principal.OrgID == "" {
			log.Printf("Principal has no orgId or orgSchemaName")
			return ErrInvalidOrgSchema
		}

		var err error
		orgSchemaName, err = s.repo.GetSchemaNameByOrgID(principal.OrgID)
		if err != nil {
			log.Printf("Failed to get schema name for orgId %s: %v", principal.OrgID, err)
			return ErrInvalidOrgSchema
		}
		log.Printf("Looked up schema name '%s' for orgId '%s'", orgSchemaName, principal.OrgID)
	}

	user, err := s.repo.GetByID(orgSchemaName, userID)
	if err != nil {
		return err
	}

	if principal.OrgID != "" && user.OrgID != principal.OrgID {
		return ErrForbidden
	}

	err = s.keycloakAdmin.DeleteUser(user.KeycloakUserID)
	if err != nil {
		return fmt.Errorf("failed to delete user from Keycloak: %w", err)
	}

	err = s.repo.Delete(principal.OrgSchemaName, user.OrgID, userID, user.Role)
	if err != nil {
		log.Printf("WARNING: User deleted from Keycloak but failed to delete from database: %s", userID)
		return fmt.Errorf("failed to delete user from database: %w", err)
	}

	log.Printf("Successfully deleted user: %s (Keycloak ID: %s, Role: %s)", user.Email, user.KeycloakUserID, user.Role)

	return nil
}

func (s *Service) hasRole(principal *auth.Principal, role string) bool {
	for _, r := range principal.Roles {
		if r == role {
			return true
		}
	}
	return false
}
