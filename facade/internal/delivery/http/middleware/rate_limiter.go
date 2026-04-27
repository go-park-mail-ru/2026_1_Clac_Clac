package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/rs/zerolog"
)

const (
	failUpdateMessage = "fail update count requests"
)

type CheckLimit interface {
	UpdateCountRequests(ctx context.Context, check domain.RateLimitCheck) (bool, error)
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

func RateLimiterMiddleware(client CheckLimit, configRateLimiter domain.RateLimitConfig, logger *zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userIP := GetUserIP(r)
			isFull, err := client.UpdateCountRequests(r.Context(), domain.RateLimitCheck{
				Limit:   configRateLimiter.Limit,
				WindowS: configRateLimiter.WindowS,
				Action:  configRateLimiter.Action,
				UserIP:  userIP,
			})

			if err != nil {
				logger.Warn().Err(err).Msg("fail update count requests")
			} else if isFull {
				api.RespondError(w, http.StatusTooManyRequests, failUpdateMessage)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
