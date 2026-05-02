package enginegrpc

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type Engine struct {
	config        Config
	logger       *zerolog.Logger
	Server       *grpc.Server
	ServerOptions []grpc.ServerOption
	OnListen     func(addr string)
}

func New(config Config, logger *zerolog.Logger, opts ...grpc.ServerOption) *Engine {
	return &Engine{
		config:        config,
		logger:        logger,
		Server:       grpc.NewServer(opts...),
		ServerOptions: opts,
	}
}

func (e *Engine) Start(ctx context.Context) error {
	// Запрашиваем порт у ОС через net.Listen
	ln, err := net.Listen("tcp", e.config.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", e.config.Addr, err)
	}

	actualAddr := ln.Addr().String()

	// Вызываем хук, когда удалось получить адрес
	if e.OnListen != nil {
		e.OnListen(actualAddr)
	}

	e.logger.Info().Str("addr", ln.Addr().String()).Msg("server started")

	// Перехватываем сигналы, чтобы программа не завершилась сразу
	interruptCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM) // SIGTERM нужен для Docker
	defer stop()

	// errgroup, чтобы вытащить ошибку из горутины
	g, gCtx := errgroup.WithContext(interruptCtx)

	// Graceful shutdown
	g.Go(e.gracefulShutdown(gCtx))
	// Если вызывать Server.Serve вне гоуртины, то не получится перехватить прерывание
	g.Go(func() error {
		if err := e.Server.Serve(ln); err != nil {
			return fmt.Errorf("error while serving: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		e.logger.Error().Err(err).Msg("engine.Start")
		return fmt.Errorf("engine.Start: %w", err)
	}

	e.logger.Info().Msg("stopped")
	return nil
}

func (e *Engine) gracefulShutdown(errgroupCtx context.Context) func() error {
	return func() error {
		<-errgroupCtx.Done()

		e.logger.Info().Msg("shutdown, wait...")

		stopped := make(chan struct{})

		go func() {
			e.Server.GracefulStop()
			close(stopped)
		}()

		timeout := time.Duration(e.config.GracefulShutdownTimeout) * time.Second
		timer := time.NewTimer(timeout)

		select {
		case <-stopped:
			e.logger.Info().Msg("server gracefully stopped all active connections")
		case <-timer.C:
			e.logger.Warn().Msgf("graceful shutdown timed out after %v", timeout)
			e.Server.Stop()
		}

		return nil
	}
}
