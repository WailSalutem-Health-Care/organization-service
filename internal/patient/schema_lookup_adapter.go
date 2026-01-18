package patient

import (
	"context"
	"database/sql"

	"github.com/WailSalutem-Health-Care/organization-service/internal/organization"
)

// DBSchemaLookup implements SchemaLookup using the database
type DBSchemaLookup struct {
	db *sql.DB
}

// NewDBSchemaLookup creates a new DBSchemaLookup
func NewDBSchemaLookup(db *sql.DB) *DBSchemaLookup {
	return &DBSchemaLookup{db: db}
}

// GetSchemaNameByOrgID looks up the schema name for an organization
func (d *DBSchemaLookup) GetSchemaNameByOrgID(ctx context.Context, orgID string) (string, error) {
	return organization.GetSchemaNameByOrgID(ctx, d.db, orgID)
}
