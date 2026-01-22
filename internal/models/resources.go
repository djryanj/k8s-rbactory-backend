// internal/models/resources.go
package models

// KubernetesResource represents a Kubernetes resource with RBAC access information
type KubernetesResource struct {
	Kind        string              `json:"kind"`
	Name        string              `json:"name"`
	Namespace   string              `json:"namespace,omitempty"`
	APIVersion  string              `json:"apiVersion"`
	CreatedAt   string              `json:"createdAt"`
	Labels      map[string]string   `json:"labels,omitempty"`
	Annotations map[string]string   `json:"annotations,omitempty"`
	AccessInfo  *ResourceAccessInfo `json:"accessInfo,omitempty"`
}

// ResourceAccessInfo contains information about who has access to a resource
type ResourceAccessInfo struct {
	DirectAccess    []AccessGrant `json:"directAccess"`
	InheritedAccess []AccessGrant `json:"inheritedAccess"`
	PrincipalCount  int           `json:"principalCount"`
}

// AccessGrant represents a single access grant to a resource
type AccessGrant struct {
	Principal AccessPrincipal `json:"principal"`
	Role      AccessRole      `json:"role"`
	Binding   AccessBinding   `json:"binding"`
	Verbs     []string        `json:"verbs"`
	Scope     string          `json:"scope"` // "namespace" or "cluster"
}

// AccessPrincipal represents the principal (user/group/serviceaccount) with access
type AccessPrincipal struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// AccessRole represents the role granting access
type AccessRole struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// AccessBinding represents the binding connecting principal to role
type AccessBinding struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// KubernetesResourceList represents a paginated list of Kubernetes resources
type KubernetesResourceList struct {
	Items      []KubernetesResource `json:"items"`
	TotalCount int                  `json:"totalCount"`
	Limit      int                  `json:"limit,omitempty"`
	Offset     int                  `json:"offset,omitempty"`
	HasMore    bool                 `json:"hasMore,omitempty"`
}

// ResourceAccessDetail contains detailed access information for a specific resource
type ResourceAccessDetail struct {
	Resource     KubernetesResource `json:"resource"`
	AccessGrants []AccessGrant      `json:"accessGrants"`
	Summary      AccessSummary      `json:"summary"`
}

// AccessSummary provides aggregate statistics about resource access
type AccessSummary struct {
	TotalPrincipals int            `json:"totalPrincipals"`
	TotalRoles      int            `json:"totalRoles"`
	TotalBindings   int            `json:"totalBindings"`
	VerbCounts      map[string]int `json:"verbCounts"`
}

// ResourceTypeList represents available resource types
type ResourceTypeList struct {
	ResourceTypes []string `json:"resourceTypes"`
}

// Common Kubernetes resource types
var CommonResourceTypes = []string{
	"secrets",
	"configmaps",
	"pods",
	"services",
	"deployments",
	"statefulsets",
	"daemonsets",
	"replicasets",
	"persistentvolumeclaims",
	"persistentvolumes",
	"serviceaccounts",
	"ingresses",
	"networkpolicies",
	"jobs",
	"cronjobs",
}
