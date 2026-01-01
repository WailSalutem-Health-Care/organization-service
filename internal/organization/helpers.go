package organization

import (
	"context"
	"database/sql"
	"sync"
)


var (
	schemaCache      = make(map[string]string)
	schemaCacheMutex sync.RWMutex
	globalDB         *sql.DB
)


func InitializeSchemaHelper(db *sql.DB) {
	globalDB = db
}


func GetSchemaNameByOrgID(ctx context.Context, db *sql.DB, orgID string) (string, error) {

	schemaCacheMutex.RLock()
	if schemaName, ok := schemaCache[orgID]; ok {
		schemaCacheMutex.RUnlock()
		return schemaName, nil
	}
	schemaCacheMutex.RUnlock()


	query := `SELECT schema_name FROM wailsalutem.organizations WHERE id = $1 AND deleted_at IS NULL`
	var schemaName string
	err := db.QueryRowContext(ctx, query, orgID).Scan(&schemaName)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}


	schemaCacheMutex.Lock()
	schemaCache[orgID] = schemaName
	schemaCacheMutex.Unlock()

	return schemaName, nil
}


func ClearSchemaCache() {
	schemaCacheMutex.Lock()
	schemaCache = make(map[string]string)
	schemaCacheMutex.Unlock()
}
