package app

import (
	"fmt"

	pubsub "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/pubsub"
	pubsubRedis "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/pubsub/redis"
	pkgredis "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/redis"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/config"
	goredis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

const (
	// TODO: Перенести в конфиг
	DefaultStreamName = "boards"
)

type Store struct {
	RedisClient      *goredis.Client
	RedisMultiplexor *pubsubRedis.RedisMultiplexor
	Subscriber       pubsub.Subscriber[common.BoardUpdateEvent]
}

func NewStore(logger *zerolog.Logger, conf config.Config) (*Store, error) {
	store := &Store{}

	if err := store.setupRedis(conf.Redis.ToPkg(), logger); err != nil {
		return nil, fmt.Errorf("store.setupRedis: %w", err)
	}

	store.setupRedisMultiplexor()
	store.setupSubscriber()

	return store, nil
}

func (s *Store) Close() error {
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

func (s *Store) setupRedisMultiplexor() {
	s.RedisMultiplexor = pubsubRedis.NewRedisMultiplexor(s.RedisClient, DefaultStreamName)
}

func (s *Store) setupSubscriber() {
	s.Subscriber = pubsubRedis.NewMuxSubscriber[common.BoardUpdateEvent](s.RedisMultiplexor)
}
