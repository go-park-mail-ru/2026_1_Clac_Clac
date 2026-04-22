package app

import (
	"net/http"

	handlerDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/auth/handler/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/middleware"
	"github.com/rs/zerolog"
)

func wrapWithLimit(auth middleware.CheckLimit, configLimiter handlerDto.RateLimitConfig,
	logger *zerolog.Logger, next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	limiter := middleware.RateLimiterMiddleware(auth, configLimiter, logger)

	return limiter(http.HandlerFunc(next))
}
