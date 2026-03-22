package app

import (
	auth "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/repository"
	board "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/repository"
	profile "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Для хранения всех репозиториев и их удобной инициализации
type Store struct {
	Auth     *auth.Repository
	Boards   *board.Repository
	Profiles *profile.Repository
}

func NewStore(pool *pgxpool.Pool, client *redis.Client) *Store {
	return &Store{
		Auth:     auth.NewRepository(pool, client),
		Boards:   board.NewRepository(pool),
		Profiles: profile.NewRepository(pool),
	}
}
