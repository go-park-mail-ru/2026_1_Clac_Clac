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

const defaultGRPCMsgSize = 4 * 1024 * 1024

type Connector struct {
	User        *clients.User
	Auth        *clients.Auth
	MailSender  *clients.MailSender
	RateLimiter *clients.RateLimiter
	Appeal      *clients.Appeal

	Board   *clients.Board
	Section *clients.Section
	Card    *clients.Card
	Poll    *clients.Poll

	logger *zerolog.Logger
	conns  []*grpc.ClientConn
}

func NewConnector(app *config.Application, config *config.Services, logger *zerolog.Logger) (*Connector, error) {
	var activeConns []*grpc.ClientConn

	connect := func(addr string, timeout time.Duration, retries int, maxSizeConn int) (*grpc.ClientConn, error) {
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
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(maxSizeConn),
				grpc.MaxCallSendMsgSize(maxSizeConn),
			),
		)

		if err != nil {
			return nil, err
		}
		activeConns = append(activeConns, conn)
		return conn, nil
	}

	userConn, err := connect(config.User.Client.Addr, config.User.Client.TimeOut, config.User.Client.Retries, int(app.MaxUploadImageSize))
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to User service: %w", err)
	}

	configUser := clients.ConfigUser{
		MaxUserAvatarBytesSize: int(app.MaxUploadImageSize),
	}

	authConn, err := connect(config.Auth.Client.Addr, config.Auth.Client.TimeOut, config.Auth.Client.Retries, defaultGRPCMsgSize)
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to Auth service: %w", err)
	}

	mailSenderConn, err := connect(config.MailSender.Addr, config.MailSender.TimeOut, config.MailSender.Retries, defaultGRPCMsgSize)
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to MailSender service: %w", err)
	}

	rateLimiterConn, err := connect(config.RateLimiters.Addr, config.RateLimiters.ClientConfig.TimeOut, config.RateLimiters.ClientConfig.Retries, defaultGRPCMsgSize)
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to RateLimiter service: %w", err)
	}

	boardConn, err := connect(config.Board.Client.Addr, config.Board.Client.TimeOut, config.Board.Client.Retries, int(app.MaxUploadImageSize))
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to Board service: %w", err)
	}

	configBoard := clients.ConfigBoard{
		MaxBackgroundBytesSize: int(app.MaxUploadImageSize),
		ChunkSize:              config.Board.Client.ChunkSize,
	}

	configCard := clients.CardConfig{
		MaxAttachmentBufferSize: int(app.MaxFileSize),
	}

	appealConn, err := connect(config.Appeal.Client.Addr, config.Appeal.Client.TimeOut, config.Appeal.Client.Retries, int(app.MaxFileSize))
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to Appeal service: %w", err)
	}

	configAppeal := clients.ConfigAppeal{
		MaxAppealAttachmentBytesSize: int(app.MaxFileSize),
	}

	return &Connector{
		User:        clients.NewUserClient(userConn, configUser),
		Auth:        clients.NewAuthClient(authConn),
		MailSender:  clients.NewMailSenderClient(mailSenderConn),
		RateLimiter: clients.NewRateLimiterClient(rateLimiterConn),
		Board:       clients.NewBoardClient(boardConn, configBoard),
		Section:     clients.NewSectionClient(boardConn),
		Card:        clients.NewCardClient(boardConn, configCard),
		Poll:        clients.NewPollClient(boardConn),
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
