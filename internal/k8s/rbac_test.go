// backend/internal/k8s/rbac_test.go
package k8s

import (
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConvertRoleToModel(t *testing.T) {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-role",
			Namespace: "default",
			Labels: map[string]string{
				"app": "test",
			},
			CreationTimestamp: metav1.Now(),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "list"},
			},
		},
	}

	model := convertRoleToModel(role)

	if model.Kind != "Role" {
		t.Errorf("expected Kind 'Role', got %s", model.Kind)
	}

	if model.Name != "test-role" {
		t.Errorf("expected Name 'test-role', got %s", model.Name)
	}

	if model.Namespace != "default" {
		t.Errorf("expected Namespace 'default', got %s", model.Namespace)
	}

	if len(model.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(model.Rules))
	}

	if model.Labels["app"] != "test" {
		t.Errorf("expected label app=test, got %s", model.Labels["app"])
	}
}

func TestConvertClusterRoleToModel(t *testing.T) {
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-cluster-role",
			Labels: map[string]string{
				"type": "system",
			},
			CreationTimestamp: metav1.Now(),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}

	model := convertClusterRoleToModel(clusterRole)

	if model.Kind != "ClusterRole" {
		t.Errorf("expected Kind 'ClusterRole', got %s", model.Kind)
	}

	if model.Name != "test-cluster-role" {
		t.Errorf("expected Name 'test-cluster-role', got %s", model.Name)
	}

	if model.Namespace != "" {
		t.Errorf("expected empty Namespace for ClusterRole, got %s", model.Namespace)
	}

	if len(model.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(model.Rules))
	}
}

func TestConvertRoleBindingToModel(t *testing.T) {
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-binding",
			Namespace:         "default",
			CreationTimestamp: metav1.Now(),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "test-sa",
				Namespace: "default",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     "test-role",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	model := convertRoleBindingToModel(roleBinding)

	if model.Kind != "RoleBinding" {
		t.Errorf("expected Kind 'RoleBinding', got %s", model.Kind)
	}

	if len(model.Subjects) != 1 {
		t.Errorf("expected 1 subject, got %d", len(model.Subjects))
	}

	if model.Subjects[0].Kind != "ServiceAccount" {
		t.Errorf("expected subject kind 'ServiceAccount', got %s", model.Subjects[0].Kind)
	}

	if model.RoleRef == nil {
		t.Fatal("expected RoleRef to be set")
	}

	if model.RoleRef.Name != "test-role" {
		t.Errorf("expected RoleRef name 'test-role', got %s", model.RoleRef.Name)
	}
}

func TestConvertClusterRoleBindingToModel(t *testing.T) {
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-cluster-binding",
			CreationTimestamp: metav1.Now(),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "User",
				Name: "test-user",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	model := convertClusterRoleBindingToModel(clusterRoleBinding)

	if model.Kind != "ClusterRoleBinding" {
		t.Errorf("expected Kind 'ClusterRoleBinding', got %s", model.Kind)
	}

	if model.Namespace != "" {
		t.Errorf("expected empty Namespace for ClusterRoleBinding, got %s", model.Namespace)
	}

	if model.RoleRef.Kind != "ClusterRole" {
		t.Errorf("expected RoleRef kind 'ClusterRole', got %s", model.RoleRef.Kind)
	}
}

func TestConvertPolicyRules(t *testing.T) {
	rules := []rbacv1.PolicyRule{
		{
			APIGroups:     []string{"", "apps"},
			Resources:     []string{"pods", "deployments"},
			ResourceNames: []string{"my-pod"},
			Verbs:         []string{"get", "list", "create"},
		},
	}

	modelRules := convertPolicyRules(rules)

	if len(modelRules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(modelRules))
	}

	rule := modelRules[0]

	if len(rule.APIGroups) != 2 {
		t.Errorf("expected 2 API groups, got %d", len(rule.APIGroups))
	}

	if len(rule.Resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(rule.Resources))
	}

	if len(rule.ResourceNames) != 1 {
		t.Errorf("expected 1 resource name, got %d", len(rule.ResourceNames))
	}

	if len(rule.Verbs) != 3 {
		t.Errorf("expected 3 verbs, got %d", len(rule.Verbs))
	}
}

func TestConvertSubjects(t *testing.T) {
	subjects := []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      "sa1",
			Namespace: "default",
		},
		{
			Kind: "User",
			Name: "user1",
		},
		{
			Kind: "Group",
			Name: "group1",
		},
	}

	modelSubjects := convertSubjects(subjects)

	if len(modelSubjects) != 3 {
		t.Fatalf("expected 3 subjects, got %d", len(modelSubjects))
	}

	// Check ServiceAccount
	if modelSubjects[0].Kind != "ServiceAccount" {
		t.Errorf("expected kind 'ServiceAccount', got %s", modelSubjects[0].Kind)
	}
	if modelSubjects[0].Namespace != "default" {
		t.Errorf("expected namespace 'default', got %s", modelSubjects[0].Namespace)
	}

	// Check User (no namespace)
	if modelSubjects[1].Kind != "User" {
		t.Errorf("expected kind 'User', got %s", modelSubjects[1].Kind)
	}
	if modelSubjects[1].Namespace != "" {
		t.Errorf("expected empty namespace for User, got %s", modelSubjects[1].Namespace)
	}

	// Check Group
	if modelSubjects[2].Kind != "Group" {
		t.Errorf("expected kind 'Group', got %s", modelSubjects[2].Kind)
	}
}

func TestConvertRoleRef(t *testing.T) {
	roleRef := rbacv1.RoleRef{
		Kind:     "Role",
		Name:     "test-role",
		APIGroup: "rbac.authorization.k8s.io",
	}

	modelRoleRef := convertRoleRef(roleRef)

	if modelRoleRef == nil {
		t.Fatal("expected non-nil RoleRef")
	}

	if modelRoleRef.Kind != "Role" {
		t.Errorf("expected Kind 'Role', got %s", modelRoleRef.Kind)
	}

	if modelRoleRef.Name != "test-role" {
		t.Errorf("expected Name 'test-role', got %s", modelRoleRef.Name)
	}

	if modelRoleRef.APIGroup != "rbac.authorization.k8s.io" {
		t.Errorf("expected APIGroup 'rbac.authorization.k8s.io', got %s", modelRoleRef.APIGroup)
	}
}
