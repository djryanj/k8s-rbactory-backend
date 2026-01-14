// backend/internal/k8s/mock_client_test.go
package k8s

import (
	"context"
	"fmt"
	"testing"

	"github.com/djryanj/k8s-rbactory-backend/internal/models"
)

func TestMockClient_ImplementsInterface(t *testing.T) {
	// Compile-time check that MockClient implements ClientInterface
	var _ ClientInterface = (*MockClient)(nil)

	t.Log("MockClient correctly implements ClientInterface")
}

func TestMockClient_DefaultBehavior(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	t.Run("ListRoles returns empty slice by default", func(t *testing.T) {
		roles, err := mock.ListRoles(ctx, "default")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(roles) != 0 {
			t.Errorf("expected empty slice, got %d items", len(roles))
		}
	})

	t.Run("GetRole returns error by default", func(t *testing.T) {
		_, err := mock.GetRole(ctx, "default", "test")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("ListClusterRoles returns empty slice by default", func(t *testing.T) {
		roles, err := mock.ListClusterRoles(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(roles) != 0 {
			t.Errorf("expected empty slice, got %d items", len(roles))
		}
	})

	t.Run("GetClusterRole returns error by default", func(t *testing.T) {
		_, err := mock.GetClusterRole(ctx, "test")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("ListRoleBindings returns empty slice by default", func(t *testing.T) {
		bindings, err := mock.ListRoleBindings(ctx, "default")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(bindings) != 0 {
			t.Errorf("expected empty slice, got %d items", len(bindings))
		}
	})

	t.Run("ListClusterRoleBindings returns empty slice by default", func(t *testing.T) {
		bindings, err := mock.ListClusterRoleBindings(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(bindings) != 0 {
			t.Errorf("expected empty slice, got %d items", len(bindings))
		}
	})

	t.Run("ListNamespaces returns default namespaces", func(t *testing.T) {
		namespaces, err := mock.ListNamespaces(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(namespaces) != 2 {
			t.Errorf("expected 2 default namespaces, got %d", len(namespaces))
		}
	})

	t.Run("GetClusterVersion returns default version", func(t *testing.T) {
		version, err := mock.GetClusterVersion(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if version != "v1.28.0" {
			t.Errorf("expected version v1.28.0, got %s", version)
		}
	})
}

func TestMockClient_CustomBehavior(t *testing.T) {
	ctx := context.Background()

	t.Run("ListRoles with custom function", func(t *testing.T) {
		mock := &MockClient{
			ListRolesFunc: func(ctx context.Context, namespace string) ([]models.RBACResource, error) {
				return []models.RBACResource{
					{Name: "role1", Namespace: namespace},
					{Name: "role2", Namespace: namespace},
				}, nil
			},
		}

		roles, err := mock.ListRoles(ctx, "test-ns")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(roles) != 2 {
			t.Errorf("expected 2 roles, got %d", len(roles))
		}
		if roles[0].Namespace != "test-ns" {
			t.Errorf("expected namespace 'test-ns', got %s", roles[0].Namespace)
		}
	})

	t.Run("GetRole with custom function", func(t *testing.T) {
		mock := &MockClient{
			GetRoleFunc: func(ctx context.Context, namespace, name string) (*models.RBACResource, error) {
				return &models.RBACResource{
					Name:      name,
					Namespace: namespace,
					Kind:      "Role",
				}, nil
			},
		}

		role, err := mock.GetRole(ctx, "default", "test-role")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if role.Name != "test-role" {
			t.Errorf("expected name 'test-role', got %s", role.Name)
		}
	})

	t.Run("ListRoles with error", func(t *testing.T) {
		expectedErr := fmt.Errorf("mock error")
		mock := &MockClient{
			ListRolesFunc: func(ctx context.Context, namespace string) ([]models.RBACResource, error) {
				return nil, expectedErr
			},
		}

		_, err := mock.ListRoles(ctx, "default")
		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}

func TestMockClient_ContextHandling(t *testing.T) {
	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		mock := &MockClient{
			ListRolesFunc: func(ctx context.Context, namespace string) ([]models.RBACResource, error) {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				default:
					return []models.RBACResource{}, nil
				}
			},
		}

		_, err := mock.ListRoles(ctx, "default")
		if err != context.Canceled {
			t.Errorf("expected context.Canceled error, got %v", err)
		}
	})
}

func TestMockClient_AllMethodsCallable(t *testing.T) {
	// Verify all interface methods can be called without panic
	mock := NewMockClient()
	ctx := context.Background()

	tests := []struct {
		name string
		fn   func() error
	}{
		{
			name: "ListRoles",
			fn: func() error {
				_, err := mock.ListRoles(ctx, "default")
				return err
			},
		},
		{
			name: "GetRole",
			fn: func() error {
				_, err := mock.GetRole(ctx, "default", "test")
				return err
			},
		},
		{
			name: "ListClusterRoles",
			fn: func() error {
				_, err := mock.ListClusterRoles(ctx)
				return err
			},
		},
		{
			name: "GetClusterRole",
			fn: func() error {
				_, err := mock.GetClusterRole(ctx, "test")
				return err
			},
		},
		{
			name: "ListRoleBindings",
			fn: func() error {
				_, err := mock.ListRoleBindings(ctx, "default")
				return err
			},
		},
		{
			name: "ListClusterRoleBindings",
			fn: func() error {
				_, err := mock.ListClusterRoleBindings(ctx)
				return err
			},
		},
		{
			name: "ListNamespaces",
			fn: func() error {
				_, err := mock.ListNamespaces(ctx)
				return err
			},
		},
		{
			name: "GetClusterVersion",
			fn: func() error {
				_, err := mock.GetClusterVersion(ctx)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			_ = tt.fn()
		})
	}
}

func BenchmarkMockClient_ListRoles(b *testing.B) {
	mock := &MockClient{
		ListRolesFunc: func(ctx context.Context, namespace string) ([]models.RBACResource, error) {
			return []models.RBACResource{
				{Name: "role1"},
				{Name: "role2"},
			}, nil
		},
	}

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = mock.ListRoles(ctx, "default")
	}
}
