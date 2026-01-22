// internal/k8s/rbac_analyzer.go
package k8s

import (
	"context"
	"strings"

	"github.com/djryanj/k8s-rbactory-backend/internal/models"
)

// RBACAnalyzer analyzes RBAC permissions for resources
type RBACAnalyzer struct {
	client ClientInterface
}

// NewRBACAnalyzer creates a new RBAC analyzer
func NewRBACAnalyzer(client ClientInterface) *RBACAnalyzer {
	return &RBACAnalyzer{
		client: client,
	}
}

// AnalyzeResourceAccess determines who has access to a specific resource
func (a *RBACAnalyzer) AnalyzeResourceAccess(
	ctx context.Context,
	resourceType string,
	namespace string,
	resourceName string,
) ([]models.AccessGrant, error) {
	var grants []models.AccessGrant

	// Get all roles and bindings
	roles, err := a.client.ListRoles(ctx, "")
	if err != nil {
		return nil, err
	}

	clusterRoles, err := a.client.ListClusterRoles(ctx)
	if err != nil {
		return nil, err
	}

	roleBindings, err := a.client.ListRoleBindings(ctx, "")
	if err != nil {
		return nil, err
	}

	clusterRoleBindings, err := a.client.ListClusterRoleBindings(ctx)
	if err != nil {
		return nil, err
	}

	// Build role map for quick lookup
	roleMap := make(map[string]models.RBACResource)
	for _, role := range roles {
		key := makeRoleKey(role.Kind, role.Namespace, role.Name)
		roleMap[key] = role
	}
	for _, role := range clusterRoles {
		key := makeRoleKey(role.Kind, "", role.Name)
		roleMap[key] = role
	}

	// Analyze RoleBindings
	for _, binding := range roleBindings {
		// Skip if binding is not in the same namespace as the resource
		if namespace != "" && binding.Namespace != namespace {
			continue
		}

		// Find the role
		var lookupKey string
		if binding.RoleRef.Kind == "ClusterRole" {
			lookupKey = makeRoleKey("ClusterRole", "", binding.RoleRef.Name)
		} else {
			lookupKey = makeRoleKey(binding.RoleRef.Kind, binding.Namespace, binding.RoleRef.Name)
		}

		role, exists := roleMap[lookupKey]
		if !exists {
			continue
		}

		// Check if role grants access to this resource
		verbs := a.getMatchingVerbs(role, resourceType, resourceName)
		if len(verbs) == 0 {
			continue
		}

		// Create grants for each subject
		for _, subject := range binding.Subjects {
			grant := models.AccessGrant{
				Principal: models.AccessPrincipal{
					Kind:      subject.Kind,
					Name:      subject.Name,
					Namespace: subject.Namespace,
				},
				Role: models.AccessRole{
					Kind:      role.Kind,
					Name:      role.Name,
					Namespace: role.Namespace,
				},
				Binding: models.AccessBinding{
					Kind:      binding.Kind,
					Name:      binding.Name,
					Namespace: binding.Namespace,
				},
				Verbs: verbs,
				Scope: "namespace",
			}
			grants = append(grants, grant)
		}
	}

	// Analyze ClusterRoleBindings
	for _, binding := range clusterRoleBindings {
		// Find the role
		lookupKey := makeRoleKey("ClusterRole", "", binding.RoleRef.Name)
		role, exists := roleMap[lookupKey]
		if !exists {
			continue
		}

		// Check if role grants access to this resource
		verbs := a.getMatchingVerbs(role, resourceType, resourceName)
		if len(verbs) == 0 {
			continue
		}

		// Create grants for each subject
		for _, subject := range binding.Subjects {
			// For namespaced resources, check if subject is in the right namespace
			if namespace != "" && subject.Kind == "ServiceAccount" && subject.Namespace != namespace {
				// ClusterRoleBinding can grant access, but we filter by subject namespace
				// for better relevance
				continue
			}

			grant := models.AccessGrant{
				Principal: models.AccessPrincipal{
					Kind:      subject.Kind,
					Name:      subject.Name,
					Namespace: subject.Namespace,
				},
				Role: models.AccessRole{
					Kind:      role.Kind,
					Name:      role.Name,
					Namespace: role.Namespace,
				},
				Binding: models.AccessBinding{
					Kind:      binding.Kind,
					Name:      binding.Name,
					Namespace: binding.Namespace,
				},
				Verbs: verbs,
				Scope: "cluster",
			}
			grants = append(grants, grant)
		}
	}

	return grants, nil
}

