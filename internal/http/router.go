package http

import (
	"database/sql"
	"net/http"

	"github.com/WailSalutem-Health-Care/organisation-service/internal/auth"
	"github.com/WailSalutem-Health-Care/organisation-service/internal/caregiver"
	"github.com/WailSalutem-Health-Care/organisation-service/internal/organization"
	"github.com/WailSalutem-Health-Care/organisation-service/internal/patient"
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

	// Initialize caregiver components
	caregiverRepo := caregiver.NewRepository(db)
	caregiverService := caregiver.NewService(caregiverRepo)
	caregiverHandler := caregiver.NewHandler(caregiverService, db)

	r := mux.NewRouter()

	// Public health endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"organisation-service"}`))
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

	// Caregiver routes (ORG_ADMIN only)
	r.Handle("/caregivers",
		auth.Middleware(verifier)(
			auth.RequirePermission("caregiver:create", perms)(
				http.HandlerFunc(caregiverHandler.CreateCaregiver),
			),
		),
	).Methods("POST")

	r.Handle("/caregivers",
		auth.Middleware(verifier)(
			auth.RequirePermission("caregiver:view", perms)(
				http.HandlerFunc(caregiverHandler.ListCaregivers),
			),
		),
	).Methods("GET")

	r.Handle("/caregivers/{id}",
		auth.Middleware(verifier)(
			auth.RequirePermission("caregiver:view", perms)(
				http.HandlerFunc(caregiverHandler.GetCaregiver),
			),
		),
	).Methods("GET")

	r.Handle("/caregivers/{id}",
		auth.Middleware(verifier)(
			auth.RequirePermission("caregiver:update", perms)(
				http.HandlerFunc(caregiverHandler.UpdateCaregiver),
			),
		),
	).Methods("PUT")

	r.Handle("/caregivers/{id}",
		auth.Middleware(verifier)(
			auth.RequirePermission("caregiver:delete", perms)(
				http.HandlerFunc(caregiverHandler.DeleteCaregiver),
			),
		),
	).Methods("DELETE")

	return r
}
