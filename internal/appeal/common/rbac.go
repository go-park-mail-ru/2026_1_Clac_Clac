package common

// type Action string

// var Actions = struct {
// 	ReadOwn   Action
// 	ReadOpen   Action
// 	ReadStats Action
// 	Invite Action
// }{
// 	View:   "view",
// 	Edit:   "edit",
// 	Delete: "delete",
// 	Invite: "invite",
// }

type Role string

var Roles = struct {
	None    Role
	Support Role
	Admin   Role
}{
	None:    "none",
	Support: "support",
	Admin:   "admin",
}

// var rbacPolicy = map[Role]map[Action]bool{
// 	Roles.None: {},
// 	Roles.Viewer: {
// 		Actions.View: true,
// 	},
// 	Roles.Editor: {
// 		Actions.View: true,
// 		Actions.Edit: true,
// 	},
// 	Roles.Admin: {
// 		Actions.View:   true,
// 		Actions.Edit:   true,
// 		Actions.Invite: true,
// 	},
// 	Roles.Creator: {
// 		Actions.View:   true,
// 		Actions.Edit:   true,
// 		Actions.Delete: true,
// 		Actions.Invite: true,
// 	},
// }

// func IsActionAllowed(role Role, action Action) bool {
// 	return rbacPolicy[role][action]
// }
