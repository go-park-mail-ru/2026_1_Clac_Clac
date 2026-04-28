package app

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/config"
	engine "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/grpcEngine"
	redisConnector "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/redis"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
)

type App struct {
	Config  *config.Config
	Logger  *zerolog.Logger
	Store   *Store
	Manager *Manager
	Engine  *engine.Engine
}

func NewApp(conf *config.Config) (*App, error) {
	logger := setupLogger(&conf.App)

	engine := setupEngine(conf.Engine, logger)

	store, err := setupStore(conf, logger)
	if err != nil {
		return nil, fmt.Errorf("setupStore: %w", err)
	}

	manager := setupManager(store, conf)

	vkOAuth := NewVKOAuth(&conf.VkOAuth)
	delivery := setupDelivery(manager, conf, vkOAuth)
	delivery.Register(engine.Server)

	return &App{
		Config:  conf,
		Logger:  logger,
		Store:   store,
		Manager: manager,
		Engine:  engine,
	}, nil
}

func (a *App) Run() {
	defer func() {
		errClose := a.Store.Close()
		if errClose != nil {
			a.Logger.Err(errClose).Msg("close store error")
		}
	}()

	if err := a.Engine.Start(context.Background()); err != nil {
		a.Logger.Err(err).Msg("engine error")
	}
}

func setupEngine(conf engine.Config, logger *zerolog.Logger) *engine.Engine {
	return engine.New(conf, logger)
}

func setupLogger(conf *config.Application) *zerolog.Logger {
	var loggerOutput io.Writer

	if config.IsDebug(conf.LogLevel) {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		loggerOutput = zerolog.ConsoleWriter{Out: os.Stdout}
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		loggerOutput = os.Stdout
	}

	logger := zerolog.New(loggerOutput).With().Timestamp().Logger()
	return &logger
}

func setupRedis(redisConnection *config.RedisConnection, logger *zerolog.Logger) (*redis.Client, error) {
	redisSettings := redis.Options{
		Addr:         fmt.Sprintf("%s:%s", redisConnection.Host, redisConnection.Port),
		Password:     redisConnection.Password,
		DB:           redisConnection.NumberDB,
		PoolSize:     redisConnection.MaxConnections,
		MinIdleConns: redisConnection.MinConnections,
	}

	covertedConfigRedis := redisConnector.Config{
		PingSleepTime: redisConnection.PingSleepTime,
		MaxRetries:    redisConnection.MaxRetries,
	}

	client, err := redisConnector.NewPoolRedis(&redisSettings, covertedConfigRedis, logger)
	if err != nil {
		return nil, fmt.Errorf("db.NewPoolRedis: %w", err)
	}

	return client, nil
}

func setupStore(conf *config.Config, logger *zerolog.Logger) (*Store, error) {
	redisClient, err := setupRedis(&conf.RedisConnection, logger)
	if err != nil {
		return nil, fmt.Errorf("setupRedis: %w", err)
	}

	return NewStore(redisClient, *conf), nil
}

func setupManager(s *Store, conf *config.Config) *Manager {
	return NewManager(s, *conf)
}

func setupDelivery(m *Manager, conf *config.Config, vkOAuth *oauth2.Config) *Delivery {
	return NewDelivery(m, conf, vkOAuth)
}
