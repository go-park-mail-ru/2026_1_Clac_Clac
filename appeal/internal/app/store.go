package app

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/config"
	appeal "github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/repository"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/appealRbac"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/postgres"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/s3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type Store struct {
	Appeal            *appeal.Repository
	PermissionChecker rbac.Repository

	PostgresPool *pgxpool.Pool
	S3Client     s3.S3Client
}

func NewStore(logger *zerolog.Logger, conf config.Config) (*Store, error) {
	store := &Store{}

	if err := store.setupPostgresPool(&conf.Database, logger); err != nil {
		return nil, fmt.Errorf("store.setupPostgresPool: %w", err)
	}

	if err := store.setupS3(&conf.S3); err != nil {
		return nil, fmt.Errorf("store.setupS3: %w", err)
	}

	store.Appeal = appeal.NewRepository(
		store.PostgresPool,
		store.S3Client.NewBucket(
			conf.S3.AppealAttachmentBucket,
			conf.S3.AppealAttachmentPrefix,
			s3.ACL.PublicRead,
		),
	)
	store.PermissionChecker = rbac.NewRepository(store.PostgresPool)

	return store, nil
}

func (s *Store) Close() error {
	s.PostgresPool.Close()
	return nil
}

func (s *Store) setupPostgresPool(conf *postgres.Config, logger *zerolog.Logger) error {
	DSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		conf.User,
		conf.Password,
		conf.Host,
		conf.Port,
		conf.Name,
	)

	pool, err := postgres.NewPoolPostgres(DSN, conf, logger)
	if err != nil {
		return fmt.Errorf("postgres.NewPoolPostgres: %w", err)
	}

	s.PostgresPool = pool
	return nil
}

func (s *Store) setupS3(conf *config.S3) error {
	connectTimeout, err := strconv.ParseInt(conf.ConnectTimeout, 10, 64)
	if err != nil {
		return fmt.Errorf("strconv.ParseInt: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(connectTimeout)*time.Second)
	defer cancel()

	client, err := s3.NewAWSClient(
		ctx,
		conf.Region,
		conf.Endpoint,
		conf.AccessKey,
		conf.SecretKey,
	)
	if err != nil {
		return fmt.Errorf("s3.NewAWSClient: %w", err)
	}

	s.S3Client = client
	return nil
}
