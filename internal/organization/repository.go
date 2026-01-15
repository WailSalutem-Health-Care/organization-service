package organization

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/WailSalutem-Health-Care/organization-service/internal/messaging"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Repository struct {
	db        *sql.DB
	publisher *messaging.Publisher
}

func NewRepository(db *sql.DB, publisher *messaging.Publisher) *Repository {
	return &Repository{
		db:        db,
		publisher: publisher,
	}
}

func (r *Repository) CreateOrganization(ctx context.Context, req CreateOrganizationRequest) (*OrganizationResponse, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	orgID := uuid.New()

	sanitizedName := sanitizeName(req.Name)
	schemaName := fmt.Sprintf("org_%s_%s", sanitizedName, orgID.String()[:8])

	query := `
        INSERT INTO wailsalutem.organizations 
        (id, name, schema_name, contact_email, contact_phone, address, status, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, 'active', $7)
        RETURNING id, name, schema_name, contact_email, contact_phone, address, status, created_at
    `

	createdAt := time.Now()
	var org OrganizationResponse

	err = tx.QueryRowContext(ctx, query,
		orgID,
		req.Name,
		schemaName,
		req.ContactEmail,
		req.ContactPhone,
		req.Address,
		createdAt,
	).Scan(
		&org.ID,
		&org.Name,
		&org.SchemaName,
		&org.ContactEmail,
		&org.ContactPhone,
		&org.Address,
		&org.Status,
		&org.CreatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				return nil, fmt.Errorf("organization with this name already exists")
			}
		}
		return nil, fmt.Errorf("failed to insert organization: %w", err)
	}

	if err := r.createOrganizationSchema(ctx, tx, schemaName); err != nil {
		return nil, fmt.Errorf("failed to create organization schema: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &org, nil
}

func (r *Repository) createOrganizationSchema(ctx context.Context, tx *sql.Tx, schemaName string) error {
	// Call the database function that creates the tenant schema
	// This ensures schema definition is maintained in migrations only
	_, err := tx.ExecContext(
		ctx,
		"SELECT wailsalutem.create_tenant_schema($1)",
		schemaName,
	)
	if err != nil {
		return fmt.Errorf("failed to create tenant schema via database function: %w", err)
	}

	log.Printf("Created tenant schema '%s' via database function", schemaName)
	return nil
}

func sanitizeName(name string) string {

	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")

	result := strings.Builder{}
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			result.WriteRune(r)
		}
	}

	if result.Len() > 20 {
		return result.String()[:20]
	}
	return result.String()
}

