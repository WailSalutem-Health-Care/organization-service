package http

import (
	"net/http"
	"os"
	"strings"
)

// CORSMiddleware adds CORS headers to allow frontend access
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get allowed origins from environment variable, default to localhost for development
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
		if allowedOrigins == "" {
			// Default allowed origins for development
			allowedOrigins = "http://localhost:3000,https://wailsalutem-web-ui.netlify.app"
		}

		origin := r.Header.Get("Origin")

		// Check if the origin is allowed
		if origin != "" {
			origins := strings.Split(allowedOrigins, ",")
			for _, allowedOrigin := range origins {
				if strings.TrimSpace(allowedOrigin) == origin || allowedOrigin == "*" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Organization-ID")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
