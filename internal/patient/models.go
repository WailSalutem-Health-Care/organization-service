package patient

import "time"

// CreatePatientRequest represents the request to create a new patient user
type CreatePatientRequest struct {
	// Authentication fields
	Username          string `json:"username" validate:"required"`
	TemporaryPassword string `json:"temporaryPassword"`
	SendResetEmail    bool   `json:"sendResetEmail"`

	// Personal information
	FirstName   string `json:"firstName" validate:"required"`
	LastName    string `json:"lastName" validate:"required"`
	Email       string `json:"email" validate:"required"`
	PhoneNumber string `json:"phoneNumber"`

	// Patient-specific fields
	DateOfBirth           string `json:"dateOfBirth" validate:"required"` // Format: YYYY-MM-DD
	Address               string `json:"address" validate:"required"`
	EmergencyContactName  string `json:"emergencyContactName"`
	EmergencyContactPhone string `json:"emergencyContactPhone"`
	MedicalNotes          string `json:"medicalNotes"`

	// Care plan fields
	CareplanType      string `json:"careplanType"`      // e.g., "basic", "intensive", "palliative"
	CareplanFrequency string `json:"careplanFrequency"` // e.g., "daily", "weekly", "monthly"
}

// UpdatePatientRequest represents the request to update a patient
type UpdatePatientRequest struct {
	FirstName             *string `json:"first_name,omitempty"`
	LastName              *string `json:"last_name,omitempty"`
	Email                 *string `json:"email,omitempty"`
	PhoneNumber           *string `json:"phone_number,omitempty"`
	DateOfBirth           *string `json:"date_of_birth,omitempty"`
	Address               *string `json:"address,omitempty"`
	EmergencyContactName  *string `json:"emergency_contact_name,omitempty"`
	EmergencyContactPhone *string `json:"emergency_contact_phone,omitempty"`
	MedicalNotes          *string `json:"medical_notes,omitempty"`
	IsActive              *bool   `json:"is_active,omitempty"`
	CareplanType          *string `json:"careplan_type,omitempty"`
	CareplanFrequency     *string `json:"careplan_frequency,omitempty"`
}

// PatientResponse represents the patient data returned to clients
type PatientResponse struct {
	ID                    string     `json:"id"`
	KeycloakUserID        string     `json:"keycloak_user_id"`
	FirstName             string     `json:"first_name"`
	LastName              string     `json:"last_name"`
	Email                 string     `json:"email"`
	PhoneNumber           string     `json:"phone_number"`
	DateOfBirth           *string    `json:"date_of_birth,omitempty"`
	Address               string     `json:"address"`
	EmergencyContactName  string     `json:"emergency_contact_name"`
	EmergencyContactPhone string     `json:"emergency_contact_phone"`
	MedicalNotes          string     `json:"medical_notes"`
	CareplanType          string     `json:"careplan_type"`
	CareplanFrequency     string     `json:"careplan_frequency"`
	IsActive              bool       `json:"is_active"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             *time.Time `json:"updated_at,omitempty"`
}
