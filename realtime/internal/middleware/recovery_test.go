package middleware_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/middleware"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecoveryMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.HandlerFunc
		expectedStatus int
		expectLog      bool
	}{
		{
			name: "NoPanic",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedStatus: http.StatusOK,
			expectLog:      false,
		},
		{
			name: "Panic",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic("something exploded")
			},
			expectedStatus: http.StatusInternalServerError,
			expectLog:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := zerolog.New(buf)

			req, err := http.NewRequest(http.MethodGet, "/", nil)
			require.NoError(t, err)
			res := httptest.NewRecorder()

			middleware.RecoveryMiddleware(&logger)(tc.handler).ServeHTTP(res, req)

			assert.Equal(t, tc.expectedStatus, res.Code)
			if tc.expectLog {
				assert.NotEmpty(t, buf.Bytes(), "panic must be logged")
			}
		})
	}
}
