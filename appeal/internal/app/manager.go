package app

import (
	appeal "github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/service"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/appealRbac"
)

type Manager struct {
	Appeal *appeal.Service
}

func NewManager(store *Store) *Manager {
	permissionChecker := rbac.NewService(store.PermissionChecker)

	return &Manager{
		Appeal: appeal.NewService(store.Appeal, permissionChecker),
	}
}
