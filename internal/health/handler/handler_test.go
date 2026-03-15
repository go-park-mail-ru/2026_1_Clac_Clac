package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHealthcheck(t *testing.T) {
	t.Run("ok status", func(t *testing.T) {
		res := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)

		require.NoError(t, err, "cannot create request")

		h := http.HandlerFunc(HealthcheckHandler)
		h.ServeHTTP(res, req)

		require.Equal(t, http.StatusOK, res.Code)
	})
}