// AnalyzeResourceTypeAccess analyzes access for all resources of a given type
func (a *RBACAnalyzer) AnalyzeResourceTypeAccess(
	ctx context.Context,
	resourceType string,
	namespace string,
) (map[string]*models.ResourceAccessInfo, error) {
	// Get all roles and bindings once
	roles, err := a.client.ListRoles(ctx, "")
	if err != nil {
		return nil, err
	}

	clusterRoles, err := a.client.ListClusterRoles(ctx)
	if err != nil {
		return nil, err
	}

	roleBindings, err := a.client.ListRoleBindings(ctx, "")
	if err != nil {
		return nil, err
	}

	clusterRoleBindings, err := a.client.ListClusterRoleBindings(ctx)
	if err != nil {
		return nil, err
	}

	// Build role map
	roleMap := make(map[string]models.RBACResource)
	for _, role := range roles {
		key := makeRoleKey(role.Kind, role.Namespace, role.Name)
		roleMap[key] = role
	}
	for _, role := range clusterRoles {
		key := makeRoleKey(role.Kind, "", role.Name)
		roleMap[key] = role
	}

	// Map to store access info: resourceKey -> AccessInfo
	accessMap := make(map[string]*models.ResourceAccessInfo)

	// Process RoleBindings
	for _, binding := range roleBindings {
		if namespace != "" && binding.Namespace != namespace {
			continue
		}

		var lookupKey string
		if binding.RoleRef.Kind == "ClusterRole" {
			lookupKey = makeRoleKey("ClusterRole", "", binding.RoleRef.Name)
		} else {
			lookupKey = makeRoleKey(binding.RoleRef.Kind, binding.Namespace, binding.RoleRef.Name)
		}

		role, exists := roleMap[lookupKey]
		if !exists {
			continue
		}

		// Check if role grants access to this resource type
		if !a.roleGrantsAccessToResourceType(role, resourceType) {
			continue
		}

		// This role grants access to the resource type
		// We'll mark all resources of this type as accessible
		for _, subject := range binding.Subjects {
			grant := models.AccessGrant{
				Principal: models.AccessPrincipal{
					Kind:      subject.Kind,
					Name:      subject.Name,
					Namespace: subject.Namespace,
				},
				Role: models.AccessRole{
					Kind:      role.Kind,
					Name:      role.Name,
					Namespace: role.Namespace,
				},
				Binding: models.AccessBinding{
					Kind:      binding.Kind,
					Name:      binding.Name,
					Namespace: binding.Namespace,
				},
				Verbs: a.getMatchingVerbs(role, resourceType, ""),
				Scope: "namespace",
			}

			// Add to access map (using a placeholder key since we don't have actual resources yet)
			key := "type-level-access"
			if accessMap[key] == nil {
				accessMap[key] = &models.ResourceAccessInfo{
					DirectAccess:    []models.AccessGrant{},
					InheritedAccess: []models.AccessGrant{},
				}
			}
			if binding.Namespace != "" {
				accessMap[key].DirectAccess = append(accessMap[key].DirectAccess, grant)
			} else {
				accessMap[key].InheritedAccess = append(accessMap[key].InheritedAccess, grant)
			}
		}
	}

	// Process ClusterRoleBindings
	for _, binding := range clusterRoleBindings {
		lookupKey := makeRoleKey("ClusterRole", "", binding.RoleRef.Name)
		role, exists := roleMap[lookupKey]
		if !exists {
			continue
		}

		if !a.roleGrantsAccessToResourceType(role, resourceType) {
			continue
		}

		for _, subject := range binding.Subjects {
			if namespace != "" && subject.Kind == "ServiceAccount" && subject.Namespace != namespace {
				continue
			}

			grant := models.AccessGrant{
				Principal: models.AccessPrincipal{
					Kind:      subject.Kind,
					Name:      subject.Name,
					Namespace: subject.Namespace,
				},
				Role: models.AccessRole{
					Kind:      role.Kind,
					Name:      role.Name,
					Namespace: role.Namespace,
				},
				Binding: models.AccessBinding{
					Kind:      binding.Kind,
					Name:      binding.Name,
					Namespace: binding.Namespace,
				},
				Verbs: a.getMatchingVerbs(role, resourceType, ""),
				Scope: "cluster",
			}

			key := "type-level-access"
			if accessMap[key] == nil {
				accessMap[key] = &models.ResourceAccessInfo{
					DirectAccess:    []models.AccessGrant{},
					InheritedAccess: []models.AccessGrant{},
				}
			}
			accessMap[key].InheritedAccess = append(accessMap[key].InheritedAccess, grant)
		}
	}

	// Calculate principal counts
	for _, info := range accessMap {
		principalSet := make(map[string]bool)
		for _, grant := range info.DirectAccess {
			key := grant.Principal.Kind + ":" + grant.Principal.Name + ":" + grant.Principal.Namespace
			principalSet[key] = true
		}
		for _, grant := range info.InheritedAccess {
			key := grant.Principal.Kind + ":" + grant.Principal.Name + ":" + grant.Principal.Namespace
			principalSet[key] = true
		}
		info.PrincipalCount = len(principalSet)
	}

	return accessMap, nil
}

// getMatchingVerbs returns verbs that grant access to the resource
func (a *RBACAnalyzer) getMatchingVerbs(role models.RBACResource, resourceType string, resourceName string) []string {
	var matchingVerbs []string
	verbSet := make(map[string]bool)

	for _, rule := range role.Rules {
		// Check if rule applies to this resource type
		if !a.ruleMatchesResource(rule, resourceType) {
			continue
		}

		// Check resource names if specified
		if len(rule.ResourceNames) > 0 && resourceName != "" {
			found := false
			for _, name := range rule.ResourceNames {
				if name == resourceName || name == "*" {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Add verbs
		for _, verb := range rule.Verbs {
			if !verbSet[verb] {
				verbSet[verb] = true
				matchingVerbs = append(matchingVerbs, verb)
			}
		}
	}

	return matchingVerbs
}

// roleGrantsAccessToResourceType checks if a role grants access to a resource type
func (a *RBACAnalyzer) roleGrantsAccessToResourceType(role models.RBACResource, resourceType string) bool {
	for _, rule := range role.Rules {
		if a.ruleMatchesResource(rule, resourceType) {
			return true
		}
	}
	return false
}

// ruleMatchesResource checks if a policy rule matches a resource type
func (a *RBACAnalyzer) ruleMatchesResource(rule models.PolicyRule, resourceType string) bool {
	for _, resource := range rule.Resources {
		if resource == "*" || resource == resourceType || strings.TrimSuffix(resource, "s") == resourceType {
			return true
		}
	}
	return false
}

// makeRoleKey generates a unique key for a role
// Changed from roleKey to makeRoleKey to avoid naming conflicts
func makeRoleKey(kind, namespace, name string) string {
	if namespace == "" {
		return kind + ":" + name
	}
	return kind + ":" + namespace + ":" + name
}
