// backend/internal/k8s/interface.go
package k8s

import (
	"context"

	"github.com/djryanj/k8s-rbactory-backend/internal/models"
)

// ClientInterface defines the interface for Kubernetes operations
type ClientInterface interface {
	// RBAC operations
	ListRoles(ctx context.Context, namespace string) ([]models.RBACResource, error)
	GetRole(ctx context.Context, namespace, name string) (*models.RBACResource, error)
	ListClusterRoles(ctx context.Context) ([]models.RBACResource, error)
	GetClusterRole(ctx context.Context, name string) (*models.RBACResource, error)
	ListRoleBindings(ctx context.Context, namespace string) ([]models.RBACResource, error)
	ListClusterRoleBindings(ctx context.Context) ([]models.RBACResource, error)

	// Cluster operations
	ListNamespaces(ctx context.Context) ([]string, error)
	GetClusterVersion(ctx context.Context) (string, error)
	GetNodeCount(ctx context.Context) (int, error)

	// Resource operations
	ListKubernetesResources(ctx context.Context, resourceType string, namespace string) ([]models.KubernetesResource, error)
	ListSecrets(ctx context.Context, namespace string) ([]models.KubernetesResource, error)
	ListConfigMaps(ctx context.Context, namespace string) ([]models.KubernetesResource, error)
	ListPods(ctx context.Context, namespace string) ([]models.KubernetesResource, error)
	ListServices(ctx context.Context, namespace string) ([]models.KubernetesResource, error)
}

// Ensure Client implements ClientInterface
var _ ClientInterface = (*Client)(nil)

// Ensure MockClient implements ClientInterface
var _ ClientInterface = (*MockClient)(nil)
