package internal

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/engine"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/handler/auth"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/handler/board"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/handler/health"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/handler/profile"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository"
	dbConnection "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/db_connection"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type App struct {
	Config   *config.Config
	Logger   *zerolog.Logger
	Engine   *engine.Engine
	Database *dbConnection.MapDatabases
	Store    *repository.Store
	Manager  *service.Manager
}

// Создает приложение, настраивает его компоненты
func NewApp(conf *config.Config) *App {
	logger := setupLogger(&conf.App)

	db := setupDatabase()
	store := setupStore(db)
	manager := setupManager(store, &conf.MailSender)

	router := setupRouter(manager, logger)

	e := setupEngine(&conf.Engine, logger, router)

	return &App{
		Config:   conf,
		Logger:   logger,
		Engine:   e,
		Database: db,
		Store:    store,
		Manager:  manager,
	}
}

// Запуск приложения
func (a *App) Run() {
	if err := a.Engine.Start(context.Background()); err != nil {
		a.Logger.Err(err).Msg("engine error")
	}
}

// Настройка рутов
func setupRouter(manager *service.Manager, logger *zerolog.Logger) *mux.Router {
	router := mux.NewRouter()

	// Добавление обищх мидлваре
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.LoggerMiddleware(logger))

	// Ручки, которым не нужна авторизация
	public := router.PathPrefix("/").Subrouter()
	public.HandleFunc("/healthcheck", health.HealthcheckHandler).Methods(http.MethodGet)
	// Добавление рутов, зависящих от сервисов
	authHandler := auth.NewAuthHandler(manager.Auth)

	public.HandleFunc("/register", authHandler.RegisterUser).Methods(http.MethodPost)
	public.HandleFunc("/login", authHandler.LogInUser).Methods(http.MethodPost)

	public.HandleFunc("/forgot-password", authHandler.SendRecoveryEmail).Methods(http.MethodPost)
	public.HandleFunc("/check-code", authHandler.CheckRecoveryCode).Methods(http.MethodPost)
	public.HandleFunc("/reset-password", authHandler.ResetUserPassword).Methods(http.MethodPost)

	// Для досутпа к этим ручкам нужна авторизация
	protected := router.PathPrefix("/").Subrouter()
	// Добавление мидлваре для авторизации
	protected.Use(middleware.AuthMiddleware(manager.Auth))
	// Руты, на которые пользователь объязательно должен быть авторизован
	boardHandler := board.NewBoardHandler(manager.Board)
	profileHandler := profile.NewProfileHandler(manager.Profile)

	protected.HandleFunc("/logout", authHandler.LogOutUser).Methods(http.MethodPost)
	protected.HandleFunc("/home", boardHandler.GetUserBoards).Methods(http.MethodGet)
	protected.HandleFunc("/profile", profileHandler.GetProfile).Methods(http.MethodGet)

	return router
}

// Настройка сервера
func setupEngine(conf *config.Engine, logger *zerolog.Logger, router *mux.Router) *engine.Engine {
	// Создание сервера
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
func setupDatabase() *dbConnection.MapDatabases {
	return dbConnection.NewMapDatabse()
}

// Настройка стора
func setupStore(db *dbConnection.MapDatabases) *repository.Store {
	return repository.NewStore(db)
}

// Настройка менеджера сервисов
func setupManager(s *repository.Store, mailSenderConf *config.MailSender) *service.Manager {
	return service.NewManager(s, mailSenderConf)
}
