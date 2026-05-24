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

func TestParseRole(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        Role
		wantErr     bool
		expectedErr error
	}{
		{
			name:    "valid viewer",
			input:   "viewer",
			want:    Roles.Viewer,
			wantErr: false,
		},
		{
			name:    "valid editor",
			input:   "editor",
			want:    Roles.Editor,
			wantErr: false,
		},
		{
			name:    "valid admin",
			input:   "admin",
			want:    Roles.Admin,
			wantErr: false,
		},
		{
			name:    "valid creator",
			input:   "creator",
			want:    Roles.Creator,
			wantErr: false,
		},
		{
			name:    "valid none",
			input:   "none",
			want:    Roles.None,
			wantErr: false,
		},
		{
			name:        "invalid role",
			input:       "superadmin",
			want:        "",
			wantErr:     true,
			expectedErr: ErrInvalidRole,
		},
		{
			name:        "empty string",
			input:       "",
			want:        "",
			wantErr:     true,
			expectedErr: ErrInvalidRole,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParseRole(test.input)
			if test.wantErr {
				assert.Error(t, err)
				if test.expectedErr != nil {
					assert.ErrorIs(t, err, test.expectedErr)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestRoleString(t *testing.T) {
	tests := []struct {
		name string
		role Role
		want string
	}{
		{"viewer string", Roles.Viewer, "viewer"},
		{"editor string", Roles.Editor, "editor"},
		{"admin string", Roles.Admin, "admin"},
		{"creator string", Roles.Creator, "creator"},
		{"none string", Roles.None, "none"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, test.role.String())
		})
	}
}
