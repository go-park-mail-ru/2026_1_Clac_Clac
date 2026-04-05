package app

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	auth "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/handler"
	handlerDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/handler/dto"
	board "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/handler"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/db"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/engine"
	health "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/health/handler"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	profile "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/handler"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/s3"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"golang.org/x/oauth2"
)

type App struct {
	Config  *config.Config
	Logger  *zerolog.Logger
	Engine  *engine.Engine
	Store   *Store
	Manager *Manager
}

// Создает приложение, настраивает его компоненты
func NewApp(conf *config.Config) *App {
	logger := setupLogger(&conf.App)

	store, err := setupStore(conf, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("cannot initialise store")
	}

	manager := setupManager(store, &conf.MailSender)

	router := setupRouter(manager, conf, logger)

	e := setupEngine(&conf.Engine, logger, router)

	return &App{
		Config:  conf,
		Logger:  logger,
		Engine:  e,
		Store:   store,
		Manager: manager,
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

	if err := a.Engine.Start(context.Background()); err != nil {
		a.Logger.Err(err).Msg("engine error")
	}
}

// Настройка рутов
func setupRouter(manager *Manager, conf *config.Config, logger *zerolog.Logger) *mux.Router {
	router := mux.NewRouter().PathPrefix("/api").Subrouter()

	// Добавление обищх мидлваре
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.LoggerMiddleware(logger))
	router.Use(middleware.CORSMiddleware(&conf.CORS))
	router.Use(middleware.LimitRequestSizeMiddleware(conf.App.MaxTextRequestSize))
	router.Use(middleware.TimeOutMiddleware(time.Second * 5))

	authHandler := auth.NewHandler(manager.Auth)
	router.HandleFunc("/csrf", authHandler.SetCSRFCookieHandler)

	csrfProtected := router.PathPrefix("/").Subrouter()
	csrfProtected.Use(middleware.CSRFMiddleware)

	// Ручки, которым не нужна авторизация
	public := csrfProtected.PathPrefix("/").Subrouter()
	public.HandleFunc("/healthcheck", health.HealthcheckHandler).Methods(http.MethodGet)
	public.Handle("/docs", http.RedirectHandler("/api/docs/", http.StatusMovedPermanently))
	public.PathPrefix("/docs").Handler(httpSwagger.WrapHandler)
	// Добавление рутов, зависящих от сервисов

	public.Handle("/login", wrapWithLimit(manager.Auth, handlerDto.RateLimitConfig(conf.DBRateLimiters.GetParameters(config.LogInUser)),
		logger, authHandler.LogInUser)).Methods(http.MethodPost)

	public.Handle("/register", wrapWithLimit(manager.Auth, handlerDto.RateLimitConfig(conf.DBRateLimiters.GetParameters(config.RegisterUser)),
		logger, authHandler.RegisterUser)).Methods(http.MethodPost)

	public.HandleFunc("/logout", authHandler.LogOutUser).Methods(http.MethodPost)

	vkOAuth := setupVKOAuth(&conf.VkOAuth)
	public.HandleFunc("/oauth/vk", authHandler.VkOAuthCallback(&conf.VkOAuth, "/", vkOAuth))

	public.HandleFunc("/forgot-password", authHandler.SendRecoveryEmail).Methods(http.MethodPost)
	public.HandleFunc("/check-code", authHandler.CheckRecoveryCode).Methods(http.MethodPost)
	public.HandleFunc("/reset-password", authHandler.ResetUserPassword).Methods(http.MethodPost)

	// Для досутпа к этим ручкам нужна авторизация
	protected := csrfProtected.PathPrefix("/").Subrouter()
	// Добавление мидлваре для авторизации
	protected.Use(middleware.AuthMiddleware(manager.Auth, logger))
	// Руты, на которые пользователь объязательно должен быть авторизован
	profileHandler := profile.NewHandler(manager.Profile)

	protected.HandleFunc("/me", authHandler.MeHandler).Methods(http.MethodGet)
	protected.HandleFunc("/profile", profileHandler.GetProfile).Methods(http.MethodGet)

	boardHandler := board.NewHandler(manager.Board)
	protected.HandleFunc("/boards", boardHandler.GetBoards).Methods(http.MethodGet)
	protected.HandleFunc("/board/{link}", boardHandler.GetBoard).Methods(http.MethodGet)
	protected.HandleFunc("/board/create", boardHandler.CreateBoard).Methods(http.MethodPost)
	protected.HandleFunc("/board/delete", boardHandler.DeleteBoard).Methods(http.MethodPost)
	protected.HandleFunc("/board/update", boardHandler.UpdateBoard).Methods(http.MethodPost)
	protected.HandleFunc("/board/update-background/{link}", boardHandler.UploadBackground).Methods(http.MethodPost)

	return router
}

func setupEngine(conf *config.Engine, logger *zerolog.Logger, router *mux.Router) *engine.Engine {
	return engine.New(conf, logger, router)
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

	pool, err := db.NewPoolPostgres(dsn, dbConnection, logger)
	if err != nil {
		return nil, fmt.Errorf("db.NewPool: %w", err)
	}

	// migrate -path ./internal/db/migrations -database [DSN] up
	//
	// err = db.RunMigrations(dsn, logger)
	// if err != nil {
	// 	return nil, fmt.Errorf("db.RunMigrations: %w", err)
	// }

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

	client, err := db.NewPoolRedis(&redisSettings, redisConnection, logger)
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
	s3ConnectTimeout, err := strconv.ParseInt(conf.S3Avatars.ConnectTimeout, intConvertationBase, intConvertationSize)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s3ConnectTimeout)*time.Second)
	defer cancel()

	s3Client, err := s3.NewAWSClient(
		ctx,
		conf.S3Avatars.Region,
		conf.S3Avatars.Endpoint,
		conf.S3Avatars.AccessKey,
		conf.S3Avatars.SecretKey,
	)
	if err != nil {
		return nil, fmt.Errorf("s3.NewAWSClient: %w", err)
	}

	return NewStore(pool, redisClient, s3Client, conf.S3Avatars, conf.S3Boards), nil
}

// Настройка менеджера сервисов
func setupManager(s *Store, mailSenderConf *config.MailSender) *Manager {
	return NewManager(s, mailSenderConf)
}

func setupVKOAuth(conf *config.VkOAuth) *oauth2.Config {
	return NewVKOAuth(conf)
}
