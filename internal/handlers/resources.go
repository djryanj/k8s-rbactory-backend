// internal/handlers/resources.go
package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/djryanj/k8s-rbactory-backend/internal/k8s"
	"github.com/djryanj/k8s-rbactory-backend/internal/logging"
	"github.com/djryanj/k8s-rbactory-backend/internal/models"
	"github.com/gorilla/mux"
)

type ResourceHandler struct {
	k8sClient k8s.ClientInterface
	analyzer  *k8s.RBACAnalyzer
	logger    *slog.Logger
}

// NewResourceHandler creates a new resource handler
func NewResourceHandler(client k8s.ClientInterface, logger *slog.Logger) *ResourceHandler {
	return &ResourceHandler{
		k8sClient: client,
		analyzer:  k8s.NewRBACAnalyzer(client),
		logger:    logger,
	}
}

// ListKubernetesResources handles GET /api/v1/resources
func (h *ResourceHandler) ListKubernetesResources(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	resourceType := r.URL.Query().Get("resourceType")
	namespace := r.URL.Query().Get("namespace")

	// Validate resource type
	if resourceType == "" {
		WriteValidationError(w, "resourceType", "resource type is required")
		return
	}

	// Validate namespace if provided
	if err := ValidateNamespace(namespace); err != nil {
		WriteValidationError(w, "namespace", err.Error())
		return
	}

	// Validate pagination
	limit, offset, err := ValidatePaginationParams(r)
	if err != nil {
		WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	logging.LogInfo(ctx, "listing kubernetes resources",
		slog.String("resourceType", resourceType),
		slog.String("namespace", namespace),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)

	// List resources
	resources, err := h.k8sClient.ListKubernetesResources(ctx, resourceType, namespace)
	if err != nil {
		logging.LogError(ctx, "failed to list kubernetes resources", err,
			slog.String("resourceType", resourceType),
			slog.String("namespace", namespace),
		)
		WriteKubernetesError(w, err)
		return
	}

	// Analyze access for all resources
	accessMap, err := h.analyzer.AnalyzeResourceTypeAccess(ctx, resourceType, namespace)
	if err != nil {
		logging.LogWarn(ctx, "failed to analyze resource access", slog.String("error", err.Error()))
		// Continue without access info rather than failing
		accessMap = make(map[string]*models.ResourceAccessInfo)
	}

	// Attach access info to resources
	// For now, we'll use the type-level access for all resources
	// In a production system, you'd analyze each resource individually
	typeAccessInfo := accessMap["type-level-access"]
	for i := range resources {
		if typeAccessInfo != nil {
			resources[i].AccessInfo = typeAccessInfo
		}
	}

	// Apply pagination
	totalCount := len(resources)
	start := min(offset, totalCount)
	end := min(offset+limit, totalCount)

	paginatedResources := resources[start:end]
	hasMore := end < totalCount

	response := models.KubernetesResourceList{
		Items:      paginatedResources,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
		HasMore:    hasMore,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.LogError(ctx, "failed to encode response", err)
		WriteJSONError(w, http.StatusInternalServerError, "Failed to encode response")
	}
}

// GetResourceAccess handles GET /api/v1/resources/{type}/{namespace}/{name}/access
func (h *ResourceHandler) GetResourceAccess(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	resourceType := vars["type"]
	namespace := vars["namespace"]
	name := vars["name"]

	// Handle cluster-scoped resources (no namespace in path)
	if name == "" {
		name = namespace
		namespace = ""
	}

	// Validate inputs
	if err := ValidateK8sName(name); err != nil {
		WriteValidationError(w, "name", err.Error())
		return
	}

	if namespace != "" {
		if err := ValidateNamespace(namespace); err != nil {
			WriteValidationError(w, "namespace", err.Error())
			return
		}
	}

	logging.LogInfo(ctx, "getting resource access",
		slog.String("resourceType", resourceType),
		slog.String("namespace", namespace),
		slog.String("name", name),
	)

	// Analyze access
	grants, err := h.analyzer.AnalyzeResourceAccess(ctx, resourceType, namespace, name)
	if err != nil {
		logging.LogError(ctx, "failed to analyze resource access", err,
			slog.String("resourceType", resourceType),
			slog.String("namespace", namespace),
			slog.String("name", name),
		)
		WriteKubernetesError(w, err)
		return
	}

	// Build summary
	principalSet := make(map[string]bool)
	roleSet := make(map[string]bool)
	bindingSet := make(map[string]bool)
	verbCounts := make(map[string]int)

	for _, grant := range grants {
		principalKey := grant.Principal.Kind + ":" + grant.Principal.Name + ":" + grant.Principal.Namespace
		principalSet[principalKey] = true

		roleKey := grant.Role.Kind + ":" + grant.Role.Name
		roleSet[roleKey] = true

		bindingKey := grant.Binding.Kind + ":" + grant.Binding.Name
		bindingSet[bindingKey] = true

		for _, verb := range grant.Verbs {
			verbCounts[verb]++
		}
	}

	summary := models.AccessSummary{
		TotalPrincipals: len(principalSet),
		TotalRoles:      len(roleSet),
		TotalBindings:   len(bindingSet),
		VerbCounts:      verbCounts,
	}

	// Create resource object
	resource := models.KubernetesResource{
		Kind:      strings.Title(resourceType),
		Name:      name,
		Namespace: namespace,
	}

	response := models.ResourceAccessDetail{
		Resource:     resource,
		AccessGrants: grants,
		Summary:      summary,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.LogError(ctx, "failed to encode response", err)
		WriteJSONError(w, http.StatusInternalServerError, "Failed to encode response")
	}
}

// ListResourceTypes handles GET /api/v1/resources/types
func (h *ResourceHandler) ListResourceTypes(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	namespace := r.URL.Query().Get("namespace")

	logging.LogInfo(ctx, "listing resource types",
		slog.String("namespace", namespace),
	)

	// Return common resource types
	// In a production system, you might query the API server for available resources
	response := models.ResourceTypeList{
		ResourceTypes: models.CommonResourceTypes,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.LogError(ctx, "failed to encode response", err)
		WriteJSONError(w, http.StatusInternalServerError, "Failed to encode response")
	}
}
