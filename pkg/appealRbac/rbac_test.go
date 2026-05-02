package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsActionAllowed(t *testing.T) {
	tests := []struct {
		name     string
		role     Role
		action   Action
		expected bool
	}{
		{
			name:     "User role - allow create",
			role:     Roles.User,
			action:   Actions.Create,
			expected: true,
		},
		{
			name:     "User role - allow view",
			role:     Roles.User,
			action:   Actions.View,
			expected: true,
		},
		{
			name:     "User role - allow edit",
			role:     Roles.User,
			action:   Actions.Edit,
			expected: true,
		},
		{
			name:     "User role - allow delete",
			role:     Roles.User,
			action:   Actions.Delete,
			expected: true,
		},
		{
			name:     "User role - deny view stats",
			role:     Roles.User,
			action:   Actions.ViewStats,
			expected: false,
		},
		{
			name:     "User role - deny change status",
			role:     Roles.User,
			action:   Actions.ChangeStatus,
			expected: false,
		},
		{
			name:     "Support role - allow view",
			role:     Roles.Support,
			action:   Actions.View,
			expected: true,
		},
		{
			name:     "Support role - allow change status",
			role:     Roles.Support,
			action:   Actions.ChangeStatus,
			expected: true,
		},
		{
			name:     "Support role - deny create",
			role:     Roles.Support,
			action:   Actions.Create,
			expected: false,
		},
		{
			name:     "Support role - deny delete",
			role:     Roles.Support,
			action:   Actions.Delete,
			expected: false,
		},
		{
			name:     "Support role - deny view stats",
			role:     Roles.Support,
			action:   Actions.ViewStats,
			expected: false,
		},
		{
			name:     "Admin role - allow view",
			role:     Roles.Admin,
			action:   Actions.View,
			expected: true,
		},
		{
			name:     "Admin role - allow change status",
			role:     Roles.Admin,
			action:   Actions.ChangeStatus,
			expected: true,
		},
		{
			name:     "Admin role - allow view stats",
			role:     Roles.Admin,
			action:   Actions.ViewStats,
			expected: true,
		},
		{
			name:     "Admin role - deny create",
			role:     Roles.Admin,
			action:   Actions.Create,
			expected: false,
		},
		{
			name:     "Admin role - deny delete",
			role:     Roles.Admin,
			action:   Actions.Delete,
			expected: false,
		},
		{
			name:     "Unknown role - deny everything",
			role:     Role("unknown"),
			action:   Actions.View,
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsActionAllowed(test.role, test.action)
			assert.Equal(t, test.expected, result)
		})
	}
}
