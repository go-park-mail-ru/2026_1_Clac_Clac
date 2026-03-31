package middleware

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"regexp"
	"sync"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/rs/zerolog"
)

var (
	ErrCannotReadRequestBody  = errors.New("cannot read request body")
	ErrCannotCloseRequestBody = errors.New("cannot close request body")
	ErrSuspiciousRequestBody  = errors.New("suspicious request body")
)

// Если тело запроса не удовлятворяет хотя бы одному
// правилу, кидаем ошибку
var xssRules = []*regexp.Regexp{
	regexp.MustCompile(`^[^<>]*$`),
}

func XSSMiddleware(next http.Handler) http.Handler {
	buffersPool := sync.Pool{
		New: func() any {
			return new(bytes.Buffer)
		},
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := zerolog.Ctx(r.Context())

		buf := buffersPool.Get().(*bytes.Buffer)
		defer func() {
			buf.Reset()
			buffersPool.Put(buf)
		}()

		defer func(requestBody io.ReadCloser) {
			if err := requestBody.Close(); err != nil {
				logger.Warn().Err(err).Msg("xss middleware")
			}
		}(r.Body)

		if _, err := buf.ReadFrom(r.Body); err != nil {
			api.RespondError(w, http.StatusBadRequest, ErrCannotReadRequestBody.Error())
			return
		}

		// Если в хендлере вызовется горутина,
		// которая будет читать тело запроса,
		// то все сломается. Но у нас такого нет
		body := buf.Bytes()

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
