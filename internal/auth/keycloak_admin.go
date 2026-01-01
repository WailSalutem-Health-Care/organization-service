package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	ErrKeycloakRequest = errors.New("keycloak request failed")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrUserNotFound    = errors.New("user not found")
	ErrRoleNotFound    = errors.New("role not found")
	ErrInvalidResponse = errors.New("invalid response from keycloak")
)

// KeycloakAdminClient handles administrative operations in Keycloak
type KeycloakAdminClient struct {
	baseURL      string
	realm        string
	clientID     string
	clientSecret string
	httpClient   *http.Client

	tokenMux    sync.RWMutex
	accessToken string
	tokenExpiry time.Time
}

// KeycloakUser represents a user in Keycloak
type KeycloakUser struct {
	ID         string              `json:"id,omitempty"`
	Username   string              `json:"username"`
	Email      string              `json:"email"`
	FirstName  string              `json:"firstName"`
	LastName   string              `json:"lastName"`
	Enabled    bool                `json:"enabled"`
	Attributes map[string][]string `json:"attributes,omitempty"`
}

// KeycloakRole represents a role in Keycloak
type KeycloakRole struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// PasswordReset represents password reset request
type PasswordReset struct {
	Type      string `json:"type"`
	Value     string `json:"value"`
	Temporary bool   `json:"temporary"`
}

// NewKeycloakAdminClient creates a new Keycloak admin client
func NewKeycloakAdminClient() (*KeycloakAdminClient, error) {
	baseURL := os.Getenv("KEYCLOAK_BASE_URL")
	realm := os.Getenv("KEYCLOAK_REALM")
	clientID := os.Getenv("KEYCLOAK_ADMIN_CLIENT_ID")
	clientSecret := os.Getenv("KEYCLOAK_ADMIN_CLIENT_SECRET")

	if baseURL == "" || realm == "" || clientID == "" || clientSecret == "" {
		return nil, errors.New("missing required Keycloak admin configuration")
	}

	// Remove trailing slash from baseURL
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &KeycloakAdminClient{
		baseURL:      baseURL,
		realm:        realm,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// getAdminToken obtains an admin access token using client credentials
func (k *KeycloakAdminClient) getAdminToken() (string, error) {
	k.tokenMux.RLock()
	if k.accessToken != "" && time.Now().Before(k.tokenExpiry) {
		token := k.accessToken
		k.tokenMux.RUnlock()
		return token, nil
	}
	k.tokenMux.RUnlock()

	k.tokenMux.Lock()
	defer k.tokenMux.Unlock()

	// Double check after acquiring write lock
	if k.accessToken != "" && time.Now().Before(k.tokenExpiry) {
		return k.accessToken, nil
	}

	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", k.baseURL, k.realm)

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", k.clientID)
	data.Set("client_secret", k.clientSecret)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Failed to get admin token: %d - %s", resp.StatusCode, string(body))
		return "", ErrUnauthorized
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	// Store token with 60 second buffer before expiry
	k.accessToken = result.AccessToken
	k.tokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn-60) * time.Second)

	log.Printf("Obtained new Keycloak admin token (expires in %d seconds)", result.ExpiresIn)

	return k.accessToken, nil
}

// CreateUser creates a new user in Keycloak
func (k *KeycloakAdminClient) CreateUser(user KeycloakUser) (string, error) {
	token, err := k.getAdminToken()
	if err != nil {
		return "", err
	}

	createURL := fmt.Sprintf("%s/admin/realms/%s/users", k.baseURL, k.realm)

	body, err := json.Marshal(user)
	if err != nil {
		return "", fmt.Errorf("failed to marshal user: %w", err)
	}

	log.Printf("Creating Keycloak user with payload: %s", string(body))

	req, err := http.NewRequest("POST", createURL, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Failed to create user: %d - %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("%w: status %d", ErrKeycloakRequest, resp.StatusCode)
	}

	// Extract user ID from Location header
	location := resp.Header.Get("Location")
	if location == "" {
		return "", ErrInvalidResponse
	}

	// Location format: .../users/{userId}
	parts := strings.Split(location, "/")
	if len(parts) == 0 {
		return "", ErrInvalidResponse
	}
	userID := parts[len(parts)-1]

	log.Printf("Created user in Keycloak: %s (ID: %s)", user.Username, userID)

	return userID, nil
}

