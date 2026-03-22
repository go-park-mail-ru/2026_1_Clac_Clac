package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	fastTime = time.Millisecond * 50
)

func TestTimeOutMoiddleware(t *testing.T) {
	t.Run("handler without error", func(t *testing.T) {
		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		var newContext context.Context

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			newContext = r.Context()

			select {
			case <-newContext.Done():
			case <-time.After(fastTime * 2):
			}
		})

		timeOutActive := TimeOutMiddleware(fastTime)
		handlerWithTimeOut := timeOutActive(handler)

		handlerWithTimeOut.ServeHTTP(res, req)

		assert.ErrorIs(t, newContext.Err(), context.DeadlineExceeded, "wait context is done")
	})
}
