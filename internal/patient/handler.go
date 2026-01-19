package patient

import (
	"encoding/json"
	"net/http"

	"github.com/WailSalutem-Health-Care/organization-service/internal/auth"
	"github.com/WailSalutem-Health-Care/organization-service/internal/pagination"
	"github.com/gorilla/mux"
)

type Handler struct {
	service      ServiceInterface
	schemaLookup SchemaLookup
}

func NewHandler(service ServiceInterface, schemaLookup SchemaLookup) *Handler {
	return &Handler{
		service:      service,
		schemaLookup: schemaLookup,
	}
}

type PatientSuccessResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Patient *PatientResponse `json:"patient,omitempty"`
}

type PatientListResponse struct {
	Success  bool              `json:"success"`
	Patients []PatientResponse `json:"patients"`
	Total    int               `json:"total"`
}

func (h *Handler) CreatePatient(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.FromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated", "User not authenticated")
		return
	}

	var orgID string
	var schemaName string
	var err error

	// Check if user is SUPER_ADMIN
	isSuperAdmin := false
	for _, role := range principal.Roles {
		if role == "SUPER_ADMIN" {
			isSuperAdmin = true
			break
		}
	}

	if isSuperAdmin {
		// SUPER_ADMIN must provide X-Organization-ID header
		orgID = r.Header.Get("X-Organization-ID")
		if orgID == "" {
			respondError(w, http.StatusBadRequest, "missing_org", "X-Organization-ID header is required for SUPER_ADMIN")
			return
		}

		schemaName, err = h.schemaLookup.GetSchemaNameByOrgID(r.Context(), orgID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "schema_lookup_failed", "Failed to lookup organization schema: "+err.Error())
			return
		}
		if schemaName == "" {
			respondError(w, http.StatusNotFound, "org_not_found", "Organization schema not found")
			return
		}
	} else {
		// ORG_ADMIN: get org ID and schema from token
		orgID = principal.OrgID
		schemaName = principal.OrgSchemaName

		if orgID == "" || schemaName == "" {
			respondError(w, http.StatusBadRequest, "missing_org_info", "Organization information not found in token")
			return
		}
	}

	var req CreatePatientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload: "+err.Error())
		return
	}

	if req.FirstName == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "First name is required")
		return
	}

	if req.LastName == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "Last name is required")
		return
	}

	patient, err := h.service.CreatePatient(r.Context(), schemaName, orgID, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "creation_failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(PatientSuccessResponse{
		Success: true,
		Message: "Patient created successfully",
		Patient: patient,
	})
}

