package caregiver

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/WailSalutem-Health-Care/organisation-service/internal/auth"
	"github.com/WailSalutem-Health-Care/organisation-service/internal/organization"
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

type CaregiverSuccessResponse struct {
	Success   bool               `json:"success"`
	Message   string             `json:"message"`
	Caregiver *CaregiverResponse `json:"caregiver,omitempty"`
}

type CaregiverListResponse struct {
	Success    bool                `json:"success"`
	Caregivers []CaregiverResponse `json:"caregivers"`
	Total      int                 `json:"total"`
}

func (h *Handler) CreateCaregiver(w http.ResponseWriter, r *http.Request) {
	pr, ok := auth.FromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated", "User not authenticated")
		return
	}

	orgID := pr.OrgID
	if orgID == "" {
		respondError(w, http.StatusBadRequest, "missing_org", "Organization ID not found in token")
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

	var req CreateCaregiverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload: "+err.Error())
		return
	}

	if req.FullName == "" || req.Email == "" || req.KeycloakUserID == "" || req.Role == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "Full name, email, keycloak_user_id, and role are required")
		return
	}

	caregiver, err := h.service.CreateCaregiver(r.Context(), schemaName, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "creation_failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CaregiverSuccessResponse{
		Success:   true,
		Message:   "Caregiver created successfully",
		Caregiver: caregiver,
	})
}

func (h *Handler) ListCaregivers(w http.ResponseWriter, r *http.Request) {
	pr, ok := auth.FromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated", "User not authenticated")
		return
	}

	orgID := pr.OrgID
	if orgID == "" {
		respondError(w, http.StatusBadRequest, "missing_org", "Organization ID not found in token")
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

	caregivers, err := h.service.ListCaregivers(r.Context(), schemaName)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CaregiverListResponse{
		Success:    true,
		Caregivers: caregivers,
		Total:      len(caregivers),
	})
}

func (h *Handler) GetCaregiver(w http.ResponseWriter, r *http.Request) {
	pr, ok := auth.FromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated", "User not authenticated")
		return
	}

	orgID := pr.OrgID
	if orgID == "" {
		respondError(w, http.StatusBadRequest, "missing_org", "Organization ID not found in token")
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
		respondError(w, http.StatusBadRequest, "validation_error", "Caregiver ID is required")
		return
	}

	caregiver, err := h.service.GetCaregiver(r.Context(), schemaName, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CaregiverSuccessResponse{
		Success:   true,
		Message:   "Caregiver retrieved successfully",
		Caregiver: caregiver,
	})
}

func (h *Handler) UpdateCaregiver(w http.ResponseWriter, r *http.Request) {
	pr, ok := auth.FromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated", "User not authenticated")
		return
	}

	orgID := pr.OrgID
	if orgID == "" {
		respondError(w, http.StatusBadRequest, "missing_org", "Organization ID not found in token")
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
		respondError(w, http.StatusBadRequest, "validation_error", "Caregiver ID is required")
		return
	}

	var req UpdateCaregiverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload: "+err.Error())
		return
	}

	caregiver, err := h.service.UpdateCaregiver(r.Context(), schemaName, id, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CaregiverSuccessResponse{
		Success:   true,
		Message:   "Caregiver updated successfully",
		Caregiver: caregiver,
	})
}

func (h *Handler) DeleteCaregiver(w http.ResponseWriter, r *http.Request) {
	pr, ok := auth.FromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated", "User not authenticated")
		return
	}

	orgID := pr.OrgID
	if orgID == "" {
		respondError(w, http.StatusBadRequest, "missing_org", "Organization ID not found in token")
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
		respondError(w, http.StatusBadRequest, "validation_error", "Caregiver ID is required")
		return
	}

	err = h.service.DeleteCaregiver(r.Context(), schemaName, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "deletion_failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Caregiver deleted successfully",
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
