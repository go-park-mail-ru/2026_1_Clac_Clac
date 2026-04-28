package common

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
