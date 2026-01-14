// backend/internal/handlers/rbac.go
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/djryanj/k8s-rbactory-backend/internal/k8s"
	"github.com/djryanj/k8s-rbactory-backend/internal/logging"
	"github.com/djryanj/k8s-rbactory-backend/internal/models"
	"github.com/gorilla/mux"
)

type RBACHandler struct {
	k8sClient k8s.ClientInterface // Changed from *k8s.Client
	logger    *slog.Logger
}

// NewRBACHandler creates a new RBAC handler
func NewRBACHandler(client k8s.ClientInterface, logger *slog.Logger) *RBACHandler {
	return &RBACHandler{
		k8sClient: client,
		logger:    logger,
	}
}

// Helper function to get pagination parameters
func getPaginationParams(r *http.Request) (limit, offset int) {
	limit = 20
	offset = 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return
}

// ListRoles handles GET /api/v1/roles with pagination and validation
func (h *RBACHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	namespace := r.URL.Query().Get("namespace")

	// Validate namespace
	if err := ValidateNamespace(namespace); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid namespace",
			ValidationError{Field: "namespace", Message: err.Error()})
		return
	}

	// Validate pagination
	limit, offset, err := ValidatePaginationParams(r)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	logging.LogInfo(ctx, "listing roles",
		slog.String("namespace", namespace),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)

	roles, err := h.k8sClient.ListRoles(ctx, namespace)
	if err != nil {
		logging.LogError(ctx, "failed to list roles", err,
			slog.String("namespace", namespace),
		)
		WriteErrorResponse(w, http.StatusInternalServerError, "Failed to list roles")
		return
	}

	// Safe pagination with bounds checking
	totalCount := len(roles)
	start := min(offset, totalCount)
	end := min(offset+limit, totalCount)

	paginatedRoles := roles[start:end]
	hasMore := end < totalCount

	response := models.RBACList{
		Items:      paginatedRoles,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
		HasMore:    hasMore,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.LogError(ctx, "failed to encode response", err)
	}
}

// GetRole handles GET /api/v1/roles/{namespace}/{name} with validation
func (h *RBACHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	namespace := vars["namespace"]
	name := vars["name"]

	// Validate inputs
	if err := ValidateNamespace(namespace); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid namespace",
			ValidationError{Field: "namespace", Message: err.Error()})
		return
	}

	if err := ValidateK8sName(name); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "Invalid name",
			ValidationError{Field: "name", Message: err.Error()})
		return
	}

	logging.LogInfo(ctx, "getting role",
		slog.String("namespace", namespace),
		slog.String("name", name),
	)

	role, err := h.k8sClient.GetRole(ctx, namespace, name)
	if err != nil {
		logging.LogError(ctx, "failed to get role", err,
			slog.String("namespace", namespace),
			slog.String("name", name),
		)
		WriteErrorResponse(w, http.StatusNotFound, "Role not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(role); err != nil {
		logging.LogError(ctx, "failed to encode response", err)
	}
}

