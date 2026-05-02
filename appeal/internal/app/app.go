package app

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/config"
	appeal "github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/delivery"
	grpcEngine "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/grpcEngine"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/interceptors"
	appealPB "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/appeal/v1"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/s3"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type App struct {
	Config  config.Config
	Logger  zerolog.Logger
	Engine  *grpcEngine.Engine
	Store   *Store
	Manager *Manager
}

func NewApp(conf config.Config) (*App, error) {
	app := &App{
		Config: conf,
	}

	app.setupLogger()

	if err := app.setupStore(&app.Logger); err != nil {
		return nil, fmt.Errorf("app.setupStore: %w", err)
	}

	app.setupManager(app.Store)
	app.setupEngine(&app.Logger)
	app.registerServices(app.Engine, app.Manager)

	return app, nil
}

func (a *App) Run() {
	defer func() {
		if err := a.Store.Close(); err != nil {
			a.Logger.Err(err).Msg("close store error")
		}
	}()

	if err := a.Engine.Start(context.Background()); err != nil {
		a.Logger.Err(err).Msg("engine error")
	}
}

func (a *App) setupLogger() {
	var loggerOutput io.Writer

	// В зависимости от режима работы разные форматы вывода
	if config.IsDebug(a.Config.App.LogLevel) {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		loggerOutput = zerolog.ConsoleWriter{Out: os.Stdout}
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		loggerOutput = os.Stdout
	}

	a.Logger = zerolog.New(loggerOutput).With().Timestamp().Logger()
}

func (a *App) setupStore(logger *zerolog.Logger) error {
	store, err := NewStore(logger, a.Config)
	if err != nil {
		return fmt.Errorf("store NewStore: %w", err)
	}

	a.Store = store
	return nil
}

func (a *App) setupManager(store *Store) {
	a.Manager = NewManager(store)
}

func (a *App) registerServices(engine *grpcEngine.Engine, manager *Manager) {
	appealPB.RegisterAppealServiceServer(
		engine.Server,
		appeal.NewHandler(manager.Appeal, appeal.Config{
			AttachmentBaseURL: s3.GetURL(a.Config.S3.Endpoint, a.Config.S3.AppealAttachmentBucket),
		}),
	)

	reflection.Register(engine.Server)
}

func (a *App) setupEngine(logger *zerolog.Logger) {
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptors.UnaryAccessLog(logger),
			interceptors.UnaryPanicRecovery(logger),
		),
		grpc.ChainStreamInterceptor(
			interceptors.StreamAccessLog(logger),
			interceptors.StreamPanicRecovery(logger),
		),
	}
	a.Engine = grpcEngine.New(a.Config.Engine, logger, opts...)
}
