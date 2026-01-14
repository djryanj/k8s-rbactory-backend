// backend/internal/handlers/benchmark_test.go
package handlers

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/djryanj/k8s-rbactory-backend/internal/k8s"
	"github.com/djryanj/k8s-rbactory-backend/internal/models"
	"github.com/djryanj/k8s-rbactory-backend/internal/testutil"
)

func BenchmarkListRoles(b *testing.B) {
	// Create mock data
	roles := make([]models.RBACResource, 100)
	for i := 0; i < 100; i++ {
		roles[i] = testutil.TestRole("default", "role"+string(rune(i)))
	}

	mockClient := &k8s.MockClient{
		ListRolesFunc: func(ctx context.Context, namespace string) ([]models.RBACResource, error) {
			return roles, nil
		},
	}

	handler := NewRBACHandler(mockClient, testutil.NewTestLogger(b))

	req := testutil.NewTestRequest(b, "GET", "/api/v1/roles?namespace=default", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler.ListRoles(rr, req)
	}
}

func BenchmarkListRoles_Parallel(b *testing.B) {
	roles := make([]models.RBACResource, 100)
	for i := 0; i < 100; i++ {
		roles[i] = testutil.TestRole("default", "role"+string(rune(i)))
	}

	mockClient := &k8s.MockClient{
		ListRolesFunc: func(ctx context.Context, namespace string) ([]models.RBACResource, error) {
			return roles, nil
		},
	}

	handler := NewRBACHandler(mockClient, testutil.NewTestLogger(b))

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		req := testutil.NewTestRequest(b, "GET", "/api/v1/roles?namespace=default", nil)
		for pb.Next() {
			rr := httptest.NewRecorder()
			handler.ListRoles(rr, req)
		}
	})
}

func BenchmarkGetRole(b *testing.B) {
	role := testutil.TestRole("default", "test-role")

	mockClient := &k8s.MockClient{
		GetRoleFunc: func(ctx context.Context, namespace, name string) (*models.RBACResource, error) {
			return &role, nil
		},
	}

	handler := NewRBACHandler(mockClient, testutil.NewTestLogger(b))

	req := testutil.NewTestRequest(b, "GET", "/api/v1/roles/default/test-role", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler.GetRole(rr, req)
	}
}

func BenchmarkGetCounts(b *testing.B) {
	mockClient := &k8s.MockClient{
		ListRolesFunc: func(ctx context.Context, namespace string) ([]models.RBACResource, error) {
			return []models.RBACResource{
				testutil.TestRole("default", "role1"),
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

	handler := NewRBACHandler(mockClient, testutil.NewTestLogger(b))

	req := testutil.NewTestRequest(b, "GET", "/api/v1/counts", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler.GetCounts(rr, req)
	}
}

func BenchmarkGetCounts_Parallel(b *testing.B) {
	mockClient := &k8s.MockClient{
		ListRolesFunc: func(ctx context.Context, namespace string) ([]models.RBACResource, error) {
			return []models.RBACResource{
				testutil.TestRole("default", "role1"),
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

	handler := NewRBACHandler(mockClient, testutil.NewTestLogger(b))

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		req := testutil.NewTestRequest(b, "GET", "/api/v1/counts", nil)
		for pb.Next() {
			rr := httptest.NewRecorder()
			handler.GetCounts(rr, req)
		}
	})
}

func BenchmarkListPrincipals(b *testing.B) {
	mockClient := &k8s.MockClient{
		ListRoleBindingsFunc: func(ctx context.Context, namespace string) ([]models.RBACResource, error) {
			return []models.RBACResource{
				testutil.TestRoleBinding("default", "binding1", "role1"),
				testutil.TestRoleBinding("default", "binding2", "role2"),
			}, nil
		},
		ListClusterRoleBindingsFunc: func(ctx context.Context) ([]models.RBACResource, error) {
			return []models.RBACResource{
				testutil.TestClusterRoleBinding("cluster-binding1", "cluster-role1"),
			}, nil
		},
	}

	handler := NewRBACHandler(mockClient, testutil.NewTestLogger(b))

	req := testutil.NewTestRequest(b, "GET", "/api/v1/principals", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler.ListPrincipals(rr, req)
	}
}

func BenchmarkValidatePaginationParams(b *testing.B) {
	req := testutil.NewTestRequest(b, "GET", "/test?limit=50&offset=10", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, _ = ValidatePaginationParams(req)
	}
}

func BenchmarkValidateK8sName(b *testing.B) {
	testName := "my-valid-role-name"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ValidateK8sName(testName)
	}
}

// Benchmark with different data sizes
func BenchmarkListRoles_10(b *testing.B)   { benchmarkListRolesN(b, 10) }
func BenchmarkListRoles_100(b *testing.B)  { benchmarkListRolesN(b, 100) }
func BenchmarkListRoles_1000(b *testing.B) { benchmarkListRolesN(b, 1000) }

func benchmarkListRolesN(b *testing.B, n int) {
	roles := make([]models.RBACResource, n)
	for i := 0; i < n; i++ {
		roles[i] = testutil.TestRole("default", "role"+string(rune(i)))
	}

	mockClient := &k8s.MockClient{
		ListRolesFunc: func(ctx context.Context, namespace string) ([]models.RBACResource, error) {
			return roles, nil
		},
	}

	handler := NewRBACHandler(mockClient, testutil.NewTestLogger(b))

	req := testutil.NewTestRequest(b, "GET", "/api/v1/roles?namespace=default", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler.ListRoles(rr, req)
	}
}
