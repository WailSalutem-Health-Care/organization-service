package http

import (
	"database/sql"
	"net/http"

	"github.com/WailSalutem-Health-Care/organization-service/internal/auth"
	"github.com/WailSalutem-Health-Care/organization-service/internal/organization"
	"github.com/WailSalutem-Health-Care/organization-service/internal/patient"
	"github.com/gorilla/mux"
)

// SetupRouter initializes all routes for the application
func SetupRouter(db *sql.DB, verifier *auth.Verifier, perms map[string][]string) *mux.Router {
	// Initialize organization components
	orgRepo := organization.NewRepository(db)
	orgService := organization.NewService(orgRepo)
	orgHandler := organization.NewHandler(orgService)

	// Initialize patient components
	patientRepo := patient.NewRepository(db)
	patientService := patient.NewService(patientRepo)
	patientHandler := patient.NewHandler(patientService, db)

	r := mux.NewRouter()

	// Public health endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"organization-service"}`))
	}).Methods("GET")

	// Organization routes (SUPER_ADMIN only)
	r.Handle("/organizations",
		auth.Middleware(verifier)(
			auth.RequirePermission("organization:create", perms)(
				http.HandlerFunc(orgHandler.CreateOrganization),
			),
		),
	).Methods("POST")

	r.Handle("/organizations",
		auth.Middleware(verifier)(
			auth.RequirePermission("organization:view", perms)(
				http.HandlerFunc(orgHandler.ListOrganizations),
			),
		),
	).Methods("GET")

	r.Handle("/organizations/{id}",
		auth.Middleware(verifier)(
			auth.RequirePermission("organization:view", perms)(
				http.HandlerFunc(orgHandler.GetOrganization),
			),
		),
	).Methods("GET")

	r.Handle("/organizations/{id}",
		auth.Middleware(verifier)(
			auth.RequirePermission("organization:update", perms)(
				http.HandlerFunc(orgHandler.UpdateOrganization),
			),
		),
	).Methods("PUT")

	r.Handle("/organizations/{id}",
		auth.Middleware(verifier)(
			auth.RequirePermission("organization:delete", perms)(
				http.HandlerFunc(orgHandler.DeleteOrganization),
			),
		),
	).Methods("DELETE")

	// Patient routes (ORG_ADMIN and CAREGIVER can view, ORG_ADMIN can manage)
	r.Handle("/patients",
		auth.Middleware(verifier)(
			auth.RequirePermission("patient:create", perms)(
				http.HandlerFunc(patientHandler.CreatePatient),
			),
		),
	).Methods("POST")

	r.Handle("/patients",
		auth.Middleware(verifier)(
			auth.RequirePermission("patient:view", perms)(
				http.HandlerFunc(patientHandler.ListPatients),
			),
		),
	).Methods("GET")

	r.Handle("/patients/{id}",
		auth.Middleware(verifier)(
			auth.RequirePermission("patient:view", perms)(
				http.HandlerFunc(patientHandler.GetPatient),
			),
		),
	).Methods("GET")

	r.Handle("/patients/{id}",
		auth.Middleware(verifier)(
			auth.RequirePermission("patient:update", perms)(
				http.HandlerFunc(patientHandler.UpdatePatient),
			),
		),
	).Methods("PUT")

	r.Handle("/patients/{id}",
		auth.Middleware(verifier)(
			auth.RequirePermission("patient:delete", perms)(
				http.HandlerFunc(patientHandler.DeletePatient),
			),
		),
	).Methods("DELETE")

	return r
}
