package messaging

import (
	"fmt"
	"time"
)

// Event routing keys as constants
const (
	// Patient events
	EventPatientCreated       = "patient.created"
	EventPatientDeleted       = "patient.deleted"
	EventPatientUpdated       = "patient.updated"
	EventPatientStatusChanged = "patient.status_changed"

	// User events (for staff: CAREGIVER, ORG_ADMIN, etc.)
	EventUserCreated       = "user.created"
	EventUserDeleted       = "user.deleted"
	EventUserStatusChanged = "user.status_changed"
	EventUserRoleChanged   = "user.role_changed"

	// Organization events
	EventOrganizationDeleted       = "organization.deleted"
	EventOrganizationStatusChanged = "organization.status_changed"
)

// BaseEvent contains common fields for all events
type BaseEvent struct {
	EventType   string    `json:"event_type"`
	EventID     string    `json:"event_id"`
	Timestamp   time.Time `json:"timestamp"`
	ServiceName string    `json:"service_name"`
}

// PatientCreatedEvent represents a patient creation event
type PatientCreatedEvent struct {
	BaseEvent
	Data PatientCreatedData `json:"data"`
}

type PatientCreatedData struct {
	PatientID      string    `json:"patient_id"`
	KeycloakUserID string    `json:"keycloak_user_id"`
	OrganizationID string    `json:"organization_id"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Email          string    `json:"email"`
	PhoneNumber    string    `json:"phone_number,omitempty"`
	DateOfBirth    string    `json:"date_of_birth,omitempty"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
}

// PatientDeletedEvent represents a patient deletion event
type PatientDeletedEvent struct {
	BaseEvent
	Data PatientDeletedData `json:"data"`
}

type PatientDeletedData struct {
	PatientID      string    `json:"patient_id"`
	OrganizationID string    `json:"organization_id"`
	DeletedAt      time.Time `json:"deleted_at"`
}

// PatientStatusChangedEvent represents a patient status change event
type PatientStatusChangedEvent struct {
	BaseEvent
	Data PatientStatusChangedData `json:"data"`
}

type PatientStatusChangedData struct {
	PatientID      string    `json:"patient_id"`
	OrganizationID string    `json:"organization_id"`
	OldStatus      string    `json:"old_status"`
	NewStatus      string    `json:"new_status"`
	ChangedAt      time.Time `json:"changed_at"`
}

// UserCreatedEvent represents a user creation event (CAREGIVER, ORG_ADMIN, etc.)
type UserCreatedEvent struct {
	BaseEvent
	Data UserCreatedData `json:"data"`
}

type UserCreatedData struct {
	UserID         string    `json:"user_id"`
	KeycloakUserID string    `json:"keycloak_user_id"`
	OrganizationID string    `json:"organization_id"`
	Email          string    `json:"email"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Role           string    `json:"role"` // CAREGIVER, ORG_ADMIN, INSURER, MUNICIPALITY
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
}

// UserDeletedEvent represents a user deletion event
type UserDeletedEvent struct {
	BaseEvent
	Data UserDeletedData `json:"data"`
}

type UserDeletedData struct {
	UserID         string    `json:"user_id"`
	OrganizationID string    `json:"organization_id"`
	Role           string    `json:"role"`
	DeletedAt      time.Time `json:"deleted_at"`
}

// UserStatusChangedEvent represents a user status change event
type UserStatusChangedEvent struct {
	BaseEvent
	Data UserStatusChangedData `json:"data"`
}

type UserStatusChangedData struct {
	UserID         string    `json:"user_id"`
	OrganizationID string    `json:"organization_id"`
	Role           string    `json:"role"`
	OldStatus      string    `json:"old_status"` // "active" or "inactive"
	NewStatus      string    `json:"new_status"`
	ChangedAt      time.Time `json:"changed_at"`
}

// UserRoleChangedEvent represents a user role change event
type UserRoleChangedEvent struct {
	BaseEvent
	Data UserRoleChangedData `json:"data"`
}

type UserRoleChangedData struct {
	UserID         string    `json:"user_id"`
	OrganizationID string    `json:"organization_id"`
	OldRole        string    `json:"old_role"`
	NewRole        string    `json:"new_role"`
	ChangedAt      time.Time `json:"changed_at"`
}

// OrganizationDeletedEvent represents an organization deletion event
type OrganizationDeletedEvent struct {
	BaseEvent
	Data OrganizationDeletedData `json:"data"`
}

type OrganizationDeletedData struct {
	OrganizationID   string    `json:"organization_id"`
	OrganizationName string    `json:"organization_name"`
	SchemaName       string    `json:"schema_name"`
	DeletedAt        time.Time `json:"deleted_at"`
}

// OrganizationStatusChangedEvent represents an organization status change event
type OrganizationStatusChangedEvent struct {
	BaseEvent
	Data OrganizationStatusChangedData `json:"data"`
}

type OrganizationStatusChangedData struct {
	OrganizationID string    `json:"organization_id"`
	OldStatus      string    `json:"old_status"` // "active", "suspended", etc.
	NewStatus      string    `json:"new_status"`
	ChangedAt      time.Time `json:"changed_at"`
}

// NewBaseEvent creates a base event with common fields
func NewBaseEvent(eventType string) BaseEvent {
	return BaseEvent{
		EventType:   eventType,
		EventID:     fmt.Sprintf("%d", time.Now().UnixNano()),
		Timestamp:   time.Now().UTC(), // Explicitly set to UTC
		ServiceName: "organization-service",
	}
}
