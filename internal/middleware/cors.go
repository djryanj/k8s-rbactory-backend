// backend/internal/middleware/cors.go
package middleware

import (
	"os"
	"strings"

	"github.com/rs/cors"
)

func CORS() *cors.Cors {
	// Get allowed origins from environment variable
	allowedOrigins := []string{"http://localhost:5173"} // vite default for dev

	if origins := os.Getenv("ALLOWED_ORIGINS"); origins != "" {
		allowedOrigins = strings.Split(origins, ",")
	}

	return cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}
