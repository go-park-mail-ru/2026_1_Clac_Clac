package app

import (
	"fmt"

	auth "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/repository"
	board "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/repository"
	card "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/card/repository"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	profile "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/repository"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/s3"
	section "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/section/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Для хранения всех репозиториев и их удобной инициализации
type Store struct {
	Auth    *auth.Repository
	Board   *board.Repository
	Profile *profile.Repository
	Section *section.Repository
	Card    *card.Repository

	s3Client     s3.S3Client
	postgresPool *pgxpool.Pool
	redisClient  *redis.Client
}

func (s *Store) Close() error {
	s.postgresPool.Close()

	err := s.redisClient.Close()
	if err != nil {
		return fmt.Errorf("cannot close redis client: %w", err)
	}

	return nil
}

func NewStore(pool *pgxpool.Pool, redisClient *redis.Client, s3Client s3.S3Client, conf config.Config) *Store {
	backgrounds := s3Client.NewBucket(conf.S3.BoardsBackgroundsBucket, conf.S3.BoardsBackgroundsPrefix, s3.ACL.PublicRead)
	avatars := s3Client.NewBucket(conf.S3.AvatarsBucket, conf.S3.AvatarsPrefix, s3.ACL.PublicRead)

	boardConfig := board.Config{
		CreateBoardDefaultUserRole: conf.Board.Repository.CreateBoardDefaultUserRole,
	}

	return &Store{
		Auth:    auth.NewRepository(pool, redisClient),
		Board:   board.NewRepository(pool, backgrounds, boardConfig),
		Profile: profile.NewRepository(pool, avatars),
		Section: section.NewRepository(pool),
		Card:    card.NewRepository(pool),

		s3Client:     s3Client,
		postgresPool: pool,
		redisClient:  redisClient,
	}
}
