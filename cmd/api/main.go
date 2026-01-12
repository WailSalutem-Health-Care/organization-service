package main

import (
	"log"
	"net/http"
	"time"

	"github.com/WailSalutem-Health-Care/organization-service/internal/auth"
	"github.com/WailSalutem-Health-Care/organization-service/internal/db"
	httpRouter "github.com/WailSalutem-Health-Care/organization-service/internal/http"
	"github.com/WailSalutem-Health-Care/organization-service/internal/messaging"
)

func main() {
	// Set timezone to CET (Central European Time - GMT+1, auto-adjusts to CEST in summer)
	loc, err := time.LoadLocation("CET")
	if err != nil {
		log.Printf("Warning: failed to load CET timezone: %v, using UTC", err)
	} else {
		time.Local = loc
		log.Println("✓ Timezone set to CET (GMT+1)")
	}

	log.Println("organization-service starting on :8080")
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

	// Initialize RabbitMQ publisher (optional dependency)
	publisher, err := messaging.NewPublisher()
	if err != nil {
		log.Printf("Warning: failed to connect to RabbitMQ: %v", err)
		log.Println("Service will continue without event publishing (RabbitMQ optional)")
		publisher = nil
	} else {
		defer publisher.Close()
		log.Println("✓ RabbitMQ publisher initialized")
	}

	// Setup router with all routes
	router := httpRouter.SetupRouter(database, ver, perms, publisher)

	log.Println("auth configured, jwks loaded, database connected, listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
