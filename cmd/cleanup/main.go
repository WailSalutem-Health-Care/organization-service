package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/WailSalutem-Health-Care/organization-service/internal/db"
	"github.com/WailSalutem-Health-Care/organization-service/internal/organization"
)

func main() {
	log.Println("Organization Cleanup Job - Starting")
	log.Println("Retention Policy: 3 years")

	// Connect to database
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		os.Exit(1)
	}
	defer database.Close()

	// Create cleanup service
	cleanupService := organization.NewCleanupService(database)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Check how many organizations are eligible for cleanup
	count, err := cleanupService.GetExpiredOrganizationsCount(ctx)
	if err != nil {
		log.Fatalf("Failed to get expired organizations count: %v", err)
		os.Exit(1)
	}

	log.Printf("Found %d organizations eligible for permanent deletion", count)

	if count == 0 {
		log.Println("No cleanup needed. Exiting.")
		os.Exit(0)
	}

	// Perform cleanup
	deletedCount, err := cleanupService.CleanupExpiredOrganizations(ctx)
	if err != nil {
		log.Fatalf("Cleanup failed: %v", err)
		os.Exit(1)
	}

	log.Printf("✓ Cleanup completed successfully: %d organizations permanently deleted", deletedCount)
	log.Println("Cleanup Job - Finished")
}
