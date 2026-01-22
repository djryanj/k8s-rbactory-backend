// internal/handlers/resources_integration_test.go
//go:build integration
// +build integration

package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/djryanj/k8s-rbactory-backend/internal/k8s"
	"github.com/djryanj/k8s-rbactory-backend/internal/models"
	"github.com/gorilla/mux"
)

// TestResourceAccessIntegration tests the full flow with a real K8s client
// Run with: go test -tags=integration ./internal/handlers
func TestResourceAccessIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Create real K8s client
	client, err := k8s.NewClient()
	if err != nil {
		t.Fatalf("failed to create k8s client: %v", err)
	}

	handler := NewResourceHandler(client, logger)

	// Test listing resource types
	t.Run("list resource types", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/resources/types", nil)
		w := httptest.NewRecorder()

		handler.ListResourceTypes(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response models.ResourceTypeList
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(response.ResourceTypes) == 0 {
			t.Error("expected resource types")
		}

		t.Logf("Found %d resource types", len(response.ResourceTypes))
	})

	// Test listing secrets
	t.Run("list secrets", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/resources?resourceType=secrets&namespace=kube-system", nil)
		w := httptest.NewRecorder()

		handler.ListKubernetesResources(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var response models.KubernetesResourceList
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		t.Logf("Found %d secrets in kube-system", len(response.Items))

		// Verify structure
		if len(response.Items) > 0 {
			secret := response.Items[0]
			if secret.Kind != "Secret" {
				t.Errorf("expected kind 'Secret', got '%s'", secret.Kind)
			}
			if secret.Name == "" {
				t.Error("expected secret name")
			}
		}
	})

	// Test getting access for a specific resource
	t.Run("get resource access", func(t *testing.T) {
		// First, list secrets to get a real secret name
		ctx := context.Background()
		secrets, err := client.ListSecrets(ctx, "kube-system")
		if err != nil {
			t.Fatalf("failed to list secrets: %v", err)
		}

		if len(secrets) == 0 {
			t.Skip("no secrets found in kube-system")
		}

		secretName := secrets[0].Name

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/v1/resources/secrets/kube-system/"+secretName+"/access",
			nil,
		)
		w := httptest.NewRecorder()

		// Use router to handle path variables
		router := setupTestRouter(handler)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var response models.ResourceAccessDetail
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		t.Logf("Found %d access grants for secret %s", len(response.AccessGrants), secretName)
		t.Logf("Summary: %d principals, %d roles, %d bindings",
			response.Summary.TotalPrincipals,
			response.Summary.TotalRoles,
			response.Summary.TotalBindings,
		)

		// Log some details
		for i, grant := range response.AccessGrants {
			if i >= 3 {
				break // Just show first 3
			}
			t.Logf("Grant %d: %s/%s via %s/%s (verbs: %v)",
				i+1,
				grant.Principal.Kind,
				grant.Principal.Name,
				grant.Role.Kind,
				grant.Role.Name,
				grant.Verbs,
			)
		}
	})
}

func setupTestRouter(handler *ResourceHandler) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/resources/types", handler.ListResourceTypes)
	router.HandleFunc("/api/v1/resources", handler.ListKubernetesResources)
	router.HandleFunc("/api/v1/resources/{type}/{namespace}/{name}/access", handler.GetResourceAccess)
	return router
}
