package users

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/WailSalutem-Health-Care/organization-service/internal/messaging"
	"github.com/google/uuid"
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

func (r *Repository) GetSchemaNameByOrgID(orgID string) (string, error) {
	query := `SELECT schema_name FROM wailsalutem.organizations WHERE id = $1`
	var schemaName string
	err := r.db.QueryRow(query, orgID).Scan(&schemaName)
	if err == sql.ErrNoRows {
		return "", ErrInvalidOrgSchema
	}
	if err != nil {
		return "", fmt.Errorf("failed to get schema name: %w", err)
	}
	return schemaName, nil
}

func (r *Repository) ValidateOrgSchema(schemaName string) error {
	query := `SELECT 1 FROM wailsalutem.organizations WHERE schema_name = $1`
	var exists int
	err := r.db.QueryRow(query, schemaName).Scan(&exists)
	if err == sql.ErrNoRows {
		return ErrInvalidOrgSchema
	}
	if err != nil {
		return fmt.Errorf("failed to validate schema: %w", err)
	}
	return nil
}

func (r *Repository) Create(user *User) error {
	if err := r.ValidateOrgSchema(user.OrgSchemaName); err != nil {
		return err
	}

	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()

	query := fmt.Sprintf(`
		INSERT INTO %s.users (id, keycloak_user_id, email, first_name, last_name, phone_number, role, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, user.OrgSchemaName)

	_, err := r.db.Exec(query,
		user.ID,
		user.KeycloakUserID,
		user.Email,
		user.FirstName,
		user.LastName,
		user.PhoneNumber,
		user.Role,
		user.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user in database: %w", err)
	}

	log.Printf("Created user in database: %s %s (schema: %s)", user.FirstName, user.LastName, user.OrgSchemaName)

	// Publish user.created event
	if r.publisher != nil {
		event := messaging.UserCreatedEvent{
			BaseEvent: messaging.NewBaseEvent(messaging.EventUserCreated),
			Data: messaging.UserCreatedData{
				UserID:         user.ID,
				KeycloakUserID: user.KeycloakUserID,
				OrganizationID: user.OrgID,
				Email:          user.Email,
				FirstName:      user.FirstName,
				LastName:       user.LastName,
				Role:           user.Role,
				IsActive:       true,
				CreatedAt:      user.CreatedAt,
			},
		}

		if err := r.publisher.Publish(nil, messaging.EventUserCreated, event); err != nil {
			log.Printf("Warning: failed to publish user.created event: %v", err)
		}
	}

	return nil
}

func (r *Repository) GetByID(schemaName, userID string) (*User, error) {
	if err := r.ValidateOrgSchema(schemaName); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT id, keycloak_user_id, email, first_name, last_name, phone_number, role, created_at, updated_at
		FROM %s.users
		WHERE id = $1
	`, schemaName)

	user := &User{}
	var updatedAt sql.NullTime
	var phoneNumber sql.NullString
	var email sql.NullString
	var firstName sql.NullString
	var lastName sql.NullString

	err := r.db.QueryRow(query, userID).Scan(
		&user.ID,
		&user.KeycloakUserID,
		&email,
		&firstName,
		&lastName,
		&phoneNumber,
		&user.Role,
		&user.CreatedAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if email.Valid {
		user.Email = email.String
	}
	if firstName.Valid {
		user.FirstName = firstName.String
	}
	if lastName.Valid {
		user.LastName = lastName.String
	}
	if phoneNumber.Valid {
		user.PhoneNumber = phoneNumber.String
	}
	if updatedAt.Valid {
		user.UpdatedAt = updatedAt.Time
	}

	user.OrgSchemaName = schemaName

	return user, nil
}

func (r *Repository) GetByKeycloakID(schemaName, keycloakUserID string) (*User, error) {
	if err := r.ValidateOrgSchema(schemaName); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT id, keycloak_user_id, email, first_name, last_name, phone_number, role, created_at, updated_at
		FROM %s.users
		WHERE keycloak_user_id = $1
	`, schemaName)

	user := &User{}
	var updatedAt sql.NullTime
	var phoneNumber sql.NullString
	var email sql.NullString
	var firstName sql.NullString
	var lastName sql.NullString

	err := r.db.QueryRow(query, keycloakUserID).Scan(
		&user.ID,
		&user.KeycloakUserID,
		&email,
		&firstName,
		&lastName,
		&phoneNumber,
		&user.Role,
		&user.CreatedAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if email.Valid {
		user.Email = email.String
	}
	if firstName.Valid {
		user.FirstName = firstName.String
	}
	if lastName.Valid {
		user.LastName = lastName.String
	}
	if phoneNumber.Valid {
		user.PhoneNumber = phoneNumber.String
	}
	if updatedAt.Valid {
		user.UpdatedAt = updatedAt.Time
	}

	user.OrgSchemaName = schemaName

	return user, nil
}

func (r *Repository) List(schemaName string) ([]User, error) {
	if err := r.ValidateOrgSchema(schemaName); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT id, keycloak_user_id, email, first_name, last_name, phone_number, role, created_at, updated_at
		FROM %s.users
		ORDER BY created_at DESC
	`, schemaName)

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		var updatedAt sql.NullTime
		var phoneNumber sql.NullString
		var email sql.NullString
		var firstName sql.NullString
		var lastName sql.NullString

		err := rows.Scan(
			&user.ID,
			&user.KeycloakUserID,
			&email,
			&firstName,
			&lastName,
			&phoneNumber,
			&user.Role,
			&user.CreatedAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		if email.Valid {
			user.Email = email.String
		}
		if firstName.Valid {
			user.FirstName = firstName.String
		}
		if lastName.Valid {
			user.LastName = lastName.String
		}
		if phoneNumber.Valid {
			user.PhoneNumber = phoneNumber.String
		}
		if updatedAt.Valid {
			user.UpdatedAt = updatedAt.Time
		}

		user.OrgSchemaName = schemaName
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// ListWithPagination retrieves users with pagination support
func (r *Repository) ListWithPagination(schemaName string, limit, offset int) ([]User, int, error) {
	if err := r.ValidateOrgSchema(schemaName); err != nil {
		return nil, 0, err
	}

	// First, get total count
	var totalCount int
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM %s.users
	`, schemaName)

	err := r.db.QueryRow(countQuery).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Then get paginated results
	query := fmt.Sprintf(`
		SELECT id, keycloak_user_id, email, first_name, last_name, phone_number, role, created_at, updated_at
		FROM %s.users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, schemaName)

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		var updatedAt sql.NullTime
		var phoneNumber sql.NullString
		var email sql.NullString
		var firstName sql.NullString
		var lastName sql.NullString

		err := rows.Scan(
			&user.ID,
			&user.KeycloakUserID,
			&email,
			&firstName,
			&lastName,
			&phoneNumber,
			&user.Role,
			&user.CreatedAt,
			&updatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}

		if email.Valid {
			user.Email = email.String
		}
		if firstName.Valid {
			user.FirstName = firstName.String
		}
		if lastName.Valid {
			user.LastName = lastName.String
		}
		if phoneNumber.Valid {
			user.PhoneNumber = phoneNumber.String
		}
		if updatedAt.Valid {
			user.UpdatedAt = updatedAt.Time
		}

		user.OrgSchemaName = schemaName
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating users: %w", err)
	}

	return users, totalCount, nil
}

