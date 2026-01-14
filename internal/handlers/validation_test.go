// backend/internal/handlers/validation_test.go
package handlers

import (
	"net/http/httptest"
	"testing"
)

func TestValidatePaginationParams(t *testing.T) {
	tests := []struct {
		name        string
		queryParams map[string]string
		wantLimit   int
		wantOffset  int
		wantErr     bool
	}{
		{
			name:        "default values",
			queryParams: map[string]string{},
			wantLimit:   20,
			wantOffset:  0,
			wantErr:     false,
		},
		{
			name: "valid custom values",
			queryParams: map[string]string{
				"limit":  "50",
				"offset": "10",
			},
			wantLimit:  50,
			wantOffset: 10,
			wantErr:    false,
		},
		{
			name: "limit too high",
			queryParams: map[string]string{
				"limit": "500",
			},
			wantErr: true,
		},
		{
			name: "negative limit",
			queryParams: map[string]string{
				"limit": "-10",
			},
			wantErr: true,
		},
		{
			name: "negative offset",
			queryParams: map[string]string{
				"offset": "-5",
			},
			wantErr: true,
		},
		{
			name: "invalid limit format",
			queryParams: map[string]string{
				"limit": "abc",
			},
			wantErr: true,
		},
		{
			name: "zero limit",
			queryParams: map[string]string{
				"limit": "0",
			},
			wantErr: true,
		},
		{
			name: "max valid limit",
			queryParams: map[string]string{
				"limit": "200",
			},
			wantLimit:  200,
			wantOffset: 0,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()

			limit, offset, err := ValidatePaginationParams(req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if limit != tt.wantLimit {
				t.Errorf("limit = %d, want %d", limit, tt.wantLimit)
			}

			if offset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", offset, tt.wantOffset)
			}
		})
	}
}

func TestValidateK8sName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid simple name",
			input:   "my-role",
			wantErr: false,
		},
		{
			name:    "valid with dots",
			input:   "my.role.name",
			wantErr: false,
		},
		{
			name:    "valid with numbers",
			input:   "role-123",
			wantErr: false,
		},
		{
			name:    "empty name",
			input:   "",
			wantErr: true,
		},
		{
			name:    "starts with hyphen",
			input:   "-invalid",
			wantErr: true,
		},
		{
			name:    "ends with hyphen",
			input:   "invalid-",
			wantErr: true,
		},
		{
			name:    "uppercase letters",
			input:   "Invalid",
			wantErr: true,
		},
		{
			name:    "special characters",
			input:   "role@name",
			wantErr: true,
		},
		{
			name:    "too long",
			input:   string(make([]byte, 254)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateK8sName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateK8sName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateNamespace(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		wantErr   bool
	}{
		{
			name:      "empty namespace (all namespaces)",
			namespace: "",
			wantErr:   false,
		},
		{
			name:      "valid namespace",
			namespace: "default",
			wantErr:   false,
		},
		{
			name:      "invalid namespace",
			namespace: "Invalid-Namespace",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNamespace(tt.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNamespace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
