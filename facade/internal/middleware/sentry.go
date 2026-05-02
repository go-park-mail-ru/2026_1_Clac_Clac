package middleware

import (
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/gorilla/mux"
)

func SentryHubMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hub := sentry.CurrentHub().Clone()

			ctx := sentry.SetHubOnContext(r.Context(), hub)

			hub.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetTag("http.method", r.Method)

				if route := mux.CurrentRoute(r); route != nil {
					if pathTemplate, err := route.GetPathTemplate(); err == nil {
						scope.SetTag("http.route", pathTemplate)
					}
				}

				scope.SetContext("HTTP Request info", map[string]interface{}{
					"client_ip":      r.RemoteAddr,
					"user_agent":     r.UserAgent(),
					"content_length": r.ContentLength,
				})
			})

			defer func() {
				if err := recover(); err != nil {
					hub.Recover(err)

					hub.Flush(sentry.DefaultFlushTimeout)

					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
