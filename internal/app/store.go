package app

import (
	"fmt"

	auth "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/repository"
	board "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/repository"
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

func NewStore(pool *pgxpool.Pool, redisClient *redis.Client, s3Client s3.S3Client, s3AvatarsConf config.S3Avatars) *Store {
	depsAuth := auth.Deps{
		Pool:        pool,
		RedisClient: redisClient,
	}

	depsProfile := profile.Deps{
		Pool:    pool,
		Avatars: s3Client.NewBucket(s3AvatarsConf.Bucket, s3AvatarsConf.Prefix, s3.ACL.PublicRead),
	}

	depsSection := section.Deps{
		Pool: pool,
	}

	return &Store{
		Auth:         auth.NewRepository(depsAuth),
		Board:        board.NewRepository(pool),
		Profile:      profile.NewRepository(depsProfile),
		Section:      section.NewRepository(depsSection),
		s3Client:     s3Client,
		postgresPool: pool,
		redisClient:  redisClient,
	}
}
