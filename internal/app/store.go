package app

import (
	auth "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/repository"
	board "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/repository"
	dbConnection "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/db"
	profile "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/repository"
)

// Для хранения всех репозиториев и их удобной инициализации
type Store struct {
	Auth     *auth.Repository
	Boards   *board.Repository
	Profiles *profile.Repository
}

func NewStore(db *dbConnection.MapDatabases) *Store {
	return &Store{
		Auth:     auth.NewRepository(db),
		Boards:   board.NewRepository(db),
		Profiles: profile.NewProfileRepository(db),
	}
}
