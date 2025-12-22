package auth

import (
	"context"
	"net/http"
	"strings"
)

type ctxKey string

const principalKey ctxKey = "auth_principal"

// Middleware validates token, injects Principal into request context.
// verifier should be created with NewVerifier.
func Middleware(ver *verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authz := r.Header.Get("Authorization")
			if authz == "" {
				http.Error(w, "missing authorization", http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(authz, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}
			tok := parts[1]
			pr, err := ver.ParseAndVerifyToken(tok)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), principalKey, pr)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission returns middleware that ensures the principal has permission.
func RequirePermission(per string, perms Permissions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pr, ok := FromContext(r.Context())
			if !ok {
				http.Error(w, "unauthenticated", http.StatusUnauthorized)
				return
			}
			if !HasPermission(pr, per, perms) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// FromContext extracts Principal from context.
func FromContext(ctx context.Context) (*Principal, bool) {
	pr, ok := ctx.Value(principalKey).(*Principal)
	return pr, ok
}

// HasPermission checks roles -> permissions mapping.
func HasPermission(pr *Principal, permission string, perms Permissions) bool {
	roleSet := map[string]struct{}{}
	for _, r := range pr.Roles {
		roleSet[r] = struct{}{}
	}
	for role := range roleSet {
		if pList, ok := perms[role]; ok {
			for _, p := range pList {
				if p == permission {
					return true
				}
			}
		}
	}
	return false
}
