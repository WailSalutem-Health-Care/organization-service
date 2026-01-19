package http

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/WailSalutem-Health-Care/organization-service/internal/auth"
	"github.com/WailSalutem-Health-Care/organization-service/internal/messaging"
	"github.com/WailSalutem-Health-Care/organization-service/internal/organization"
	"github.com/WailSalutem-Health-Care/organization-service/internal/patient"
	"github.com/WailSalutem-Health-Care/organization-service/internal/telemetry"
	"github.com/WailSalutem-Health-Care/organization-service/internal/users"
	"github.com/gorilla/mux"
)

// SetupRouter initializes all routes for the application
func SetupRouter(db *sql.DB, verifier *auth.Verifier, perms map[string][]string, publisher messaging.PublisherInterface, metrics *telemetry.Metrics) *mux.Router {
	// Initialize Keycloak admin client
	keycloakAdmin, err := auth.NewKeycloakAdminClient()
	if err != nil {
		log.Fatalf("failed to initialize Keycloak admin client: %v", err)
	}

	return SetupRouterWithKeycloak(db, verifier, perms, publisher, keycloakAdmin, metrics)
}

// SetupRouterWithKeycloak initializes all routes with a provided Keycloak client
// This is useful for testing where you can pass a mock Keycloak client
func SetupRouterWithKeycloak(db *sql.DB, verifier *auth.Verifier, perms map[string][]string, publisher messaging.PublisherInterface, keycloakAdmin interface{}, metrics *telemetry.Metrics) *mux.Router {
	// Initialize organization components
	orgRepo := organization.NewRepository(db, publisher)
	orgService := organization.NewService(orgRepo)
	orgHandler := organization.NewHandler(orgService)

	// Cast keycloakAdmin to the appropriate interface types
	// For users and patients, they use their own KeycloakAdminInterface
	var userKeycloak users.KeycloakAdminInterface
	var patientKeycloak patient.KeycloakAdminInterface

	if keycloakAdmin != nil {
		// Try to cast to users.KeycloakAdminInterface
		if kc, ok := keycloakAdmin.(users.KeycloakAdminInterface); ok {
			userKeycloak = kc
		}
		// Try to cast to patient.KeycloakAdminInterface
		if kc, ok := keycloakAdmin.(patient.KeycloakAdminInterface); ok {
			patientKeycloak = kc
		}
	}

	// Initialize patient components
	patientRepo := patient.NewRepository(db, publisher)
	patientService := patient.NewService(patientRepo, patientKeycloak)
	patientSchemaLookup := patient.NewDBSchemaLookup(db)
	patientHandler := patient.NewHandler(patientService, patientSchemaLookup)

	// Initialize user components
	userRepo := users.NewRepository(db, publisher)
	userService := users.NewService(userRepo, userKeycloak)
	userHandler := users.NewHandler(userService)

	r := mux.NewRouter()

	// Public health endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"organization-service"}`))
	}).Methods("GET")

	r.Handle("/organizations",
		auth.MiddlewareWithMetrics(verifier, metrics)(
			auth.RequirePermissionWithMetrics("organization:create", perms, metrics)(
				http.HandlerFunc(orgHandler.CreateOrganization),
			),
		),
	).Methods("POST")

	r.Handle("/organizations",
		auth.MiddlewareWithMetrics(verifier, metrics)(
			auth.RequirePermissionWithMetrics("organization:view", perms, metrics)(
				http.HandlerFunc(orgHandler.ListOrganizations),
			),
		),
	).Methods("GET")

	r.Handle("/organizations/{id}",
		auth.MiddlewareWithMetrics(verifier, metrics)(
			auth.RequirePermissionWithMetrics("organization:view", perms, metrics)(
				http.HandlerFunc(orgHandler.GetOrganization),
			),
		),
	).Methods("GET")

	r.Handle("/organizations/{id}",
		auth.MiddlewareWithMetrics(verifier, metrics)(
			auth.RequirePermissionWithMetrics("organization:update", perms, metrics)(
				http.HandlerFunc(orgHandler.UpdateOrganization),
			),
		),
	).Methods("PUT", "PATCH")

	r.Handle("/organizations/{id}",
		auth.MiddlewareWithMetrics(verifier, metrics)(
			auth.RequirePermissionWithMetrics("organization:delete", perms, metrics)(
				http.HandlerFunc(orgHandler.DeleteOrganization),
			),
		),
	).Methods("DELETE")

	r.Handle("/organization/patients",
		auth.MiddlewareWithMetrics(verifier, metrics)(
			auth.RequirePermissionWithMetrics("patient:create", perms, metrics)(
				http.HandlerFunc(patientHandler.CreatePatient),
			),
		),
	).Methods("POST")

	r.Handle("/organization/patients",
		auth.MiddlewareWithMetrics(verifier, metrics)(
			auth.RequirePermissionWithMetrics("patient:view", perms, metrics)(
				http.HandlerFunc(patientHandler.ListPatients),
			),
		),
	).Methods("GET")

	r.Handle("/organization/patients/active",
		auth.MiddlewareWithMetrics(verifier, metrics)(
			auth.RequirePermissionWithMetrics("patient:view", perms, metrics)(
				http.HandlerFunc(patientHandler.ListActivePatients),
			),
		),
	).Methods("GET")

	r.Handle("/organization/patients/{id}",
		auth.MiddlewareWithMetrics(verifier, metrics)(
			auth.RequirePermissionWithMetrics("patient:view", perms, metrics)(
				http.HandlerFunc(patientHandler.GetPatient),
			),
		),
	).Methods("GET")

	r.Handle("/organization/patients/{id}",
		auth.MiddlewareWithMetrics(verifier, metrics)(
			auth.RequirePermissionWithMetrics("patient:update", perms, metrics)(
				http.HandlerFunc(patientHandler.UpdatePatient),
			),
		),
	).Methods("PUT", "PATCH")

	r.Handle("/organization/patients/{id}",
		auth.MiddlewareWithMetrics(verifier, metrics)(
			auth.RequirePermissionWithMetrics("patient:delete", perms, metrics)(
				http.HandlerFunc(patientHandler.DeletePatient),
			),
		),
	).Methods("DELETE")

	r.Handle("/organization/users",
		auth.MiddlewareWithMetrics(verifier, metrics)(
			auth.RequirePermissionWithMetrics("user:create", perms, metrics)(
				http.HandlerFunc(userHandler.CreateUser),
			),
		),
	).Methods("POST")

	r.Handle("/organization/users",
		auth.MiddlewareWithMetrics(verifier, metrics)(
			auth.RequirePermissionWithMetrics("user:view", perms, metrics)(
				http.HandlerFunc(userHandler.ListUsers),
			),
		),
	).Methods("GET")

	r.Handle("/organization/users/caregivers/active",
		auth.MiddlewareWithMetrics(verifier, metrics)(
			auth.RequirePermissionWithMetrics("user:view", perms, metrics)(
				http.HandlerFunc(userHandler.ListActiveCaregivers),
			),
		),
	).Methods("GET")

	r.Handle("/organization/users/municipality/active",
		auth.MiddlewareWithMetrics(verifier, metrics)(
			auth.RequirePermissionWithMetrics("user:view", perms, metrics)(
				http.HandlerFunc(userHandler.ListActiveMunicipality),
			),
		),
	).Methods("GET")

	r.Handle("/organization/users/insurers/active",
		auth.MiddlewareWithMetrics(verifier, metrics)(
			auth.RequirePermissionWithMetrics("user:view", perms, metrics)(
				http.HandlerFunc(userHandler.ListActiveInsurers),
			),
		),
	).Methods("GET")

	r.Handle("/organization/users/org-admins/active",
		auth.MiddlewareWithMetrics(verifier, metrics)(
			auth.RequirePermissionWithMetrics("user:view", perms, metrics)(
				http.HandlerFunc(userHandler.ListActiveOrgAdmins),
			),
		),
	).Methods("GET")

	r.Handle("/organization/users/me",
		auth.MiddlewareWithMetrics(verifier, metrics)(
			http.HandlerFunc(userHandler.GetMyProfile),
		),
	).Methods("GET")

	r.Handle("/organization/users/me",
		auth.MiddlewareWithMetrics(verifier, metrics)(
			http.HandlerFunc(userHandler.UpdateMyProfile),
		),
	).Methods("PATCH")

	r.Handle("/organization/patients/me",
		auth.MiddlewareWithMetrics(verifier, metrics)(
			auth.RequirePermissionWithMetrics("patient:view", perms, metrics)(
				http.HandlerFunc(patientHandler.GetMyPatient),
			),
		),
	).Methods("GET")

	r.Handle("/organization/users/{id}",
		auth.MiddlewareWithMetrics(verifier, metrics)(
			auth.RequirePermissionWithMetrics("user:view", perms, metrics)(
				http.HandlerFunc(userHandler.GetUser),
			),
		),
	).Methods("GET")

	r.Handle("/organization/users/{id}",
		auth.MiddlewareWithMetrics(verifier, metrics)(
			auth.RequirePermissionWithMetrics("user:update", perms, metrics)(
				http.HandlerFunc(userHandler.UpdateUser),
			),
		),
	).Methods("PATCH")

	r.Handle("/organization/users/{id}/reset-password",
		auth.MiddlewareWithMetrics(verifier, metrics)(
			auth.RequirePermissionWithMetrics("user:update", perms, metrics)(
				http.HandlerFunc(userHandler.ResetPassword),
			),
		),
	).Methods("POST")

	r.Handle("/organization/users/{id}",
		auth.MiddlewareWithMetrics(verifier, metrics)(
			auth.RequirePermissionWithMetrics("user:delete", perms, metrics)(
				http.HandlerFunc(userHandler.DeleteUser),
			),
		),
	).Methods("DELETE")

	return r
}
