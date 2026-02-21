package engine

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type Engine struct {
	server *http.Server
	router *mux.Router
	logger *zerolog.Logger
}

func New(logger *zerolog.Logger) *Engine {
	engine := &Engine{
		router: mux.NewRouter(),
		logger: logger,
	}

	// Установка middleware
	engine.Use(engine.recoveryMiddleware)
	engine.Use(engine.loggerMiddleware)

	return engine
}

func (e *Engine) Start(addr string) {
	e.server = &http.Server{
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      e.router,
	}

	// Graceful shutdown
	go func() {
		e.logger.Info().Msg("server started")

		if err := e.server.ListenAndServe(); err != nil {
			e.logger.Error().Err(err).Msg("error when shutdown the http.Server")
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM) // SIGTERM нужен для Docker
	defer stop()

	<-ctx.Done()

	e.logger.Info().Msg("shutdown started, wait...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	e.server.Shutdown(ctx)

	e.logger.Info().Msg("server stopped")
}

// Добавления обработчика GET запросов
func (e *Engine) GET(path string, handler func(w http.ResponseWriter, r *http.Request)) {
	e.router.HandleFunc(path, handler).Methods(http.MethodGet)
}

// Добавления обработчика POST запросов
func (e *Engine) POST(path string, handler func(w http.ResponseWriter, r *http.Request)) {
	e.router.HandleFunc(path, handler).Methods(http.MethodPost)
}

// Метод для добавления своих Middleware
func (e *Engine) Use(middleware ...mux.MiddlewareFunc) {
	e.router.Use(middleware...)
}

// Middleware для пробрасывания логгера в запрос.
// Чтобы получить логгер в обработчике, используйте engine.GetLoggerFromRequest
func (e *Engine) loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := e.logger.WithContext(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Middleware для отлова паники
func (e *Engine) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				e.logger.Error().Interface("panic", err).Msg("it's toast")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
