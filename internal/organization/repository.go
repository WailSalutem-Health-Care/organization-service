package organization

import (
	"context"
	"database/sql"
	"encoding/json"
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

	var settingsJSON []byte
	if req.Settings != nil {
		settingsJSON, err = json.Marshal(req.Settings)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal settings: %w", err)
		}
	} else {

		settingsJSON = []byte("{}")
	}

	query := `
        INSERT INTO wailsalutem.organizations 
        (id, name, schema_name, contact_email, contact_phone, address, status, settings, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, 'active', $7, $8)
        RETURNING id, name, schema_name, contact_email, contact_phone, address, status, settings, created_at
    `

	createdAt := time.Now()
	var org OrganizationResponse
	var settingsStr sql.NullString

	err = tx.QueryRowContext(ctx, query,
		orgID,
		req.Name,
		schemaName,
		req.ContactEmail,
		req.ContactPhone,
		req.Address,
		settingsJSON,
		createdAt,
	).Scan(
		&org.ID,
		&org.Name,
		&org.SchemaName,
		&org.ContactEmail,
		&org.ContactPhone,
		&org.Address,
		&org.Status,
		&settingsStr,
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

	if settingsStr.Valid && settingsStr.String != "" {
		if err := json.Unmarshal([]byte(settingsStr.String), &org.Settings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
		}
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

	_, err := tx.ExecContext(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", pq.QuoteIdentifier(schemaName)))
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	usersTable := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s.users (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            keycloak_user_id UUID NOT NULL,
            full_name VARCHAR(255),
            email VARCHAR(255),
            phone_number VARCHAR(50),
            role VARCHAR(50),
            is_active BOOLEAN DEFAULT true,
            created_at TIMESTAMP DEFAULT now(),
            updated_at TIMESTAMP,
            deleted_at TIMESTAMP
        )
    `, pq.QuoteIdentifier(schemaName))

	if _, err := tx.ExecContext(ctx, usersTable); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	patientsTable := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s.patients (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            keycloak_user_id UUID NOT NULL,
            first_name VARCHAR(255),
            last_name VARCHAR(255),
            email VARCHAR(255),
            phone_number VARCHAR(50),
            date_of_birth DATE,
            address TEXT,
            emergency_contact_name VARCHAR(255),
            emergency_contact_phone VARCHAR(50),
            medical_notes TEXT,
            careplan_type VARCHAR(100),
            careplan_frequency VARCHAR(50),
            is_active BOOLEAN DEFAULT true,
            created_at TIMESTAMP DEFAULT now(),
            updated_at TIMESTAMP,
            deleted_at TIMESTAMP
        )
    `, pq.QuoteIdentifier(schemaName))

	if _, err := tx.ExecContext(ctx, patientsTable); err != nil {
		return fmt.Errorf("failed to create patients table: %w", err)
	}

	careSessionsTable := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s.care_sessions (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            patient_id UUID REFERENCES %s.patients(id),
            caregiver_id UUID,
            check_in_time TIMESTAMP,
            check_out_time TIMESTAMP,
            status VARCHAR(50),
            caregiver_notes TEXT,
            created_at TIMESTAMP DEFAULT now(),
            updated_at TIMESTAMP,
            deleted_at TIMESTAMP
        )
    `, pq.QuoteIdentifier(schemaName), pq.QuoteIdentifier(schemaName))

	if _, err := tx.ExecContext(ctx, careSessionsTable); err != nil {
		return fmt.Errorf("failed to create care_sessions table: %w", err)
	}

	nfcTagsTable := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s.nfc_tags (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            tag_id VARCHAR(100) UNIQUE,
            patient_id UUID REFERENCES %s.patients(id),
            issued_at TIMESTAMP DEFAULT now(),
            status VARCHAR(50),
            deactivated_at TIMESTAMP,
            created_at TIMESTAMP DEFAULT now(),
            updated_at TIMESTAMP
        )
    `, pq.QuoteIdentifier(schemaName), pq.QuoteIdentifier(schemaName))

	if _, err := tx.ExecContext(ctx, nfcTagsTable); err != nil {
		return fmt.Errorf("failed to create nfc_tags table: %w", err)
	}

	feedbackTable := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s.feedback (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            care_session_id UUID UNIQUE,
            patient_id UUID,
            rating INTEGER CHECK (rating BETWEEN 1 AND 5),
            patient_feedback TEXT,
            created_at TIMESTAMP DEFAULT now(),
            deleted_at TIMESTAMP
        )
    `, pq.QuoteIdentifier(schemaName))

	if _, err := tx.ExecContext(ctx, feedbackTable); err != nil {
		return fmt.Errorf("failed to create feedback table: %w", err)
	}

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
		SELECT id, name, schema_name, contact_email, contact_phone, address, status, settings, created_at
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
		var settingsStr sql.NullString
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
			&settingsStr,
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

		if settingsStr.Valid && settingsStr.String != "" {
			if err := json.Unmarshal([]byte(settingsStr.String), &org.Settings); err != nil {
				return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
			}
		}

		orgs = append(orgs, org)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating organizations: %w", err)
	}

	return orgs, nil
}

// ListOrganizationsWithPagination retrieves organizations with pagination support
func (r *Repository) ListOrganizationsWithPagination(ctx context.Context, limit, offset int) ([]OrganizationResponse, int, error) {
	// First, get total count
	var totalCount int
	countQuery := `
		SELECT COUNT(*) 
		FROM wailsalutem.organizations
		WHERE deleted_at IS NULL
	`
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count organizations: %w", err)
	}

	// Then get paginated results
	query := `
		SELECT id, name, schema_name, contact_email, contact_phone, address, status, settings, created_at
		FROM wailsalutem.organizations
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query organizations: %w", err)
	}
	defer rows.Close()

	var orgs []OrganizationResponse
	for rows.Next() {
		var org OrganizationResponse
		var settingsStr sql.NullString
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
			&settingsStr,
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

		if settingsStr.Valid && settingsStr.String != "" {
			if err := json.Unmarshal([]byte(settingsStr.String), &org.Settings); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal settings: %w", err)
			}
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
		SELECT id, name, schema_name, contact_email, contact_phone, address, status, settings, created_at
		FROM wailsalutem.organizations
		WHERE id = $1 AND deleted_at IS NULL
	`

	var org OrganizationResponse
	var settingsStr sql.NullString
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
		&settingsStr,
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

	if settingsStr.Valid && settingsStr.String != "" {
		if err := json.Unmarshal([]byte(settingsStr.String), &org.Settings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
		}
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
	if req.Settings != nil {
		settingsJSON, err := json.Marshal(*req.Settings)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal settings: %w", err)
		}
		updates = append(updates, fmt.Sprintf("settings = $%d", argIndex))
		args = append(args, settingsJSON)
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
		RETURNING id, name, schema_name, contact_email, contact_phone, address, status, settings, created_at
	`, strings.Join(updates, ", "), argIndex)

	var org OrganizationResponse
	var settingsStr sql.NullString
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
		&settingsStr,
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

	if settingsStr.Valid && settingsStr.String != "" {
		if err := json.Unmarshal([]byte(settingsStr.String), &org.Settings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
		}
	}

	return &org, nil
}

func (r *Repository) DeleteOrganization(ctx context.Context, id string) error {
	// Get organization details before deleting (for event)
	org, err := r.GetOrganization(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get organization for deletion: %w", err)
	}

	// Soft delete: Set deleted_at timestamp and update status
	query := `
		UPDATE wailsalutem.organizations
		SET deleted_at = $1,
		    status = 'deleted',
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
