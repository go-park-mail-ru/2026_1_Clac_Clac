package app

import (
	"fmt"

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

	Board   *clients.Board
	Section *clients.Section
	Card    *clients.Card

	logger *zerolog.Logger
	conns  []*grpc.ClientConn
}

func NewConnector(config *config.Services, logger *zerolog.Logger) (*Connector, error) {
	var activeConns []*grpc.ClientConn

	connect := func(addr string) (*grpc.ClientConn, error) {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		activeConns = append(activeConns, conn)
		return conn, nil
	}

	userConn, err := connect(config.User.Client.Addr)
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to User service: %w", err)
	}

	authConn, err := connect(config.Auth.Client.Addr)
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to Auth service: %w", err)
	}

	mailSenderConn, err := connect(config.MailSender.Addr)
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to MailSender service: %w", err)
	}

	rateLimiterConn, err := connect(config.RateLimiters.Addr)
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to RateLimiter service: %w", err)
	}

	boardConn, err := connect(config.Board.Client.Addr)
	if err != nil {
		closeAll(activeConns, logger)
		return nil, fmt.Errorf("failed to connect to Board service: %w", err)
	}

	return &Connector{
		User:        clients.NewUserClient(userConn),
		Auth:        clients.NewAuthClient(authConn),
		MailSender:  clients.NewMailSenderClient(mailSenderConn),
		RateLimiter: clients.NewRateLimiterClient(rateLimiterConn),
		Board:       clients.NewBoardClient(boardConn),
		Section:     clients.NewSectionClient(boardConn),
		Card:        clients.NewCardClient(boardConn),
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
