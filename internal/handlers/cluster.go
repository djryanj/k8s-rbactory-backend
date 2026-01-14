package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/djryanj/k8s-rbactory-backend/internal/k8s"
	"github.com/djryanj/k8s-rbactory-backend/internal/logging"
	"github.com/djryanj/k8s-rbactory-backend/internal/models"
)

type ClusterHandler struct {
	k8sClient k8s.ClientInterface // Changed from *k8s.Client
	logger    *slog.Logger
}

// NewClusterHandler creates a new cluster handler
func NewClusterHandler(client k8s.ClientInterface, logger *slog.Logger) *ClusterHandler {
	return &ClusterHandler{
		k8sClient: client,
		logger:    logger,
	}
}

// GetClusterInfo handles GET /api/v1/cluster/info
func (h *ClusterHandler) GetClusterInfo(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logging.LogInfo(ctx, "getting cluster info")

	version, err := h.k8sClient.GetClusterVersion(ctx)
	if err != nil {
		logging.LogError(ctx, "failed to get cluster version", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	namespaces, err := h.k8sClient.ListNamespaces(ctx)
	if err != nil {
		logging.LogError(ctx, "failed to list namespaces", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	nodeCount, err := h.k8sClient.GetNodeCount(ctx)
	if err != nil {
		logging.LogError(ctx, "failed to get node count", err)
		nodeCount = 0 // Don't fail the whole request
	}

	info := models.ClusterInfo{
		Version:    version,
		Namespaces: namespaces,
		NodeCount:  nodeCount,
	}

	logging.LogInfo(ctx, "cluster info retrieved",
		slog.String("version", version),
		slog.Int("namespaces", len(namespaces)),
	)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		logging.LogError(ctx, "failed to encode response", err)
	}
}

// ListNamespaces handles GET /api/v1/namespaces
func (h *ClusterHandler) ListNamespaces(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	logging.LogInfo(ctx, "listing namespaces")

	namespaces, err := h.k8sClient.ListNamespaces(ctx)
	if err != nil {
		logging.LogError(ctx, "failed to list namespaces", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.NamespaceList{
		Namespaces: namespaces,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.LogError(ctx, "failed to encode response", err)
	}
}
