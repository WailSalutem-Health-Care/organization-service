package testutil

import (
	"crypto/rsa"
	"testing"

	"github.com/WailSalutem-Health-Care/organization-service/internal/auth"
)

// CreateTestVerifier creates a verifier configured for E2E testing
// It returns the verifier and the private key to sign test tokens
func CreateTestVerifier(t *testing.T) (*auth.Verifier, *rsa.PrivateKey) {
	t.Helper()

	// Generate test key pair
	privateKey, publicKey := GenerateTestKeyPair(t)

	// Create test JWKS that will accept our test tokens
	// We use the auth package's JWKS struct directly
	testJWKS := auth.NewTestJWKS(publicKey)

	// Create config matching test tokens
	cfg := auth.Config{
		Issuer: "https://test-keycloak.com/realms/test",
	}

	// Create verifier
	verifier := auth.NewVerifier(cfg, testJWKS)

	return verifier, privateKey
}
