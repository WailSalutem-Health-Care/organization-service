package patient

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

func (r *Repository) CreatePatient(ctx context.Context, schemaName string, req CreatePatientRequest) (*PatientResponse, error) {
	patientID := uuid.New()
	createdAt := time.Now()

	query := fmt.Sprintf(`
		INSERT INTO %s.patients 
		(id, first_name, last_name, email, phone_number, date_of_birth, address, emergency_contact_name, emergency_contact_phone, medical_notes, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, true, $11)
		RETURNING id, first_name, last_name, email, phone_number, date_of_birth, address, emergency_contact_name, emergency_contact_phone, medical_notes, is_active, created_at
	`, pq.QuoteIdentifier(schemaName))

	var patient PatientResponse
	var dob sql.NullString
	var email sql.NullString
	var phoneNumber sql.NullString
	var address sql.NullString
	var emergencyContactName sql.NullString
	var emergencyContactPhone sql.NullString
	var medicalNotes sql.NullString

	err := r.db.QueryRowContext(ctx, query,
		patientID,
		req.FirstName,
		req.LastName,
		req.Email,
		req.PhoneNumber,
		req.DateOfBirth,
		req.Address,
		req.EmergencyContactName,
		req.EmergencyContactPhone,
		req.MedicalNotes,
		createdAt,
	).Scan(
		&patient.ID,
		&patient.FirstName,
		&patient.LastName,
		&email,
		&phoneNumber,
		&dob,
		&address,
		&emergencyContactName,
		&emergencyContactPhone,
		&medicalNotes,
		&patient.IsActive,
		&patient.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to insert patient: %w", err)
	}

	if dob.Valid {
		patient.DateOfBirth = &dob.String
	}
	if email.Valid {
		patient.Email = email.String
	}
	if phoneNumber.Valid {
		patient.PhoneNumber = phoneNumber.String
	}
	if address.Valid {
		patient.Address = address.String
	}
	if emergencyContactName.Valid {
		patient.EmergencyContactName = emergencyContactName.String
	}
	if emergencyContactPhone.Valid {
		patient.EmergencyContactPhone = emergencyContactPhone.String
	}
	if medicalNotes.Valid {
		patient.MedicalNotes = medicalNotes.String
	}

	return &patient, nil
}

func (r *Repository) ListPatients(ctx context.Context, schemaName string) ([]PatientResponse, error) {
	query := fmt.Sprintf(`
		SELECT id, first_name, last_name, email, phone_number, date_of_birth, address, emergency_contact_name, emergency_contact_phone, medical_notes, is_active, created_at, updated_at
		FROM %s.patients
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`, pq.QuoteIdentifier(schemaName))

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query patients: %w", err)
	}
	defer rows.Close()

	var patients []PatientResponse
	for rows.Next() {
		var patient PatientResponse
		var dob sql.NullString
		var email sql.NullString
		var phoneNumber sql.NullString
		var address sql.NullString
		var emergencyContactName sql.NullString
		var emergencyContactPhone sql.NullString
		var medicalNotes sql.NullString
		var updatedAt sql.NullTime

		err := rows.Scan(
			&patient.ID,
			&patient.FirstName,
			&patient.LastName,
			&email,
			&phoneNumber,
			&dob,
			&address,
			&emergencyContactName,
			&emergencyContactPhone,
			&medicalNotes,
			&patient.IsActive,
			&patient.CreatedAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan patient: %w", err)
		}

		if dob.Valid {
			patient.DateOfBirth = &dob.String
		}
		if email.Valid {
			patient.Email = email.String
		}
		if phoneNumber.Valid {
			patient.PhoneNumber = phoneNumber.String
		}
		if address.Valid {
			patient.Address = address.String
		}
		if emergencyContactName.Valid {
			patient.EmergencyContactName = emergencyContactName.String
		}
		if emergencyContactPhone.Valid {
			patient.EmergencyContactPhone = emergencyContactPhone.String
		}
		if medicalNotes.Valid {
			patient.MedicalNotes = medicalNotes.String
		}
		if updatedAt.Valid {
			patient.UpdatedAt = &updatedAt.Time
		}

		patients = append(patients, patient)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating patients: %w", err)
	}

	return patients, nil
}