func (h *Handler) ListPatients(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.FromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated", "User not authenticated")
		return
	}

	var orgID string
	var schemaName string
	var err error

	// Check if user is SUPER_ADMIN
	isSuperAdmin := false
	for _, role := range principal.Roles {
		if role == "SUPER_ADMIN" {
			isSuperAdmin = true
			break
		}
	}

	if isSuperAdmin {
		// SUPER_ADMIN must provide X-Organization-ID header
		orgID = r.Header.Get("X-Organization-ID")
		if orgID == "" {
			respondError(w, http.StatusBadRequest, "missing_org", "X-Organization-ID header is required for SUPER_ADMIN")
			return
		}

		schemaName, err = h.schemaLookup.GetSchemaNameByOrgID(r.Context(), orgID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "schema_lookup_failed", "Failed to lookup organization schema: "+err.Error())
			return
		}
		if schemaName == "" {
			respondError(w, http.StatusNotFound, "org_not_found", "Organization schema not found")
			return
		}
	} else {
		// ORG_ADMIN: get org ID and schema from token
		orgID = principal.OrgID
		schemaName = principal.OrgSchemaName

		if orgID == "" || schemaName == "" {
			respondError(w, http.StatusBadRequest, "missing_org_info", "Organization information not found in token")
			return
		}
	}

	// Parse pagination parameters from query string
	params := pagination.ParseParams(r)

	// Get paginated patients
	response, err := h.service.ListPatientsWithPagination(r.Context(), schemaName, params)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) ListActivePatients(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.FromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated", "User not authenticated")
		return
	}

	var orgID string
	var schemaName string
	var err error

	// Check if user is SUPER_ADMIN
	isSuperAdmin := false
	for _, role := range principal.Roles {
		if role == "SUPER_ADMIN" {
			isSuperAdmin = true
			break
		}
	}

	if isSuperAdmin {
		// SUPER_ADMIN must provide X-Organization-ID header
		orgID = r.Header.Get("X-Organization-ID")
		if orgID == "" {
			respondError(w, http.StatusBadRequest, "missing_org", "X-Organization-ID header is required for SUPER_ADMIN")
			return
		}

		schemaName, err = h.schemaLookup.GetSchemaNameByOrgID(r.Context(), orgID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "schema_lookup_failed", "Failed to lookup organization schema: "+err.Error())
			return
		}
		if schemaName == "" {
			respondError(w, http.StatusNotFound, "org_not_found", "Organization schema not found")
			return
		}
	} else {
		// ORG_ADMIN: get org ID and schema from token
		orgID = principal.OrgID
		schemaName = principal.OrgSchemaName

		if orgID == "" || schemaName == "" {
			respondError(w, http.StatusBadRequest, "missing_org_info", "Organization information not found in token")
			return
		}
	}

	// Parse pagination parameters from query string
	params := pagination.ParseParams(r)

	// Get paginated active patients
	response, err := h.service.ListActivePatientsWithPagination(r.Context(), schemaName, params)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetPatient(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.FromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated", "User not authenticated")
		return
	}

	var orgID string
	var schemaName string
	var err error

	// Check if user is SUPER_ADMIN
	isSuperAdmin := false
	for _, role := range principal.Roles {
		if role == "SUPER_ADMIN" {
			isSuperAdmin = true
			break
		}
	}

	if isSuperAdmin {
		// SUPER_ADMIN must provide X-Organization-ID header
		orgID = r.Header.Get("X-Organization-ID")
		if orgID == "" {
			respondError(w, http.StatusBadRequest, "missing_org", "X-Organization-ID header is required for SUPER_ADMIN")
			return
		}

		schemaName, err = h.schemaLookup.GetSchemaNameByOrgID(r.Context(), orgID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "schema_lookup_failed", "Failed to lookup organization schema: "+err.Error())
			return
		}
		if schemaName == "" {
			respondError(w, http.StatusNotFound, "org_not_found", "Organization schema not found")
			return
		}
	} else {
		// ORG_ADMIN: get org ID and schema from token
		orgID = principal.OrgID
		schemaName = principal.OrgSchemaName

		if orgID == "" || schemaName == "" {
			respondError(w, http.StatusBadRequest, "missing_org_info", "Organization information not found in token")
			return
		}
	}

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "Patient ID is required")
		return
	}

	patient, err := h.service.GetPatient(r.Context(), schemaName, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PatientSuccessResponse{
		Success: true,
		Message: "Patient retrieved successfully",
		Patient: patient,
	})
}

func (h *Handler) GetMyPatient(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.FromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated", "User not authenticated")
		return
	}

	var orgID string
	var schemaName string
	var err error

	// Check if user is SUPER_ADMIN
	isSuperAdmin := false
	for _, role := range principal.Roles {
		if role == "SUPER_ADMIN" {
			isSuperAdmin = true
			break
		}
	}

	if isSuperAdmin {
		// SUPER_ADMIN must provide X-Organization-ID header
		orgID = r.Header.Get("X-Organization-ID")
		if orgID == "" {
			respondError(w, http.StatusBadRequest, "missing_org", "X-Organization-ID header is required for SUPER_ADMIN")
			return
		}

		schemaName, err = h.schemaLookup.GetSchemaNameByOrgID(r.Context(), orgID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "schema_lookup_failed", "Failed to lookup organization schema: "+err.Error())
			return
		}
		if schemaName == "" {
			respondError(w, http.StatusNotFound, "org_not_found", "Organization schema not found")
			return
		}
	} else {
		// Get org ID and schema from token
		orgID = principal.OrgID
		schemaName = principal.OrgSchemaName

		if orgID == "" || schemaName == "" {
			respondError(w, http.StatusBadRequest, "missing_org_info", "Organization information not found in token")
			return
		}
	}

	patient, err := h.service.GetMyPatient(r.Context(), schemaName, principal.UserID)
	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PatientSuccessResponse{
		Success: true,
		Message: "Patient retrieved successfully",
		Patient: patient,
	})
}

