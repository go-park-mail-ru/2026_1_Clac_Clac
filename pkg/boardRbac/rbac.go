package rbac

import "errors"

var (
	ErrActionDenied = errors.New("action denied")
	ErrInvalidRole  = errors.New("invalid role")
)

type Action string

var Actions = struct {
	View   Action
	Edit   Action
	Delete Action
	Invite Action
}{
	View:   "view",
	Edit:   "edit",
	Delete: "delete",
	Invite: "invite",
}

type Role string

func (r Role) String() string {
	return string(r)
}

var Roles = struct {
	None    Role
	Viewer  Role
	Editor  Role
	Admin   Role
	Creator Role
}{
	None:    "none",
	Viewer:  "viewer",
	Editor:  "editor",
	Admin:   "admin",
	Creator: "creator",
}

var rbacPolicy = map[Role]map[Action]bool{
	Roles.None: {},
	Roles.Viewer: {
		Actions.View: true,
	},
	Roles.Editor: {
		Actions.View: true,
		Actions.Edit: true,
	},
	Roles.Admin: {
		Actions.View:   true,
		Actions.Edit:   true,
		Actions.Invite: true,
	},
	Roles.Creator: {
		Actions.View:   true,
		Actions.Edit:   true,
		Actions.Delete: true,
		Actions.Invite: true,
	},
}

func IsActionAllowed(role Role, action Action) bool {
	return rbacPolicy[role][action]
}

func ParseRole(s string) (Role, error) {
	switch Role(s) {
	case Roles.None, Roles.Viewer, Roles.Editor, Roles.Admin, Roles.Creator:
		return Role(s), nil
	default:
		return "", ErrInvalidRole
	}
}
