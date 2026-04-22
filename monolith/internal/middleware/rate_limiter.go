package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/api"
	handlerDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/auth/handler/dto"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/auth/service/dto"
	"github.com/rs/zerolog"
)

const (
	failUpdateMessage = "fail update count requests"
)

type CheckLimit interface {
	UpdateCountRequests(ctx context.Context, configLimiter serviceDto.RateLimiterConfig) (bool, error)
}

func GetUserIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	if xRealIP := r.Header.Get("X-Real-IP"); xRealIP != "" {
		return xRealIP
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

func RateLimiterMiddleware(srv CheckLimit, configRateLimiter handlerDto.RateLimitConfig, logger *zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userIP := GetUserIP(r)

			isFull, err := srv.UpdateCountRequests(r.Context(), serviceDto.RateLimiterConfig{
				Limit:  configRateLimiter.Limit,
				Window: configRateLimiter.Window,
				Action: configRateLimiter.Action,
				UserIP: userIP,
			})

			if err != nil {
				logger.Warn().Err(err).Msg("fail Redis update count requests")
			} else if isFull {
				api.RespondError(w, http.StatusTooManyRequests, failUpdateMessage)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