// SetPassword sets or resets a user's password
func (k *KeycloakAdminClient) SetPassword(userID string, password string, temporary bool) error {
	token, err := k.getAdminToken()
	if err != nil {
		return err
	}

	resetURL := fmt.Sprintf("%s/admin/realms/%s/users/%s/reset-password", k.baseURL, k.realm, userID)

	passwordReset := PasswordReset{
		Type:      "password",
		Value:     password,
		Temporary: temporary,
	}

	body, err := json.Marshal(passwordReset)
	if err != nil {
		return fmt.Errorf("failed to marshal password reset: %w", err)
	}

	req, err := http.NewRequest("PUT", resetURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to set password: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Failed to set password: %d - %s", resp.StatusCode, string(body))
		return fmt.Errorf("%w: status %d", ErrKeycloakRequest, resp.StatusCode)
	}

	log.Printf("Set password for user %s (temporary: %v)", userID, temporary)

	return nil
}

// GetRole fetches a realm role by name
func (k *KeycloakAdminClient) GetRole(roleName string) (*KeycloakRole, error) {
	token, err := k.getAdminToken()
	if err != nil {
		return nil, err
	}

	roleURL := fmt.Sprintf("%s/admin/realms/%s/roles/%s", k.baseURL, k.realm, roleName)

	req, err := http.NewRequest("GET", roleURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrRoleNotFound
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Failed to get role: %d - %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("%w: status %d", ErrKeycloakRequest, resp.StatusCode)
	}

	var role KeycloakRole
	if err := json.NewDecoder(resp.Body).Decode(&role); err != nil {
		return nil, fmt.Errorf("failed to decode role: %w", err)
	}

	return &role, nil
}

// AssignRole assigns a realm role to a user
func (k *KeycloakAdminClient) AssignRole(userID string, role KeycloakRole) error {
	token, err := k.getAdminToken()
	if err != nil {
		return err
	}

	assignURL := fmt.Sprintf("%s/admin/realms/%s/users/%s/role-mappings/realm", k.baseURL, k.realm, userID)

	// Must be an array of roles
	roles := []KeycloakRole{role}
	body, err := json.Marshal(roles)
	if err != nil {
		return fmt.Errorf("failed to marshal role: %w", err)
	}

	req, err := http.NewRequest("POST", assignURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Failed to assign role: %d - %s", resp.StatusCode, string(body))
		return fmt.Errorf("%w: status %d", ErrKeycloakRequest, resp.StatusCode)
	}

	log.Printf("Assigned role %s to user %s", role.Name, userID)

	return nil
}

// DeleteUser deletes a user from Keycloak (for rollback)
func (k *KeycloakAdminClient) DeleteUser(userID string) error {
	token, err := k.getAdminToken()
	if err != nil {
		return err
	}

	deleteURL := fmt.Sprintf("%s/admin/realms/%s/users/%s", k.baseURL, k.realm, userID)

	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Failed to delete user: %d - %s", resp.StatusCode, string(body))
		return fmt.Errorf("%w: status %d", ErrKeycloakRequest, resp.StatusCode)
	}

	log.Printf("Deleted user from Keycloak: %s", userID)

	return nil
}

// SendEmailAction sends an email action to a user (e.g., UPDATE_PASSWORD)
func (k *KeycloakAdminClient) SendEmailAction(userID string, actions []string) error {
	token, err := k.getAdminToken()
	if err != nil {
		return err
	}

	emailURL := fmt.Sprintf("%s/admin/realms/%s/users/%s/execute-actions-email", k.baseURL, k.realm, userID)

	body, err := json.Marshal(actions)
	if err != nil {
		return fmt.Errorf("failed to marshal actions: %w", err)
	}

	req, err := http.NewRequest("PUT", emailURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email action: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Failed to send email action: %d - %s", resp.StatusCode, string(body))
		return fmt.Errorf("%w: status %d", ErrKeycloakRequest, resp.StatusCode)
	}

	log.Printf("Sent email action to user %s: %v", userID, actions)

	return nil
}
