package patient

import "github.com/WailSalutem-Health-Care/organization-service/internal/auth"

// KeycloakAdminInterface defines the contract for Keycloak operations
type KeycloakAdminInterface interface {
	CreateUser(user auth.KeycloakUser) (string, error)
	SetPassword(userID, password string, temporary bool) error
	SendEmailAction(userID string, actions []string) error
	GetRole(roleName string) (*auth.KeycloakRole, error)
	AssignRole(userID string, role auth.KeycloakRole) error
	DeleteUser(userID string) error
}

// Ensure KeycloakAdminClient implements KeycloakAdminInterface
var _ KeycloakAdminInterface = (*auth.KeycloakAdminClient)(nil)
