package app

import (
	"context"
	"fmt"
	"strconv"
	"time"

	board "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/repository"
	card "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/repository"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/config"
	section "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/repository"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/boardRbac"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/postgres"
	pkgredis "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/redis"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/s3"
	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Store struct {
	Board             *board.Repository
	Section           *section.Repository
	Card              *card.Repository
	PermissionChecker rbac.Repository

	PostgresPool *pgxpool.Pool
	S3Client     s3.S3Client
	RedisClient  *goredis.Client
}

func NewStore(logger *zerolog.Logger, conf config.Config) (*Store, error) {
	store := &Store{}

	if err := store.setupPostgresPool(conf.Database.ToPkg(), logger); err != nil {
		return nil, fmt.Errorf("store.setupPostgresPool: %w", err)
	}

	if err := store.setupRedis(conf.Redis.ToPkg(), logger); err != nil {
		return nil, fmt.Errorf("store.setupRedis: %w", err)
	}

	if err := store.setupS3(&conf.S3); err != nil {
		return nil, fmt.Errorf("store.setupS3: %w", err)
	}

	store.Board = board.NewRepository(
		store.PostgresPool,
		store.S3Client.NewBucket(
			conf.S3.BoardsBackgroundsBucket,
			conf.S3.BoardsBackgroundsPrefix,
			s3.ACL.PublicRead,
		),
		board.Config{
			CreateBoardDefaultUserRole: conf.Board.Repository.CreateBoardDefaultUserRole,
		},
	)
	store.Section = section.NewRepository(store.PostgresPool)
	store.Card = card.NewRepository(store.PostgresPool)
	store.PermissionChecker = rbac.NewRepository(store.PostgresPool)

	return store, nil
}

func (s *Store) Close() error {
	s.PostgresPool.Close()
	return s.RedisClient.Close()
}

func (s *Store) setupRedis(conf pkgredis.Config, logger *zerolog.Logger) error {
	client, err := pkgredis.NewPoolRedis(&goredis.Options{
		Addr:     fmt.Sprintf("%s:%s", conf.Host, conf.Port),
		Password: conf.Password,
		DB:       conf.NumberDB,
	}, conf, logger)
	if err != nil {
		return fmt.Errorf("pkgredis.NewPoolRedis: %w", err)
	}

	s.RedisClient = client
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
