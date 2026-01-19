package patient

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
	publisher messaging.PublisherInterface
}

func NewRepository(db *sql.DB, publisher messaging.PublisherInterface) *Repository {
	return &Repository{
		db:        db,
		publisher: publisher,
	}
}

// generatePatientID generates a sequential patient ID like PT-0001, PT-0002, etc.
func (r *Repository) generatePatientID(ctx context.Context, schemaName string) (string, error) {
	query := fmt.Sprintf(`
		SELECT patient_id FROM %s.patients 
		WHERE patient_id IS NOT NULL 
		ORDER BY patient_id DESC 
		LIMIT 1
	`, pq.QuoteIdentifier(schemaName))

	var lastPatientID sql.NullString
	err := r.db.QueryRowContext(ctx, query).Scan(&lastPatientID)

	if err == sql.ErrNoRows || !lastPatientID.Valid {
		// First patient in this organization
		return "PT-0001", nil
	}

	if err != nil {
		return "", fmt.Errorf("failed to get last patient ID: %w", err)
	}

	// Parse the number from PT-XXXX format
	var currentNum int
	_, err = fmt.Sscanf(lastPatientID.String, "PT-%d", &currentNum)
	if err != nil {
		// If parsing fails, start from 1
		return "PT-0001", nil
	}

	// Increment and format with leading zeros
	nextNum := currentNum + 1
	return fmt.Sprintf("PT-%04d", nextNum), nil
}

