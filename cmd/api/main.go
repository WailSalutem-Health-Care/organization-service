package main

import (
	"log"
	"net/http"
	"time"

	"github.com/WailSalutem-Health-Care/organisation-service/internal/auth"
	"github.com/WailSalutem-Health-Care/organisation-service/internal/db"
	httpRouter "github.com/WailSalutem-Health-Care/organisation-service/internal/http"
)

func main() {
	log.Println("organisation-service starting on :8080")

	// Connect to database
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

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

	// Setup router with all routes
	router := httpRouter.SetupRouter(database, ver, perms)

	log.Println("auth configured, jwks loaded, database connected, listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
