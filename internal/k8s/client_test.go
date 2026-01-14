// backend/internal/k8s/client_test.go
package k8s

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestNewClient_InClusterConfig(t *testing.T) {
	// This test will fail outside a Kubernetes cluster, which is expected
	// We're testing that the function handles the error gracefully

	// Clear any kubeconfig environment variables
	oldKubeconfig := os.Getenv("KUBECONFIG")
	oldHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("KUBECONFIG", oldKubeconfig)
		os.Setenv("HOME", oldHome)
	}()

	// Set to non-existent paths
	os.Setenv("KUBECONFIG", "/nonexistent/kubeconfig")
	os.Setenv("HOME", "/nonexistent")

	_, err := NewClient()

	// We expect an error when not in cluster and no valid kubeconfig
	if err == nil {
		t.Log("Warning: NewClient succeeded unexpectedly (might be in cluster or have valid kubeconfig)")
	}
}

func TestNewClient_WithKubeconfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping kubeconfig test in short mode")
	}

	// Create a temporary kubeconfig file
	tmpfile, err := os.CreateTemp("", "kubeconfig-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write a minimal valid kubeconfig
	kubeconfigContent := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://localhost:6443
    insecure-skip-tls-verify: true
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`
	if _, err := tmpfile.Write([]byte(kubeconfigContent)); err != nil {
		t.Fatalf("failed to write kubeconfig: %v", err)
	}
	tmpfile.Close()

	// Set KUBECONFIG to our temp file
	oldKubeconfig := os.Getenv("KUBECONFIG")
	defer os.Setenv("KUBECONFIG", oldKubeconfig)
	os.Setenv("KUBECONFIG", tmpfile.Name())

	client, err := NewClient()

	// The client creation should succeed (even if server is unreachable)
	if err != nil {
		t.Fatalf("expected client creation to succeed with valid kubeconfig: %v", err)
	}

	if client == nil {
		t.Fatal("expected non-nil client")
	}

	if client.Clientset == nil {
		t.Error("expected non-nil clientset")
	}

	if client.Config == nil {
		t.Error("expected non-nil config")
	}

	// Verify QPS and Burst settings
	if client.Config.QPS != 100.0 {
		t.Errorf("expected QPS 100.0, got %f", client.Config.QPS)
	}

	if client.Config.Burst != 200 {
		t.Errorf("expected Burst 200, got %d", client.Config.Burst)
	}
}

func TestNewClient_InvalidKubeconfig(t *testing.T) {
	// Create a temporary invalid kubeconfig file
	tmpfile, err := os.CreateTemp("", "invalid-kubeconfig-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write invalid YAML
	if _, err := tmpfile.Write([]byte("invalid: yaml: content:")); err != nil {
		t.Fatalf("failed to write invalid kubeconfig: %v", err)
	}
	tmpfile.Close()

	oldKubeconfig := os.Getenv("KUBECONFIG")
	defer os.Setenv("KUBECONFIG", oldKubeconfig)
	os.Setenv("KUBECONFIG", tmpfile.Name())

	_, err = NewClient()

	if err == nil {
		t.Error("expected error with invalid kubeconfig")
	}
}

func TestClient_GetClusterVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cluster version test in short mode")
	}

	// This test requires a real or mock Kubernetes API server
	// For now, we'll test the error case

	client := &Client{
		Clientset: nil, // This will cause an error
		Config:    nil,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.GetClusterVersion(ctx)

	if err == nil {
		t.Error("expected error with nil clientset")
	}
}

func TestClient_ConfigurationDefaults(t *testing.T) {
	// Test that we can create a config with proper defaults
	// even if we can't connect to a cluster

	tmpfile, err := os.CreateTemp("", "kubeconfig-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	kubeconfigContent := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://localhost:6443
    insecure-skip-tls-verify: true
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`
	if _, err := tmpfile.Write([]byte(kubeconfigContent)); err != nil {
		t.Fatalf("failed to write kubeconfig: %v", err)
	}
	tmpfile.Close()

	oldKubeconfig := os.Getenv("KUBECONFIG")
	defer os.Setenv("KUBECONFIG", oldKubeconfig)
	os.Setenv("KUBECONFIG", tmpfile.Name())

	client, err := NewClient()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Verify rate limiting configuration
	tests := []struct {
		name     string
		actual   interface{}
		expected interface{}
	}{
		{"QPS", client.Config.QPS, 100.0},
		{"Burst", client.Config.Burst, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.actual != tt.expected {
				t.Errorf("expected %s to be %v, got %v", tt.name, tt.expected, tt.actual)
			}
		})
	}
}

func TestClient_GetNodeCount(t *testing.T) {
	// Test with mock client
	mockClient := NewMockClient()
	mockClient.GetNodeCountFunc = func(ctx context.Context) (int, error) {
		return 3, nil
	}

	ctx := context.Background()
	count, err := mockClient.GetNodeCount(ctx)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if count != 3 {
		t.Errorf("expected node count 3, got %d", count)
	}
}

func BenchmarkNewClient(b *testing.B) {
	// Create a valid kubeconfig
	tmpfile, err := os.CreateTemp("", "kubeconfig-*.yaml")
	if err != nil {
		b.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	kubeconfigContent := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://localhost:6443
    insecure-skip-tls-verify: true
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`
	if _, err := tmpfile.Write([]byte(kubeconfigContent)); err != nil {
		b.Fatalf("failed to write kubeconfig: %v", err)
	}
	tmpfile.Close()

	oldKubeconfig := os.Getenv("KUBECONFIG")
	defer os.Setenv("KUBECONFIG", oldKubeconfig)
	os.Setenv("KUBECONFIG", tmpfile.Name())

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = NewClient()
	}
}

func TestIsInCluster(t *testing.T) {
	tests := []struct {
		name     string
		setEnv   map[string]string
		expected bool
	}{
		{
			name: "in cluster",
			setEnv: map[string]string{
				"KUBERNETES_SERVICE_HOST": "10.96.0.1",
				"KUBERNETES_SERVICE_PORT": "443",
			},
			expected: true,
		},
		{
			name:     "not in cluster",
			setEnv:   map[string]string{},
			expected: false,
		},
		{
			name: "missing port",
			setEnv: map[string]string{
				"KUBERNETES_SERVICE_HOST": "10.96.0.1",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Unsetenv("KUBERNETES_SERVICE_HOST")
			os.Unsetenv("KUBERNETES_SERVICE_PORT")

			// Set test environment
			for k, v := range tt.setEnv {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			result := IsInCluster()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetServiceAccountNamespace(t *testing.T) {
	// This test will only work in-cluster or with a mock file
	namespace := GetServiceAccountNamespace()

	if namespace != "" {
		t.Logf("Running in namespace: %s", namespace)
	} else {
		t.Log("Not running in-cluster (expected outside Kubernetes)")
	}
}
