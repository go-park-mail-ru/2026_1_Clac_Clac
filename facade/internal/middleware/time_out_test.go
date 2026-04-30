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

func TestTimeOutMiddlewareContextCancelled(t *testing.T) {
	var capturedCtx context.Context

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
		select {
		case <-capturedCtx.Done():
		case <-time.After(fastTimeout * 3):
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()

	TimeOutMiddleware(fastTimeout)(h).ServeHTTP(res, req)

	assert.ErrorIs(t, capturedCtx.Err(), context.DeadlineExceeded)
}

func TestTimeOutMiddlewareFastHandler(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()

	TimeOutMiddleware(fastTimeout)(h).ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
}

func TestTimeOutMiddlewareContextPropagated(t *testing.T) {
	type key struct{}
	ctx := context.WithValue(context.Background(), key{}, "test-value")

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val := r.Context().Value(key{})
		if val != "test-value" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	res := httptest.NewRecorder()

	TimeOutMiddleware(fastTimeout)(h).ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
}
