// backend/test/e2e_test.go
package test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"log/slog"
	"os"

	"github.com/djryanj/k8s-rbactory-backend/internal/handlers"
	"github.com/djryanj/k8s-rbactory-backend/internal/k8s"
	"github.com/djryanj/k8s-rbactory-backend/internal/middleware"
	"github.com/djryanj/k8s-rbactory-backend/internal/models"
	"github.com/djryanj/k8s-rbactory-backend/internal/testutil"
	"github.com/gorilla/mux"
)

// TestE2E_CompleteWorkflow tests a complete user workflow
func TestE2E_CompleteWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping end-to-end test in short mode")
	}

	// Setup
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mockClient := setupMockClient()
	server := setupTestServer(t, mockClient, logger)
	defer server.Close()

	client := server.Client()

	t.Run("workflow: discover cluster -> list resources -> get details", func(t *testing.T) {
		// Step 1: Check health
		resp, err := client.Get(server.URL + "/api/v1/healthz")
		if err != nil {
			t.Fatalf("health check failed: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		// Step 2: Get cluster info
		resp, err = client.Get(server.URL + "/api/v1/cluster/info")
		if err != nil {
			t.Fatalf("cluster info request failed: %v", err)
		}
		defer resp.Body.Close()

		var clusterInfo models.ClusterInfo
		if err := json.NewDecoder(resp.Body).Decode(&clusterInfo); err != nil {
			t.Fatalf("failed to decode cluster info: %v", err)
		}

		if clusterInfo.Version == "" {
			t.Error("expected cluster version to be set")
		}

		// Step 3: Get resource counts
		resp, err = client.Get(server.URL + "/api/v1/counts")
		if err != nil {
			t.Fatalf("counts request failed: %v", err)
		}
		defer resp.Body.Close()

		var counts models.ResourceCounts
		if err := json.NewDecoder(resp.Body).Decode(&counts); err != nil {
			t.Fatalf("failed to decode counts: %v", err)
		}

		if counts.Roles == 0 {
			t.Error("expected some roles in counts")
		}

		// Step 4: List roles
		resp, err = client.Get(server.URL + "/api/v1/roles?namespace=default&limit=10")
		if err != nil {
			t.Fatalf("list roles request failed: %v", err)
		}
		defer resp.Body.Close()

		var roleList models.RBACList
		if err := json.NewDecoder(resp.Body).Decode(&roleList); err != nil {
			t.Fatalf("failed to decode role list: %v", err)
		}

		if len(roleList.Items) == 0 {
			t.Error("expected some roles in list")
		}

		// Step 5: Get specific role
		roleName := roleList.Items[0].Name
		resp, err = client.Get(server.URL + "/api/v1/roles/default/" + roleName)
		if err != nil {
			t.Fatalf("get role request failed: %v", err)
		}
		defer resp.Body.Close()

		var role models.RBACResource
		if err := json.NewDecoder(resp.Body).Decode(&role); err != nil {
			t.Fatalf("failed to decode role: %v", err)
		}

		if role.Name != roleName {
			t.Errorf("expected role name %s, got %s", roleName, role.Name)
		}

		// Verify request ID is present in all responses
		if reqID := resp.Header.Get("X-Request-ID"); reqID == "" {
			t.Error("expected X-Request-ID header in response")
		}
	})
}

func setupMockClient() *k8s.MockClient {
	return &k8s.MockClient{
		GetClusterVersionFunc: func(ctx context.Context) (string, error) {
			return "v1.28.0", nil
		},
		ListNamespacesFunc: func(ctx context.Context) ([]string, error) {
			return []string{"default", "kube-system"}, nil
		},
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
}

func setupTestServer(t *testing.T, mockClient *k8s.MockClient, logger *slog.Logger) *httptest.Server {
	rbacHandler := handlers.NewRBACHandler(mockClient, logger)
	clusterHandler := handlers.NewClusterHandler(mockClient, logger)

	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1").Subrouter()

	// Register routes
	api.HandleFunc("/healthz", handlers.HealthHandler).Methods(http.MethodGet)
	api.HandleFunc("/cluster/info", clusterHandler.GetClusterInfo).Methods(http.MethodGet)
	api.HandleFunc("/counts", rbacHandler.GetCounts).Methods(http.MethodGet)
	api.HandleFunc("/roles", rbacHandler.ListRoles).Methods(http.MethodGet)
	api.HandleFunc("/roles/{namespace}/{name}", rbacHandler.GetRole).Methods(http.MethodGet)

	// Apply middleware
	handler := middleware.RequestID(r)
	handler = middleware.Logging(logger)(handler)
	handler = middleware.Timeout(logger, 30*time.Second)(handler)

	return httptest.NewServer(handler)
}
