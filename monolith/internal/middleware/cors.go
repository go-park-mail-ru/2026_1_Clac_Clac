package middleware

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/config"
	"github.com/gorilla/mux"
)

const (
	corsCredentialsHeader = "Access-Control-Allow-Credentials"
	corsOriginHeader      = "Access-Control-Allow-Origin"
	corsMethodsHeader     = "Access-Control-Allow-Methods"
	corsHeadersHeader     = "Access-Control-Allow-Headers"
	corsMaxAgeHeader      = "Access-Control-Max-Age"
)

func CORSMiddleware(conf *config.CORS) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(corsCredentialsHeader, conf.Credentials)
			w.Header().Set(corsOriginHeader, conf.Origin)

			if r.Method == http.MethodOptions {
				w.Header().Set(corsMethodsHeader, conf.Methods)
				w.Header().Set(corsHeadersHeader, conf.Headers)
				w.Header().Set(corsMaxAgeHeader, conf.MaxAge)

				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
