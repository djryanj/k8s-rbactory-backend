// backend/internal/handlers/integration_test.go
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"log/slog"
	"os"

	"github.com/djryanj/k8s-rbactory-backend/internal/k8s"
	"github.com/djryanj/k8s-rbactory-backend/internal/middleware"
	"github.com/djryanj/k8s-rbactory-backend/internal/models"
	"github.com/djryanj/k8s-rbactory-backend/internal/testutil"
	"github.com/gorilla/mux"
)

// setupTestServer creates a test server with all middleware
func setupTestServer(t *testing.T, mockClient *k8s.MockClient) *httptest.Server {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	rbacHandler := NewRBACHandler(mockClient, logger)
	clusterHandler := NewClusterHandler(mockClient, logger)

	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1").Subrouter()

	// Register routes
	api.HandleFunc("/healthz", HealthHandler).Methods(http.MethodGet)
	api.HandleFunc("/cluster/info", clusterHandler.GetClusterInfo).Methods(http.MethodGet)
	api.HandleFunc("/namespaces", clusterHandler.ListNamespaces).Methods(http.MethodGet)
	api.HandleFunc("/counts", rbacHandler.GetCounts).Methods(http.MethodGet)
	api.HandleFunc("/roles", rbacHandler.ListRoles).Methods(http.MethodGet)
	api.HandleFunc("/roles/{namespace}/{name}", rbacHandler.GetRole).Methods(http.MethodGet)
	api.HandleFunc("/clusterroles", rbacHandler.ListClusterRoles).Methods(http.MethodGet)
	api.HandleFunc("/clusterroles/{name}", rbacHandler.GetClusterRole).Methods(http.MethodGet)

	// Apply middleware
	handler := middleware.RequestID(r)
	handler = middleware.Logging(logger)(handler)

	return httptest.NewServer(handler)
}

func TestIntegration_FullWorkflow(t *testing.T) {
	// Setup mock data
	mockClient := &k8s.MockClient{
		ListRolesFunc: func(ctx context.Context, namespace string) ([]models.RBACResource, error) {
			return []models.RBACResource{
				testutil.TestRole("default", "role1"),
				testutil.TestRole("default", "role2"),
			}, nil
		},
		GetRoleFunc: func(ctx context.Context, namespace, name string) (*models.RBACResource, error) {
			role := testutil.TestRole(namespace, name)
			return &role, nil
		},
		ListClusterRolesFunc: func(ctx context.Context) ([]models.RBACResource, error) {
			return []models.RBACResource{
				testutil.TestClusterRole("cluster-role1"),
			}, nil
		},
		ListNamespacesFunc: func(ctx context.Context) ([]string, error) {
			return []string{"default", "kube-system"}, nil
		},
		GetClusterVersionFunc: func(ctx context.Context) (string, error) {
			return "v1.28.0", nil
		},
		ListRoleBindingsFunc: func(ctx context.Context, namespace string) ([]models.RBACResource, error) {
			return []models.RBACResource{
				testutil.TestRoleBinding("default", "binding1", "role1"),
			}, nil
		},
		ListClusterRoleBindingsFunc: func(ctx context.Context) ([]models.RBACResource, error) {
			return []models.RBACResource{
				testutil.TestClusterRoleBinding("cluster-binding1", "cluster-role1"),
			}, nil
		},
	}

	server := setupTestServer(t, mockClient)
	defer server.Close()

	client := server.Client()

	t.Run("health check", func(t *testing.T) {
		resp, err := client.Get(server.URL + "/api/v1/healthz")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var health HealthResponse
		if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if health.Status != "healthy" {
			t.Errorf("expected status 'healthy', got %s", health.Status)
		}
	})

	t.Run("list roles", func(t *testing.T) {
		resp, err := client.Get(server.URL + "/api/v1/roles?namespace=default")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var roleList models.RBACList
		if err := json.NewDecoder(resp.Body).Decode(&roleList); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(roleList.Items) != 2 {
			t.Errorf("expected 2 roles, got %d", len(roleList.Items))
		}

		// Check request ID in response header
		if reqID := resp.Header.Get("X-Request-ID"); reqID == "" {
			t.Error("expected X-Request-ID header")
		}
	})

	t.Run("get specific role", func(t *testing.T) {
		resp, err := client.Get(server.URL + "/api/v1/roles/default/role1")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var role models.RBACResource
		if err := json.NewDecoder(resp.Body).Decode(&role); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if role.Name != "role1" {
			t.Errorf("expected role name 'role1', got %s", role.Name)
		}
	})

	t.Run("get counts", func(t *testing.T) {
		resp, err := client.Get(server.URL + "/api/v1/counts")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var counts models.ResourceCounts
		if err := json.NewDecoder(resp.Body).Decode(&counts); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if counts.Roles != 2 {
			t.Errorf("expected 2 roles, got %d", counts.Roles)
		}
		if counts.ClusterRoles != 1 {
			t.Errorf("expected 1 cluster role, got %d", counts.ClusterRoles)
		}
	})

	t.Run("cluster info", func(t *testing.T) {
		resp, err := client.Get(server.URL + "/api/v1/cluster/info")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var info models.ClusterInfo
		if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if info.Version != "v1.28.0" {
			t.Errorf("expected version 'v1.28.0', got %s", info.Version)
		}

		if len(info.Namespaces) != 2 {
			t.Errorf("expected 2 namespaces, got %d", len(info.Namespaces))
		}
	})
}

func TestIntegration_ErrorHandling(t *testing.T) {
	mockClient := &k8s.MockClient{
		ListRolesFunc: func(ctx context.Context, namespace string) ([]models.RBACResource, error) {
			return nil, context.DeadlineExceeded
		},
	}

	server := setupTestServer(t, mockClient)
	defer server.Close()

	client := server.Client()

	t.Run("handles timeout", func(t *testing.T) {
		resp, err := client.Get(server.URL + "/api/v1/roles")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", resp.StatusCode)
		}
	})
}
