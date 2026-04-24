package app

import (
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/s3"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/config"
	userRepo "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/user/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	User *userRepo.Repository

	s3Client     s3.S3Client
	postgresPool *pgxpool.Pool
}

func (s *Store) Close() error {
	s.postgresPool.Close()
	return nil
}

func NewStore(pool *pgxpool.Pool, s3Client s3.S3Client, conf config.Config) *Store {
	avatars := s3Client.NewBucket(conf.S3.AvatarsBucket, conf.S3.AvatarsPrefix, s3.ACL.PublicRead)

	return &Store{
		User: userRepo.NewRepository(pool, avatars),

		s3Client:     s3Client,
		postgresPool: pool,
	}
}
