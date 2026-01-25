// backend/internal/middleware/cors.go
package middleware

import (
	"os"
	"strings"

	"github.com/rs/cors"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns the default CORS configuration for development
// Uses wildcard for headers to simplify local development.
// WARNING: This should be overridden in production with explicit headers.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"http://localhost:5173"}, // Vite default for dev
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		// Wildcard allows all headers - convenient for development
		// Override in production via ALLOWED_HEADERS environment variable
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300, // 5 minutes
	}
}

// ProductionCORSConfig returns a secure CORS configuration template for production
// This uses explicit header allowlisting for better security.
// Customize the AllowedOrigins and AllowedHeaders for your deployment.
func ProductionCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{}, // Must be explicitly set via ALLOWED_ORIGINS
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		// Explicit headers for production security
		// Only allow headers that your application actually needs
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-Request-ID",
			"X-CSRF-Token",
		},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}
}

// CORS creates CORS middleware with configuration from environment variables
// This is the convenience function for production use.
//
// Environment variables:
//   - ALLOWED_ORIGINS: Comma-separated list of allowed origins (required in production)
//   - ALLOWED_HEADERS: Comma-separated list of allowed headers (optional, uses wildcard if not set)
//
// Example:
//
//	ALLOWED_ORIGINS=https://myapp.com,https://www.myapp.com
//	ALLOWED_HEADERS=Accept,Authorization,Content-Type,X-Request-ID
func CORS() *cors.Cors {
	config := DefaultCORSConfig()

	// Override allowed origins from environment if set
	if origins := getAllowedOriginsFromEnv(); len(origins) > 0 {
		config.AllowedOrigins = origins
	}

	// Override allowed headers from environment if set
	// This allows production to use explicit headers instead of wildcard
	if headers := getAllowedHeadersFromEnv(); len(headers) > 0 {
		config.AllowedHeaders = headers
	}

	return CORSWithConfig(config)
}

// CORSWithConfig creates CORS middleware with custom configuration
// This is useful for testing and when you need explicit control over CORS settings.
//
// Note: The rs/cors library requires either:
//   - Wildcard "*" to allow all headers (convenient but less secure)
//   - Explicit list of headers (more secure but requires maintenance)
//
// The library does NOT support case-insensitive matching or automatic header
// normalization, which is why we recommend using "*" for development.
func CORSWithConfig(config CORSConfig) *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins:     config.AllowedOrigins,
		AllowedMethods:     config.AllowedMethods,
		AllowedHeaders:     config.AllowedHeaders,
		ExposedHeaders:     config.ExposedHeaders,
		AllowCredentials:   config.AllowCredentials,
		MaxAge:             config.MaxAge,
		OptionsPassthrough: false,
		Debug:              false,
	})
}

// getAllowedOriginsFromEnv reads allowed origins from environment variable
// Format: comma-separated list of origins
// Example: "https://app.example.com,https://www.example.com"
func getAllowedOriginsFromEnv() []string {
	if origins := os.Getenv("ALLOWED_ORIGINS"); origins != "" {
		parts := strings.Split(origins, ",")
		result := make([]string, 0, len(parts))
		for _, origin := range parts {
			if trimmed := strings.TrimSpace(origin); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return nil
}

// getAllowedHeadersFromEnv reads allowed headers from environment variable
// Format: comma-separated list of header names
// Example: "Accept,Authorization,Content-Type,X-Request-ID"
//
// If not set, the default configuration uses "*" (wildcard) for development convenience.
// In production, set this to an explicit list of headers for better security.
func getAllowedHeadersFromEnv() []string {
	if headers := os.Getenv("ALLOWED_HEADERS"); headers != "" {
		parts := strings.Split(headers, ",")
		result := make([]string, 0, len(parts))
		for _, header := range parts {
			if trimmed := strings.TrimSpace(header); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return nil
}
