package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/WailSalutem-Health-Care/organization-service/internal/auth"
	"github.com/WailSalutem-Health-Care/organization-service/internal/db"
	httpRouter "github.com/WailSalutem-Health-Care/organization-service/internal/http"
	"github.com/WailSalutem-Health-Care/organization-service/internal/messaging"
	"github.com/WailSalutem-Health-Care/organization-service/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
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

	// Initialize OpenTelemetry (tracer and meter providers)
	ctx := context.Background()
	otelCfg := telemetry.LoadConfig()
	otelProvider, err := telemetry.InitProvider(ctx, otelCfg)
	if err != nil {
		log.Printf("Warning: failed to initialize OpenTelemetry: %v", err)
		log.Println("Service will continue without observability")
	} else {
		// Ensure telemetry is flushed on shutdown
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := otelProvider.Shutdown(shutdownCtx); err != nil {
				log.Printf("Error shutting down OpenTelemetry: %v", err)
			}
		}()
	}

	// Initialize custom metrics
	metrics, err := telemetry.InitMetrics()
	if err != nil {
		log.Printf("Warning: failed to initialize custom metrics: %v", err)
		metrics = nil
	}

	// Connect to database with OpenTelemetry instrumentation
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
	router := httpRouter.SetupRouter(database, ver, perms, publisher, metrics)

	// Wrap router with OpenTelemetry instrumentation
	router.Use(otelmux.Middleware("organization-service"))

	log.Println("auth configured, jwks loaded, database connected, observability enabled, listening on :8080")

	// Create HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      httpRouter.CORSMiddleware(router),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("Server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-stop
	log.Println("Shutting down server gracefully...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
