package users

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
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

func (r *Repository) Delete(schemaName, userID string) error {
	if err := r.ValidateOrgSchema(schemaName); err != nil {
		return err
	}

	query := fmt.Sprintf(`DELETE FROM %s.users WHERE id = $1`, schemaName)

	result, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrUserNotFound
	}

	log.Printf("Deleted user from database: %s (schema: %s)", userID, schemaName)

	return nil
}
