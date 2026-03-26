package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

// Middleware для пробрасывания логгера в запрос.
func LoggerMiddleware(logger *zerolog.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestLogger := logger.With().Str("reqId", uuid.New().String()).Logger()
			ctx := requestLogger.WithContext(r.Context())

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
