package patient

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/WailSalutem-Health-Care/organization-service/internal/auth"
	"github.com/WailSalutem-Health-Care/organization-service/internal/organization"
	"github.com/gorilla/mux"
)

type Handler struct {
	service *Service
	db      *sql.DB
}

func NewHandler(service *Service, db *sql.DB) *Handler {
	return &Handler{
		service: service,
		db:      db,
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
	_, ok := auth.FromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated", "User not authenticated")
		return
	}

	// Get organization ID from header
	orgID := r.Header.Get("X-Organization-ID")
	if orgID == "" {
		respondError(w, http.StatusBadRequest, "missing_org", "X-Organization-ID header is required")
		return
	}

	schemaName, err := organization.GetSchemaNameByOrgID(r.Context(), h.db, orgID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "schema_lookup_failed", "Failed to lookup organization schema: "+err.Error())
		return
	}
	if schemaName == "" {
		respondError(w, http.StatusNotFound, "org_not_found", "Organization schema not found")
		return
	}

	var req CreatePatientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload: "+err.Error())
		return
	}

	if req.FullName == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "Full name is required")
		return
	}

	patient, err := h.service.CreatePatient(r.Context(), schemaName, req)
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
	_, ok := auth.FromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated", "User not authenticated")
		return
	}

	// Get organization ID from header
	orgID := r.Header.Get("X-Organization-ID")
	if orgID == "" {
		respondError(w, http.StatusBadRequest, "missing_org", "X-Organization-ID header is required")
		return
	}

	schemaName, err := organization.GetSchemaNameByOrgID(r.Context(), h.db, orgID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "schema_lookup_failed", "Failed to lookup organization schema: "+err.Error())
		return
	}
	if schemaName == "" {
		respondError(w, http.StatusNotFound, "org_not_found", "Organization schema not found")
		return
	}

	patients, err := h.service.ListPatients(r.Context(), schemaName)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PatientListResponse{
		Success:  true,
		Patients: patients,
		Total:    len(patients),
	})
}

func (h *Handler) GetPatient(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.FromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated", "User not authenticated")
		return
	}

	// Get organization ID from header
	orgID := r.Header.Get("X-Organization-ID")
	if orgID == "" {
		respondError(w, http.StatusBadRequest, "missing_org", "X-Organization-ID header is required")
		return
	}

	schemaName, err := organization.GetSchemaNameByOrgID(r.Context(), h.db, orgID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "schema_lookup_failed", "Failed to lookup organization schema: "+err.Error())
		return
	}
	if schemaName == "" {
		respondError(w, http.StatusNotFound, "org_not_found", "Organization schema not found")
		return
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

func (h *Handler) UpdatePatient(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.FromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated", "User not authenticated")
		return
	}

	// Get organization ID from header
	orgID := r.Header.Get("X-Organization-ID")
	if orgID == "" {
		respondError(w, http.StatusBadRequest, "missing_org", "X-Organization-ID header is required")
		return
	}

	schemaName, err := organization.GetSchemaNameByOrgID(r.Context(), h.db, orgID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "schema_lookup_failed", "Failed to lookup organization schema: "+err.Error())
		return
	}
	if schemaName == "" {
		respondError(w, http.StatusNotFound, "org_not_found", "Organization schema not found")
		return
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
	_, ok := auth.FromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated", "User not authenticated")
		return
	}

	// Get organization ID from header
	orgID := r.Header.Get("X-Organization-ID")
	if orgID == "" {
		respondError(w, http.StatusBadRequest, "missing_org", "X-Organization-ID header is required")
		return
	}

	schemaName, err := organization.GetSchemaNameByOrgID(r.Context(), h.db, orgID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "schema_lookup_failed", "Failed to lookup organization schema: "+err.Error())
		return
	}
	if schemaName == "" {
		respondError(w, http.StatusNotFound, "org_not_found", "Organization schema not found")
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "Patient ID is required")
		return
	}

	err = h.service.DeletePatient(r.Context(), schemaName, id)
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
