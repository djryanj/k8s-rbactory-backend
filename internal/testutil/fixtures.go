// backend/internal/testutil/fixtures.go
package testutil

import (
	"github.com/djryanj/k8s-rbactory-backend/internal/models"
)

// TestRole creates a test Role
func TestRole(namespace, name string) models.RBACResource {
	return models.RBACResource{
		Kind:      "Role",
		Name:      name,
		Namespace: namespace,
		Rules: []models.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "list"},
			},
		},
		Labels: map[string]string{
			"app": "test",
		},
		CreatedAt: "2024-01-01T00:00:00Z",
	}
}

// TestClusterRole creates a test ClusterRole
func TestClusterRole(name string) models.RBACResource {
	return models.RBACResource{
		Kind: "ClusterRole",
		Name: name,
		Rules: []models.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"get", "list"},
			},
		},
		Labels: map[string]string{
			"app": "test",
		},
		CreatedAt: "2024-01-01T00:00:00Z",
	}
}

// TestRoleBinding creates a test RoleBinding
func TestRoleBinding(namespace, name, roleName string) models.RBACResource {
	return models.RBACResource{
		Kind:      "RoleBinding",
		Name:      name,
		Namespace: namespace,
		Subjects: []models.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "test-sa",
				Namespace: namespace,
			},
		},
		RoleRef: &models.RoleRef{
			Kind:     "Role",
			Name:     roleName,
			APIGroup: "rbac.authorization.k8s.io",
		},
		CreatedAt: "2024-01-01T00:00:00Z",
	}
}

// TestClusterRoleBinding creates a test ClusterRoleBinding
func TestClusterRoleBinding(name, roleName string) models.RBACResource {
	return models.RBACResource{
		Kind: "ClusterRoleBinding",
		Name: name,
		Subjects: []models.Subject{
			{
				Kind: "User",
				Name: "test-user",
			},
		},
		RoleRef: &models.RoleRef{
			Kind:     "ClusterRole",
			Name:     roleName,
			APIGroup: "rbac.authorization.k8s.io",
		},
		CreatedAt: "2024-01-01T00:00:00Z",
	}
}
