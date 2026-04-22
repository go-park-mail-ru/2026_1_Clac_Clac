package app

import (
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/s3"
	auth "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/auth/repository"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/config"
	profile "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/profile/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Для хранения всех репозиториев и их удобной инициализации
type Store struct {
	Auth    *auth.Repository
	Profile *profile.Repository

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
	avatars := s3Client.NewBucket(conf.S3.AvatarsBucket, conf.S3.AvatarsPrefix, s3.ACL.PublicRead)

	return &Store{
		Auth:    auth.NewRepository(pool, redisClient),
		Profile: profile.NewRepository(pool, avatars),

		s3Client:     s3Client,
		postgresPool: pool,
		redisClient:  redisClient,
	}
}
