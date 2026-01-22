// internal/models/rbac.go
package models

// RBACResource represents a complete RBAC configuration
type RBACResource struct {
	Kind      string            `json:"kind"`
	Name      string            `json:"name"`
	Namespace string            `json:"namespace,omitempty"`
	Rules     []PolicyRule      `json:"rules,omitempty"`
	Subjects  []Subject         `json:"subjects,omitempty"`
	RoleRef   *RoleRef          `json:"roleRef,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	CreatedAt string            `json:"createdAt"`
}

// PolicyRule represents a Kubernetes RBAC policy rule
type PolicyRule struct {
	APIGroups     []string `json:"apiGroups"`
	Resources     []string `json:"resources"`
	ResourceNames []string `json:"resourceNames,omitempty"`
	Verbs         []string `json:"verbs"`
}

// Subject represents a subject in a role binding
type Subject struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// RoleRef represents a reference to a role
type RoleRef struct {
	Kind     string `json:"kind"`
	Name     string `json:"name"`
	APIGroup string `json:"apiGroup"`
}

// RBACList represents a list of RBAC resources with pagination support
type RBACList struct {
	Items      []RBACResource `json:"items"`
	TotalCount int            `json:"totalCount"`
	Limit      int            `json:"limit,omitempty"`
	Offset     int            `json:"offset,omitempty"`
	HasMore    bool           `json:"hasMore,omitempty"`
}

// Principal represents a unique identity (User, Group, or ServiceAccount)
type Principal struct {
	Kind         string   `json:"kind"`
	Name         string   `json:"name"`
	Namespace    string   `json:"namespace,omitempty"`
	BindingCount int      `json:"bindingCount"`
	RoleCount    int      `json:"roleCount"`
	Bindings     []string `json:"bindings,omitempty"`
	Roles        []string `json:"roles,omitempty"`
}

// PrincipalList represents a list of principals with pagination support
type PrincipalList struct {
	Items      []Principal `json:"items"`
	TotalCount int         `json:"totalCount"`
	Limit      int         `json:"limit,omitempty"`
	Offset     int         `json:"offset,omitempty"`
	HasMore    bool        `json:"hasMore,omitempty"`
}

// ResourceCounts contains counts of all RBAC resources
type ResourceCounts struct {
	Roles               int `json:"roles"`
	ClusterRoles        int `json:"clusterRoles"`
	RoleBindings        int `json:"roleBindings"`
	ClusterRoleBindings int `json:"clusterRoleBindings"`
	Principals          int `json:"principals"`
	Resources           int `json:"resources"` // Add this field
}

// RelationshipResponse contains complete relationship data for visualization
type RelationshipResponse struct {
	Role            *RBACResource  `json:"role,omitempty"`
	Binding         *RBACResource  `json:"binding,omitempty"`
	RelatedBindings []RBACResource `json:"relatedBindings"`
	RelatedRoles    []RBACResource `json:"relatedRoles"`
}

// NamespaceList represents a list of namespaces
type NamespaceList struct {
	Namespaces []string `json:"namespaces"`
}

// ClusterInfo represents cluster information
type ClusterInfo struct {
	Version    string   `json:"version"`
	Namespaces []string `json:"namespaces"`
	NodeCount  int      `json:"nodeCount"`
}

// ApplyRequest represents a request to apply RBAC configuration
type ApplyRequest struct {
	YAML string `json:"yaml"`
}

// ApplyResponse represents the response from applying RBAC
type ApplyResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Applied []string `json:"applied,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}
