package middleware

import (
	"context"
	"net/http"
	"time"
)

func TimeOutMiddleware(timeOut time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			timeOutCtx, cancel := context.WithTimeout(r.Context(), timeOut)
			defer cancel()

			contextWithTimeOut := r.WithContext(timeOutCtx)

			next.ServeHTTP(w, contextWithTimeOut)
		})
	}
}
