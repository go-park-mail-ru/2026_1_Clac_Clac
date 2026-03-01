package engine

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

type Engine struct {
	config *config.Engine
	server *http.Server
	router *mux.Router
	logger *zerolog.Logger
	// Хук, вызывается когда порт успешно открыт
	OnListen func(addr string)
}

func New(config *config.Engine, logger *zerolog.Logger, router *mux.Router) *Engine {
	return &Engine{
		config: config,
		logger: logger,
		router: router,
	}
}

func (e *Engine) Start(ctx context.Context) error {
	// Запрашиваем порт у ОС через net.Listen
	ln, err := net.Listen("tcp", e.config.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", e.config.Addr, err)
	}

	actualAddr := ln.Addr().String()

	e.server = &http.Server{
		Addr:         actualAddr,
		WriteTimeout: time.Duration(e.config.WriteTimeout) * time.Second,
		ReadTimeout:  time.Duration(e.config.ReadTimeout) * time.Second,
		IdleTimeout:  time.Duration(e.config.IdleTimeout) * time.Second,
		Handler:      e.router,
	}

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
		if err := e.server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(e.config.GracefulShutdownTimeout)*time.Second)
		defer cancel()

		if err := e.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("error while graceful shutdown: %w", err)
		}

		return nil
	}
}