// ListActiveUsersWithPagination retrieves active users (not soft deleted) with pagination support
func (r *Repository) ListActiveUsersWithPagination(schemaName string, limit, offset int) ([]User, int, error) {
	if err := r.ValidateOrgSchema(schemaName); err != nil {
		return nil, 0, err
	}

	// First, get total count of active users
	var totalCount int
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM %s.users
		WHERE deleted_at IS NULL
	`, schemaName)

	err := r.db.QueryRow(countQuery).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count active users: %w", err)
	}

	// Then get paginated results
	query := fmt.Sprintf(`
		SELECT id, keycloak_user_id, email, first_name, last_name, phone_number, role, created_at, updated_at
		FROM %s.users
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, schemaName)

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list active users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		var updatedAt sql.NullTime
		var phoneNumber sql.NullString
		var email sql.NullString
		var firstName sql.NullString
		var lastName sql.NullString

		err := rows.Scan(
			&user.ID,
			&user.KeycloakUserID,
			&email,
			&firstName,
			&lastName,
			&phoneNumber,
			&user.Role,
			&user.CreatedAt,
			&updatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}

		if email.Valid {
			user.Email = email.String
		}
		if firstName.Valid {
			user.FirstName = firstName.String
		}
		if lastName.Valid {
			user.LastName = lastName.String
		}
		if phoneNumber.Valid {
			user.PhoneNumber = phoneNumber.String
		}
		if updatedAt.Valid {
			user.UpdatedAt = updatedAt.Time
		}

		user.OrgSchemaName = schemaName
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating users: %w", err)
	}

	return users, totalCount, nil
}

func (r *Repository) Update(user *User) error {
	if err := r.ValidateOrgSchema(user.OrgSchemaName); err != nil {
		return err
	}

	user.UpdatedAt = time.Now()

	query := fmt.Sprintf(`
		UPDATE %s.users
		SET email = $1, first_name = $2, last_name = $3, phone_number = $4, updated_at = $5
		WHERE id = $6
	`, user.OrgSchemaName)

	result, err := r.db.Exec(query,
		user.Email,
		user.FirstName,
		user.LastName,
		user.PhoneNumber,
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrUserNotFound
	}

	log.Printf("Updated user in database: %s %s (schema: %s)", user.FirstName, user.LastName, user.OrgSchemaName)

	return nil
}

func (r *Repository) Delete(schemaName, orgID, userID string, role string) error {
	if err := r.ValidateOrgSchema(schemaName); err != nil {
		return err
	}

	// Soft delete: Set deleted_at timestamp
	query := fmt.Sprintf(`
		UPDATE %s.users 
		SET deleted_at = $1,
		    updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`, schemaName)

	deletedAt := time.Now()
	result, err := r.db.Exec(query, deletedAt, userID)
	if err != nil {
		return fmt.Errorf("failed to soft delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrUserNotFound
	}

	// Publish user.deleted event
	if r.publisher != nil {
		event := messaging.UserDeletedEvent{
			BaseEvent: messaging.NewBaseEvent(messaging.EventUserDeleted),
			Data: messaging.UserDeletedData{
				UserID:         userID,
				OrganizationID: orgID,
				Role:           role,
				DeletedAt:      deletedAt,
			},
		}

		if err := r.publisher.Publish(nil, messaging.EventUserDeleted, event); err != nil {
			log.Printf("Warning: failed to publish user.deleted event: %v", err)
		}
	}

	log.Printf("Deleted user from database: %s (schema: %s)", userID, schemaName)

	return nil
}
