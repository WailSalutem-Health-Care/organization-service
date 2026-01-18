package auth

import "context"

// ContextWithPrincipal adds a principal to the context for testing purposes
// This is exported to allow other packages to create test contexts
func ContextWithPrincipal(ctx context.Context, principal *Principal) context.Context {
	return context.WithValue(ctx, principalKey, principal)
}
