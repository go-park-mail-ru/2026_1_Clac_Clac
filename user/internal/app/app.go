package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	db "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/db"
	engine "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/grpcEngine"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/s3"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type App struct {
	Config  *config.Config
	Logger  *zerolog.Logger
	Store   *Store
	Manager *Manager
	Engine  *engine.Engine
}

// Создает приложение, настраивает его компоненты
func NewApp(conf *config.Config) (*App, error) {
	logger := setupLogger(&conf.App)

	engine := setupEngine(conf.Engine, logger)

	store, err := setupStore(conf, logger)
	if err != nil {
		return nil, fmt.Errorf("setupStore: %w", err)
	}

	manager := setupManager(store, conf)

	delivery := setupDelivery(manager, conf)
	delivery.Register(engine.Server)

	return &App{
		Config:  conf,
		Logger:  logger,
		Store:   store,
		Manager: manager,
		Engine:  engine,
	}, nil
}

// Запуск приложения
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

func setupEngine(conf engine.Config, logger *zerolog.Logger) *engine.Engine {
	return engine.New(conf, logger)
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

// Настройка стора
func setupStore(conf *config.Config, logger *zerolog.Logger) (*Store, error) {
	pool, err := setupDatabase(&conf.DBConnection, logger)
	if err != nil {
		return nil, fmt.Errorf("setupDatabase: %w", err)
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

	return NewStore(pool, s3Client, *conf), nil
}

// Настройка менеджера сервисов
func setupManager(s *Store, conf *config.Config) *Manager {
	return NewManager(s, *conf)
}

func setupDelivery(m *Manager, conf *config.Config) *Delivery {
	return NewDelivery(m, conf)
}
