package app

import (
	"context"
	"fmt"
	"io"
	"os"

	enginegrpc "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/grpcEngine"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/interceptors"
	redisConnector "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/redis"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/rate_limiter/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type App struct {
	Config  *config.Config
	Logger  *zerolog.Logger
	Engine  *enginegrpc.Engine
	Store   *Store
	Manager *Manager
}

func NewApp(conf *config.Config) (*App, error) {
	logger := setupLogger(&conf.App)

	engine := setupEngine(conf.Engine, logger)

	store, err := setupStore(conf, logger)
	if err != nil {
		return nil, fmt.Errorf("setupStore: %w", err)
	}

	manager := setupManager(store)

	delivery := setupDelivery(manager)
	delivery.Register(engine.Server)

	reflection.Register(engine.Server)

	return &App{
		Config:  conf,
		Logger:  logger,
		Engine:  engine,
		Store:   store,
		Manager: manager,
	}, nil
}

func (a *App) Run() {
	defer func() {
		if err := a.Store.Close(); err != nil {
			a.Logger.Err(err).Msg("close store error")
		}
	}()

	if err := a.Engine.Start(context.Background()); err != nil {
		a.Logger.Fatal().Err(err).Msg("engine.Start")
	}
}

func setupEngine(conf enginegrpc.Config, logger *zerolog.Logger) *enginegrpc.Engine {
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptors.UnaryAccessLog(logger)),
		grpc.StreamInterceptor(interceptors.StreamAccessLog(logger)),
		grpc.UnaryInterceptor(interceptors.UnaryPanicRecovery(logger)),
		grpc.StreamInterceptor(interceptors.StreamPanicRecovery(logger)),
	}
	return enginegrpc.New(conf, logger, opts...)
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

	convertedConfigRedis := redisConnector.Config{
		PingSleepTime: redisConnection.PingSleepTime,
		MaxRetries:    redisConnection.MaxRetries,
	}

	client, err := redisConnector.NewPoolRedis(&redisSettings, convertedConfigRedis, logger)
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

	return NewStore(redisClient), nil
}

func setupManager(s *Store) *Manager {
	return NewManager(s)
}

func setupDelivery(m *Manager) *Delivery {
	return NewDelivery(m)
}
