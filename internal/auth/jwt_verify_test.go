package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// TestVerifier_ParseAndVerifyToken_Success tests successful token parsing
func TestVerifier_ParseAndVerifyToken_Success(t *testing.T) {
	// Generate a test RSA key pair
	privateKey, publicKey := generateTestKeyPair(t)

	// Create mock JWKS
	mockJWKS := newMockJWKS(publicKey)

	cfg := Config{
		Issuer: "https://test-keycloak.com/realms/test",
	}

	verifier := NewVerifier(cfg, mockJWKS)

	// Create a valid token
	claims := jwt.MapClaims{
		"sub": "user-123",
		"iss": cfg.Issuer,
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"realm_access": map[string]interface{}{
			"roles": []interface{}{"SUPER_ADMIN", "ORG_ADMIN"},
		},
		"organizationID": "org-456",
		"orgSchemaName":  "org_456_schema",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key-id"

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	// Parse and verify
	principal, err := verifier.ParseAndVerifyToken(tokenString)

	// Verify results
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if principal == nil {
		t.Fatal("Expected principal, got nil")
	}
	if principal.UserID != "user-123" {
		t.Errorf("Expected UserID 'user-123', got '%s'", principal.UserID)
	}
	if len(principal.Roles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(principal.Roles))
	}
	if principal.Roles[0] != "SUPER_ADMIN" {
		t.Errorf("Expected first role 'SUPER_ADMIN', got '%s'", principal.Roles[0])
	}
	if principal.OrgID != "org-456" {
		t.Errorf("Expected OrgID 'org-456', got '%s'", principal.OrgID)
	}
	if principal.OrgSchemaName != "org_456_schema" {
		t.Errorf("Expected OrgSchemaName 'org_456_schema', got '%s'", principal.OrgSchemaName)
	}
}

// TestVerifier_ParseAndVerifyToken_EmptyToken tests empty token
func TestVerifier_ParseAndVerifyToken_EmptyToken(t *testing.T) {
	cfg := Config{Issuer: "https://test.com"}
	verifier := NewVerifier(cfg, nil)

	principal, err := verifier.ParseAndVerifyToken("")

	if err != ErrNoToken {
		t.Errorf("Expected ErrNoToken, got: %v", err)
	}
	if principal != nil {
		t.Error("Expected nil principal")
	}
}

// TestVerifier_ParseAndVerifyToken_InvalidIssuer tests wrong issuer
func TestVerifier_ParseAndVerifyToken_InvalidIssuer(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	mockJWKS := newMockJWKS(publicKey)

	cfg := Config{
		Issuer: "https://correct-issuer.com/realms/test",
	}
	verifier := NewVerifier(cfg, mockJWKS)

	// Create token with wrong issuer
	claims := jwt.MapClaims{
		"sub": "user-123",
		"iss": "https://wrong-issuer.com/realms/test",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key-id"

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	principal, err := verifier.ParseAndVerifyToken(tokenString)

	if err != ErrInvalidIssuer {
		t.Errorf("Expected ErrInvalidIssuer, got: %v", err)
	}
	if principal != nil {
		t.Error("Expected nil principal")
	}
}

// TestVerifier_ParseAndVerifyToken_ExpiredToken tests expired token
func TestVerifier_ParseAndVerifyToken_ExpiredToken(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	mockJWKS := newMockJWKS(publicKey)

	cfg := Config{
		Issuer: "https://test-keycloak.com/realms/test",
	}
	verifier := NewVerifier(cfg, mockJWKS)

	// Create expired token
	claims := jwt.MapClaims{
		"sub": "user-123",
		"iss": cfg.Issuer,
		"exp": time.Now().Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
		"iat": time.Now().Add(-2 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key-id"

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	principal, err := verifier.ParseAndVerifyToken(tokenString)

	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got: %v", err)
	}
	if principal != nil {
		t.Error("Expected nil principal")
	}
}

// TestVerifier_ParseAndVerifyToken_MissingSubClaim tests missing sub claim
func TestVerifier_ParseAndVerifyToken_MissingSubClaim(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	mockJWKS := newMockJWKS(publicKey)

	cfg := Config{
		Issuer: "https://test-keycloak.com/realms/test",
	}
	verifier := NewVerifier(cfg, mockJWKS)

	// Create token without sub claim
	claims := jwt.MapClaims{
		"iss": cfg.Issuer,
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key-id"

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	principal, err := verifier.ParseAndVerifyToken(tokenString)

	if err != ErrMissingSub {
		t.Errorf("Expected ErrMissingSub, got: %v", err)
	}
	if principal != nil {
		t.Error("Expected nil principal")
	}
}

// TestVerifier_ParseAndVerifyToken_NoKid tests token without key ID
func TestVerifier_ParseAndVerifyToken_NoKid(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	mockJWKS := newMockJWKS(publicKey)

	cfg := Config{
		Issuer: "https://test-keycloak.com/realms/test",
	}
	verifier := NewVerifier(cfg, mockJWKS)

	// Create token without kid in header
	claims := jwt.MapClaims{
		"sub": "user-123",
		"iss": cfg.Issuer,
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	// Don't set kid in header

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	principal, err := verifier.ParseAndVerifyToken(tokenString)

	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got: %v", err)
	}
	if principal != nil {
		t.Error("Expected nil principal")
	}
}

// TestVerifier_ParseAndVerifyToken_NoRoles tests token without realm_access roles
func TestVerifier_ParseAndVerifyToken_NoRoles(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	mockJWKS := newMockJWKS(publicKey)

	cfg := Config{
		Issuer: "https://test-keycloak.com/realms/test",
	}
	verifier := NewVerifier(cfg, mockJWKS)

	// Create token without realm_access
	claims := jwt.MapClaims{
		"sub": "user-123",
		"iss": cfg.Issuer,
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key-id"

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	principal, err := verifier.ParseAndVerifyToken(tokenString)

	// Should succeed but with empty roles
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if principal == nil {
		t.Fatal("Expected principal, got nil")
	}
	if len(principal.Roles) != 0 {
		t.Errorf("Expected 0 roles, got %d", len(principal.Roles))
	}
}

// TestVerifier_ParseAndVerifyToken_NoOrgClaims tests token without org claims
func TestVerifier_ParseAndVerifyToken_NoOrgClaims(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	mockJWKS := newMockJWKS(publicKey)

	cfg := Config{
		Issuer: "https://test-keycloak.com/realms/test",
	}
	verifier := NewVerifier(cfg, mockJWKS)

	// Create token without org claims
	claims := jwt.MapClaims{
		"sub": "user-123",
		"iss": cfg.Issuer,
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"realm_access": map[string]interface{}{
			"roles": []interface{}{"PATIENT"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key-id"

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	principal, err := verifier.ParseAndVerifyToken(tokenString)

	// Should succeed but with empty org fields
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if principal == nil {
		t.Fatal("Expected principal, got nil")
	}
	if principal.OrgID != "" {
		t.Errorf("Expected empty OrgID, got '%s'", principal.OrgID)
	}
	if principal.OrgSchemaName != "" {
		t.Errorf("Expected empty OrgSchemaName, got '%s'", principal.OrgSchemaName)
	}
}

// Helper functions

// generateTestKeyPair generates an RSA key pair for testing
func generateTestKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()
	
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}
	return privateKey, &privateKey.PublicKey
}

// newMockJWKS creates a mock JWKS for testing
func newMockJWKS(publicKey *rsa.PublicKey) *JWKS {
	return &JWKS{
		keys: map[string]*rsa.PublicKey{
			"test-key-id": publicKey,
		},
	}
}
