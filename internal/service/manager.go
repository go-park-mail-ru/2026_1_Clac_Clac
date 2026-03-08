package service

import (
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/store"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/board"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/profile"
)

type Manager struct {
	Auth    *auth.AuthUserService
	Board   *board.BoardService
	Profile *profile.ProfileService
}

func NewManager(s *store.Store) *Manager {
	return &Manager{
		Auth:    auth.NewAuthService(s.Auth, auth.HashPassword, auth.CheckPassword, auth.GenerateSessionID),
		Board:   board.NewBoardService(s.Boards),
		Profile: profile.NewProfileService(s.Profiles),
	}
}
