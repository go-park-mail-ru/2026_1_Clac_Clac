package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

const (
	xRealIPHeader       = "X-Real-IP"
	xForwardedForHeader = "X-Forwarded-For"
)

var (
	ErrInteralServerError = errors.New("something went wrong")
)

type LoggerResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *LoggerResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func GetRealIP(r *http.Request) string {
	if ip := r.Header.Get(xRealIPHeader); ip != "" {
		return ip
	}

	if ip := r.Header.Get(xForwardedForHeader); ip != "" {
		return strings.Split(ip, ",")[0]
	}

	return r.RemoteAddr
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

			requestDuration := time.Since(requestStart)

			// TODO: решить вопрос с request.Body
			requestLogger.Info().
				Str("method", r.Method).
				Str("remote_addr", r.RemoteAddr).
				Str("url", r.URL.Path).
				Dur("work_time", requestDuration).
				Int("status", responseWriter.statusCode).
				Str("user_agent", r.UserAgent()).
				Str("host", r.Host).
				Str("real_ip", GetRealIP(r)).
				Int64("content_length", r.ContentLength).
				Str("start_time", requestStart.Format(time.RFC3339)).
				Str("duration_human", requestDuration.String()).
				Int64("duration_ms", requestDuration.Milliseconds()).
				Msg("request processed")
		})
	}
}
