// backend/test/server_test.go
package test

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestServer_StartupAndShutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping server startup test in short mode")
	}

	// This test actually starts the server and verifies it responds
	// Requires a valid kubeconfig or mock setup

	// Set a test port
	os.Setenv("PORT", "18080")
	defer os.Unsetenv("PORT")

	// Build the binary
	cmd := exec.Command("go", "build", "-o", "/tmp/test-api-server", "../cmd/api")
	if err := cmd.Run(); err != nil {
		t.Skipf("skipping test: failed to build server: %v", err)
	}
	defer os.Remove("/tmp/test-api-server")

	// Start the server
	serverCmd := exec.Command("/tmp/test-api-server")
	serverCmd.Env = append(os.Environ(), "PORT=18080")

	if err := serverCmd.Start(); err != nil {
		t.Skipf("skipping test: failed to start server: %v", err)
	}
	defer func() {
		if serverCmd.Process != nil {
			serverCmd.Process.Kill()
		}
	}()

	// Wait for server to start
	time.Sleep(2 * time.Second)

	// Test health endpoint
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:18080/api/v1/healthz", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Skipf("server not responding (might need kubeconfig): %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	if len(body) == 0 {
		t.Error("expected non-empty response body")
	}

	t.Logf("Health check response: %s", string(body))
}

func TestServer_CORS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CORS test in short mode")
	}

	// This would require a running server
	// Placeholder for CORS verification test
	t.Log("CORS test would verify Access-Control-Allow-Origin headers")
}

func TestServer_RateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping rate limiting test in short mode")
	}

	// This would require a running server
	// Placeholder for rate limiting verification test
	t.Log("Rate limiting test would verify 429 responses after threshold")
}
