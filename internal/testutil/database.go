package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
)

// SetupTestDB creates a connection to the test database
// This connects to the local wailsalutem_test database
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	connStr := "host=localhost port=5432 user=localadmin password=Stoplying! dbname=wailsalutem_test sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	return db
}

// SetupTestTransaction creates a test database connection and begins a transaction
// The transaction is automatically rolled back when the test ends
// This ensures test isolation without needing cleanup
func SetupTestTransaction(t *testing.T) (*sql.DB, *sql.Tx) {
	t.Helper()

	db := SetupTestDB(t)

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		db.Close()
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Ensure transaction is rolled back when test ends
	t.Cleanup(func() {
		tx.Rollback()
		db.Close()
	})

	return db, tx
}

// CleanupTestDB cleans up test data from the database
// Use this if you're not using transactions
func CleanupTestDB(t *testing.T, db *sql.DB) {
	t.Helper()

	// Clean up organizations (cascade will clean tenant schemas)
	_, err := db.Exec("TRUNCATE TABLE wailsalutem.organizations CASCADE")
	if err != nil {
		t.Logf("Warning: Failed to clean up organizations: %v", err)
	}

	// Drop any tenant schemas that were created during tests
	rows, err := db.Query(`
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name LIKE 'org_%'
	`)
	if err != nil {
		t.Logf("Warning: Failed to query tenant schemas: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var schemaName string
		if err := rows.Scan(&schemaName); err != nil {
			continue
		}
		_, err := db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
		if err != nil {
			t.Logf("Warning: Failed to drop schema %s: %v", schemaName, err)
		}
	}
}

// CreateTestOrg creates a test organization and returns its schema name
// This is a helper for tests that need a tenant schema
func CreateTestOrg(t *testing.T, db *sql.DB, name string) (orgID, schemaName string) {
	t.Helper()

	query := `
		INSERT INTO wailsalutem.organizations 
		(name, schema_name, status, created_at)
		VALUES ($1, $2, 'active', NOW())
		RETURNING id, schema_name
	`

	schemaName = fmt.Sprintf("org_test_%s", name)
	err := db.QueryRow(query, name, schemaName).Scan(&orgID, &schemaName)
	if err != nil {
		t.Fatalf("Failed to create test organization: %v", err)
	}

	// Create the tenant schema
	_, err = db.Exec(fmt.Sprintf("SELECT wailsalutem.create_tenant_schema('%s')", schemaName))
	if err != nil {
		t.Fatalf("Failed to create tenant schema: %v", err)
	}

	return orgID, schemaName
}
