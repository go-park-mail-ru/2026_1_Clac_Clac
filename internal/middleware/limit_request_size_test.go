package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestLimitRequestSizeMiddleware(t *testing.T) {
	reader := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		_, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	t.Run("limit did not reach", func(t *testing.T) {
		const maxRequestSize = 16 // 16 байт

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)

		m := middleware.LimitRequestSizeMiddleware(maxRequestSize)
		h := m(reader)

		h.ServeHTTP(res, req)

		r := res.Result()

		assert.Equal(t, http.StatusOK, r.StatusCode, "must return 200")
	})

	t.Run("reach limit", func(t *testing.T) {
		const maxRequestSize = 16 // 16 байт

		res := httptest.NewRecorder()

		requestBody := strings.NewReader("hello, world! hello, world! hello, world! hello, world!")
		req := httptest.NewRequest(http.MethodGet, "/", requestBody)

		m := middleware.LimitRequestSizeMiddleware(maxRequestSize)
		h := m(reader)

		h.ServeHTTP(res, req)

		r := res.Result()

		assert.Equal(t, http.StatusBadRequest, r.StatusCode, "must return 200")
	})
}
