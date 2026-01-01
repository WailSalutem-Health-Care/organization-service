package patient

import "time"

// CreatePatientRequest represents the request to create a new patient
type CreatePatientRequest struct {
	FullName              string `json:"full_name" validate:"required"`
	Email                 string `json:"email"`
	PhoneNumber           string `json:"phone_number"`
	DateOfBirth           string `json:"date_of_birth"` // Format: YYYY-MM-DD
	Address               string `json:"address"`
	EmergencyContactName  string `json:"emergency_contact_name"`
	EmergencyContactPhone string `json:"emergency_contact_phone"`
	MedicalNotes          string `json:"medical_notes"`
}

// UpdatePatientRequest represents the request to update a patient
type UpdatePatientRequest struct {
	FullName              *string `json:"full_name,omitempty"`
	Email                 *string `json:"email,omitempty"`
	PhoneNumber           *string `json:"phone_number,omitempty"`
	DateOfBirth           *string `json:"date_of_birth,omitempty"`
	Address               *string `json:"address,omitempty"`
	EmergencyContactName  *string `json:"emergency_contact_name,omitempty"`
	EmergencyContactPhone *string `json:"emergency_contact_phone,omitempty"`
	MedicalNotes          *string `json:"medical_notes,omitempty"`
	IsActive              *bool   `json:"is_active,omitempty"`
}

// PatientResponse represents the patient data returned to clients
type PatientResponse struct {
	ID                    string     `json:"id"`
	FullName              string     `json:"full_name"`
	Email                 string     `json:"email"`
	PhoneNumber           string     `json:"phone_number"`
	DateOfBirth           *string    `json:"date_of_birth,omitempty"`
	Address               string     `json:"address"`
	EmergencyContactName  string     `json:"emergency_contact_name"`
	EmergencyContactPhone string     `json:"emergency_contact_phone"`
	MedicalNotes          string     `json:"medical_notes"`
	IsActive              bool       `json:"is_active"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             *time.Time `json:"updated_at,omitempty"`
}
