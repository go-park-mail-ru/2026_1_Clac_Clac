package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const fastTimeout = 50 * time.Millisecond

func TestTimeOutMiddleware(t *testing.T) {
	type ctxKey struct{}

	tests := []struct {
		name           string
		handler        http.HandlerFunc
		requestCtx     context.Context
		expectedStatus int
		checkDeadline  bool
	}{
		{
			name: "FastHandler",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			requestCtx:     context.Background(),
			expectedStatus: http.StatusOK,
		},
		{
			name: "ContextPropagated",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Context().Value(ctxKey{}) != "test-value" {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
			},
			requestCtx:     context.WithValue(context.Background(), ctxKey{}, "test-value"),
			expectedStatus: http.StatusOK,
		},
		{
			name: "ContextCancelledAfterTimeout",
			handler: func(w http.ResponseWriter, r *http.Request) {
				select {
				case <-r.Context().Done():
				case <-time.After(fastTimeout * 3):
				}
			},
			requestCtx:    context.Background(),
			checkDeadline: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var capturedCtx context.Context

			wrapped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedCtx = r.Context()
				tc.handler(w, r)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(tc.requestCtx)
			res := httptest.NewRecorder()

			TimeOutMiddleware(fastTimeout)(wrapped).ServeHTTP(res, req)

			if tc.checkDeadline {
				assert.ErrorIs(t, capturedCtx.Err(), context.DeadlineExceeded)
			} else {
				assert.Equal(t, tc.expectedStatus, res.Code)
			}
		})
	}
}
