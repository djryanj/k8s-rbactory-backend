// backend/internal/k8s/mock_client.go
package k8s

import (
	"context"
	"fmt"

	"github.com/djryanj/k8s-rbactory-backend/internal/models"
)

// MockClient is a mock Kubernetes client for testing
type MockClient struct {
	ListRolesFunc               func(ctx context.Context, namespace string) ([]models.RBACResource, error)
	GetRoleFunc                 func(ctx context.Context, namespace, name string) (*models.RBACResource, error)
	ListClusterRolesFunc        func(ctx context.Context) ([]models.RBACResource, error)
	GetClusterRoleFunc          func(ctx context.Context, name string) (*models.RBACResource, error)
	ListRoleBindingsFunc        func(ctx context.Context, namespace string) ([]models.RBACResource, error)
	ListClusterRoleBindingsFunc func(ctx context.Context) ([]models.RBACResource, error)
	ListNamespacesFunc          func(ctx context.Context) ([]string, error)
	GetClusterVersionFunc       func(ctx context.Context) (string, error)
	GetNodeCountFunc            func(ctx context.Context) (int, error)
}

// Ensure MockClient implements ClientInterface at compile time
var _ ClientInterface = (*MockClient)(nil)

func (m *MockClient) ListRoles(ctx context.Context, namespace string) ([]models.RBACResource, error) {
	if m.ListRolesFunc != nil {
		return m.ListRolesFunc(ctx, namespace)
	}
	return []models.RBACResource{}, nil
}

func (m *MockClient) GetRole(ctx context.Context, namespace, name string) (*models.RBACResource, error) {
	if m.GetRoleFunc != nil {
		return m.GetRoleFunc(ctx, namespace, name)
	}
	return nil, fmt.Errorf("role not found")
}

func (m *MockClient) ListClusterRoles(ctx context.Context) ([]models.RBACResource, error) {
	if m.ListClusterRolesFunc != nil {
		return m.ListClusterRolesFunc(ctx)
	}
	return []models.RBACResource{}, nil
}

func (m *MockClient) GetClusterRole(ctx context.Context, name string) (*models.RBACResource, error) {
	if m.GetClusterRoleFunc != nil {
		return m.GetClusterRoleFunc(ctx, name)
	}
	return nil, fmt.Errorf("cluster role not found")
}

func (m *MockClient) ListRoleBindings(ctx context.Context, namespace string) ([]models.RBACResource, error) {
	if m.ListRoleBindingsFunc != nil {
		return m.ListRoleBindingsFunc(ctx, namespace)
	}
	return []models.RBACResource{}, nil
}

func (m *MockClient) ListClusterRoleBindings(ctx context.Context) ([]models.RBACResource, error) {
	if m.ListClusterRoleBindingsFunc != nil {
		return m.ListClusterRoleBindingsFunc(ctx)
	}
	return []models.RBACResource{}, nil
}

func (m *MockClient) ListNamespaces(ctx context.Context) ([]string, error) {
	if m.ListNamespacesFunc != nil {
		return m.ListNamespacesFunc(ctx)
	}
	return []string{"default", "kube-system"}, nil
}

func (m *MockClient) GetClusterVersion(ctx context.Context) (string, error) {
	if m.GetClusterVersionFunc != nil {
		return m.GetClusterVersionFunc(ctx)
	}
	return "v1.28.0", nil
}

// NewMockClient creates a mock client with default implementations
func NewMockClient() *MockClient {
	return &MockClient{}
}

func (m *MockClient) GetNodeCount(ctx context.Context) (int, error) {
	if m.GetNodeCountFunc != nil {
		return m.GetNodeCountFunc(ctx)
	}
	return 3, nil // default mock value
}
