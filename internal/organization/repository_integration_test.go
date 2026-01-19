// go:build integration
//go:build integration

package organization

import (
	"context"
	"testing"

	"github.com/WailSalutem-Health-Care/organization-service/internal/testutil"
)

// TestRepositoryCreateOrganization_Integration tests creating an organization with real database
func TestRepositoryCreateOrganization_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	repo := NewRepository(db, nil)

	req := CreateOrganizationRequest{
		Name:         "Test Hospital",
		ContactEmail: "test@hospital.com",
		ContactPhone: "+1234567890",
		Address:      "123 Test St",
	}

	org, err := repo.CreateOrganization(context.Background(), req)

	if err != nil {
		t.Fatalf("CreateOrganization failed: %v", err)
	}

	if org.ID == "" {
		t.Error("Expected organization ID to be set")
	}

	if org.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, org.Name)
	}

	if org.SchemaName == "" {
		t.Error("Expected schema name to be set")
	}

	if org.Status != "active" {
		t.Errorf("Expected status 'active', got %s", org.Status)
	}
}

// TestRepositoryCreateOrganization_MultipleOrgs_Integration tests creating multiple organizations
func TestRepositoryCreateOrganization_MultipleOrgs_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	repo := NewRepository(db, nil)

	// Create multiple organizations with same name (allowed since name is not unique)
	for i := 1; i <= 3; i++ {
		req := CreateOrganizationRequest{
			Name:         "Test Hospital",  // Same name is allowed
			ContactEmail: "test@hospital.com",
		}

		org, err := repo.CreateOrganization(context.Background(), req)
		if err != nil {
			t.Fatalf("CreateOrganization %d failed: %v", i, err)
		}

		if org.SchemaName == "" {
			t.Errorf("Expected unique schema name for org %d", i)
		}
	}
}

// TestRepositoryGetOrganization_Integration tests retrieving an organization
func TestRepositoryGetOrganization_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	repo := NewRepository(db, nil)

	// Create organization
	created, err := repo.CreateOrganization(context.Background(), CreateOrganizationRequest{
		Name:         "Get Test Hospital",
		ContactEmail: "get@test.com",
	})
	if err != nil {
		t.Fatalf("CreateOrganization failed: %v", err)
	}

	// Get organization
	org, err := repo.GetOrganization(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetOrganization failed: %v", err)
	}

	if org.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, org.ID)
	}

	if org.Name != created.Name {
		t.Errorf("Expected name %s, got %s", created.Name, org.Name)
	}
}

// TestRepositoryGetOrganization_NotFound_Integration tests getting non-existent organization
func TestRepositoryGetOrganization_NotFound_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	repo := NewRepository(db, nil)

	_, err := repo.GetOrganization(context.Background(), "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Error("Expected error for non-existent organization, got nil")
	}
}

// TestRepositoryListOrganizations_Integration tests listing organizations
func TestRepositoryListOrganizations_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	repo := NewRepository(db, nil)

	// Create multiple organizations
	for i := 1; i <= 3; i++ {
		_, err := repo.CreateOrganization(context.Background(), CreateOrganizationRequest{
			Name:         "List Test Hospital " + string(rune('A'+i-1)),
			ContactEmail: "list@test.com",
		})
		if err != nil {
			t.Fatalf("CreateOrganization %d failed: %v", i, err)
		}
	}

	// List organizations
	orgs, err := repo.ListOrganizations(context.Background())
	if err != nil {
		t.Fatalf("ListOrganizations failed: %v", err)
	}

	if len(orgs) < 3 {
		t.Errorf("Expected at least 3 organizations, got %d", len(orgs))
	}
}

