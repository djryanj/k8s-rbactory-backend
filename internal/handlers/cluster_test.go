// backend/internal/handlers/cluster_test.go
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
)

func TestClusterHandler_GetClusterInfo(t *testing.T) {
	tests := []struct {
		name             string
		mockVersion      string
		mockVersionError error
		mockNamespaces   []string
		mockNSError      error
		expectedStatus   int
		expectedVersion  string
		expectedNSCount  int
	}{
		{
			name:            "successful cluster info",
			mockVersion:     "v1.28.0",
			mockNamespaces:  []string{"default", "kube-system", "kube-public"},
			expectedStatus:  http.StatusOK,
			expectedVersion: "v1.28.0",
			expectedNSCount: 3,
		},
		{
			name:             "version error",
			mockVersionError: fmt.Errorf("API error"),
			expectedStatus:   http.StatusInternalServerError,
		},
		{
			name:           "namespace error",
			mockVersion:    "v1.28.0",
			mockNSError:    fmt.Errorf("API error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &k8s.MockClient{
				GetClusterVersionFunc: func(ctx context.Context) (string, error) {
					if tt.mockVersionError != nil {
						return "", tt.mockVersionError
					}
					return tt.mockVersion, nil
				},
				ListNamespacesFunc: func(ctx context.Context) ([]string, error) {
					if tt.mockNSError != nil {
						return nil, tt.mockNSError
					}
					return tt.mockNamespaces, nil
				},
			}

			handler := NewClusterHandler(mockClient, testutil.NewTestLogger(t))

			req := testutil.NewTestRequest(t, "GET", "/api/v1/cluster/info", nil)
			rr := httptest.NewRecorder()

			handler.GetClusterInfo(rr, req)

			testutil.AssertStatusCode(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response models.ClusterInfo
				testutil.AssertJSONResponse(t, rr, http.StatusOK, &response)

				if response.Version != tt.expectedVersion {
					t.Errorf("expected version %s, got %s", tt.expectedVersion, response.Version)
				}

				if len(response.Namespaces) != tt.expectedNSCount {
					t.Errorf("expected %d namespaces, got %d", tt.expectedNSCount, len(response.Namespaces))
				}
			}
		})
	}
}

func TestClusterHandler_ListNamespaces(t *testing.T) {
	mockClient := &k8s.MockClient{
		ListNamespacesFunc: func(ctx context.Context) ([]string, error) {
			return []string{"default", "kube-system"}, nil
		},
	}

	handler := NewClusterHandler(mockClient, testutil.NewTestLogger(t))

	req := testutil.NewTestRequest(t, "GET", "/api/v1/namespaces", nil)
	rr := httptest.NewRecorder()

	handler.ListNamespaces(rr, req)

	var response models.NamespaceList
	testutil.AssertJSONResponse(t, rr, http.StatusOK, &response)

	if len(response.Namespaces) != 2 {
		t.Errorf("expected 2 namespaces, got %d", len(response.Namespaces))
	}
}
