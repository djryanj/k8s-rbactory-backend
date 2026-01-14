// backend/internal/k8s/rbac.go
package k8s

import (
	"context"

	"github.com/djryanj/k8s-rbactory-backend/internal/models"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListRoles returns all roles in a namespace or all namespaces
func (c *Client) ListRoles(ctx context.Context, namespace string) ([]models.RBACResource, error) {
	var roles []models.RBACResource

	if namespace == "" {
		// List all namespaces
		namespaces, err := c.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		for _, ns := range namespaces.Items {
			nsRoles, err := c.listRolesInNamespace(ctx, ns.Name)
			if err != nil {
				continue // Skip namespaces we can't access
			}
			roles = append(roles, nsRoles...)
		}
	} else {
		return c.listRolesInNamespace(ctx, namespace)
	}

	return roles, nil
}

func (c *Client) listRolesInNamespace(ctx context.Context, namespace string) ([]models.RBACResource, error) {
	roleList, err := c.Clientset.RbacV1().Roles(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var roles []models.RBACResource
	for _, role := range roleList.Items {
		roles = append(roles, convertRoleToModel(&role))
	}

	return roles, nil
}

// ListClusterRoles returns all cluster roles
func (c *Client) ListClusterRoles(ctx context.Context) ([]models.RBACResource, error) {
	clusterRoleList, err := c.Clientset.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var clusterRoles []models.RBACResource
	for _, cr := range clusterRoleList.Items {
		clusterRoles = append(clusterRoles, convertClusterRoleToModel(&cr))
	}

	return clusterRoles, nil
}

// ListRoleBindings returns all role bindings in a namespace or all namespaces
func (c *Client) ListRoleBindings(ctx context.Context, namespace string) ([]models.RBACResource, error) {
	var bindings []models.RBACResource

	if namespace == "" {
		namespaces, err := c.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		for _, ns := range namespaces.Items {
			nsBindings, err := c.listRoleBindingsInNamespace(ctx, ns.Name)
			if err != nil {
				continue
			}
			bindings = append(bindings, nsBindings...)
		}
	} else {
		return c.listRoleBindingsInNamespace(ctx, namespace)
	}

	return bindings, nil
}

func (c *Client) listRoleBindingsInNamespace(ctx context.Context, namespace string) ([]models.RBACResource, error) {
	bindingList, err := c.Clientset.RbacV1().RoleBindings(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var bindings []models.RBACResource
	for _, binding := range bindingList.Items {
		bindings = append(bindings, convertRoleBindingToModel(&binding))
	}

	return bindings, nil
}

// ListClusterRoleBindings returns all cluster role bindings
func (c *Client) ListClusterRoleBindings(ctx context.Context) ([]models.RBACResource, error) {
	bindingList, err := c.Clientset.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var bindings []models.RBACResource
	for _, binding := range bindingList.Items {
		bindings = append(bindings, convertClusterRoleBindingToModel(&binding))
	}

	return bindings, nil
}

// GetRole returns a specific role
func (c *Client) GetRole(ctx context.Context, namespace, name string) (*models.RBACResource, error) {
	role, err := c.Clientset.RbacV1().Roles(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	resource := convertRoleToModel(role)
	return &resource, nil
}

// GetClusterRole returns a specific cluster role
func (c *Client) GetClusterRole(ctx context.Context, name string) (*models.RBACResource, error) {
	clusterRole, err := c.Clientset.RbacV1().ClusterRoles().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	resource := convertClusterRoleToModel(clusterRole)
	return &resource, nil
}

// ListNamespaces returns all namespaces
func (c *Client) ListNamespaces(ctx context.Context) ([]string, error) {
	namespaces, err := c.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var names []string
	for _, ns := range namespaces.Items {
		names = append(names, ns.Name)
	}

	return names, nil
}

// Converter functions
func convertRoleToModel(role *rbacv1.Role) models.RBACResource {
	return models.RBACResource{
		Kind:      "Role",
		Name:      role.Name,
		Namespace: role.Namespace,
		Rules:     convertPolicyRules(role.Rules),
		Labels:    role.Labels,
		CreatedAt: role.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}
}

func convertClusterRoleToModel(cr *rbacv1.ClusterRole) models.RBACResource {
	return models.RBACResource{
		Kind:      "ClusterRole",
		Name:      cr.Name,
		Rules:     convertPolicyRules(cr.Rules),
		Labels:    cr.Labels,
		CreatedAt: cr.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}
}

func convertRoleBindingToModel(rb *rbacv1.RoleBinding) models.RBACResource {
	return models.RBACResource{
		Kind:      "RoleBinding",
		Name:      rb.Name,
		Namespace: rb.Namespace,
		Subjects:  convertSubjects(rb.Subjects),
		RoleRef:   convertRoleRef(rb.RoleRef),
		Labels:    rb.Labels,
		CreatedAt: rb.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}
}

func convertClusterRoleBindingToModel(crb *rbacv1.ClusterRoleBinding) models.RBACResource {
	return models.RBACResource{
		Kind:      "ClusterRoleBinding",
		Name:      crb.Name,
		Subjects:  convertSubjects(crb.Subjects),
		RoleRef:   convertRoleRef(crb.RoleRef),
		Labels:    crb.Labels,
		CreatedAt: crb.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}
}

func convertPolicyRules(rules []rbacv1.PolicyRule) []models.PolicyRule {
	var modelRules []models.PolicyRule
	for _, rule := range rules {
		modelRules = append(modelRules, models.PolicyRule{
			APIGroups:     rule.APIGroups,
			Resources:     rule.Resources,
			ResourceNames: rule.ResourceNames,
			Verbs:         rule.Verbs,
		})
	}
	return modelRules
}

func convertSubjects(subjects []rbacv1.Subject) []models.Subject {
	var modelSubjects []models.Subject
	for _, subject := range subjects {
		modelSubjects = append(modelSubjects, models.Subject{
			Kind:      subject.Kind,
			Name:      subject.Name,
			Namespace: subject.Namespace,
		})
	}
	return modelSubjects
}

func convertRoleRef(roleRef rbacv1.RoleRef) *models.RoleRef {
	return &models.RoleRef{
		Kind:     roleRef.Kind,
		Name:     roleRef.Name,
		APIGroup: roleRef.APIGroup,
	}
}
