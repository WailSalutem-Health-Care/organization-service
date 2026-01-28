package auth

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type ctxKey string

const principalKey ctxKey = "auth_principal"

var tracer = otel.Tracer("github.com/WailSalutem-Health-Care/organization-service/auth")

// MetricsRecorder interface for recording auth metrics
type MetricsRecorder interface {
	RecordAuthFailure(ctx context.Context, reason string)
}

// Middleware validates token, injects Principal into request context.
// verifier should be created with NewVerifier.
func Middleware(ver *Verifier) func(http.Handler) http.Handler {
	return MiddlewareWithMetrics(ver, nil)
}

// MiddlewareWithMetrics validates token with metrics recording
func MiddlewareWithMetrics(ver *Verifier, metrics MetricsRecorder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx, span := tracer.Start(ctx, "auth.Middleware",
				trace.WithSpanKind(trace.SpanKindInternal),
			)
			defer span.End()

			authz := r.Header.Get("Authorization")
			if authz == "" {
				span.SetStatus(codes.Error, "missing authorization")
				span.SetAttributes(attribute.String("error.type", "missing_authorization"))
				if metrics != nil {
					metrics.RecordAuthFailure(ctx, "missing_authorization")
				}
				http.Error(w, "missing authorization", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authz, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				span.SetStatus(codes.Error, "invalid authorization header")
				span.SetAttributes(attribute.String("error.type", "invalid_header_format"))
				if metrics != nil {
					metrics.RecordAuthFailure(ctx, "invalid_header_format")
				}
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}

			tok := parts[1]
			pr, err := ver.ParseAndVerifyToken(tok)
			if err != nil {
				log.Printf("[ERROR] Token validation failed: %v", err)
				span.SetStatus(codes.Error, "token validation failed")
				span.SetAttributes(
					attribute.String("error.type", "invalid_token"),
					attribute.String("error.message", err.Error()),
				)
				if metrics != nil {
					metrics.RecordAuthFailure(ctx, "invalid_token")
				}
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			// Add principal information to span
			email := ""
			if emailClaim, ok := pr.Claims["email"].(string); ok {
				email = emailClaim
			}
			span.SetAttributes(
				attribute.String("user.id", pr.UserID),
				attribute.String("user.email", email),
				attribute.StringSlice("user.roles", pr.Roles),
				attribute.String("organization.id", pr.OrgID),
			)
			span.SetStatus(codes.Ok, "authentication successful")

			ctx = context.WithValue(ctx, principalKey, pr)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// PermissionMetricsRecorder interface for recording permission check metrics
type PermissionMetricsRecorder interface {
	RecordPermissionCheck(ctx context.Context, permission string, durationMs float64, allowed bool)
}

// RequirePermission returns middleware that ensures the principal has permission.
func RequirePermission(per string, perms Permissions) func(http.Handler) http.Handler {
	return RequirePermissionWithMetrics(per, perms, nil)
}

// RequirePermissionWithMetrics returns middleware with metrics recording
func RequirePermissionWithMetrics(per string, perms Permissions, metrics PermissionMetricsRecorder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			start := time.Now()
			ctx, span := tracer.Start(ctx, "auth.RequirePermission",
				trace.WithSpanKind(trace.SpanKindInternal),
				trace.WithAttributes(attribute.String("permission.required", per)),
			)
			defer span.End()

			pr, ok := FromContext(ctx)
			if !ok {
				span.SetStatus(codes.Error, "unauthenticated")
				if metrics != nil {
					metrics.RecordPermissionCheck(ctx, per, float64(time.Since(start).Milliseconds()), false)
				}
				http.Error(w, "unauthenticated", http.StatusUnauthorized)
				return
			}

			allowed := HasPermission(pr, per, perms)
			duration := float64(time.Since(start).Milliseconds())

			span.SetAttributes(
				attribute.Bool("permission.allowed", allowed),
				attribute.String("user.id", pr.UserID),
				attribute.StringSlice("user.roles", pr.Roles),
			)

			if metrics != nil {
				metrics.RecordPermissionCheck(ctx, per, duration, allowed)
			}

			if !allowed {
				log.Printf("[PERMISSION DENIED] User: %s, Roles: %v, Required Permission: %s, Available Permissions: %v", 
					pr.UserID, pr.Roles, per, perms)
				span.SetStatus(codes.Error, "forbidden")
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			span.SetStatus(codes.Ok, "permission granted")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FromContext extracts Principal from context.
func FromContext(ctx context.Context) (*Principal, bool) {
	pr, ok := ctx.Value(principalKey).(*Principal)
	return pr, ok
}

// HasPermission checks roles -> permissions mapping.
// Role lookup is case-insensitive so Keycloak realm roles (e.g. "patient") match permissions.yml (e.g. "PATIENT").
func HasPermission(pr *Principal, permission string, perms Permissions) bool {
	roleSet := map[string]struct{}{}
	for _, r := range pr.Roles {
		roleSet[r] = struct{}{}
	}
	for role := range roleSet {
		// Try exact match first, then uppercase (permissions.yml uses PATIENT, ORG_ADMIN, etc.)
		pList, ok := perms[role]
		if !ok {
			pList, ok = perms[strings.ToUpper(role)]
		}
		if ok {
			for _, p := range pList {
				if p == permission {
					return true
				}
			}
		}
	}
	return false
}
