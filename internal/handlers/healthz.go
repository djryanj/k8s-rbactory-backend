// internal/handlers/healthz.go
package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/djryanj/k8s-rbactory-backend/internal/k8s"
	"github.com/djryanj/k8s-rbactory-backend/internal/logging"
)

// HealthStatus represents the overall health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    HealthStatus           `json:"status"`
	Version   string                 `json:"version"`
	Timestamp string                 `json:"timestamp"`
	Checks    map[string]HealthCheck `json:"checks"`
	Message   string                 `json:"message,omitempty"`
}

// HealthCheck represents an individual health check
type HealthCheck struct {
	Status  HealthStatus `json:"status"`
	Message string       `json:"message,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// HealthHandler handles health check requests with actual cluster validation
type HealthHandler struct {
	k8sClient k8s.ClientInterface
	logger    *slog.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(client k8s.ClientInterface, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		k8sClient: client,
		logger:    logger,
	}
}

// ServeHTTP handles the health check request
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	logging.LogInfo(ctx, "health check requested")

	response := HealthResponse{
		Status:    HealthStatusHealthy,
		Version:   "1.0.0",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    make(map[string]HealthCheck),
	}

	// Check Kubernetes connectivity
	clusterCheck := h.checkClusterConnectivity(ctx)
	response.Checks["cluster_connectivity"] = clusterCheck

	// Check RBAC permissions
	rbacCheck := h.checkRBACPermissions(ctx)
	response.Checks["rbac_permissions"] = rbacCheck

	// Determine overall status
	if clusterCheck.Status == HealthStatusUnhealthy || rbacCheck.Status == HealthStatusUnhealthy {
		response.Status = HealthStatusUnhealthy
		response.Message = "Service is unhealthy - cluster connectivity or permissions issues detected"
	} else if clusterCheck.Status == HealthStatusDegraded || rbacCheck.Status == HealthStatusDegraded {
		response.Status = HealthStatusDegraded
		response.Message = "Service is degraded - some functionality may be limited"
	}

	// Set appropriate HTTP status code
	statusCode := http.StatusOK
	if response.Status == HealthStatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.LogError(ctx, "failed to encode health response", err)
	}
}

// checkClusterConnectivity verifies we can connect to the cluster
func (h *HealthHandler) checkClusterConnectivity(ctx context.Context) HealthCheck {
	_, err := h.k8sClient.GetClusterVersion(ctx)
	if err != nil {
		logging.LogError(ctx, "cluster connectivity check failed", err)
		return HealthCheck{
			Status:  HealthStatusUnhealthy,
			Message: "Cannot connect to Kubernetes cluster",
			Error:   err.Error(),
		}
	}

	return HealthCheck{
		Status:  HealthStatusHealthy,
		Message: "Cluster is reachable",
	}
}

// checkRBACPermissions verifies we have necessary RBAC permissions
func (h *HealthHandler) checkRBACPermissions(ctx context.Context) HealthCheck {
	// Try to list namespaces as a basic permission check
	_, err := h.k8sClient.ListNamespaces(ctx)
	if err != nil {
		logging.LogError(ctx, "RBAC permissions check failed", err)

		// Determine if it's a permission issue or other error
		errMsg := err.Error()
		if containsAny(errMsg, "forbidden", "is forbidden", "cannot list") {
			return HealthCheck{
				Status:  HealthStatusUnhealthy,
				Message: "Insufficient RBAC permissions - service account cannot list namespaces",
				Error:   errMsg,
			}
		}

		return HealthCheck{
			Status:  HealthStatusUnhealthy,
			Message: "Failed to verify RBAC permissions",
			Error:   errMsg,
		}
	}

	// Check if we can list roles (another common permission)
	_, err = h.k8sClient.ListRoles(ctx, "")
	if err != nil {
		logging.LogWarn(ctx, "limited RBAC permissions detected", slog.String("error", err.Error()))
		return HealthCheck{
			Status:  HealthStatusDegraded,
			Message: "Limited RBAC permissions - some features may not work",
			Error:   err.Error(),
		}
	}

	return HealthCheck{
		Status:  HealthStatusHealthy,
		Message: "RBAC permissions verified",
	}
}

// SimpleHealthHandler provides a simple health check without K8s validation
// Use this for basic liveness probes
func SimpleHealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
