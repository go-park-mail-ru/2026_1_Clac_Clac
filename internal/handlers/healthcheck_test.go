package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/handlers"
	"github.com/stretchr/testify/require"
)

func TestHealthcheck(t *testing.T) {
	res := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatalf("error when creating request: %v", err)
	}

	h := http.HandlerFunc(handlers.HealthcheckHandler)
	h.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
}
