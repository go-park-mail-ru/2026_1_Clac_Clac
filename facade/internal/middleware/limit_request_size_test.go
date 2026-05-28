package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/middleware"
	"github.com/stretchr/testify/assert"
)

var bodyReader = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()
	if _, err := io.ReadAll(r.Body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
})

func TestLimitRequestSize(t *testing.T) {
	tests := []struct {
		name           string
		maxSize        int64
		body           string
		expectedStatus int
	}{
		{
			name:           "UnderLimit",
			maxSize:        64,
			body:           "small body",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "OverLimit",
			maxSize:        16,
			body:           strings.Repeat("x", 100),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "EmptyBody",
			maxSize:        16,
			body:           "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ExactLimit",
			maxSize:        10,
			body:           "1234567890",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.body))
			res := httptest.NewRecorder()

			middleware.LimitRequestSizeMiddleware(tc.maxSize)(bodyReader).ServeHTTP(res, req)

			assert.Equal(t, tc.expectedStatus, res.Code)
		})
	}
}
