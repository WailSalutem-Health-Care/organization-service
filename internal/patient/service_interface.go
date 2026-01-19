package patient

import (
	"context"

	"github.com/WailSalutem-Health-Care/organization-service/internal/pagination"
)

// ServiceInterface defines the contract for patient business logic operations
type ServiceInterface interface {
	CreatePatient(ctx context.Context, schemaName, orgID string, req CreatePatientRequest) (*PatientResponse, error)
	GetPatient(ctx context.Context, schemaName, id string) (*PatientResponse, error)
	GetMyPatient(ctx context.Context, schemaName string, keycloakUserID string) (*PatientResponse, error)
	ListPatients(ctx context.Context, schemaName string) ([]PatientResponse, error)
	ListPatientsWithPagination(ctx context.Context, schemaName string, params pagination.Params) (*PaginatedPatientListResponse, error)
	ListActivePatientsWithPagination(ctx context.Context, schemaName string, params pagination.Params) (*PaginatedPatientListResponse, error)
	UpdatePatient(ctx context.Context, schemaName, id string, req UpdatePatientRequest) (*PatientResponse, error)
	DeletePatient(ctx context.Context, schemaName, orgID, id string) error
}

// SchemaLookup defines the contract for looking up organization schemas
type SchemaLookup interface {
	GetSchemaNameByOrgID(ctx context.Context, orgID string) (string, error)
}
