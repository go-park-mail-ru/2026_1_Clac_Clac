package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestPrometheusMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		handlerStatus  int
		expectedStatus int
	}{
		{
			name:           "GET 200",
			method:         http.MethodGet,
			handlerStatus:  http.StatusOK,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST 201",
			method:         http.MethodPost,
			handlerStatus:  http.StatusCreated,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "DELETE 404",
			method:         http.MethodDelete,
			handlerStatus:  http.StatusNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "PUT 500",
			method:         http.MethodPut,
			handlerStatus:  http.StatusInternalServerError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.handlerStatus)
			})

			req := httptest.NewRequest(tc.method, "/test", nil)
			rec := httptest.NewRecorder()

			middleware.PrometheusMiddleware()(handler).ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Code)
		})
	}
}

func TestPrometheusMiddlewareDefaultStatus(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// WriteHeader не вызывается явно — должен быть 200 по умолчанию
		_, _ = w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	middleware.PrometheusMiddleware()(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestPrometheusMiddlewarePassThrough(t *testing.T) {
	const wantBody = "hello"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(wantBody))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	middleware.PrometheusMiddleware()(handler).ServeHTTP(rec, req)

	assert.Equal(t, wantBody, rec.Body.String())
}