func (r *Repository) CreatePatient(ctx context.Context, schemaName string, orgID string, keycloakUserID string, req CreatePatientRequest) (*PatientResponse, error) {
	patientID := uuid.New()
	createdAt := time.Now()

	// Generate sequential patient ID
	patientDisplayID, err := r.generatePatientID(ctx, schemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate patient ID: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s.patients 
		(id, patient_id, keycloak_user_id, first_name, last_name, email, phone_number, date_of_birth, address, 
		 emergency_contact_name, emergency_contact_phone, medical_notes, careplan_type, careplan_frequency, 
		 is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, true, $15)
		RETURNING id, patient_id, keycloak_user_id, first_name, last_name, email, phone_number, date_of_birth, address, 
				  emergency_contact_name, emergency_contact_phone, medical_notes, careplan_type, 
				  careplan_frequency, is_active, created_at
	`, pq.QuoteIdentifier(schemaName))

	var patient PatientResponse
	var dob sql.NullString
	var email sql.NullString
	var phoneNumber sql.NullString
	var address sql.NullString
	var emergencyContactName sql.NullString
	var emergencyContactPhone sql.NullString
	var medicalNotes sql.NullString
	var careplanType sql.NullString
	var careplanFrequency sql.NullString
	var patientIDStr sql.NullString

	err = r.db.QueryRowContext(ctx, query,
		patientID,
		patientDisplayID, // Use generated patient ID like PT-0001
		keycloakUserID,
		req.FirstName,
		req.LastName,
		req.Email,
		req.PhoneNumber,
		req.DateOfBirth,
		req.Address,
		req.EmergencyContactName,
		req.EmergencyContactPhone,
		req.MedicalNotes,
		req.CareplanType,
		req.CareplanFrequency,
		createdAt,
	).Scan(
		&patient.ID,
		&patientIDStr,
		&patient.KeycloakUserID,
		&patient.FirstName,
		&patient.LastName,
		&email,
		&phoneNumber,
		&dob,
		&address,
		&emergencyContactName,
		&emergencyContactPhone,
		&medicalNotes,
		&careplanType,
		&careplanFrequency,
		&patient.IsActive,
		&patient.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to insert patient: %w", err)
	}

	if patientIDStr.Valid {
		patient.PatientID = patientIDStr.String
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
	if careplanType.Valid {
		patient.CareplanType = careplanType.String
	}
	if careplanFrequency.Valid {
		patient.CareplanFrequency = careplanFrequency.String
	}

	// Publish patient.created event
	if r.publisher != nil {
		event := messaging.PatientCreatedEvent{
			BaseEvent: messaging.NewBaseEvent(messaging.EventPatientCreated),
			Data: messaging.PatientCreatedData{
				PatientID:      patient.ID,
				KeycloakUserID: keycloakUserID,
				OrganizationID: orgID,
				FirstName:      patient.FirstName,
				LastName:       patient.LastName,
				Email:          patient.Email,
				PhoneNumber:    patient.PhoneNumber,
				DateOfBirth:    getStringValue(patient.DateOfBirth),
				IsActive:       patient.IsActive,
				CreatedAt:      patient.CreatedAt,
			},
		}

		if err := r.publisher.Publish(ctx, messaging.EventPatientCreated, event); err != nil {
			log.Printf("Warning: failed to publish patient.created event: %v", err)
		}
	}

	return &patient, nil
}

// Helper function to safely get string value from pointer
func getStringValue(ptr *string) string {
	if ptr != nil {
		return *ptr
	}
	return ""
}

func (r *Repository) ListPatients(ctx context.Context, schemaName string) ([]PatientResponse, error) {
	query := fmt.Sprintf(`
		SELECT id, patient_id, keycloak_user_id, first_name, last_name, email, phone_number, date_of_birth, address, 
			   emergency_contact_name, emergency_contact_phone, medical_notes, careplan_type, 
			   careplan_frequency, is_active, created_at, updated_at
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
		var careplanType sql.NullString
		var careplanFrequency sql.NullString
		var updatedAt sql.NullTime
		var patientIDStr sql.NullString

		err := rows.Scan(
			&patient.ID,
			&patientIDStr,
			&patient.KeycloakUserID,
			&patient.FirstName,
			&patient.LastName,
			&email,
			&phoneNumber,
			&dob,
			&address,
			&emergencyContactName,
			&emergencyContactPhone,
			&medicalNotes,
			&careplanType,
			&careplanFrequency,
			&patient.IsActive,
			&patient.CreatedAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan patient: %w", err)
		}

		if patientIDStr.Valid {
			patient.PatientID = patientIDStr.String
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
		if careplanType.Valid {
			patient.CareplanType = careplanType.String
		}
		if careplanFrequency.Valid {
			patient.CareplanFrequency = careplanFrequency.String
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

// ListPatientsWithPagination retrieves patients with pagination support
func (r *Repository) ListPatientsWithPagination(ctx context.Context, schemaName string, limit, offset int, search string) ([]PatientResponse, int, error) {
	// Build WHERE clause for search
	whereClause := "WHERE deleted_at IS NULL"
	countWhereClause := "WHERE deleted_at IS NULL"
	args := []interface{}{limit, offset}

	if search != "" {
		whereClause += ` AND (first_name ILIKE $3 OR last_name ILIKE $3 OR email ILIKE $3)`
		countWhereClause += ` AND (first_name ILIKE $1 OR last_name ILIKE $1 OR email ILIKE $1)`
		args = append(args, "%"+search+"%")
	}

	// First, get total count
	var totalCount int
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM %s.patients
		%s
	`, pq.QuoteIdentifier(schemaName), countWhereClause)

	if search != "" {
		err := r.db.QueryRowContext(ctx, countQuery, "%"+search+"%").Scan(&totalCount)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to count patients: %w", err)
		}
	} else {
		err := r.db.QueryRowContext(ctx, countQuery).Scan(&totalCount)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to count patients: %w", err)
		}
	}

	// Then get paginated results
	query := fmt.Sprintf(`
		SELECT id, patient_id, keycloak_user_id, first_name, last_name, email, phone_number, date_of_birth, address, 
			   emergency_contact_name, emergency_contact_phone, medical_notes, careplan_type, 
			   careplan_frequency, is_active, created_at, updated_at
		FROM %s.patients
		%s
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, pq.QuoteIdentifier(schemaName), whereClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query patients: %w", err)
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
		var careplanType sql.NullString
		var careplanFrequency sql.NullString
		var updatedAt sql.NullTime
		var patientIDStr sql.NullString

		err := rows.Scan(
			&patient.ID,
			&patientIDStr,
			&patient.KeycloakUserID,
			&patient.FirstName,
			&patient.LastName,
			&email,
			&phoneNumber,
			&dob,
			&address,
			&emergencyContactName,
			&emergencyContactPhone,
			&medicalNotes,
			&careplanType,
			&careplanFrequency,
			&patient.IsActive,
			&patient.CreatedAt,
			&updatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan patient: %w", err)
		}

		if patientIDStr.Valid {
			patient.PatientID = patientIDStr.String
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
		if careplanType.Valid {
			patient.CareplanType = careplanType.String
		}
		if careplanFrequency.Valid {
			patient.CareplanFrequency = careplanFrequency.String
		}
		if updatedAt.Valid {
			patient.UpdatedAt = &updatedAt.Time
		}

		patients = append(patients, patient)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating patients: %w", err)
	}

	return patients, totalCount, nil
}

// ListActivePatientsWithPagination retrieves active patients (not soft deleted and is_active = true) with pagination support
func (r *Repository) ListActivePatientsWithPagination(ctx context.Context, schemaName string, limit, offset int, search string) ([]PatientResponse, int, error) {
	// Build WHERE clause for search
	whereClause := "WHERE deleted_at IS NULL AND is_active = true"
	countWhereClause := "WHERE deleted_at IS NULL AND is_active = true"
	args := []interface{}{limit, offset}

	if search != "" {
		whereClause += ` AND (first_name ILIKE $3 OR last_name ILIKE $3 OR email ILIKE $3)`
		countWhereClause += ` AND (first_name ILIKE $1 OR last_name ILIKE $1 OR email ILIKE $1)`
		args = append(args, "%"+search+"%")
	}

	// First, get total count of active patients
	var totalCount int
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM %s.patients
		%s
	`, pq.QuoteIdentifier(schemaName), countWhereClause)

	if search != "" {
		err := r.db.QueryRowContext(ctx, countQuery, "%"+search+"%").Scan(&totalCount)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to count active patients: %w", err)
		}
	} else {
		err := r.db.QueryRowContext(ctx, countQuery).Scan(&totalCount)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to count active patients: %w", err)
		}
	}

	// Then get paginated results
	query := fmt.Sprintf(`
		SELECT id, patient_id, keycloak_user_id, first_name, last_name, email, phone_number, date_of_birth, address, 
			   emergency_contact_name, emergency_contact_phone, medical_notes, careplan_type, 
			   careplan_frequency, is_active, created_at, updated_at
		FROM %s.patients
		%s
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, pq.QuoteIdentifier(schemaName), whereClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query active patients: %w", err)
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
		var careplanType sql.NullString
		var careplanFrequency sql.NullString
		var updatedAt sql.NullTime
		var patientIDStr sql.NullString

		err := rows.Scan(
			&patient.ID,
			&patientIDStr,
			&patient.KeycloakUserID,
			&patient.FirstName,
			&patient.LastName,
			&email,
			&phoneNumber,
			&dob,
			&address,
			&emergencyContactName,
			&emergencyContactPhone,
			&medicalNotes,
			&careplanType,
			&careplanFrequency,
			&patient.IsActive,
			&patient.CreatedAt,
			&updatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan patient: %w", err)
		}

		if patientIDStr.Valid {
			patient.PatientID = patientIDStr.String
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
		if careplanType.Valid {
			patient.CareplanType = careplanType.String
		}
		if careplanFrequency.Valid {
			patient.CareplanFrequency = careplanFrequency.String
		}
		if updatedAt.Valid {
			patient.UpdatedAt = &updatedAt.Time
		}

		patients = append(patients, patient)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating patients: %w", err)
	}

	return patients, totalCount, nil
}

func (r *Repository) GetPatient(ctx context.Context, schemaName string, id string) (*PatientResponse, error) {
	query := fmt.Sprintf(`
		SELECT id, patient_id, keycloak_user_id, first_name, last_name, email, phone_number, date_of_birth, address, 
			   emergency_contact_name, emergency_contact_phone, medical_notes, careplan_type, 
			   careplan_frequency, is_active, created_at, updated_at
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
	var careplanType sql.NullString
	var careplanFrequency sql.NullString
	var updatedAt sql.NullTime
	var patientIDStr sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&patient.ID,
		&patientIDStr,
		&patient.KeycloakUserID,
		&patient.FirstName,
		&patient.LastName,
		&email,
		&phoneNumber,
		&dob,
		&address,
		&emergencyContactName,
		&emergencyContactPhone,
		&medicalNotes,
		&careplanType,
		&careplanFrequency,
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

	if patientIDStr.Valid {
		patient.PatientID = patientIDStr.String
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
	if careplanType.Valid {
		patient.CareplanType = careplanType.String
	}
	if careplanFrequency.Valid {
		patient.CareplanFrequency = careplanFrequency.String
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
	if req.CareplanType != nil {
		updates = append(updates, fmt.Sprintf("careplan_type = $%d", argIndex))
		args = append(args, *req.CareplanType)
		argIndex++
	}
	if req.CareplanFrequency != nil {
		updates = append(updates, fmt.Sprintf("careplan_frequency = $%d", argIndex))
		args = append(args, *req.CareplanFrequency)
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
		RETURNING id, patient_id, keycloak_user_id, first_name, last_name, email, phone_number, date_of_birth, address, 
				  emergency_contact_name, emergency_contact_phone, medical_notes, careplan_type, 
				  careplan_frequency, is_active, created_at, updated_at
	`, pq.QuoteIdentifier(schemaName), strings.Join(updates, ", "), argIndex)

	var patient PatientResponse
	var dob sql.NullString
	var address sql.NullString
	var email sql.NullString
	var phoneNumber sql.NullString
	var emergencyContactName sql.NullString
	var emergencyContactPhone sql.NullString
	var medicalNotes sql.NullString
	var careplanType sql.NullString
	var careplanFrequency sql.NullString
	var updatedAt sql.NullTime
	var patientIDStr sql.NullString

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&patient.ID,
		&patientIDStr,
		&patient.KeycloakUserID,
		&patient.FirstName,
		&patient.LastName,
		&email,
		&phoneNumber,
		&dob,
		&address,
		&emergencyContactName,
		&emergencyContactPhone,
		&medicalNotes,
		&careplanType,
		&careplanFrequency,
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

	// Handle nullable fields
	if patientIDStr.Valid {
		patient.PatientID = patientIDStr.String
	}
	if email.Valid {
		patient.Email = email.String
	}
	if phoneNumber.Valid {
		patient.PhoneNumber = phoneNumber.String
	}
	if dob.Valid {
		patient.DateOfBirth = &dob.String
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
	if careplanType.Valid {
		patient.CareplanType = careplanType.String
	}
	if careplanFrequency.Valid {
		patient.CareplanFrequency = careplanFrequency.String
	}
	if updatedAt.Valid {
		patient.UpdatedAt = &updatedAt.Time
	}

	return &patient, nil
}

func (r *Repository) DeletePatient(ctx context.Context, schemaName string, orgID string, id string) error {
	query := fmt.Sprintf(`
		UPDATE %s.patients
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`, pq.QuoteIdentifier(schemaName))

	deletedAt := time.Now()
	result, err := r.db.ExecContext(ctx, query, deletedAt, id)
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

	// Publish patient.deleted event
	if r.publisher != nil {
		event := messaging.PatientDeletedEvent{
			BaseEvent: messaging.NewBaseEvent(messaging.EventPatientDeleted),
			Data: messaging.PatientDeletedData{
				PatientID:      id,
				OrganizationID: orgID,
				DeletedAt:      deletedAt,
			},
		}

		if err := r.publisher.Publish(ctx, messaging.EventPatientDeleted, event); err != nil {
			log.Printf("Warning: failed to publish patient.deleted event: %v", err)
		}
	}

	return nil
}