func (h *Handler) UpdatePatient(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.FromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated", "User not authenticated")
		return
	}

	var orgID string
	var schemaName string
	var err error

	// Check if user is SUPER_ADMIN
	isSuperAdmin := false
	for _, role := range principal.Roles {
		if role == "SUPER_ADMIN" {
			isSuperAdmin = true
			break
		}
	}

	if isSuperAdmin {
		// SUPER_ADMIN must provide X-Organization-ID header
		orgID = r.Header.Get("X-Organization-ID")
		if orgID == "" {
			respondError(w, http.StatusBadRequest, "missing_org", "X-Organization-ID header is required for SUPER_ADMIN")
			return
		}

		schemaName, err = h.schemaLookup.GetSchemaNameByOrgID(r.Context(), orgID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "schema_lookup_failed", "Failed to lookup organization schema: "+err.Error())
			return
		}
		if schemaName == "" {
			respondError(w, http.StatusNotFound, "org_not_found", "Organization schema not found")
			return
		}
	} else {
		// ORG_ADMIN: get org ID and schema from token
		orgID = principal.OrgID
		schemaName = principal.OrgSchemaName

		if orgID == "" || schemaName == "" {
			respondError(w, http.StatusBadRequest, "missing_org_info", "Organization information not found in token")
			return
		}
	}

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "Patient ID is required")
		return
	}

	var req UpdatePatientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload: "+err.Error())
		return
	}

	patient, err := h.service.UpdatePatient(r.Context(), schemaName, id, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PatientSuccessResponse{
		Success: true,
		Message: "Patient updated successfully",
		Patient: patient,
	})
}

func (h *Handler) DeletePatient(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.FromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated", "User not authenticated")
		return
	}

	var orgID string
	var schemaName string
	var err error

	// Check if user is SUPER_ADMIN
	isSuperAdmin := false
	for _, role := range principal.Roles {
		if role == "SUPER_ADMIN" {
			isSuperAdmin = true
			break
		}
	}

	if isSuperAdmin {
		// SUPER_ADMIN must provide X-Organization-ID header
		orgID = r.Header.Get("X-Organization-ID")
		if orgID == "" {
			respondError(w, http.StatusBadRequest, "missing_org", "X-Organization-ID header is required for SUPER_ADMIN")
			return
		}

		schemaName, err = h.schemaLookup.GetSchemaNameByOrgID(r.Context(), orgID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "schema_lookup_failed", "Failed to lookup organization schema: "+err.Error())
			return
		}
		if schemaName == "" {
			respondError(w, http.StatusNotFound, "org_not_found", "Organization schema not found")
			return
		}
	} else {
		// ORG_ADMIN: get org ID and schema from token
		orgID = principal.OrgID
		schemaName = principal.OrgSchemaName

		if orgID == "" || schemaName == "" {
			respondError(w, http.StatusBadRequest, "missing_org_info", "Organization information not found in token")
			return
		}
	}

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "Patient ID is required")
		return
	}

	err = h.service.DeletePatient(r.Context(), schemaName, orgID, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "deletion_failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Patient deleted successfully",
	})
}

func respondError(w http.ResponseWriter, statusCode int, errorType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   errorType,
		"message": message,
	})
}
