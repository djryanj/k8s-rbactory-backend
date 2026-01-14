// backend/internal/k8s/client.go
package k8s

import (
	"context"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps the Kubernetes client
type Client struct {
	Clientset *kubernetes.Clientset
	Config    *rest.Config
}

// NewClient creates a new Kubernetes client
func NewClient() (*Client, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = os.Getenv("HOME") + "/.kube/config"
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create kubernetes config: %w", err)
		}
	}

	config.QPS = 100.0 // Queries per second
	config.Burst = 200 // Burst capacity

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &Client{
		Clientset: clientset,
		Config:    config,
	}, nil
}

// GetClusterVersion returns the Kubernetes cluster version
func (c *Client) GetClusterVersion(ctx context.Context) (string, error) {
	version, err := c.Clientset.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return version.GitVersion, nil
}

// GetNodeCount returns the number of nodes in the cluster
func (c *Client) GetNodeCount(ctx context.Context) (int, error) {
	nodes, err := c.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to list nodes: %w", err)
	}
	return len(nodes.Items), nil
}

// IsInCluster returns true if running inside a Kubernetes cluster
func IsInCluster() bool {
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	port := os.Getenv("KUBERNETES_SERVICE_PORT")
	return host != "" && port != ""
}

// GetServiceAccountNamespace returns the namespace of the service account
// when running in-cluster, or empty string if not in cluster
func GetServiceAccountNamespace() string {
	data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return ""
	}
	return string(data)
}
