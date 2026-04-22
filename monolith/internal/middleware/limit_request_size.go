package middleware

import (
	"net/http"

	"github.com/gorilla/mux"
)

func LimitRequestSizeMiddleware(maxRequestSize int64) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)
			next.ServeHTTP(w, r)
		})
	}
}
