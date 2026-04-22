package app

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"

	db "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/db"
	redisConnector "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/redis"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/s3"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
)

type App struct {
	Config     *config.Config
	Logger     *zerolog.Logger
	Store      *Store
	Manager    *Manager
	GRPCServer *grpc.Server
}

// Создает приложение, настраивает его компоненты
func NewApp(conf *config.Config) *App {
	logger := setupLogger(&conf.App)

	grpcServer := grpc.NewServer()

	store, err := setupStore(conf, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("cannot initialise store")
	}

	manager := setupManager(store, conf)

	vkOAuth := setupVKOAuth(&conf.VkOAuth)
	delivery := setupDelivery(manager, conf, vkOAuth)

	delivery.Register(grpcServer)

	return &App{
		Config:     conf,
		Logger:     logger,
		Store:      store,
		Manager:    manager,
		GRPCServer: grpcServer,
	}
}

// Запуск приложения
func (a *App) Run() {
	defer func() {
		errClose := a.Store.Close()
		if errClose != nil {
			a.Logger.Err(errClose).Msg("close strore error")
		}
	}()

	address := ":" + a.Config.GRPC.Port
	listener, err := net.Listen("tcp", address)
	if err != nil {
		a.Logger.Fatal().Err(err).Msgf("Fail to connect listener, %s", err.Error())
	}

	a.Logger.Info().Msgf("user gRPC Server is listening on %s", address)

	if err := a.GRPCServer.Serve(listener); err != nil {
		a.Logger.Fatal().Err(err).Msg("Failed to serve gRPC")
	}
}

// Настройка логера
func setupLogger(conf *config.Application) *zerolog.Logger {
	var loggerOutput io.Writer

	// В зависимости от режима работы разные форматы вывода
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

// Настройка подключения к базе данных
func setupDatabase(dbConnection *config.DatabaseConnection, logger *zerolog.Logger) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbConnection.User,
		dbConnection.Password,
		dbConnection.Host,
		dbConnection.Port,
		dbConnection.Name)

	convertedConfigDB := db.Config{
		User:     dbConnection.User,
		Password: dbConnection.Password,
		Host:     dbConnection.Host,
		Port:     dbConnection.Port,
		Name:     dbConnection.Name,

		MinConnections:        dbConnection.MinConnections,
		MaxConnections:        dbConnection.MaxConnections,
		MaxConnectionLifetime: dbConnection.MaxConnectionLifetime,
		MaxHealthCheckPeriod:  dbConnection.MaxHealthCheckPeriod,
		PingSleepTime:         dbConnection.PingSleepTime,
		TimeOut:               dbConnection.TimeOut,
		MaxRetries:            dbConnection.MaxRetries,
	}

	pool, err := db.NewPoolPostgres(dsn, convertedConfigDB, logger)
	if err != nil {
		return nil, fmt.Errorf("db.NewPool: %w", err)
	}

	return pool, nil
}

// Настройка подключения к Redis
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

// Настройка стора
func setupStore(conf *config.Config, logger *zerolog.Logger) (*Store, error) {
	pool, err := setupDatabase(&conf.DBConnection, logger)
	if err != nil {
		return nil, fmt.Errorf("setupDatabase: %w", err)
	}

	redisClient, err := setupRedis(&conf.RedisConnection, logger)
	if err != nil {
		return nil, fmt.Errorf("setupRedis: %w", err)
	}

	const intConvertationBase = 10
	const intConvertationSize = 64
	s3ConnectTimeout, err := strconv.ParseInt(conf.S3.ConnectTimeout, intConvertationBase, intConvertationSize)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s3ConnectTimeout)*time.Second)
	defer cancel()

	s3Client, err := s3.NewAWSClient(
		ctx,
		conf.S3.Region,
		conf.S3.Endpoint,
		conf.S3.AccessKey,
		conf.S3.SecretKey,
	)
	if err != nil {
		return nil, fmt.Errorf("s3.NewAWSClient: %w", err)
	}

	return NewStore(pool, redisClient, s3Client, *conf), nil
}

// Настройка менеджера сервисов
func setupManager(s *Store, conf *config.Config) *Manager {
	return NewManager(s, *conf)
}

func setupDelivery(m *Manager, conf *config.Config, vkOAuth *oauth2.Config) *Delivery {
	return NewDelivery(m, conf, vkOAuth)
}

func setupVKOAuth(conf *config.VkOAuth) *oauth2.Config {
	return NewVKOAuth(conf)
}