// TestRepositoryListOrganizationsWithPagination_Integration tests paginated listing
func TestRepositoryListOrganizationsWithPagination_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	repo := NewRepository(db, nil)

	// Create 5 organizations
	for i := 1; i <= 5; i++ {
		_, err := repo.CreateOrganization(context.Background(), CreateOrganizationRequest{
			Name:         "Pagination Test " + string(rune('A'+i-1)),
			ContactEmail: "page@test.com",
		})
		if err != nil {
			t.Fatalf("CreateOrganization %d failed: %v", i, err)
		}
	}

	// Get first page (limit 2)
	orgs, total, err := repo.ListOrganizationsWithPagination(context.Background(), 2, 0, "", "")
	if err != nil {
		t.Fatalf("ListOrganizationsWithPagination failed: %v", err)
	}

	if len(orgs) != 2 {
		t.Errorf("Expected 2 organizations on first page, got %d", len(orgs))
	}

	if total < 5 {
		t.Errorf("Expected total >= 5, got %d", total)
	}

	// Get second page (limit 2, offset 2)
	orgs, _, err = repo.ListOrganizationsWithPagination(context.Background(), 2, 2, "", "")
	if err != nil {
		t.Fatalf("ListOrganizationsWithPagination page 2 failed: %v", err)
	}

	if len(orgs) != 2 {
		t.Errorf("Expected 2 organizations on second page, got %d", len(orgs))
	}
}

// TestRepositoryUpdateOrganization_Integration tests updating an organization
func TestRepositoryUpdateOrganization_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	repo := NewRepository(db, nil)

	// Create organization
	created, err := repo.CreateOrganization(context.Background(), CreateOrganizationRequest{
		Name:         "Update Test Hospital",
		ContactEmail: "update@test.com",
	})
	if err != nil {
		t.Fatalf("CreateOrganization failed: %v", err)
	}

	// Update organization
	name := "Updated Hospital Name"
	email := "updated@test.com"
	phone := "+9876543210"
	updateReq := UpdateOrganizationRequest{
		Name:         &name,
		ContactEmail: &email,
		ContactPhone: &phone,
	}

	updated, err := repo.UpdateOrganization(context.Background(), created.ID, updateReq)
	if err != nil {
		t.Fatalf("UpdateOrganization failed: %v", err)
	}

	if updated.Name != *updateReq.Name {
		t.Errorf("Expected name %s, got %s", *updateReq.Name, updated.Name)
	}

	if updated.ContactEmail != *updateReq.ContactEmail {
		t.Errorf("Expected email %s, got %s", *updateReq.ContactEmail, updated.ContactEmail)
	}
}

// TestRepositoryDeleteOrganization_Integration tests soft deleting an organization
func TestRepositoryDeleteOrganization_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	repo := NewRepository(db, nil)

	// Create organization
	created, err := repo.CreateOrganization(context.Background(), CreateOrganizationRequest{
		Name:         "Delete Test Hospital",
		ContactEmail: "delete@test.com",
	})
	if err != nil {
		t.Fatalf("CreateOrganization failed: %v", err)
	}

	// Delete organization
	err = repo.DeleteOrganization(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("DeleteOrganization failed: %v", err)
	}

	// Verify organization is soft deleted (should still exist but with deleted_at set)
	var deletedAt *string
	err = db.QueryRow("SELECT deleted_at FROM wailsalutem.organizations WHERE id = $1", created.ID).Scan(&deletedAt)
	if err != nil {
		t.Fatalf("Failed to query deleted organization: %v", err)
	}

	if deletedAt == nil {
		t.Error("Expected deleted_at to be set after deletion")
	}
}

// TestRepositorySchemaCreation_Integration tests that tenant schema is created
func TestRepositorySchemaCreation_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	defer testutil.CleanupTestDB(t, db)

	repo := NewRepository(db, nil)

	// Create organization
	org, err := repo.CreateOrganization(context.Background(), CreateOrganizationRequest{
		Name:         "Schema Test Hospital",
		ContactEmail: "schema@test.com",
	})
	if err != nil {
		t.Fatalf("CreateOrganization failed: %v", err)
	}

	// Verify schema exists
	var schemaExists bool
	err = db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM information_schema.schemata 
			WHERE schema_name = $1
		)
	`, org.SchemaName).Scan(&schemaExists)
	if err != nil {
		t.Fatalf("Failed to check schema existence: %v", err)
	}

	if !schemaExists {
		t.Errorf("Expected schema %s to exist", org.SchemaName)
	}

	// Verify schema has required tables
	var tableCount int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = $1
	`, org.SchemaName).Scan(&tableCount)
	if err != nil {
		t.Fatalf("Failed to count tables in schema: %v", err)
	}

	if tableCount == 0 {
		t.Errorf("Expected schema %s to have tables, got 0", org.SchemaName)
	}
}
