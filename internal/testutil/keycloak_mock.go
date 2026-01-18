package testutil

import (
	"fmt"
	"sync"

	"github.com/WailSalutem-Health-Care/organization-service/internal/auth"
	"github.com/google/uuid"
)

// MockKeycloakAdmin is a mock implementation of Keycloak admin client for testing
// It stores all data in memory and doesn't make any real HTTP calls to Keycloak
type MockKeycloakAdmin struct {
	mu    sync.RWMutex
	users map[string]*auth.KeycloakUser // userID -> user
	roles map[string]*auth.KeycloakRole // roleName -> role
}

// NewMockKeycloakAdmin creates a new mock Keycloak admin client
func NewMockKeycloakAdmin() *MockKeycloakAdmin {
	mock := &MockKeycloakAdmin{
		users: make(map[string]*auth.KeycloakUser),
		roles: make(map[string]*auth.KeycloakRole),
	}

	// Pre-populate standard roles
	mock.roles["SUPER_ADMIN"] = &auth.KeycloakRole{ID: "role-super-admin", Name: "SUPER_ADMIN"}
	mock.roles["ORG_ADMIN"] = &auth.KeycloakRole{ID: "role-org-admin", Name: "ORG_ADMIN"}
	mock.roles["CAREGIVER"] = &auth.KeycloakRole{ID: "role-caregiver", Name: "CAREGIVER"}
	mock.roles["PATIENT"] = &auth.KeycloakRole{ID: "role-patient", Name: "PATIENT"}
	mock.roles["MUNICIPALITY"] = &auth.KeycloakRole{ID: "role-municipality", Name: "MUNICIPALITY"}
	mock.roles["INSURER"] = &auth.KeycloakRole{ID: "role-insurer", Name: "INSURER"}

	return mock
}

// CreateUser creates a user in the mock Keycloak (in-memory only)
func (m *MockKeycloakAdmin) CreateUser(user auth.KeycloakUser) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate a fake Keycloak user ID
	userID := uuid.New().String()

	// Store user in memory
	user.ID = userID
	user.Enabled = true // Default to enabled
	m.users[userID] = &user

	return userID, nil
}

// SetPassword sets a password for a user (no-op in mock)
func (m *MockKeycloakAdmin) SetPassword(userID, password string, temporary bool) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Verify user exists
	if _, exists := m.users[userID]; !exists {
		return auth.ErrUserNotFound
	}

	// In mock, we don't actually store passwords, just return success
	return nil
}

// SendEmailAction sends an email action to a user (no-op in mock)
func (m *MockKeycloakAdmin) SendEmailAction(userID string, actions []string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Verify user exists
	if _, exists := m.users[userID]; !exists {
		return auth.ErrUserNotFound
	}

	// In mock, we don't actually send emails, just return success
	return nil
}

// GetRole retrieves a role by name
func (m *MockKeycloakAdmin) GetRole(roleName string) (*auth.KeycloakRole, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	role, exists := m.roles[roleName]
	if !exists {
		return nil, auth.ErrRoleNotFound
	}

	return role, nil
}

// AssignRole assigns a role to a user (no-op in mock, just validates)
func (m *MockKeycloakAdmin) AssignRole(userID string, role auth.KeycloakRole) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Verify user exists
	if _, exists := m.users[userID]; !exists {
		return auth.ErrUserNotFound
	}

	// In mock, we don't track role assignments, just return success
	return nil
}

// DeleteUser deletes a user from mock Keycloak (removes from memory)
func (m *MockKeycloakAdmin) DeleteUser(userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Verify user exists
	if _, exists := m.users[userID]; !exists {
		return auth.ErrUserNotFound
	}

	// Remove from memory
	delete(m.users, userID)

	return nil
}

// UpdateUser updates a user in mock Keycloak
func (m *MockKeycloakAdmin) UpdateUser(userID string, user auth.KeycloakUser) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Verify user exists
	existing, exists := m.users[userID]
	if !exists {
		return auth.ErrUserNotFound
	}

	// Update fields
	if user.Email != "" {
		existing.Email = user.Email
	}
	if user.FirstName != "" {
		existing.FirstName = user.FirstName
	}
	if user.LastName != "" {
		existing.LastName = user.LastName
	}
	if user.Username != "" {
		existing.Username = user.Username
	}

	return nil
}

// GetUser retrieves a user by ID
func (m *MockKeycloakAdmin) GetUser(userID string) (*auth.KeycloakUser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[userID]
	if !exists {
		return nil, auth.ErrUserNotFound
	}

	// Return a copy to prevent external modifications
	userCopy := *user
	return &userCopy, nil
}

// Helper methods for testing

// GetAllUsers returns all users (for test verification)
func (m *MockKeycloakAdmin) GetAllUsers() map[string]*auth.KeycloakUser {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy
	users := make(map[string]*auth.KeycloakUser, len(m.users))
	for k, v := range m.users {
		userCopy := *v
		users[k] = &userCopy
	}
	return users
}

// GetUserCount returns the number of users (for test verification)
func (m *MockKeycloakAdmin) GetUserCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.users)
}

// UserExists checks if a user exists (for test verification)
func (m *MockKeycloakAdmin) UserExists(userID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.users[userID]
	return exists
}

// Reset clears all users (for test cleanup)
func (m *MockKeycloakAdmin) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.users = make(map[string]*auth.KeycloakUser)
}

// Ensure MockKeycloakAdmin implements both interfaces
var _ interface {
	CreateUser(user auth.KeycloakUser) (string, error)
	SetPassword(userID, password string, temporary bool) error
	SendEmailAction(userID string, actions []string) error
	GetRole(roleName string) (*auth.KeycloakRole, error)
	AssignRole(userID string, role auth.KeycloakRole) error
	DeleteUser(userID string) error
	UpdateUser(userID string, user auth.KeycloakUser) error
	GetUser(userID string) (*auth.KeycloakUser, error)
} = (*MockKeycloakAdmin)(nil)

// Error wrapper for better error messages in tests
func keycloakError(operation, detail string) error {
	return fmt.Errorf("mock keycloak %s failed: %s", operation, detail)
}
