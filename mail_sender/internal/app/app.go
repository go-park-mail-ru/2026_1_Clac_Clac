package app

import (
	"io"
	"net"
	"os"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/mail_sender/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/mail_sender/internal/manager"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	grpc_delivery "github.com/go-park-mail-ru/2026_1_Clac_Clac/mail_sender/internal/delivery/grpc"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/mail"
)

type App struct {
	Config      *config.Config
	Logger      *zerolog.Logger
	MailManager *manager.MailManager
	GRPCServer  *grpc.Server
}

// Создает приложение, настраивает его компоненты
func NewApp(conf *config.Config) *App {
	logger := setupLogger(&conf.App)

	mailManager := setupMailManager(&conf.MailSender)

	mailHandler := setupMailHandler(mailManager)
	grpcServer := grpc.NewServer()

	pb.RegisterMailServiceServer(grpcServer, mailHandler)

	return &App{
		Config:      conf,
		Logger:      logger,
		MailManager: mailManager,
		GRPCServer:  grpcServer,
	}
}

func (a *App) Run() {
	address := ":" + a.Config.GRPC.Port
	listener, err := net.Listen("tcp", address)
	if err != nil {
		a.Logger.Fatal().Err(err).Msgf("Fail to connect listener, %s", err.Error())
	}

	a.Logger.Info().Msgf("MailSender gRPC Server is listening on %s", address)

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

func setupMailManager(conf *config.MailSender) *manager.MailManager {
	return manager.NewMailSender(conf)
}

func setupMailHandler(mailManager *manager.MailManager) *grpc_delivery.MailHandler {
	return grpc_delivery.NewMailHandler(mailManager)
}
