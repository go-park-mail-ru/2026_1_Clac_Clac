package common

import (
	"errors"

	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/appealRbac"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/appeal/v1"
)

var (
	ErrUnexpectedRole = errors.New("unexpected role")
)

func ParseProtoRole(pbRole pb.Role) (rbac.Role, error) {
	switch pbRole {
	case pb.Role_ROLE_USER:
		return rbac.Roles.User, nil
	case pb.Role_ROLE_SUPPORT:
		return rbac.Roles.Support, nil
	case pb.Role_ROLE_ADMIN:
		return rbac.Roles.Admin, nil
	}

	return "", ErrUnexpectedRole
}

func ToProtoRole(role rbac.Role) pb.Role {
	switch role {
	case rbac.Roles.User:
		return pb.Role_ROLE_USER
	case rbac.Roles.Support:
		return pb.Role_ROLE_SUPPORT
	case rbac.Roles.Admin:
		return pb.Role_ROLE_ADMIN
	}

	return pb.Role_ROLE_UNSPECIFIED
}
