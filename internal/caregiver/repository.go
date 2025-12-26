package caregiver

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateCaregiver(ctx context.Context, schemaName string, req CreateCaregiverRequest) (*CaregiverResponse, error) {
	caregiverID := uuid.New()
	createdAt := time.Now()

	query := fmt.Sprintf(`
		INSERT INTO %s.users 
		(id, keycloak_user_id, full_name, email, phone_number, role, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, true, $7)
		RETURNING id, keycloak_user_id, full_name, email, phone_number, role, is_active, created_at
	`, pq.QuoteIdentifier(schemaName))

	var caregiver CaregiverResponse

	err := r.db.QueryRowContext(ctx, query,
		caregiverID,
		req.KeycloakUserID,
		req.FullName,
		req.Email,
		req.PhoneNumber,
		req.Role,
		createdAt,
	).Scan(
		&caregiver.ID,
		&caregiver.KeycloakUserID,
		&caregiver.FullName,
		&caregiver.Email,
		&caregiver.PhoneNumber,
		&caregiver.Role,
		&caregiver.IsActive,
		&caregiver.CreatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { 
				return nil, fmt.Errorf("caregiver with this keycloak user ID already exists")
			}
		}
		return nil, fmt.Errorf("failed to insert caregiver: %w", err)
	}

	return &caregiver, nil
}

func (r *Repository) ListCaregivers(ctx context.Context, schemaName string) ([]CaregiverResponse, error) {
	query := fmt.Sprintf(`
		SELECT id, keycloak_user_id, full_name, email, phone_number, role, is_active, created_at, updated_at
		FROM %s.users
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`, pq.QuoteIdentifier(schemaName))

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query caregivers: %w", err)
	}
	defer rows.Close()

	var caregivers []CaregiverResponse
	for rows.Next() {
		var caregiver CaregiverResponse
		var updatedAt sql.NullTime

		err := rows.Scan(
			&caregiver.ID,
			&caregiver.KeycloakUserID,
			&caregiver.FullName,
			&caregiver.Email,
			&caregiver.PhoneNumber,
			&caregiver.Role,
			&caregiver.IsActive,
			&caregiver.CreatedAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan caregiver: %w", err)
		}

		if updatedAt.Valid {
			caregiver.UpdatedAt = &updatedAt.Time
		}

		caregivers = append(caregivers, caregiver)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating caregivers: %w", err)
	}

	return caregivers, nil
}

func (r *Repository) GetCaregiver(ctx context.Context, schemaName string, id string) (*CaregiverResponse, error) {
	query := fmt.Sprintf(`
		SELECT id, keycloak_user_id, full_name, email, phone_number, role, is_active, created_at, updated_at
		FROM %s.users
		WHERE id = $1 AND deleted_at IS NULL
	`, pq.QuoteIdentifier(schemaName))

	var caregiver CaregiverResponse
	var updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&caregiver.ID,
		&caregiver.KeycloakUserID,
		&caregiver.FullName,
		&caregiver.Email,
		&caregiver.PhoneNumber,
		&caregiver.Role,
		&caregiver.IsActive,
		&caregiver.CreatedAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("caregiver not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query caregiver: %w", err)
	}

	if updatedAt.Valid {
		caregiver.UpdatedAt = &updatedAt.Time
	}

	return &caregiver, nil
}

func (r *Repository) UpdateCaregiver(ctx context.Context, schemaName string, id string, req UpdateCaregiverRequest) (*CaregiverResponse, error) {
	var updates []string
	var args []interface{}
	argIndex := 1

	if req.FullName != nil {
		updates = append(updates, fmt.Sprintf("full_name = $%d", argIndex))
		args = append(args, *req.FullName)
		argIndex++
	}
	if req.Email != nil {
		updates = append(updates, fmt.Sprintf("email = $%d", argIndex))
		args = append(args, *req.Email)
		argIndex++
	}
	if req.PhoneNumber != nil {
		updates = append(updates, fmt.Sprintf("phone_number = $%d", argIndex))
		args = append(args, *req.PhoneNumber)
		argIndex++
	}
	if req.Role != nil {
		updates = append(updates, fmt.Sprintf("role = $%d", argIndex))
		args = append(args, *req.Role)
		argIndex++
	}
	if req.IsActive != nil {
		updates = append(updates, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	updates = append(updates, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE %s.users
		SET %s
		WHERE id = $%d AND deleted_at IS NULL
		RETURNING id, keycloak_user_id, full_name, email, phone_number, role, is_active, created_at, updated_at
	`, pq.QuoteIdentifier(schemaName), strings.Join(updates, ", "), argIndex)

	var caregiver CaregiverResponse
	var updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&caregiver.ID,
		&caregiver.KeycloakUserID,
		&caregiver.FullName,
		&caregiver.Email,
		&caregiver.PhoneNumber,
		&caregiver.Role,
		&caregiver.IsActive,
		&caregiver.CreatedAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("caregiver not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update caregiver: %w", err)
	}

	if updatedAt.Valid {
		caregiver.UpdatedAt = &updatedAt.Time
	}

	return &caregiver, nil
}

func (r *Repository) DeleteCaregiver(ctx context.Context, schemaName string, id string) error {
	query := fmt.Sprintf(`
		UPDATE %s.users
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`, pq.QuoteIdentifier(schemaName))

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete caregiver: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("caregiver not found")
	}

	return nil
}
