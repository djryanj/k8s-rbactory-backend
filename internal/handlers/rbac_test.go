// backend/internal/handlers/rbac_test.go
package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/djryanj/k8s-rbactory-backend/internal/k8s"
	"github.com/djryanj/k8s-rbactory-backend/internal/models"
	"github.com/djryanj/k8s-rbactory-backend/internal/testutil"
	"github.com/gorilla/mux"
)

func TestRBACHandler_ListRoles(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		limit          string
		offset         string
		mockRoles      []models.RBACResource
		mockError      error
		expectedStatus int
		expectedCount  int
	}{
		{
			name:      "successful list with default pagination",
			namespace: "default",
			mockRoles: []models.RBACResource{
				testutil.TestRole("default", "role1"),
				testutil.TestRole("default", "role2"),
				testutil.TestRole("default", "role3"),
			},
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:      "successful list with custom pagination",
			namespace: "default",
			limit:     "2",
			offset:    "1",
			mockRoles: []models.RBACResource{
				testutil.TestRole("default", "role1"),
				testutil.TestRole("default", "role2"),
				testutil.TestRole("default", "role3"),
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "empty result",
			namespace:      "empty",
			mockRoles:      []models.RBACResource{},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:           "kubernetes error",
			namespace:      "default",
			mockError:      fmt.Errorf("kubernetes API error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "invalid limit",
			namespace:      "default",
			limit:          "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "negative offset",
			namespace:      "default",
			offset:         "-1",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock client
			mockClient := &k8s.MockClient{
				ListRolesFunc: func(ctx context.Context, namespace string) ([]models.RBACResource, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockRoles, nil
				},
			}

			handler := NewRBACHandler(mockClient, testutil.NewTestLogger(t))

			// Create request
			url := "/api/v1/roles?namespace=" + tt.namespace
			if tt.limit != "" {
				url += "&limit=" + tt.limit
			}
			if tt.offset != "" {
				url += "&offset=" + tt.offset
			}

			req := testutil.NewTestRequest(t, "GET", url, nil)
			rr := httptest.NewRecorder()

			// Execute
			handler.ListRoles(rr, req)

			// Assert status code
			testutil.AssertStatusCode(t, tt.expectedStatus, rr.Code)

			// Assert response body for successful requests
			if tt.expectedStatus == http.StatusOK {
				var response models.RBACList
				testutil.AssertJSONResponse(t, rr, http.StatusOK, &response)

				if len(response.Items) != tt.expectedCount {
					t.Errorf("expected %d items, got %d", tt.expectedCount, len(response.Items))
				}
			}
		})
	}
}

func TestRBACHandler_GetRole(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		roleName       string
		mockRole       *models.RBACResource
		mockError      error
		expectedStatus int
	}{
		{
			name:           "successful get",
			namespace:      "default",
			roleName:       "test-role",
			mockRole:       &models.RBACResource{Kind: "Role", Name: "test-role", Namespace: "default"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "role not found",
			namespace:      "default",
			roleName:       "nonexistent",
			mockError:      fmt.Errorf("not found"),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid namespace",
			namespace:      "Invalid-Namespace",
			roleName:       "test-role",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid role name",
			namespace:      "default",
			roleName:       "Invalid@Role",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock client
			mockClient := &k8s.MockClient{
				GetRoleFunc: func(ctx context.Context, namespace, name string) (*models.RBACResource, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockRole, nil
				},
			}

			handler := NewRBACHandler(mockClient, testutil.NewTestLogger(t))

			// Create request with mux vars
			req := testutil.NewTestRequest(t, "GET", "/api/v1/roles/"+tt.namespace+"/"+tt.roleName, nil)
			req = mux.SetURLVars(req, map[string]string{
				"namespace": tt.namespace,
				"name":      tt.roleName,
			})
			rr := httptest.NewRecorder()

			// Execute
			handler.GetRole(rr, req)

			// Assert
			testutil.AssertStatusCode(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK && tt.mockRole != nil {
				var response models.RBACResource
				testutil.AssertJSONResponse(t, rr, http.StatusOK, &response)

				if response.Name != tt.mockRole.Name {
					t.Errorf("expected role name %s, got %s", tt.mockRole.Name, response.Name)
				}
			}
		})
	}
}

func TestRBACHandler_GetCounts(t *testing.T) {
	mockClient := &k8s.MockClient{
		ListRolesFunc: func(ctx context.Context, namespace string) ([]models.RBACResource, error) {
			return []models.RBACResource{
				testutil.TestRole("default", "role1"),
				testutil.TestRole("default", "role2"),
			}, nil
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

	handler := NewRBACHandler(mockClient, testutil.NewTestLogger(t))

	req := testutil.NewTestRequest(t, "GET", "/api/v1/counts", nil)
	rr := httptest.NewRecorder()

	handler.GetCounts(rr, req)

	var response models.ResourceCounts
	testutil.AssertJSONResponse(t, rr, http.StatusOK, &response)

	if response.Roles != 2 {
		t.Errorf("expected 2 roles, got %d", response.Roles)
	}
	if response.ClusterRoles != 1 {
		t.Errorf("expected 1 cluster role, got %d", response.ClusterRoles)
	}
	if response.RoleBindings != 1 {
		t.Errorf("expected 1 role binding, got %d", response.RoleBindings)
	}
	if response.ClusterRoleBindings != 1 {
		t.Errorf("expected 1 cluster role binding, got %d", response.ClusterRoleBindings)
	}
}
