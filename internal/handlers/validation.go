// internal/handlers/validation.go
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
)

var (
	// Kubernetes resource name validation (RFC 1123 subdomain)
	k8sNameRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error      string                 `json:"error"`
	Message    string                 `json:"message,omitempty"`
	StatusCode int                    `json:"statusCode"`
	Errors     []ValidationError      `json:"errors,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

// ValidatePaginationParams validates and returns pagination parameters
func ValidatePaginationParams(r *http.Request) (limit, offset int, err error) {
	limit = 20 // default
	offset = 0

	if l := r.URL.Query().Get("limit"); l != "" {
		parsed, parseErr := strconv.Atoi(l)
		if parseErr != nil {
			return 0, 0, fmt.Errorf("invalid limit: must be an integer")
		}
		if parsed < 1 {
			return 0, 0, fmt.Errorf("invalid limit: must be greater than 0")
		}
		if parsed > 200 {
			return 0, 0, fmt.Errorf("invalid limit: maximum is 200")
		}
		limit = parsed
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		parsed, parseErr := strconv.Atoi(o)
		if parseErr != nil {
			return 0, 0, fmt.Errorf("invalid offset: must be an integer")
		}
		if parsed < 0 {
			return 0, 0, fmt.Errorf("invalid offset: must be non-negative")
		}
		offset = parsed
	}

	return limit, offset, nil
}

// ValidateK8sName validates a Kubernetes resource name
func ValidateK8sName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if len(name) > 253 {
		return fmt.Errorf("name too long: maximum 253 characters")
	}
	if !k8sNameRegex.MatchString(name) {
		return fmt.Errorf("invalid name format: must be a valid Kubernetes resource name")
	}
	return nil
}

// ValidateNamespace validates a namespace parameter
func ValidateNamespace(namespace string) error {
	if namespace == "" {
		return nil // Empty namespace is valid (means all namespaces)
	}
	return ValidateK8sName(namespace)
}

// WriteErrorResponse writes a JSON error response (legacy - use WriteJSONError instead)
func WriteErrorResponse(w http.ResponseWriter, statusCode int, message string, validationErrors ...ValidationError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:      http.StatusText(statusCode),
		Message:    message,
		StatusCode: statusCode,
		Errors:     validationErrors,
	}

	json.NewEncoder(w).Encode(response)
}
