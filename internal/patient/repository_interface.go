package patient

import "context"

// RepositoryInterface defines the contract for patient data access
type RepositoryInterface interface {
	CreatePatient(ctx context.Context, schemaName string, orgID string, keycloakUserID string, req CreatePatientRequest) (*PatientResponse, error)
	ListPatients(ctx context.Context, schemaName string) ([]PatientResponse, error)
	ListPatientsWithPagination(ctx context.Context, schemaName string, limit, offset int, search string) ([]PatientResponse, int, error)
	ListActivePatientsWithPagination(ctx context.Context, schemaName string, limit, offset int, search string) ([]PatientResponse, int, error)
	GetPatient(ctx context.Context, schemaName string, id string) (*PatientResponse, error)
	GetByKeycloakID(ctx context.Context, schemaName string, keycloakUserID string) (*PatientResponse, error)
	UpdatePatient(ctx context.Context, schemaName string, id string, req UpdatePatientRequest) (*PatientResponse, error)
	DeletePatient(ctx context.Context, schemaName string, orgID string, id string) error
}

// Ensure Repository implements RepositoryInterface
var _ RepositoryInterface = (*Repository)(nil)
