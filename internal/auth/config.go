package auth

import (
	"fmt"
	"log"
	"os"
)

// Config holds auth configuration
type Config struct {
	Issuer  string
	JWKSURL string
}

func LoadConfig() Config {
	// Read Keycloak base URL and realm from environment
	keycloakBaseURL := os.Getenv("KEYCLOAK_BASE_URL")
	if keycloakBaseURL == "" {
		log.Fatal("KEYCLOAK_BASE_URL environment variable is required")
	}

	realm := os.Getenv("KEYCLOAK_REALM")
	if realm == "" {
		log.Fatal("KEYCLOAK_REALM environment variable is required")
	}

	// Build issuer and JWKS URL from base URL and realm
	issuer := fmt.Sprintf("%s/realms/%s", keycloakBaseURL, realm)
	jwks := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", keycloakBaseURL, realm)

	log.Printf("Keycloak configured - Issuer: %s", issuer)

	return Config{
		Issuer:  issuer,
		JWKSURL: jwks,
	}
}