// Helper function for min (Go 1.21+)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ListClusterRoles handles GET /api/v1/clusterroles with pagination
func (h *RBACHandler) ListClusterRoles(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	limit, offset := getPaginationParams(r)

	logging.LogInfo(ctx, "listing cluster roles",
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)

	clusterRoles, err := h.k8sClient.ListClusterRoles(ctx)
	if err != nil {
		logging.LogError(ctx, "failed to list cluster roles", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Apply pagination
	totalCount := len(clusterRoles)
	start := offset
	end := offset + limit

	if start > totalCount {
		start = totalCount
	}
	if end > totalCount {
		end = totalCount
	}

	paginatedRoles := clusterRoles[start:end]
	hasMore := end < totalCount

	response := models.RBACList{
		Items:      paginatedRoles,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
		HasMore:    hasMore,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.LogError(ctx, "failed to encode response", err)
	}
}

// ListRoleBindings handles GET /api/v1/rolebindings with pagination
func (h *RBACHandler) ListRoleBindings(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	namespace := r.URL.Query().Get("namespace")
	limit, offset := getPaginationParams(r)

	logging.LogInfo(ctx, "listing role bindings",
		slog.String("namespace", namespace),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)

	bindings, err := h.k8sClient.ListRoleBindings(ctx, namespace)
	if err != nil {
		logging.LogError(ctx, "failed to list role bindings", err,
			slog.String("namespace", namespace),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Apply pagination
	totalCount := len(bindings)
	start := offset
	end := offset + limit

	if start > totalCount {
		start = totalCount
	}
	if end > totalCount {
		end = totalCount
	}

	paginatedBindings := bindings[start:end]
	hasMore := end < totalCount

	response := models.RBACList{
		Items:      paginatedBindings,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
		HasMore:    hasMore,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.LogError(ctx, "failed to encode response", err)
	}
}

// ListClusterRoleBindings handles GET /api/v1/clusterrolebindings with pagination
func (h *RBACHandler) ListClusterRoleBindings(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	limit, offset := getPaginationParams(r)

	logging.LogInfo(ctx, "listing cluster role bindings",
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)

	bindings, err := h.k8sClient.ListClusterRoleBindings(ctx)
	if err != nil {
		logging.LogError(ctx, "failed to list cluster role bindings", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Apply pagination
	totalCount := len(bindings)
	start := offset
	end := offset + limit

	if start > totalCount {
		start = totalCount
	}
	if end > totalCount {
		end = totalCount
	}

	paginatedBindings := bindings[start:end]
	hasMore := end < totalCount

	response := models.RBACList{
		Items:      paginatedBindings,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
		HasMore:    hasMore,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.LogError(ctx, "failed to encode response", err)
	}
}

// GetClusterRole handles GET /api/v1/clusterroles/{name}
func (h *RBACHandler) GetClusterRole(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	name := vars["name"]

	logging.LogInfo(ctx, "getting cluster role",
		slog.String("name", name),
	)

	clusterRole, err := h.k8sClient.GetClusterRole(ctx, name)
	if err != nil {
		logging.LogError(ctx, "failed to get cluster role", err,
			slog.String("name", name),
		)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(clusterRole); err != nil {
		logging.LogError(ctx, "failed to encode response", err)
	}
}

// ListPrincipals handles GET /api/v1/principals with pagination
func (h *RBACHandler) ListPrincipals(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	namespace := r.URL.Query().Get("namespace")
	limit, offset := getPaginationParams(r)

	logging.LogInfo(ctx, "listing principals",
		slog.String("namespace", namespace),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)

	// Fetch all bindings using existing methods
	roleBindings, err := h.k8sClient.ListRoleBindings(ctx, "")
	if err != nil {
		logging.LogError(ctx, "failed to list role bindings", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	clusterRoleBindings, err := h.k8sClient.ListClusterRoleBindings(ctx)
	if err != nil {
		logging.LogError(ctx, "failed to list cluster role bindings", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Compute principals
	principalMap := make(map[string]*models.Principal)

	// Process role bindings
	for _, rb := range roleBindings {
		if namespace != "" && rb.Namespace != namespace {
			continue
		}

		for _, subject := range rb.Subjects {
			key := subject.Kind + ":" + subject.Name + ":" + subject.Namespace

			if _, exists := principalMap[key]; !exists {
				principalMap[key] = &models.Principal{
					Kind:      subject.Kind,
					Name:      subject.Name,
					Namespace: subject.Namespace,
					Bindings:  []string{},
					Roles:     []string{},
				}
			}

			principal := principalMap[key]
			principal.Bindings = append(principal.Bindings, rb.Name)
			principal.BindingCount++

			if rb.RoleRef != nil {
				roleKey := rb.RoleRef.Kind + ":" + rb.RoleRef.Name
				if !contains(principal.Roles, roleKey) {
					principal.Roles = append(principal.Roles, roleKey)
					principal.RoleCount++
				}
			}
		}
	}

	// Process cluster role bindings
	for _, crb := range clusterRoleBindings {
		for _, subject := range crb.Subjects {
			if namespace != "" && subject.Kind == "ServiceAccount" && subject.Namespace != namespace {
				continue
			}

			key := subject.Kind + ":" + subject.Name + ":" + subject.Namespace

			if _, exists := principalMap[key]; !exists {
				principalMap[key] = &models.Principal{
					Kind:      subject.Kind,
					Name:      subject.Name,
					Namespace: subject.Namespace,
					Bindings:  []string{},
					Roles:     []string{},
				}
			}

			principal := principalMap[key]
			principal.Bindings = append(principal.Bindings, crb.Name)
			principal.BindingCount++

			if crb.RoleRef != nil {
				roleKey := crb.RoleRef.Kind + ":" + crb.RoleRef.Name
				if !contains(principal.Roles, roleKey) {
					principal.Roles = append(principal.Roles, roleKey)
					principal.RoleCount++
				}
			}
		}
	}

	// Convert map to slice
	principals := make([]models.Principal, 0, len(principalMap))
	for _, p := range principalMap {
		principals = append(principals, *p)
	}

	// Apply pagination
	totalCount := len(principals)
	start := offset
	end := offset + limit

	if start > totalCount {
		start = totalCount
	}
	if end > totalCount {
		end = totalCount
	}

	paginatedPrincipals := principals[start:end]
	hasMore := end < totalCount

	response := models.PrincipalList{
		Items:      paginatedPrincipals,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
		HasMore:    hasMore,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.LogError(ctx, "failed to encode response", err)
	}
}

// GetCounts handles GET /api/v1/counts
// backend/internal/handlers/rbac.go

// GetCounts handles GET /api/v1/counts with proper goroutine management
func (h *RBACHandler) GetCounts(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logging.LogInfo(ctx, "getting resource counts")

	type countResult struct {
		name  string
		count int
		err   error
	}

	// Use errgroup for better goroutine management
	// var (
	// 	rolesCount               int
	// 	clusterRolesCount        int
	// 	roleBindingsCount        int
	// 	clusterRoleBindingsCount int
	// 	principalsCount          int
	// )

	// Create a wait group to ensure all goroutines complete
	var wg sync.WaitGroup
	results := make(chan countResult, 5)

	// Helper to safely send results
	sendResult := func(name string, count int, err error) {
		defer wg.Done()
		select {
		case results <- countResult{name, count, err}:
		case <-ctx.Done():
			// Context cancelled, don't block
		}
	}

	// Fetch roles count
	wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logging.LogError(ctx, "panic in roles count", fmt.Errorf("%v", r))
				sendResult("roles", 0, fmt.Errorf("panic: %v", r))
			}
		}()
		roles, err := h.k8sClient.ListRoles(ctx, "")
		count := 0
		if err == nil {
			count = len(roles)
		} else {
			logging.LogError(ctx, "failed to count roles", err)
		}
		sendResult("roles", count, err)
	}()

	// Fetch cluster roles count
	wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logging.LogError(ctx, "panic in cluster roles count", fmt.Errorf("%v", r))
				sendResult("clusterRoles", 0, fmt.Errorf("panic: %v", r))
			}
		}()
		clusterRoles, err := h.k8sClient.ListClusterRoles(ctx)
		count := 0
		if err == nil {
			count = len(clusterRoles)
		} else {
			logging.LogError(ctx, "failed to count cluster roles", err)
		}
		sendResult("clusterRoles", count, err)
	}()

	// Fetch role bindings count
	wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logging.LogError(ctx, "panic in role bindings count", fmt.Errorf("%v", r))
				sendResult("roleBindings", 0, fmt.Errorf("panic: %v", r))
			}
		}()
		roleBindings, err := h.k8sClient.ListRoleBindings(ctx, "")
		count := 0
		if err == nil {
			count = len(roleBindings)
		} else {
			logging.LogError(ctx, "failed to count role bindings", err)
		}
		sendResult("roleBindings", count, err)
	}()

	// Fetch cluster role bindings count
	wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logging.LogError(ctx, "panic in cluster role bindings count", fmt.Errorf("%v", r))
				sendResult("clusterRoleBindings", 0, fmt.Errorf("panic: %v", r))
			}
		}()
		clusterRoleBindings, err := h.k8sClient.ListClusterRoleBindings(ctx)
		count := 0
		if err == nil {
			count = len(clusterRoleBindings)
		} else {
			logging.LogError(ctx, "failed to count cluster role bindings", err)
		}
		sendResult("clusterRoleBindings", count, err)
	}()

	// Compute principals count
	wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logging.LogError(ctx, "panic in principals count", fmt.Errorf("%v", r))
				sendResult("principals", 0, fmt.Errorf("panic: %v", r))
			}
		}()
		roleBindings, _ := h.k8sClient.ListRoleBindings(ctx, "")
		clusterRoleBindings, _ := h.k8sClient.ListClusterRoleBindings(ctx)

		principalMap := make(map[string]bool)

		for _, rb := range roleBindings {
			for _, subject := range rb.Subjects {
				key := subject.Kind + ":" + subject.Name + ":" + subject.Namespace
				principalMap[key] = true
			}
		}

		for _, crb := range clusterRoleBindings {
			for _, subject := range crb.Subjects {
				key := subject.Kind + ":" + subject.Name + ":" + subject.Namespace
				principalMap[key] = true
			}
		}

		sendResult("principals", len(principalMap), nil)
	}()

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results with timeout protection
	counts := models.ResourceCounts{}
	for result := range results {
		if result.err != nil {
			logging.LogError(ctx, "failed to get count", result.err,
				slog.String("type", result.name),
			)
		}

		switch result.name {
		case "roles":
			counts.Roles = result.count
		case "clusterRoles":
			counts.ClusterRoles = result.count
		case "roleBindings":
			counts.RoleBindings = result.count
		case "clusterRoleBindings":
			counts.ClusterRoleBindings = result.count
		case "principals":
			counts.Principals = result.count
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(counts); err != nil {
		logging.LogError(ctx, "failed to encode response", err)
	}
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetRelationships handles GET /api/v1/relationships/{kind}/{namespace}/{name}
func (h *RBACHandler) GetRelationships(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	kind := vars["kind"]
	namespace := vars["namespace"]
	name := vars["name"]

	if name == "" {
		name = namespace
		namespace = ""
	}

	logging.LogInfo(ctx, "getting relationships",
		slog.String("kind", kind),
		slog.String("namespace", namespace),
		slog.String("name", name),
	)

	var response models.RelationshipResponse

	switch kind {
	case "Role", "ClusterRole":
		response = h.getRoleRelationships(ctx, kind, namespace, name)
	case "RoleBinding", "ClusterRoleBinding":
		response = h.getBindingRelationships(ctx, kind, namespace, name)
	case "User", "Group", "ServiceAccount":
		response = h.getPrincipalRelationships(ctx, kind, namespace, name)
	default:
		http.Error(w, "Invalid resource kind", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.LogError(ctx, "failed to encode response", err)
	}
}

func (h *RBACHandler) getRoleRelationships(ctx context.Context, kind, namespace, name string) models.RelationshipResponse {
	allRoleBindings, _ := h.k8sClient.ListRoleBindings(ctx, "")
	allClusterRoleBindings, _ := h.k8sClient.ListClusterRoleBindings(ctx)

	var relatedBindings []models.RBACResource

	for _, rb := range allRoleBindings {
		if rb.RoleRef != nil && rb.RoleRef.Name == name && rb.RoleRef.Kind == kind {
			relatedBindings = append(relatedBindings, rb)
		}
	}

	for _, crb := range allClusterRoleBindings {
		if crb.RoleRef != nil && crb.RoleRef.Name == name && crb.RoleRef.Kind == kind {
			relatedBindings = append(relatedBindings, crb)
		}
	}

	return models.RelationshipResponse{
		RelatedBindings: relatedBindings,
		RelatedRoles:    []models.RBACResource{},
	}
}

func (h *RBACHandler) getBindingRelationships(ctx context.Context, kind, namespace, name string) models.RelationshipResponse {
	var binding models.RBACResource
	var roleRef *models.RoleRef
	bindingFound := false

	if kind == "RoleBinding" {
		bindings, _ := h.k8sClient.ListRoleBindings(ctx, namespace)
		for _, rb := range bindings {
			if rb.Name == name {
				binding = rb
				roleRef = rb.RoleRef
				bindingFound = true
				break
			}
		}
	} else {
		bindings, _ := h.k8sClient.ListClusterRoleBindings(ctx)
		for _, crb := range bindings {
			if crb.Name == name {
				binding = crb
				roleRef = crb.RoleRef
				bindingFound = true
				break
			}
		}
	}

	var role *models.RBACResource
	if roleRef != nil {
		if roleRef.Kind == "Role" {
			r, err := h.k8sClient.GetRole(ctx, namespace, roleRef.Name)
			if err == nil {
				role = r
			}
		} else if roleRef.Kind == "ClusterRole" {
			cr, err := h.k8sClient.GetClusterRole(ctx, roleRef.Name)
			if err == nil {
				role = cr
			}
		}
	}

	var relatedBindings []models.RBACResource
	if roleRef != nil {
		allRoleBindings, _ := h.k8sClient.ListRoleBindings(ctx, "")
		allClusterRoleBindings, _ := h.k8sClient.ListClusterRoleBindings(ctx)

		for _, rb := range allRoleBindings {
			if rb.RoleRef != nil && rb.RoleRef.Name == roleRef.Name && rb.RoleRef.Kind == roleRef.Kind && rb.Name != name {
				relatedBindings = append(relatedBindings, rb)
			}
		}

		for _, crb := range allClusterRoleBindings {
			if crb.RoleRef != nil && crb.RoleRef.Name == roleRef.Name && crb.RoleRef.Kind == roleRef.Kind && crb.Name != name {
				relatedBindings = append(relatedBindings, crb)
			}
		}
	}

	var relatedRoles []models.RBACResource
	if role != nil && role.Name != "" {
		relatedRoles = append(relatedRoles, *role)
	}

	var bindingPtr *models.RBACResource
	if bindingFound {
		bindingPtr = &binding
	}

	return models.RelationshipResponse{
		Role:            role,
		Binding:         bindingPtr,
		RelatedBindings: relatedBindings,
		RelatedRoles:    relatedRoles,
	}
}

func (h *RBACHandler) getPrincipalRelationships(ctx context.Context, kind, namespace, name string) models.RelationshipResponse {
	allRoleBindings, _ := h.k8sClient.ListRoleBindings(ctx, "")
	allClusterRoleBindings, _ := h.k8sClient.ListClusterRoleBindings(ctx)

	var relatedBindings []models.RBACResource
	roleMap := make(map[string]models.RBACResource)

	for _, rb := range allRoleBindings {
		for _, subject := range rb.Subjects {
			matchesKind := subject.Kind == kind
			matchesName := subject.Name == name
			matchesNamespace := namespace == "" || subject.Namespace == namespace

			if matchesKind && matchesName && matchesNamespace {
				relatedBindings = append(relatedBindings, rb)

				if rb.RoleRef != nil {
					roleKey := rb.RoleRef.Kind + ":" + rb.RoleRef.Name
					if _, exists := roleMap[roleKey]; !exists {
						if rb.RoleRef.Kind == "Role" {
							if rolePtr, err := h.k8sClient.GetRole(ctx, rb.Namespace, rb.RoleRef.Name); err == nil {
								roleMap[roleKey] = *rolePtr
							}
						} else if rb.RoleRef.Kind == "ClusterRole" {
							if rolePtr, err := h.k8sClient.GetClusterRole(ctx, rb.RoleRef.Name); err == nil {
								roleMap[roleKey] = *rolePtr
							}
						}
					}
				}
				break
			}
		}
	}

	for _, crb := range allClusterRoleBindings {
		for _, subject := range crb.Subjects {
			matchesKind := subject.Kind == kind
			matchesName := subject.Name == name
			matchesNamespace := namespace == "" || subject.Namespace == namespace

			if matchesKind && matchesName && matchesNamespace {
				relatedBindings = append(relatedBindings, crb)

				if crb.RoleRef != nil {
					roleKey := crb.RoleRef.Kind + ":" + crb.RoleRef.Name
					if _, exists := roleMap[roleKey]; !exists {
						if crb.RoleRef.Kind == "ClusterRole" {
							if rolePtr, err := h.k8sClient.GetClusterRole(ctx, crb.RoleRef.Name); err == nil {
								roleMap[roleKey] = *rolePtr
							}
						}
					}
				}
				break
			}
		}
	}

	var relatedRoles []models.RBACResource
	for _, role := range roleMap {
		relatedRoles = append(relatedRoles, role)
	}

	return models.RelationshipResponse{
		RelatedBindings: relatedBindings,
		RelatedRoles:    relatedRoles,
	}
}
