// internal/handlers/resources_test.go
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

func TestListKubernetesResources(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	tests := []struct {
		name           string
		resourceType   string
		namespace      string
		mockResources  []models.KubernetesResource
		mockError      error
		expectedStatus int
		expectedCount  int
	}{
		{
			name:         "list secrets successfully",
			resourceType: "secrets",
			namespace:    "default",
			mockResources: []models.KubernetesResource{
				{
					Kind:      "Secret",
					Name:      "test-secret",
					Namespace: "default",
				},
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "missing resource type",
			resourceType:   "",
			namespace:      "default",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &k8s.MockClient{
				K8sResourcesFunc: func(ctx context.Context, resourceType string, namespace string) ([]models.KubernetesResource, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockResources, nil
				},
				ListRolesFunc: func(ctx context.Context, namespace string) ([]models.RBACResource, error) {
					return []models.RBACResource{}, nil
				},
				ListClusterRolesFunc: func(ctx context.Context) ([]models.RBACResource, error) {
					return []models.RBACResource{}, nil
				},
				ListRoleBindingsFunc: func(ctx context.Context, namespace string) ([]models.RBACResource, error) {
					return []models.RBACResource{}, nil
				},
				ListClusterRoleBindingsFunc: func(ctx context.Context) ([]models.RBACResource, error) {
					return []models.RBACResource{}, nil
				},
			}

			handler := NewResourceHandler(mockClient, logger)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/resources?resourceType="+tt.resourceType+"&namespace="+tt.namespace, nil)
			w := httptest.NewRecorder()

			handler.ListKubernetesResources(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response models.KubernetesResourceList
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if len(response.Items) != tt.expectedCount {
					t.Errorf("expected %d items, got %d", tt.expectedCount, len(response.Items))
				}
			}
		})
	}
}

func TestGetResourceAccess(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	mockClient := &k8s.MockClient{
		ListRolesFunc: func(ctx context.Context, namespace string) ([]models.RBACResource, error) {
			return []models.RBACResource{
				{
					Kind:      "Role",
					Name:      "secret-reader",
					Namespace: "default",
					Rules: []models.PolicyRule{
						{
							APIGroups: []string{""},
							Resources: []string{"secrets"},
							Verbs:     []string{"get", "list"},
						},
					},
				},
			}, nil
		},
		ListClusterRolesFunc: func(ctx context.Context) ([]models.RBACResource, error) {
			return []models.RBACResource{}, nil
		},
		ListRoleBindingsFunc: func(ctx context.Context, namespace string) ([]models.RBACResource, error) {
			return []models.RBACResource{
				{
					Kind:      "RoleBinding",
					Name:      "read-secrets",
					Namespace: "default",
					Subjects: []models.Subject{
						{
							Kind:      "ServiceAccount",
							Name:      "app-sa",
							Namespace: "default",
						},
					},
					RoleRef: &models.RoleRef{
						Kind:     "Role",
						Name:     "secret-reader",
						APIGroup: "rbac.authorization.k8s.io",
					},
				},
			}, nil
		},
		ListClusterRoleBindingsFunc: func(ctx context.Context) ([]models.RBACResource, error) {
			return []models.RBACResource{}, nil
		},
	}

	handler := NewResourceHandler(mockClient, logger)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/resources/secrets/default/test-secret/access", nil)
	w := httptest.NewRecorder()

	// Setup mux to extract path variables
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/resources/{type}/{namespace}/{name}/access", handler.GetResourceAccess)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.ResourceAccessDetail
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.AccessGrants) == 0 {
		t.Error("expected at least one access grant")
	}

	if response.Summary.TotalPrincipals == 0 {
		t.Error("expected at least one principal")
	}
}

func TestListResourceTypes(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mockClient := &k8s.MockClient{}

	handler := NewResourceHandler(mockClient, logger)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/resources/types", nil)
	w := httptest.NewRecorder()

	handler.ListResourceTypes(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.ResourceTypeList
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.ResourceTypes) == 0 {
		t.Error("expected at least one resource type")
	}

	// Check for common types
	hasSecrets := false
	for _, rt := range response.ResourceTypes {
		if rt == "secrets" {
			hasSecrets = true
			break
		}
	}

	if !hasSecrets {
		t.Error("expected 'secrets' in resource types")
	}
}
