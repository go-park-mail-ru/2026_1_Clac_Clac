package app

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	sentrygrpc "github.com/getsentry/sentry-go/grpc"
	board "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/delivery"
	card "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/delivery"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/config"
	section "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/delivery"
	grpcEngine "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/grpcEngine"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/interceptors"
	sentryLogger "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/logger"
	boardPB "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/board/v1"
	cardPB "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/card/v1"
	sectionPB "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/section/v1"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/s3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type App struct {
	Config        config.Config
	Logger        zerolog.Logger
	Engine        *grpcEngine.Engine
	Store         *Store
	Manager       *Manager
	MetricsServer *http.Server
}

func NewApp(conf config.Config) (*App, error) {
	// TODO: delete this comment
	app := &App{
		Config: conf,
	}

	app.setupLogger()

	if err := app.setupSentry(); err != nil {
		return nil, fmt.Errorf("app.setupSentry: %w", err)
	}

	go app.setupMetricsServer()

	if err := app.setupStore(&app.Logger); err != nil {
		sentryLogger.CaptureError(err, "Setup connector", map[string]interface{}{"component": "store"})
		return nil, fmt.Errorf("app.setupStore: %w", err)
	}

	app.setupManager(app.Store, &conf)
	app.setupEngine(&app.Logger)
	app.registerServices(app.Engine, app.Manager)

	return app, nil
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

func (a *App) setupMetricsServer() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	a.MetricsServer = &http.Server{
		Addr:    a.Config.Metrics.MetricsPort,
		Handler: mux,
	}

	a.Logger.Info().Msg(fmt.Sprintf("Metrics server listening on: %s", a.Config.Metrics.MetricsPort))

	if err := a.MetricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		sentryLogger.CaptureError(err, "listen and serve Prometheous", map[string]interface{}{"component": "prometheous"})
	}
}

func (a *App) setupLogger() {
	var loggerOutput io.Writer

	if config.IsDebug(a.Config.App.LogLevel) {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		loggerOutput = zerolog.ConsoleWriter{Out: os.Stdout}
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		loggerOutput = os.Stdout
	}

	a.Logger = zerolog.New(loggerOutput).With().Timestamp().Logger()
}

func (a *App) setupSentry() error {
	return sentryLogger.Init(a.Config.Sentry)
}

func (a *App) setupStore(logger *zerolog.Logger) error {
	store, err := NewStore(logger, a.Config)
	if err != nil {
		return fmt.Errorf("store NewStore: %w", err)
	}

	a.Store = store
	return nil
}

func (a *App) setupManager(store *Store, conf *config.Config) {
	a.Manager = NewManager(store, conf)
}

func (a *App) registerServices(engine *grpcEngine.Engine, manager *Manager) {
	boardPB.RegisterBoardServiceServer(
		engine.Server,
		board.NewHandler(manager.Board, board.Config{
			BaseBackgroundURL:          s3.GetURL(a.Config.S3.Endpoint, a.Config.S3.BoardsBackgroundsBucket),
			MultipartBackgroundFileKey: a.Config.Board.Handler.MultipartBackgroundFileKey,
			MaxBackgroundSize:          a.Config.Board.Handler.MaxBackgroundSize,
		}),
	)
	sectionPB.RegisterSectionServiceServer(
		engine.Server,
		section.NewHandler(manager.Section, section.Config(a.Config.Section.Handler)),
	)
	cardPB.RegisterCardServiceServer(
		engine.Server,
		card.NewHandler(manager.Card, card.Config(a.Config.Card.Handler)),
	)

	reflection.Register(engine.Server)
}

func (a *App) setupEngine(logger *zerolog.Logger) {
	sentryOpts := sentrygrpc.ServerOptions{
		Repanic: a.Config.Sentry.Repanic,
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
		grpc.MaxRecvMsgSize(int(a.Config.App.MaxUploadImageSize)),
		grpc.MaxSendMsgSize(int(a.Config.App.MaxUploadImageSize)),
	}
	a.Engine = grpcEngine.New(a.Config.Engine, logger, opts...)
}
