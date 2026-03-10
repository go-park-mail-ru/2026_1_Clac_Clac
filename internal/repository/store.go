package repository

import (
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/auth"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/board"
	dbConnection "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/db_connection"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/profile"
)

// Для хранения всех репозиториев и их удобной инициализации
type Store struct {
	Auth     *auth.AuthRepository
	Boards   *board.BoardRepository
	Profiles *profile.ProfileRepository
}

func NewStore(db *dbConnection.MapDatabases) *Store {
	return &Store{
		Auth:     auth.NewAuthRepository(db),
		Boards:   board.NewBoardRepository(db),
		Profiles: profile.NewProfileRepository(db),
	}
}
