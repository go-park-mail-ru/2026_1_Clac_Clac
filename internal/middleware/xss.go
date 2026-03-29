package middleware

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"regexp"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
)

var (
	ErrCannotReadRequestBody = errors.New("bad request")
	ErrSuspiciousRequestBody = errors.New("bad request")
)

// Если тело запроса не удовлятворяет хотя бы одному
// правилу, кидаем ошибку
var xssRules = []*regexp.Regexp{
	regexp.MustCompile(`^[^<>]*$`),
}

func XSSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			api.RespondError(w, http.StatusBadRequest, ErrCannotReadRequestBody.Error())
			return
		}

		err = r.Body.Close()
		if err != nil {
			api.RespondError(w, http.StatusBadRequest, ErrCannotReadRequestBody.Error())
			return
		}

		for _, rule := range xssRules {
			if !rule.Match(body) {
				api.RespondError(w, http.StatusBadRequest, ErrSuspiciousRequestBody.Error())
				return
			}
		}

		r.Body = io.NopCloser(bytes.NewReader(body))

		next.ServeHTTP(w, r)
	})
}