func (r *Repository) GetPatient(ctx context.Context, schemaName string, id string) (*PatientResponse, error) {
	query := fmt.Sprintf(`
		SELECT id, first_name, last_name, email, phone_number, date_of_birth, address, emergency_contact_name, emergency_contact_phone, medical_notes, is_active, created_at, updated_at
		FROM %s.patients
		WHERE id = $1 AND deleted_at IS NULL
	`, pq.QuoteIdentifier(schemaName))

	var patient PatientResponse
	var dob sql.NullString
	var email sql.NullString
	var phoneNumber sql.NullString
	var address sql.NullString
	var emergencyContactName sql.NullString
	var emergencyContactPhone sql.NullString
	var medicalNotes sql.NullString
	var updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&patient.ID,
		&patient.FirstName,
		&patient.LastName,
		&email,
		&phoneNumber,
		&dob,
		&address,
		&emergencyContactName,
		&emergencyContactPhone,
		&medicalNotes,
		&patient.IsActive,
		&patient.CreatedAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("patient not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query patient: %w", err)
	}

	if dob.Valid {
		patient.DateOfBirth = &dob.String
	}
	if email.Valid {
		patient.Email = email.String
	}
	if phoneNumber.Valid {
		patient.PhoneNumber = phoneNumber.String
	}
	if address.Valid {
		patient.Address = address.String
	}
	if emergencyContactName.Valid {
		patient.EmergencyContactName = emergencyContactName.String
	}
	if emergencyContactPhone.Valid {
		patient.EmergencyContactPhone = emergencyContactPhone.String
	}
	if medicalNotes.Valid {
		patient.MedicalNotes = medicalNotes.String
	}
	if updatedAt.Valid {
		patient.UpdatedAt = &updatedAt.Time
	}

	return &patient, nil
}

func (r *Repository) UpdatePatient(ctx context.Context, schemaName string, id string, req UpdatePatientRequest) (*PatientResponse, error) {

	var updates []string
	var args []interface{}
	argIndex := 1

	if req.FirstName != nil {
		updates = append(updates, fmt.Sprintf("first_name = $%d", argIndex))
		args = append(args, *req.FirstName)
		argIndex++
	}
	if req.LastName != nil {
		updates = append(updates, fmt.Sprintf("last_name = $%d", argIndex))
		args = append(args, *req.LastName)
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
	if req.DateOfBirth != nil {
		updates = append(updates, fmt.Sprintf("date_of_birth = $%d", argIndex))
		args = append(args, *req.DateOfBirth)
		argIndex++
	}
	if req.Address != nil {
		updates = append(updates, fmt.Sprintf("address = $%d", argIndex))
		args = append(args, *req.Address)
		argIndex++
	}
	if req.EmergencyContactName != nil {
		updates = append(updates, fmt.Sprintf("emergency_contact_name = $%d", argIndex))
		args = append(args, *req.EmergencyContactName)
		argIndex++
	}
	if req.EmergencyContactPhone != nil {
		updates = append(updates, fmt.Sprintf("emergency_contact_phone = $%d", argIndex))
		args = append(args, *req.EmergencyContactPhone)
		argIndex++
	}
	if req.MedicalNotes != nil {
		updates = append(updates, fmt.Sprintf("medical_notes = $%d", argIndex))
		args = append(args, *req.MedicalNotes)
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
		UPDATE %s.patients
		SET %s
		WHERE id = $%d AND deleted_at IS NULL
		RETURNING id, first_name, last_name, email, phone_number, date_of_birth, address, emergency_contact_name, emergency_contact_phone, medical_notes, is_active, created_at, updated_at
	`, pq.QuoteIdentifier(schemaName), strings.Join(updates, ", "), argIndex)

	var patient PatientResponse
	var dob sql.NullString
	var updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&patient.ID,
		&patient.FirstName,
		&patient.LastName,
		&patient.Email,
		&patient.PhoneNumber,
		&dob,
		&patient.Address,
		&patient.EmergencyContactName,
		&patient.EmergencyContactPhone,
		&patient.MedicalNotes,
		&patient.IsActive,
		&patient.CreatedAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("patient not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update patient: %w", err)
	}

	if dob.Valid {
		patient.DateOfBirth = &dob.String
	}
	if updatedAt.Valid {
		patient.UpdatedAt = &updatedAt.Time
	}

	return &patient, nil
}

func (r *Repository) DeletePatient(ctx context.Context, schemaName string, id string) error {
	query := fmt.Sprintf(`
		UPDATE %s.patients
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`, pq.QuoteIdentifier(schemaName))

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete patient: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("patient not found")
	}

	return nil
}
