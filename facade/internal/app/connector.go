package app

import (
	"fmt"
	"time"

	sentrygrpc "github.com/getsentry/sentry-go/grpc"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/clients"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/config"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Connector struct {
	User        *clients.User
	Auth        *clients.Auth
	MailSender  *clients.MailSender
	RateLimiter *clients.RateLimiter
	Appeal      *clients.Appeal

	Board   *clients.Board
	Section *clients.Section
	Card    *clients.Card

	logger *zerolog.Logger
	conns  []*grpc.ClientConn
}

func NewConnector(config *config.Config, logger *zerolog.Logger) (*Connector, error) {
	var activeConns []*grpc.ClientConn

	connect := func(addr string, timeout time.Duration, retries int) (*grpc.ClientConn, error) {
		serviceConfig := fmt.Sprintf(`{
			"methodConfig": [{
				"name": [{"service": ""}],
				"timeout": "%fs",
				"retryPolicy": {
					"MaxAttempts": %d,
					"InitialBackoff": "0.1s",
					"MaxBackoff": "2s",
					"BackoffMultiplier": 2.0,
					"RetryableStatusCodes": ["UNAVAILABLE", "DEADLINE_EXCEEDED"]
				}
			}]
		}`, timeout.Seconds(), retries)

		conn, err := grpc.NewClient(
			addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithUnaryInterceptor(sentrygrpc.UnaryClientInterceptor()),
			grpc.WithDefaultServiceConfig(serviceConfig),
		)

		if err != nil {
			return nil, err
		}
		activeConns = append(activeConns, conn)
		return conn, nil
	}

	userConn, err := connect(config.Services.User.Client.Addr, config.Services.User.Client.TimeOut, config.Services.User.Client.Retries)
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to User service: %w", err)
	}

	configUser := clients.ConfigUser{
		MaxUserAvatarBytesSize: int(config.App.MaxUploadImageSize),
	}

	authConn, err := connect(config.Services.Auth.Client.Addr, config.Services.Auth.Client.TimeOut, config.Services.Auth.Client.Retries)
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to Auth service: %w", err)
	}

	mailSenderConn, err := connect(config.Services.MailSender.Addr, config.Services.MailSender.TimeOut, config.Services.MailSender.Retries)
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to MailSender service: %w", err)
	}

	rateLimiterConn, err := connect(config.Services.RateLimiters.Addr, config.Services.RateLimiters.ClientConfig.TimeOut, config.Services.RateLimiters.ClientConfig.Retries)
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to RateLimiter service: %w", err)
	}

	boardConn, err := connect(config.Services.Board.Client.Addr, config.Services.Board.Client.TimeOut, config.Services.Board.Client.Retries)
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to Board service: %w", err)
	}

	configBoard := clients.ConfigBoard{
		MaxBackgroundBytesSize: int(config.App.MaxUploadImageSize),
	}

	configCard := clients.CardConfig{
		MaxAttachmentBufferSize: int(config.App.MaxFileSize),
	}

	appealConn, err := connect(config.Services.Appeal.Client.Addr, config.Services.Appeal.Client.TimeOut, config.Services.Appeal.Client.Retries)
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to Appeal service: %w", err)
	}

	configAppeal := clients.ConfigAppeal{
		MaxAppealAttachmentBytesSize: int(config.App.MaxFileSize),
	}

	return &Connector{
		User:        clients.NewUserClient(userConn, configUser),
		Auth:        clients.NewAuthClient(authConn),
		MailSender:  clients.NewMailSenderClient(mailSenderConn),
		RateLimiter: clients.NewRateLimiterClient(rateLimiterConn),
		Board:       clients.NewBoardClient(boardConn, configBoard),
		Section:     clients.NewSectionClient(boardConn),
		Card:        clients.NewCardClient(boardConn, configCard),
		Appeal:      clients.NewAppealClient(appealConn, configAppeal),
		logger:      logger,
		conns:       activeConns,
	}, nil
}

func (c *Connector) Close() {
	closeAll(c.conns, c.logger)
}

func closeAll(conns []*grpc.ClientConn, logger *zerolog.Logger) {
	for _, conn := range conns {
		if conn != nil {
			if err := conn.Close(); err != nil {
				logger.Warn().Err(err).Msg("failed to close gRPC connection cleanly")
			}
		}
	}
}
