package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/middleware"
	"github.com/stretchr/testify/assert"
)

var bodyReader = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if _, err := io.ReadAll(r.Body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
})

func TestLimitRequestSizeUnderLimit(t *testing.T) {
	const maxSize int64 = 64

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("small body"))
	res := httptest.NewRecorder()

	middleware.LimitRequestSizeMiddleware(maxSize)(bodyReader).ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
}

func TestLimitRequestSizeOverLimit(t *testing.T) {
	const maxSize int64 = 16

	big := strings.Repeat("x", 100)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(big))
	res := httptest.NewRecorder()

	middleware.LimitRequestSizeMiddleware(maxSize)(bodyReader).ServeHTTP(res, req)

	assert.Equal(t, http.StatusBadRequest, res.Code)
}

func TestLimitRequestSizeEmptyBody(t *testing.T) {
	const maxSize int64 = 16

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	res := httptest.NewRecorder()

	middleware.LimitRequestSizeMiddleware(maxSize)(bodyReader).ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
}

func TestLimitRequestSizeExactLimit(t *testing.T) {
	const maxSize int64 = 10

	body := strings.NewReader("1234567890")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	res := httptest.NewRecorder()

	middleware.LimitRequestSizeMiddleware(maxSize)(bodyReader).ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
}
