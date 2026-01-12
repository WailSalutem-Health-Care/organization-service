package organization

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
)

// RetentionPeriod defines how long deleted organizations are retained (3 years)
const RetentionPeriod = 3 * 365 * 24 * time.Hour

// CleanupService handles permanent deletion of expired soft-deleted organizations
type CleanupService struct {
	db *sql.DB
}

// NewCleanupService creates a new cleanup service
func NewCleanupService(db *sql.DB) *CleanupService {
	return &CleanupService{db: db}
}

// CleanupExpiredOrganizations permanently deletes organizations that have been soft-deleted
// for more than 3 years, including dropping their schemas
func (s *CleanupService) CleanupExpiredOrganizations(ctx context.Context) (int, error) {
	cutoffDate := time.Now().Add(-RetentionPeriod)
	log.Printf("Starting cleanup of organizations deleted before %s", cutoffDate.Format(time.RFC3339))

	// Find organizations that have been soft-deleted for more than 3 years
	query := `
		SELECT id, schema_name 
		FROM wailsalutem.organizations 
		WHERE deleted_at IS NOT NULL 
		AND deleted_at < $1
		ORDER BY deleted_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query, cutoffDate)
	if err != nil {
		return 0, fmt.Errorf("failed to query expired organizations: %w", err)
	}
	defer rows.Close()

	var expiredOrgs []struct {
		ID         string
		SchemaName string
	}

	for rows.Next() {
		var org struct {
			ID         string
			SchemaName string
		}
		if err := rows.Scan(&org.ID, &org.SchemaName); err != nil {
			return 0, fmt.Errorf("failed to scan organization: %w", err)
		}
		expiredOrgs = append(expiredOrgs, org)
	}

	if err = rows.Err(); err != nil {
		return 0, fmt.Errorf("error iterating organizations: %w", err)
	}

	if len(expiredOrgs) == 0 {
		log.Println("No expired organizations found for cleanup")
		return 0, nil
	}

	log.Printf("Found %d organizations to permanently delete", len(expiredOrgs))

	// Process each expired organization
	deletedCount := 0
	for _, org := range expiredOrgs {
		if err := s.permanentlyDeleteOrganization(ctx, org.ID, org.SchemaName); err != nil {
			log.Printf("Failed to delete organization %s: %v", org.ID, err)
			continue
		}
		deletedCount++
	}

	log.Printf("Successfully cleaned up %d/%d expired organizations", deletedCount, len(expiredOrgs))
	return deletedCount, nil
}

// permanentlyDeleteOrganization performs hard delete of an organization and drops its schema
func (s *CleanupService) permanentlyDeleteOrganization(ctx context.Context, orgID, schemaName string) error {
	// Start transaction for atomic operation
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Hard delete the organization record
	deleteQuery := `
		DELETE FROM wailsalutem.organizations 
		WHERE id = $1 AND deleted_at IS NOT NULL
	`
	result, err := tx.ExecContext(ctx, deleteQuery, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete organization record: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("organization not found or not soft-deleted")
	}

	// Drop the organization's schema (CASCADE will drop all tables)
	dropSchemaQuery := fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", pq.QuoteIdentifier(schemaName))
	if _, err := tx.ExecContext(ctx, dropSchemaQuery); err != nil {
		return fmt.Errorf("failed to drop schema %s: %w", schemaName, err)
	}

	// Clear from cache
	schemaCacheMutex.Lock()
	delete(schemaCache, orgID)
	schemaCacheMutex.Unlock()

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Permanently deleted organization %s and dropped schema %s", orgID, schemaName)
	return nil
}

// GetExpiredOrganizationsCount returns count of organizations eligible for cleanup
func (s *CleanupService) GetExpiredOrganizationsCount(ctx context.Context) (int, error) {
	cutoffDate := time.Now().Add(-RetentionPeriod)

	var count int
	query := `
		SELECT COUNT(*) 
		FROM wailsalutem.organizations 
		WHERE deleted_at IS NOT NULL 
		AND deleted_at < $1
	`

	err := s.db.QueryRowContext(ctx, query, cutoffDate).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count expired organizations: %w", err)
	}

	return count, nil
}
