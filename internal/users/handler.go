package users

import (
	"encoding/json"
	"log"
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

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.FromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	targetOrgID := r.Header.Get("X-Organization-ID")

	user, err := h.service.CreateUser(req, principal, targetOrgID)
	if err != nil {
		log.Printf("Failed to create user: %v", err)

		switch err {
		case ErrMissingUsername, ErrMissingEmail, ErrMissingFirstName, ErrMissingLastName, ErrMissingRole, ErrMissingPassword:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case ErrRoleNotAllowed:
			http.Error(w, err.Error(), http.StatusForbidden)
		case ErrInvalidOrgSchema:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "failed to create user", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.FromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	targetOrgID := r.Header.Get("X-Organization-ID")

	users, err := h.service.ListUsers(principal, targetOrgID)
	if err != nil {
		log.Printf("Failed to list users: %v", err)

		if err == ErrInvalidOrgSchema {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if err == ErrForbidden {
			http.Error(w, err.Error(), http.StatusForbidden)
		} else {
			http.Error(w, "failed to list users", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": users,
		"count": len(users),
	})
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.FromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	userID := vars["id"]

	targetOrgID := r.Header.Get("X-Organization-ID")

	user, err := h.service.GetUser(userID, principal, targetOrgID)
	if err != nil {
		log.Printf("Failed to get user: %v", err)

		switch err {
		case ErrUserNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case ErrForbidden:
			http.Error(w, err.Error(), http.StatusForbidden)
		case ErrInvalidOrgSchema:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "failed to get user", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.FromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	userID := vars["id"]

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	targetOrgID := r.Header.Get("X-Organization-ID")

	user, err := h.service.UpdateUser(userID, req, principal, targetOrgID)
	if err != nil {
		log.Printf("Failed to update user: %v", err)

		switch err {
		case ErrUserNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case ErrForbidden:
			http.Error(w, err.Error(), http.StatusForbidden)
		case ErrInvalidOrgSchema:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "failed to update user", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.FromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	userID := vars["id"]

	targetOrgID := r.Header.Get("X-Organization-ID")

	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err := h.service.ResetPassword(userID, req, principal, targetOrgID)
	if err != nil {
		log.Printf("Failed to reset password: %v", err)

		switch err {
		case ErrUserNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case ErrForbidden:
			http.Error(w, err.Error(), http.StatusForbidden)
		case ErrInvalidOrgSchema:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "failed to reset password", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "password reset successfully",
	})
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.FromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	userID := vars["id"]

	err := h.service.DeleteUser(userID, principal)
	if err != nil {
		log.Printf("Failed to delete user: %v", err)

		switch err {
		case ErrUserNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case ErrForbidden:
			http.Error(w, err.Error(), http.StatusForbidden)
		case ErrInvalidOrgSchema:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "failed to delete user", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.FromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.service.UpdateMyProfile(req, principal)
	if err != nil {
		log.Printf("Failed to update profile: %v", err)

		switch err {
		case ErrUserNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case ErrInvalidOrgSchema:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "failed to update profile", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
