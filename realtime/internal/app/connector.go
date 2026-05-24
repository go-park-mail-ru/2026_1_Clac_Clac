package app

import (
	"fmt"
	"time"

	sentrygrpc "github.com/getsentry/sentry-go/grpc"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/clients"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/config"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Connector struct {
	Auth   *clients.Auth
	Board  *clients.Board
	logger *zerolog.Logger
	conns  []*grpc.ClientConn
}

func NewConnector(app *config.Application, services *config.Services, logger *zerolog.Logger) (*Connector, error) {
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

	msgSize := int(app.MaxMessageSize)

	authConn, err := connect(services.Auth.Client.Addr, services.Auth.Client.TimeOut, services.Auth.Client.Retries, msgSize)
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to Auth service: %w", err)
	}

	boardConn, err := connect(services.Board.Client.Addr, services.Board.Client.TimeOut, services.Board.Client.Retries, msgSize)
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to Board service: %w", err)
	}

	return &Connector{
		Auth:   clients.NewAuthClient(authConn),
		Board:  clients.NewBoardClient(boardConn),
		logger: logger,
		conns:  activeConns,
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
