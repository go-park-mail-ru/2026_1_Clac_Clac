package app

import (
	"context"
	"io"
	"net/http"
	"os"

	auth "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/handler"
	board "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/handler"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/db"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/engine"
	health "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/health/handler"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	profile "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/handler"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"golang.org/x/oauth2"
)

type App struct {
	Config   *config.Config
	Logger   *zerolog.Logger
	Engine   *engine.Engine
	Database *db.MapDatabases
	Store    *Store
	Manager  *Manager
}

// Создает приложение, настраивает его компоненты
func NewApp(conf *config.Config) *App {
	logger := setupLogger(&conf.App)

	db := setupDatabase()
	store := setupStore(db)
	manager := setupManager(store, &conf.MailSender)
	createDemoUser(store, manager, logger)

	router := setupRouter(manager, logger, &conf.VkOAuth)

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
func setupRouter(manager *Manager, logger *zerolog.Logger, vkOAuthConf *config.VkOAuth) *mux.Router {
	router := mux.NewRouter()

	// Добавление обищх мидлваре
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.LoggerMiddleware(logger))

	// Ручки, которым не нужна авторизация
	public := router.PathPrefix("/").Subrouter()
	public.HandleFunc("/healthcheck", health.HealthcheckHandler).Methods(http.MethodGet)
	public.PathPrefix("/docs/").Handler(httpSwagger.WrapHandler)
	public.Handle("/docs", http.RedirectHandler("/docs/", http.StatusMovedPermanently))
	// Добавление рутов, зависящих от сервисов
	authHandler := auth.NewAuthHandler(manager.Auth)

	public.HandleFunc("/register", authHandler.RegisterUser).Methods(http.MethodPost)
	public.HandleFunc("/login", authHandler.LogInUser).Methods(http.MethodPost)
	public.HandleFunc("/logout", authHandler.LogOutUser).Methods(http.MethodPost)

	vkOAuth := setupVKOAuth(vkOAuthConf)
	public.HandleFunc("/oauth/vk", authHandler.VkOAuthCallback(vkOAuthConf, "/", vkOAuth))

	public.HandleFunc("/forgot-password", authHandler.SendRecoveryEmail).Methods(http.MethodPost)
	public.HandleFunc("/check-code", authHandler.CheckRecoveryCode).Methods(http.MethodPost)
	public.HandleFunc("/reset-password", authHandler.ResetUserPassword).Methods(http.MethodPost)

	// Для досутпа к этим ручкам нужна авторизация
	protected := router.PathPrefix("/").Subrouter()
	// Добавление мидлваре для авторизации
	protected.Use(middleware.AuthMiddleware(manager.Auth))
	// Руты, на которые пользователь объязательно должен быть авторизован
	boardHandler := board.NewHandler(manager.Board)
	profileHandler := profile.NewHandler(manager.Profile)

	protected.HandleFunc("/me", authHandler.MeHandler).Methods(http.MethodGet)
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
func setupDatabase() *db.MapDatabases {
	return db.NewMapDatabse()
}

// Настройка стора
func setupStore(db *db.MapDatabases) *Store {
	return NewStore(db)
}

// Настройка менеджера сервисов
func setupManager(s *Store, mailSenderConf *config.MailSender) *Manager {
	return NewManager(s, mailSenderConf)
}

func setupVKOAuth(conf *config.VkOAuth) *oauth2.Config {
	return NewVKOAuth(conf)
}

func createDemoUser(s *Store, m *Manager, logger *zerolog.Logger) {
	user, _, err := m.Auth.Register(context.Background(), "Demo", "12345678", "demo@demo.ru")
	if err != nil {
		logger.Err(err).Msg("cannot create demo user")
	} else {
		logger.Info().Msg("demo user created")
	}

	err = m.Board.CreateEmptyBoard(context.Background(), user.ID)
	if err != nil {
		logger.Err(err).Msg("cannot create demo board")
	} else {
		logger.Info().Msg("demo board created")
	}
}
