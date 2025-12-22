package auth

import "os"

// Config holds auth configuration
type Config struct {
	Issuer   string
	JWKSURL  string
	Audience string
}

var (
	// Defaults are the Keycloak realm values from the brief.
	DefaultIssuer  = "https://keycloak-wailsalutem-suite.apps.inholland-minor.openshift.eu/realms/wailsalutem"
	DefaultJWKSURL = "https://keycloak-wailsalutem-suite.apps.inholland-minor.openshift.eu/realms/wailsalutem/protocol/openid-connect/certs"
)

// LoadConfig reads config from env with sensible defaults.
// You can override with AUTH_ISSUER and AUTH_JWKS_URL and AUTH_AUD.
func LoadConfig() Config {
	issuer := os.Getenv("AUTH_ISSUER")
	if issuer == "" {
		issuer = DefaultIssuer
	}
	jwks := os.Getenv("AUTH_JWKS_URL")
	if jwks == "" {
		jwks = DefaultJWKSURL
	}
	aud := os.Getenv("AUTH_AUD") // optional
	return Config{
		Issuer:   issuer,
		JWKSURL:  jwks,
		Audience: aud,
	}
}
