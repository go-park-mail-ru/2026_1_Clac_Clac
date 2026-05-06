package middleware

import (
	"net/http"
	"strconv"
	"time"

	httpMetric "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/metrics/http"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

func PrometheusMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rec := &statusRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rec, r)

			duration := time.Since(start).Seconds()
			statusStrc := strconv.Itoa(rec.statusCode)

			httpMetric.HttpRequestDuration.WithLabelValues(r.Method).Observe(duration)
			httpMetric.HttpRequestTotal.WithLabelValues(statusStrc, r.Method).Inc()
		})
	}
}
