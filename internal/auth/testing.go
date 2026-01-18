package auth

import (
	"context"
	"crypto/rsa"
)

// ContextWithPrincipal adds a principal to the context for testing purposes
// This is exported to allow other packages to create test contexts
func ContextWithPrincipal(ctx context.Context, principal *Principal) context.Context {
	return context.WithValue(ctx, principalKey, principal)
}

// NewTestJWKS creates a mock JWKS for testing that accepts tokens signed with the given public key
// This is exported to allow E2E tests to create test verifiers
func NewTestJWKS(publicKey *rsa.PublicKey) *JWKS {
	return &JWKS{
		keys: map[string]*rsa.PublicKey{
			"test-key-id": publicKey,
		},
	}
}
