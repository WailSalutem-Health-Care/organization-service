//go:build integration

package e2e

import (
	"crypto/rsa"
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/WailSalutem-Health-Care/organization-service/internal/auth"
	httpserver "github.com/WailSalutem-Health-Care/organization-service/internal/http"
	"github.com/WailSalutem-Health-Care/organization-service/internal/testutil"
)

// TestServer represents a complete E2E test environment
type TestServer struct {
	Server        *httptest.Server
	DB            *sql.DB
	MockPublisher *testutil.MockPublisher
	Verifier      *auth.Verifier
	PrivateKey    *rsa.PrivateKey
	MockKeycloak  *testutil.MockKeycloakAdmin
}

// SetupE2ETest creates a complete test environment for E2E testing
// This includes:
// - Real PostgreSQL database
// - Real HTTP server with all routes
// - Mock/nil RabbitMQ publisher (for now)
// - Test JWT verifier and signing key
func SetupE2ETest(t *testing.T) *TestServer {
	t.Helper()

	// Setup real database
	db := testutil.SetupTestDB(t)

	// Create mock RabbitMQ publisher (in-memory only, no real RabbitMQ calls)
	mockPublisher := testutil.NewMockPublisher()

	// Create mock Keycloak admin client (in-memory only, no real Keycloak calls)
	mockKeycloak := testutil.NewMockKeycloakAdmin()

	// Load permissions from file
	perms, err := auth.LoadPermissions("../../permissions.yml")
	if err != nil {
		t.Fatalf("Failed to load permissions: %v", err)
	}

	// Create test verifier and get private key for signing tokens
	verifier, privateKey := testutil.CreateTestVerifier(t)

	// Setup router with real dependencies, mock Keycloak, and mock publisher
	// We need to convert mockPublisher to *messaging.Publisher interface
	// Since mockPublisher implements the same Publish() method, we can use it directly
	router := httpserver.SetupRouterWithKeycloak(db, verifier, perms, mockPublisher, mockKeycloak)

	// Create test HTTP server
	server := httptest.NewServer(router)

	return &TestServer{
		Server:        server,
		DB:            db,
		MockPublisher: mockPublisher,
		Verifier:      verifier,
		PrivateKey:    privateKey,
		MockKeycloak:  mockKeycloak,
	}
}

// Cleanup cleans up all test resources
func (ts *TestServer) Cleanup(t *testing.T) {
	t.Helper()

	// Close HTTP server
	ts.Server.Close()

	// Clean up database
	testutil.CleanupTestDB(t, ts.DB)
	ts.DB.Close()
}

// GenerateSuperAdminToken generates a SUPER_ADMIN token for this test server
func (ts *TestServer) GenerateSuperAdminToken(t *testing.T) string {
	t.Helper()
	return testutil.GenerateSuperAdminToken(t, ts.PrivateKey)
}

// GenerateOrgAdminToken generates an ORG_ADMIN token for this test server
func (ts *TestServer) GenerateOrgAdminToken(t *testing.T, orgID, orgSchemaName string) string {
	t.Helper()
	return testutil.GenerateOrgAdminToken(t, ts.PrivateKey, orgID, orgSchemaName)
}

// GenerateCaregiverToken generates a CAREGIVER token for this test server
func (ts *TestServer) GenerateCaregiverToken(t *testing.T, orgID, orgSchemaName string) string {
	t.Helper()
	return testutil.GenerateCaregiverToken(t, ts.PrivateKey, orgID, orgSchemaName)
}

// NewClient creates a new HTTP test client for this server with the given token
func (ts *TestServer) NewClient(token string) *testutil.HTTPTestClient {
	return testutil.NewHTTPTestClient(ts.Server.URL, token)
}
