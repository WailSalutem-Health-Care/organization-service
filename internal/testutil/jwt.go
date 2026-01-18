package testutil

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// GenerateTestKeyPair generates an RSA key pair for testing JWT tokens
func GenerateTestKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}
	return privateKey, &privateKey.PublicKey
}

// GenerateTestJWT creates a valid JWT token for E2E testing
// This generates a token with the specified user ID, org ID, and roles
func GenerateTestJWT(t *testing.T, privateKey *rsa.PrivateKey, userID, orgID, orgSchemaName string, roles []string) string {
	t.Helper()

	// Create claims matching your auth system
	claims := jwt.MapClaims{
		"sub": userID,
		"iss": "https://test-keycloak.com/realms/test",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"realm_access": map[string]interface{}{
			"roles": interfaceSlice(roles),
		},
	}

	// Add org claims if provided
	if orgID != "" {
		claims["organizationID"] = orgID
	}
	if orgSchemaName != "" {
		claims["orgSchemaName"] = orgSchemaName
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key-id"

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	return tokenString
}

// GenerateSuperAdminToken creates a SUPER_ADMIN token for testing
func GenerateSuperAdminToken(t *testing.T, privateKey *rsa.PrivateKey) string {
	t.Helper()
	return GenerateTestJWT(t, privateKey, "admin-123", "", "", []string{"SUPER_ADMIN"})
}

// GenerateOrgAdminToken creates an ORG_ADMIN token for testing
func GenerateOrgAdminToken(t *testing.T, privateKey *rsa.PrivateKey, orgID, orgSchemaName string) string {
	t.Helper()
	return GenerateTestJWT(t, privateKey, "orgadmin-123", orgID, orgSchemaName, []string{"ORG_ADMIN"})
}

// GenerateCaregiverToken creates a CAREGIVER token for testing
func GenerateCaregiverToken(t *testing.T, privateKey *rsa.PrivateKey, orgID, orgSchemaName string) string {
	t.Helper()
	return GenerateTestJWT(t, privateKey, "caregiver-123", orgID, orgSchemaName, []string{"CAREGIVER"})
}

// GeneratePatientToken creates a PATIENT token for testing
func GeneratePatientToken(t *testing.T, privateKey *rsa.PrivateKey, orgID, orgSchemaName string) string {
	t.Helper()
	return GenerateTestJWT(t, privateKey, "patient-123", orgID, orgSchemaName, []string{"PATIENT"})
}

// interfaceSlice converts []string to []interface{} for JWT claims
func interfaceSlice(strings []string) []interface{} {
	result := make([]interface{}, len(strings))
	for i, s := range strings {
		result[i] = s
	}
	return result
}
