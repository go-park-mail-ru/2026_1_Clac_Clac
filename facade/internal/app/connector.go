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

func NewConnector(config *config.Services, logger *zerolog.Logger) (*Connector, error) {
	var activeConns []*grpc.ClientConn

	connect := func(addr string, timeout time.Duration, retries int) (*grpc.ClientConn, error) {
		serviceConfig := fmt.Sprintf(`{
			"methodConfig": [{
				"name": [{"service": ""}],
				"timeout": "%fs",
				"retryPolicy": {
					"maxAttempts": %d,
					"initialBackoff": "0.1s",
					"maxBackoff": "2s",
					"backoffMultiplier": 2.0,
					"retryableStatusCodes": ["UNAVAILABLE", "DEADLINE_EXCEEDED"]
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

	userConn, err := connect(config.User.Client.Addr, config.User.Client.TimeOut, config.User.Client.Retries)
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to User service: %w", err)
	}

	authConn, err := connect(config.Auth.Client.Addr, config.Auth.Client.TimeOut, config.Auth.Client.Retries)
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to Auth service: %w", err)
	}

	mailSenderConn, err := connect(config.MailSender.Addr, config.MailSender.TimeOut, config.MailSender.Retries)
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to MailSender service: %w", err)
	}

	rateLimiterConn, err := connect(config.RateLimiters.Addr, config.RateLimiters.ClientConfig.TimeOut, config.RateLimiters.ClientConfig.Retries)
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to RateLimiter service: %w", err)
	}

	boardConn, err := connect(config.Board.Client.Addr, config.Board.Client.TimeOut, config.Board.Client.Retries)
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to Board service: %w", err)
	}
	appealConn, err := connect(config.Appeal.Client.Addr, config.Appeal.Client.TimeOut, config.Appeal.Client.Retries)
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to Appeal service: %w", err)
	}

	return &Connector{
		User:        clients.NewUserClient(userConn),
		Auth:        clients.NewAuthClient(authConn),
		MailSender:  clients.NewMailSenderClient(mailSenderConn),
		RateLimiter: clients.NewRateLimiterClient(rateLimiterConn),
		Board:       clients.NewBoardClient(boardConn),
		Section:     clients.NewSectionClient(boardConn),
		Card:        clients.NewCardClient(boardConn),
		Appeal:      clients.NewAppealClient(appealConn),
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
