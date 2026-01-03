package organization

import (
	"encoding/json"
	"net/http"

	"github.com/WailSalutem-Health-Care/organization-service/internal/auth"
	"github.com/gorilla/mux"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Success      bool                  `json:"success"`
	Message      string                `json:"message"`
	Organization *OrganizationResponse `json:"organization,omitempty"`
}

type ListResponse struct {
	Success       bool                   `json:"success"`
	Organizations []OrganizationResponse `json:"organizations"`
	Total         int                    `json:"total"`
}

func (h *Handler) CreateOrganization(w http.ResponseWriter, r *http.Request) {

	_, ok := auth.FromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated", "User not authenticated")
		return
	}

	var req CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload: "+err.Error())
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "Organization name is required")
		return
	}

	org, err := h.service.CreateOrganization(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "creation_failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(SuccessResponse{
		Success:      true,
		Message:      "Organization created successfully with dedicated schema",
		Organization: org,
	})
}

func (h *Handler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.FromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated", "User not authenticated")
		return
	}

	orgs, err := h.service.ListOrganizations(r.Context(), principal)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "fetch_failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ListResponse{
		Success:       true,
		Organizations: orgs,
		Total:         len(orgs),
	})
}

func (h *Handler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.FromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated", "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "Organization ID is required")
		return
	}

	org, err := h.service.GetOrganization(r.Context(), id, principal)
	if err != nil {
		if err.Error() == "forbidden" {
			respondError(w, http.StatusForbidden, "forbidden", "You don't have permission to view this organization")
			return
		}
		respondError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SuccessResponse{
		Success:      true,
		Message:      "Organization retrieved successfully",
		Organization: org,
	})
}

func (h *Handler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.FromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated", "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "Organization ID is required")
		return
	}

	var req UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload: "+err.Error())
		return
	}

	org, err := h.service.UpdateOrganization(r.Context(), id, req, principal)
	if err != nil {
		if err.Error() == "forbidden" {
			respondError(w, http.StatusForbidden, "forbidden", "You don't have permission to update this organization")
			return
		}
		respondError(w, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SuccessResponse{
		Success:      true,
		Message:      "Organization updated successfully",
		Organization: org,
	})
}

func (h *Handler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.FromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated", "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		respondError(w, http.StatusBadRequest, "validation_error", "Organization ID is required")
		return
	}

	err := h.service.DeleteOrganization(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "deletion_failed", err.Error())
		return
	}

	// Return 204 No Content on successful deletion
	w.WriteHeader(http.StatusNoContent)
}

func respondError(w http.ResponseWriter, statusCode int, errorType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   errorType,
		Message: message,
	})
}
