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
	// Resource fields
	SecretsFunc      func(ctx context.Context, namespace string) ([]models.KubernetesResource, error)
	ConfigMapsFunc   func(ctx context.Context, namespace string) ([]models.KubernetesResource, error)
	PodsFunc         func(ctx context.Context, namespace string) ([]models.KubernetesResource, error)
	ServicesFunc     func(ctx context.Context, namespace string) ([]models.KubernetesResource, error)
	K8sResourcesFunc func(ctx context.Context, resourceType string, namespace string) ([]models.KubernetesResource, error)
}

// ListSecrets mock implementation
func (m *MockClient) ListSecrets(ctx context.Context, namespace string) ([]models.KubernetesResource, error) {
	if m.SecretsFunc != nil {
		return m.SecretsFunc(ctx, namespace)
	}
	return []models.KubernetesResource{}, nil
}

// ListConfigMaps mock implementation
func (m *MockClient) ListConfigMaps(ctx context.Context, namespace string) ([]models.KubernetesResource, error) {
	if m.ConfigMapsFunc != nil {
		return m.ConfigMapsFunc(ctx, namespace)
	}
	return []models.KubernetesResource{}, nil
}

// ListPods mock implementation
func (m *MockClient) ListPods(ctx context.Context, namespace string) ([]models.KubernetesResource, error) {
	if m.PodsFunc != nil {
		return m.PodsFunc(ctx, namespace)
	}
	return []models.KubernetesResource{}, nil
}

// ListServices mock implementation
func (m *MockClient) ListServices(ctx context.Context, namespace string) ([]models.KubernetesResource, error) {
	if m.ServicesFunc != nil {
		return m.ServicesFunc(ctx, namespace)
	}
	return []models.KubernetesResource{}, nil
}

// ListKubernetesResources mock implementation
func (m *MockClient) ListKubernetesResources(ctx context.Context, resourceType string, namespace string) ([]models.KubernetesResource, error) {
	if m.K8sResourcesFunc != nil {
		return m.K8sResourcesFunc(ctx, resourceType, namespace)
	}
	return []models.KubernetesResource{}, nil
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
