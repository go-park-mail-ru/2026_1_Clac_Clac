package middleware

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

// Middleware для отлова паники.
func RecoveryMiddleware(logger *zerolog.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error().Interface("panic", err).Msg("it's toast")
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
