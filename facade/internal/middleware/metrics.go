package middleware

import (
	"net/http"
	"strconv"
	"time"

	httpMetric "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/metrics/http"
	"github.com/gorilla/mux"
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

			path := "unknown"
			if route := mux.CurrentRoute(r); route != nil {
				if tmpl, err := route.GetPathTemplate(); err == nil {
					path = tmpl
				}
			}

			next.ServeHTTP(rec, r)

			duration := time.Since(start).Seconds()
			statusStrc := strconv.Itoa(rec.statusCode)

			httpMetric.HttpRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
			httpMetric.HttpRequestTotal.WithLabelValues(statusStrc, r.Method, path).Inc()
		})
	}
}
