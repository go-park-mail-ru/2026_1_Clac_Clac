package common

import rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/boardRbac"

type InviteStatus string

func (i InviteStatus) String() string {
	return string(i)
}

var InviteStatuses = struct {
	Active InviteStatus
	Closed InviteStatus
}{
	Active: "active",
	Closed: "closed",
}

var InviteAssignableRoles = map[rbac.Role]bool{
	rbac.Roles.Viewer: true,
	rbac.Roles.Editor: true,
	rbac.Roles.Admin:  true,
}
