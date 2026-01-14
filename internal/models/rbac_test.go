// backend/internal/models/rbac_test.go
package models

import (
	"encoding/json"
	"testing"
)

func TestRBACResource_JSONSerialization(t *testing.T) {
	resource := RBACResource{
		Kind:      "Role",
		Name:      "test-role",
		Namespace: "default",
		Rules: []PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "list"},
			},
		},
		Labels: map[string]string{
			"app": "test",
		},
		CreatedAt: "2024-01-01T00:00:00Z",
	}

	// Marshal to JSON
	data, err := json.Marshal(resource)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal back
	var decoded RBACResource
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify
	if decoded.Kind != resource.Kind {
		t.Errorf("expected Kind %s, got %s", resource.Kind, decoded.Kind)
	}

	if decoded.Name != resource.Name {
		t.Errorf("expected Name %s, got %s", resource.Name, decoded.Name)
	}

	if len(decoded.Rules) != len(resource.Rules) {
		t.Errorf("expected %d rules, got %d", len(resource.Rules), len(decoded.Rules))
	}
}

func TestRBACList_Pagination(t *testing.T) {
	list := RBACList{
		Items: []RBACResource{
			{Name: "role1"},
			{Name: "role2"},
		},
		TotalCount: 10,
		Limit:      2,
		Offset:     0,
		HasMore:    true,
	}

	data, err := json.Marshal(list)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded RBACList
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.TotalCount != 10 {
		t.Errorf("expected TotalCount 10, got %d", decoded.TotalCount)
	}

	if !decoded.HasMore {
		t.Error("expected HasMore to be true")
	}
}

func TestPrincipal_JSONSerialization(t *testing.T) {
	principal := Principal{
		Kind:         "ServiceAccount",
		Name:         "test-sa",
		Namespace:    "default",
		BindingCount: 2,
		RoleCount:    1,
		Bindings:     []string{"binding1", "binding2"},
		Roles:        []string{"role1"},
	}

	data, err := json.Marshal(principal)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded Principal
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.BindingCount != 2 {
		t.Errorf("expected BindingCount 2, got %d", decoded.BindingCount)
	}

	if len(decoded.Bindings) != 2 {
		t.Errorf("expected 2 bindings, got %d", len(decoded.Bindings))
	}
}

func TestResourceCounts_JSONSerialization(t *testing.T) {
	counts := ResourceCounts{
		Roles:               10,
		ClusterRoles:        5,
		RoleBindings:        20,
		ClusterRoleBindings: 8,
		Principals:          15,
	}

	data, err := json.Marshal(counts)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ResourceCounts
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Roles != 10 {
		t.Errorf("expected Roles 10, got %d", decoded.Roles)
	}

	if decoded.Principals != 15 {
		t.Errorf("expected Principals 15, got %d", decoded.Principals)
	}
}
