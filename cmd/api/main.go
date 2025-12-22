package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/WailSalutem-Health-Care/organisation-service/internal/auth"
	"github.com/gorilla/mux"
)

func main() {
	log.Println("organisation-service starting on :8080")

	// Load auth config
	cfg := auth.LoadConfig()

	// Load permissions.yml
	perms, err := auth.LoadPermissions("permissions.yml")
	if err != nil {
		log.Fatalf("failed to load permissions.yml: %v", err)
	}
	log.Printf("loaded permissions for %d roles", len(perms))

	// Initialize JWKS (cached, auto-refreshed every 15 min)
	jwks, err := auth.NewJWKS(cfg.JWKSURL, 15*time.Minute)
	if err != nil {
		log.Fatalf("failed to initialize JWKS: %v", err)
	}
	defer jwks.Close()

	// Create token verifier
	ver := auth.NewVerifier(cfg, jwks)

	// Setup router
	r := mux.NewRouter()

	// Public health endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"organisation-service"}`))
	}).Methods("GET")

	// Protected route example: POST /organizations (requires organization:create)
	createOrgHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Principal is available in context; handlers MUST use org_id from token for multi-tenancy
		pr, _ := auth.FromContext(req.Context())

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"result": "created",
			"org_id": pr.OrgID,
			"user":   pr.UserID,
		})
	})

	// Wrap with auth middleware and permission guard
	r.Handle("/organizations",
		auth.Middleware(ver)(
			auth.RequirePermission("organization:create", perms)(createOrgHandler),
		),
	).Methods("POST")

	// Example: GET /organizations (requires organization:view)
	listOrgHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		pr, _ := auth.FromContext(req.Context())

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"organizations": []string{},
			"org_id":        pr.OrgID,
			"user":          pr.UserID,
		})
	})

	r.Handle("/organizations",
		auth.Middleware(ver)(
			auth.RequirePermission("organization:view", perms)(listOrgHandler),
		),
	).Methods("GET")

	log.Println("auth configured, jwks loaded, listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
