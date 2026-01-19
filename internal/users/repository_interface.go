package users

// RepositoryInterface defines the contract for user data access
type RepositoryInterface interface {
	GetSchemaNameByOrgID(orgID string) (string, error)
	ValidateOrgSchema(schemaName string) error
	Create(user *User) error
	GetByID(schemaName, userID string) (*User, error)
	GetByKeycloakID(schemaName, keycloakUserID string) (*User, error)
	List(schemaName string) ([]User, error)
	ListWithPagination(schemaName string, limit, offset int, search string) ([]User, int, error)
	ListActiveUsersByRoleWithPagination(schemaName string, role string, limit, offset int, search string) ([]User, int, error)
	Update(user *User) error
	Delete(schemaName, orgID, userID string, role string) error
}

// Ensure Repository implements RepositoryInterface
var _ RepositoryInterface = (*Repository)(nil)
