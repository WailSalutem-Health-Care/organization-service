package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/XSAM/otelsql"
	_ "github.com/lib/pq"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// Connect creates a connection to PostgreSQL with OpenTelemetry instrumentation
func Connect() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	if host == "" || user == "" || password == "" || dbname == "" {
		return nil, fmt.Errorf("missing required database environment variables")
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=CET",
		host, port, user, password, dbname,
	)

	// Open database connection with OpenTelemetry instrumentation
	db, err := otelsql.Open("postgres", connStr,
		otelsql.WithAttributes(
			semconv.DBSystemPostgreSQL,
			semconv.DBName(dbname),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Register database stats for metrics
	_, err = otelsql.RegisterDBStatsMetrics(db,
		otelsql.WithAttributes(
			semconv.DBSystemPostgreSQL,
			semconv.DBName(dbname),
		),
	)
	if err != nil {
		log.Printf("Warning: failed to register database stats metrics: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set timezone to CET (GMT+1)
	_, err = db.Exec("SET TIME ZONE 'CET'")
	if err != nil {
		return nil, fmt.Errorf("failed to set timezone: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	log.Println("✓ Connected to PostgreSQL database (CET timezone, OpenTelemetry enabled)")
	return db, nil
}
