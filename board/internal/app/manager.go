package app

import (
	board "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/service"
	card "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/service"
	section "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/service"
)

type Manager struct {
	Board   *board.Service
	Section *section.Service
	Card    *card.Service
}

func NewManager(store *Store) *Manager {
	return &Manager{
		Board:   board.NewService(store.Board),
		Section: section.NewService(store.Section),
		Card:    card.NewService(store.Card),
	}
}
