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
			name:     "None role - deny everything",
			role:     Roles.None,
			action:   Actions.View,
			expected: false,
		},
		{
			name:     "Viewer role - allow view",
			role:     Roles.Viewer,
			action:   Actions.View,
			expected: true,
		},
		{
			name:     "Viewer role - deny edit",
			role:     Roles.Viewer,
			action:   Actions.Edit,
			expected: false,
		},
		{
			name:     "Editor role - allow edit",
			role:     Roles.Editor,
			action:   Actions.Edit,
			expected: true,
		},
		{
			name:     "Editor role - deny invite",
			role:     Roles.Editor,
			action:   Actions.Invite,
			expected: false,
		},
		{
			name:     "Admin role - allow invite",
			role:     Roles.Admin,
			action:   Actions.Invite,
			expected: true,
		},
		{
			name:     "Admin role - deny delete",
			role:     Roles.Admin,
			action:   Actions.Delete,
			expected: false,
		},
		{
			name:     "Creator role - allow delete",
			role:     Roles.Creator,
			action:   Actions.Delete,
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsActionAllowed(test.role, test.action)
			assert.Equal(t, test.expected, result)
		})
	}
}
