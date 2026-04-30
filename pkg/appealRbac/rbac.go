package rbac

import "errors"

var (
	ErrActionDenied = errors.New("action denied")
)

type Action string

var Actions = struct {
	Create       Action
	View         Action
	Edit         Action
	Delete       Action
	ViewStats    Action
	ChangeStatus Action
}{
	Create:       "create",
	View:         "view",
	Edit:         "edit",
	Delete:       "delete",
	ViewStats:    "view_stats",
	ChangeStatus: "change_status",
}

type Role string

var Roles = struct {
	User    Role
	Support Role
	Admin   Role
}{
	User:    "user",
	Support: "support",
	Admin:   "admin",
}

var rbacPolicy = map[Role]map[Action]bool{
	Roles.User: {
		Actions.Create: true,
		Actions.View:   true,
		Actions.Edit:   true,
		Actions.Delete: true,
	},
	Roles.Support: {
		Actions.View:         true,
		Actions.ChangeStatus: true,
	},
	Roles.Admin: {
		Actions.View:         true,
		Actions.ChangeStatus: true,
		Actions.ViewStats:    true,
	},
}

func IsActionAllowed(role Role, action Action) bool {
	return rbacPolicy[role][action]
}
