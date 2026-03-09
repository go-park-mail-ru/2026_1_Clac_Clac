package service

import (
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/board"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/profile"
)

type Manager struct {
	Auth    *auth.AuthUserService
	Board   *board.BoardService
	Profile *profile.ProfileService
}

func NewManager(s *repository.Store, sender auth.SenderLetters) *Manager {
	return &Manager{
		Auth:    auth.NewAuthService(s.Auth, sender, auth.HashPassword, auth.CheckPassword, auth.GenerateSessionID, auth.GeneratorCode),
		Board:   board.NewBoardService(s.Boards),
		Profile: profile.NewProfileService(s.Profiles),
	}
}