func (r *Repository) ListOrganizations(ctx context.Context) ([]OrganizationResponse, error) {
	query := `
		SELECT id, name, schema_name, contact_email, contact_phone, address, status, created_at
		FROM wailsalutem.organizations
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query organizations: %w", err)
	}
	defer rows.Close()

	var orgs []OrganizationResponse
	for rows.Next() {
		var org OrganizationResponse
		var contactEmail sql.NullString
		var contactPhone sql.NullString
		var address sql.NullString

		err := rows.Scan(
			&org.ID,
			&org.Name,
			&org.SchemaName,
			&contactEmail,
			&contactPhone,
			&address,
			&org.Status,
			&org.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}

		if contactEmail.Valid {
			org.ContactEmail = contactEmail.String
		}
		if contactPhone.Valid {
			org.ContactPhone = contactPhone.String
		}
		if address.Valid {
			org.Address = address.String
		}

		orgs = append(orgs, org)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating organizations: %w", err)
	}

	return orgs, nil
}

// ListOrganizationsWithPagination retrieves organizations with pagination support
func (r *Repository) ListOrganizationsWithPagination(ctx context.Context, limit, offset int, search string, status string) ([]OrganizationResponse, int, error) {
	// Build WHERE clause
	whereClause := "WHERE deleted_at IS NULL"
	args := []interface{}{limit, offset}
	countArgs := []interface{}{}
	argIndex := 3

	if search != "" {
		whereClause += fmt.Sprintf(` AND (name ILIKE $%d OR contact_email ILIKE $%d)`, argIndex, argIndex)
		args = append(args, "%"+search+"%")
		countArgs = append(countArgs, "%"+search+"%")
		argIndex++
	}

	if status != "" && status != "all" {
		whereClause += fmt.Sprintf(` AND status = $%d`, argIndex)
		args = append(args, status)
		countArgs = append(countArgs, status)
	}

	// First, get total count
	var totalCount int
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM wailsalutem.organizations
		%s
	`, whereClause)

	err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count organizations: %w", err)
	}

	// Then get paginated results
	query := fmt.Sprintf(`
		SELECT id, name, schema_name, contact_email, contact_phone, address, status, created_at
		FROM wailsalutem.organizations
		%s
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, whereClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query organizations: %w", err)
	}
	defer rows.Close()

	var orgs []OrganizationResponse
	for rows.Next() {
		var org OrganizationResponse
		var contactEmail sql.NullString
		var contactPhone sql.NullString
		var address sql.NullString

		err := rows.Scan(
			&org.ID,
			&org.Name,
			&org.SchemaName,
			&contactEmail,
			&contactPhone,
			&address,
			&org.Status,
			&org.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan organization: %w", err)
		}

		if contactEmail.Valid {
			org.ContactEmail = contactEmail.String
		}
		if contactPhone.Valid {
			org.ContactPhone = contactPhone.String
		}
		if address.Valid {
			org.Address = address.String
		}

		orgs = append(orgs, org)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating organizations: %w", err)
	}

	return orgs, totalCount, nil
}

func (r *Repository) GetOrganization(ctx context.Context, id string) (*OrganizationResponse, error) {
	query := `
		SELECT id, name, schema_name, contact_email, contact_phone, address, status, created_at
		FROM wailsalutem.organizations
		WHERE id = $1 AND deleted_at IS NULL
	`

	var org OrganizationResponse
	var contactEmail sql.NullString
	var contactPhone sql.NullString
	var address sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&org.ID,
		&org.Name,
		&org.SchemaName,
		&contactEmail,
		&contactPhone,
		&address,
		&org.Status,
		&org.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("organization not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query organization: %w", err)
	}

	if contactEmail.Valid {
		org.ContactEmail = contactEmail.String
	}
	if contactPhone.Valid {
		org.ContactPhone = contactPhone.String
	}
	if address.Valid {
		org.Address = address.String
	}

	return &org, nil
}

func (r *Repository) UpdateOrganization(ctx context.Context, id string, req UpdateOrganizationRequest) (*OrganizationResponse, error) {
	// Build dynamic update query
	var updates []string
	var args []interface{}
	argIndex := 1

	if req.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *req.Name)
		argIndex++
	}
	if req.ContactEmail != nil {
		updates = append(updates, fmt.Sprintf("contact_email = $%d", argIndex))
		args = append(args, *req.ContactEmail)
		argIndex++
	}
	if req.ContactPhone != nil {
		updates = append(updates, fmt.Sprintf("contact_phone = $%d", argIndex))
		args = append(args, *req.ContactPhone)
		argIndex++
	}
	if req.Address != nil {
		updates = append(updates, fmt.Sprintf("address = $%d", argIndex))
		args = append(args, *req.Address)
		argIndex++
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	// Add updated_at timestamp
	updates = append(updates, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add ID parameter
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE wailsalutem.organizations
		SET %s
		WHERE id = $%d AND deleted_at IS NULL
		RETURNING id, name, schema_name, contact_email, contact_phone, address, status, created_at
	`, strings.Join(updates, ", "), argIndex)

	var org OrganizationResponse
	var contactEmail sql.NullString
	var contactPhone sql.NullString
	var address sql.NullString

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&org.ID,
		&org.Name,
		&org.SchemaName,
		&contactEmail,
		&contactPhone,
		&address,
		&org.Status,
		&org.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("organization not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	if contactEmail.Valid {
		org.ContactEmail = contactEmail.String
	}
	if contactPhone.Valid {
		org.ContactPhone = contactPhone.String
	}
	if address.Valid {
		org.Address = address.String
	}

	return &org, nil
}

func (r *Repository) DeleteOrganization(ctx context.Context, id string) error {
	// Get organization details before deleting (for event)
	org, err := r.GetOrganization(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get organization for deletion: %w", err)
	}

	// Soft delete: Set deleted_at timestamp and update status to inactive
	query := `
		UPDATE wailsalutem.organizations
		SET deleted_at = $1,
		    status = 'inactive',
		    updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	deletedAt := time.Now()
	result, err := r.db.ExecContext(ctx, query, deletedAt, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete organization: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("organization not found or already deleted")
	}

	// Clear the schema from cache to prevent access to deleted organization
	schemaCacheMutex.Lock()
	delete(schemaCache, id)
	schemaCacheMutex.Unlock()

	// Publish organization.deleted event
	if r.publisher != nil {
		event := messaging.OrganizationDeletedEvent{
			BaseEvent: messaging.NewBaseEvent(messaging.EventOrganizationDeleted),
			Data: messaging.OrganizationDeletedData{
				OrganizationID:   org.ID,
				OrganizationName: org.Name,
				SchemaName:       org.SchemaName,
				DeletedAt:        deletedAt,
			},
		}

		if err := r.publisher.Publish(ctx, messaging.EventOrganizationDeleted, event); err != nil {
			log.Printf("Warning: failed to publish organization.deleted event: %v", err)
			// Don't fail the delete if event publishing fails
		}
	}

	// NOTE: Schema and all data are retained for 3 years as per retention policy
	// Run cleanup job periodically to purge organizations deleted more than 3 years ago

	return nil
}
