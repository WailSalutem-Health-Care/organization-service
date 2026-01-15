package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// TestMiddleware_ValidToken tests that a valid token allows the request to proceed
func TestMiddleware_ValidToken(t *testing.T) {
	// Setup - create a real verifier with mock JWKS
	privateKey, publicKey := generateTestKeyPair(t)
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
		"realm_access": map[string]interface{}{
			"roles": []interface{}{"SUPER_ADMIN"},
		},
		"organizationID": "org-456",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key-id"
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	// Create middleware
	middleware := Middleware(verifier)

	// Create test handler that checks if principal was set
	called := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		principal, ok := FromContext(r.Context())
		if !ok {
			t.Error("Expected principal in context, got none")
			return
		}
		if principal.UserID != "user-123" {
			t.Errorf("Expected UserID 'user-123', got '%s'", principal.UserID)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Wrap handler with middleware
	handler := middleware(testHandler)

	// Create request with Authorization header
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(rec, req)

	// Verify
	if !called {
		t.Error("Expected handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

// TestMiddleware_MissingAuthorizationHeader tests that missing header returns 401
func TestMiddleware_MissingAuthorizationHeader(t *testing.T) {
	_, publicKey := generateTestKeyPair(t)
	mockJWKS := newMockJWKS(publicKey)
	cfg := Config{Issuer: "https://test.com"}
	verifier := NewVerifier(cfg, mockJWKS)
	
	middleware := Middleware(verifier)

	called := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// No Authorization header set
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if called {
		t.Error("Expected handler NOT to be called")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
	expectedBody := "missing authorization\n"
	if rec.Body.String() != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, rec.Body.String())
	}
}

// TestMiddleware_InvalidAuthorizationHeader tests malformed headers
func TestMiddleware_InvalidAuthorizationHeader(t *testing.T) {
	_, publicKey := generateTestKeyPair(t)
	mockJWKS := newMockJWKS(publicKey)
	cfg := Config{Issuer: "https://test.com"}
	verifier := NewVerifier(cfg, mockJWKS)
	
	testCases := []struct {
		name   string
		header string
	}{
		{"No Bearer prefix", "some-token"},
		{"Wrong prefix", "Basic dXNlcjpwYXNz"},
		{"Only Bearer", "Bearer"},
		{"Empty after Bearer", "Bearer "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			middleware := Middleware(verifier)

			called := false
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
			})

			handler := middleware(testHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tc.header)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if called {
				t.Error("Expected handler NOT to be called")
			}
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("Expected status 401, got %d", rec.Code)
			}
		})
	}
}

// TestMiddleware_InvalidToken tests that invalid tokens are rejected
func TestMiddleware_InvalidToken(t *testing.T) {
	_, publicKey := generateTestKeyPair(t)
	mockJWKS := newMockJWKS(publicKey)
	cfg := Config{Issuer: "https://test.com"}
	verifier := NewVerifier(cfg, mockJWKS)
	
	middleware := Middleware(verifier)

	called := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	handler := middleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if called {
		t.Error("Expected handler NOT to be called")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

// TestFromContext tests extracting principal from context
func TestFromContext(t *testing.T) {
	t.Run("Principal exists in context", func(t *testing.T) {
		expected := &Principal{
			UserID: "test-user",
			Roles:  []string{"ADMIN"},
		}
		ctx := context.WithValue(context.Background(), principalKey, expected)

		principal, ok := FromContext(ctx)

		if !ok {
			t.Error("Expected principal to be found")
		}
		if principal.UserID != expected.UserID {
			t.Errorf("Expected UserID '%s', got '%s'", expected.UserID, principal.UserID)
		}
	})

	t.Run("No principal in context", func(t *testing.T) {
		ctx := context.Background()

		principal, ok := FromContext(ctx)

		if ok {
			t.Error("Expected no principal to be found")
		}
		if principal != nil {
			t.Error("Expected nil principal")
		}
	})
}

// TestHasPermission tests the permission checking logic
func TestHasPermission(t *testing.T) {
	perms := Permissions{
		"SUPER_ADMIN": {"organization:create", "organization:view", "user:delete"},
		"ORG_ADMIN":   {"organization:view", "user:create"},
		"PATIENT":     {"patient:view"},
	}

	testCases := []struct {
		name       string
		principal  *Principal
		permission string
		expected   bool
	}{
		{
			name: "Single role with permission",
			principal: &Principal{
				Roles: []string{"SUPER_ADMIN"},
			},
			permission: "organization:create",
			expected:   true,
		},
		{
			name: "Single role without permission",
			principal: &Principal{
				Roles: []string{"PATIENT"},
			},
			permission: "organization:create",
			expected:   false,
		},
		{
			name: "Multiple roles, permission in first role",
			principal: &Principal{
				Roles: []string{"ORG_ADMIN", "PATIENT"},
			},
			permission: "user:create",
			expected:   true,
		},
		{
			name: "Multiple roles, permission in second role",
			principal: &Principal{
				Roles: []string{"PATIENT", "SUPER_ADMIN"},
			},
			permission: "user:delete",
			expected:   true,
		},
		{
			name: "No roles",
			principal: &Principal{
				Roles: []string{},
			},
			permission: "organization:view",
			expected:   false,
		},
		{
			name: "Unknown role",
			principal: &Principal{
				Roles: []string{"UNKNOWN_ROLE"},
			},
			permission: "organization:view",
			expected:   false,
		},
		{
			name: "Permission exists in multiple roles",
			principal: &Principal{
				Roles: []string{"SUPER_ADMIN", "ORG_ADMIN"},
			},
			permission: "organization:view",
			expected:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := HasPermission(tc.principal, tc.permission, perms)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestRequirePermission tests the permission enforcement middleware
func TestRequirePermission(t *testing.T) {
	perms := Permissions{
		"SUPER_ADMIN": {"organization:create", "organization:delete"},
		"ORG_ADMIN":   {"organization:view"},
	}

	t.Run("User has required permission", func(t *testing.T) {
		principal := &Principal{
			UserID: "user-123",
			Roles:  []string{"SUPER_ADMIN"},
		}

		middleware := RequirePermission("organization:create", perms)

		called := false
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		handler := middleware(testHandler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		ctx := context.WithValue(req.Context(), principalKey, principal)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if !called {
			t.Error("Expected handler to be called")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
	})

	t.Run("User lacks required permission", func(t *testing.T) {
		principal := &Principal{
			UserID: "user-123",
			Roles:  []string{"ORG_ADMIN"},
		}

		middleware := RequirePermission("organization:delete", perms)

		called := false
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		})

		handler := middleware(testHandler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		ctx := context.WithValue(req.Context(), principalKey, principal)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if called {
			t.Error("Expected handler NOT to be called")
		}
		if rec.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", rec.Code)
		}
	})

	t.Run("No principal in context", func(t *testing.T) {
		middleware := RequirePermission("organization:create", perms)

		called := false
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		})

		handler := middleware(testHandler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		// No principal in context
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if called {
			t.Error("Expected handler NOT to be called")
		}
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", rec.Code)
		}
	})
}

// Helper functions are defined in jwt_verify_test.go to avoid duplication
