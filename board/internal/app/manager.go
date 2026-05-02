package app

import (
	board "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/service"
	card "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/service"
	section "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/service"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/boardRbac"
)

type Manager struct {
	Board   *board.Service
	Section *section.Service
	Card    *card.Service
}

func NewManager(store *Store) *Manager {
	permissionChecker := rbac.NewCachedService(store.PermissionChecker, store.RedisClient)

	return &Manager{
		Board:   board.NewService(store.Board, permissionChecker),
		Section: section.NewService(store.Section, permissionChecker),
		Card:    card.NewService(store.Card, permissionChecker),
	}
}
