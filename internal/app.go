package internal

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	_ "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/docs"
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
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/vk"
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
func setupRouter(manager *service.Manager, logger *zerolog.Logger, vkOAuthConf *config.VkOAuth) *mux.Router {
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
	boardHandler := board.NewBoardHandler(manager.Board)
	profileHandler := profile.NewProfileHandler(manager.Profile)

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

func setupVKOAuth(conf *config.VkOAuth) *oauth2.Config {
	const emailKey = "email"
	var vkOAuthScopes = []string{emailKey}

	return &oauth2.Config{
		ClientID:     conf.AppID,
		ClientSecret: conf.AppKey,
		RedirectURL:  conf.RedirectURL,
		Scopes:       vkOAuthScopes,
		Endpoint:     vk.Endpoint,
	}
}

func createDemoUser(s *repository.Store, m *service.Manager, logger *zerolog.Logger) {
	user, _, err := m.Auth.Register(context.Background(), "Demo", "12345678", "demo@demo.ru")
	if err != nil {
		logger.Err(err).Msg("cannot create demo user")
	} else {
		logger.Info().Msg("demo user created")
	}

	err = s.Boards.CreateEmptyBoard(context.Background(), user.ID)
	if err != nil {
		logger.Err(err).Msg("cannot create demo board")
	} else {
		logger.Info().Msg("demo board created")
	}
}
