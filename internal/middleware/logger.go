package middleware

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type LoggerResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *LoggerResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func LoggerMiddleware(logger *zerolog.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestStart := time.Now()

			requestLogger := logger.With().Str("request-id", uuid.New().String()).Logger()
			ctx := requestLogger.WithContext(r.Context())

			responseWriter := &LoggerResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			next.ServeHTTP(responseWriter, r.WithContext(ctx))

			requestLogger.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", responseWriter.statusCode).
				Dur("latency", time.Since(requestStart)).
				Msg("request processed")
		})
	}
}
