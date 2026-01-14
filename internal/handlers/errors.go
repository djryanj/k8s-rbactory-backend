// internal/handlers/errors.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/djryanj/k8s-rbactory-backend/internal/logging"
)

// WriteJSONError writes a standardized JSON error response
func WriteJSONError(w http.ResponseWriter, statusCode int, message string, details ...interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:      http.StatusText(statusCode),
		Message:    message,
		StatusCode: statusCode,
	}

	// Add details if provided
	if len(details) > 0 {
		response.Details = make(map[string]interface{})
		for i := 0; i < len(details); i += 2 {
			if i+1 < len(details) {
				if key, ok := details[i].(string); ok {
					response.Details[key] = details[i+1]
				}
			}
		}
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Fallback if JSON encoding fails
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// WriteValidationError writes a validation error response
func WriteValidationError(w http.ResponseWriter, field, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	response := ErrorResponse{
		Error:      http.StatusText(http.StatusBadRequest),
		Message:    "Validation failed",
		StatusCode: http.StatusBadRequest,
		Errors: []ValidationError{
			{
				Field:   field,
				Message: message,
			},
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// WriteKubernetesError writes a Kubernetes API error response
// This handles the specific case of K8s RBAC errors
func WriteKubernetesError(w http.ResponseWriter, err error) {
	errMsg := err.Error()

	// Determine status code based on error message
	statusCode := http.StatusInternalServerError

	// Check for specific Kubernetes error patterns
	lowerErr := strings.ToLower(errMsg)
	if containsAny(lowerErr, "forbidden", "is forbidden", "cannot list", "cannot get", "cannot watch") {
		statusCode = http.StatusForbidden
	} else if containsAny(lowerErr, "unauthorized", "authentication", "unauthenticated") {
		statusCode = http.StatusUnauthorized
	} else if containsAny(lowerErr, "not found") {
		statusCode = http.StatusNotFound
	} else if containsAny(lowerErr, "timeout", "deadline exceeded") {
		statusCode = http.StatusGatewayTimeout
	} else if containsAny(lowerErr, "connection refused", "connection reset") {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:      http.StatusText(statusCode),
		Message:    errMsg,
		StatusCode: statusCode,
	}

	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		// If we can't encode JSON, fall back to plain text
		http.Error(w, errMsg, statusCode)
	}
}

// WriteInternalError writes an internal server error with logging
func WriteInternalError(w http.ResponseWriter, r *http.Request, message string, err error) {
	logging.LogError(r.Context(), message, err)
	WriteJSONError(w, http.StatusInternalServerError, message)
}

// containsAny checks if a string contains any of the given substrings
func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}
