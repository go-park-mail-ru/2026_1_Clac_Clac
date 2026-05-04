package app

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	sentrygrpc "github.com/getsentry/sentry-go/grpc"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/config"
	enginegrpc "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/grpcEngine"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/interceptors"
	sentryLogger "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/logger"
	redisConnector "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/redis"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
)

type App struct {
	Config        *config.Config
	Logger        *zerolog.Logger
	Store         *Store
	Manager       *Manager
	Engine        *enginegrpc.Engine
	MetricsServer *http.Server
}

func NewApp(conf *config.Config) (*App, error) {
	logger := setupLogger(&conf.App)

	if err := setupSentry(conf); err != nil {
		return nil, fmt.Errorf("setupSentry: %w", err)
	}

	metricsServer := setupMetricsServer(conf)
	go func() {
		logger.Info().Msg(fmt.Sprintf("Metrics server listening on: %s", metricsServer.Addr))
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sentryLogger.CaptureError(err, "listen and serve Prometheous", map[string]interface{}{"component": "prometheous"})
		}
	}()

	engine := setupEngine(conf.Engine, conf.Sentry, logger)

	store, err := setupStore(conf, logger)
	if err != nil {
		sentryLogger.CaptureError(err, "Setup connector", map[string]interface{}{"component": "store"})
		return nil, fmt.Errorf("setupStore: %w", err)
	}

	manager := setupManager(store, conf)

	vkOAuth := NewVKOAuth(&conf.VkOAuth)
	delivery := setupDelivery(manager, conf, vkOAuth)
	delivery.Register(engine.Server)

	return &App{
		Config:        conf,
		Logger:        logger,
		Store:         store,
		Manager:       manager,
		Engine:        engine,
		MetricsServer: metricsServer,
	}, nil
}

func (a *App) Run() {
	defer func() {
		if a.MetricsServer != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := a.MetricsServer.Shutdown(ctx); err != nil {
				a.Logger.Err(err).Msg("metrics server shutdown error")
			}
		}

		if err := a.Store.Close(); err != nil {
			a.Logger.Err(err).Msg("close store error")
		}
	}()

	if err := a.Engine.Start(context.Background()); err != nil {
		a.Logger.Err(err).Msg("engine error")
	}
}

func setupEngine(conf enginegrpc.Config, sentryConf sentryLogger.Sentry, logger *zerolog.Logger) *enginegrpc.Engine {
	sentryOpts := sentrygrpc.ServerOptions{
		Repanic: sentryConf.Repanic,
	}

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptors.PrometheusUnaryInterceptor(),
			interceptors.UnaryAccessLog(logger),
			interceptors.UnaryPanicRecovery(logger),
			sentrygrpc.UnaryServerInterceptor(sentryOpts),
		),
		grpc.ChainStreamInterceptor(
			interceptors.PrometheusStreamInterceptor(),
			interceptors.StreamAccessLog(logger),
			interceptors.StreamPanicRecovery(logger),
			sentrygrpc.StreamServerInterceptor(sentryOpts),
		),
	}
	return enginegrpc.New(conf, logger, opts...)
}

func setupMetricsServer(conf *config.Config) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	metricsServer := &http.Server{
		Addr:    conf.Metrics.MetricsPort,
		Handler: mux,
	}

	return metricsServer
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

func setupSentry(conf *config.Config) error {
	return sentryLogger.Init(conf.Sentry)
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
