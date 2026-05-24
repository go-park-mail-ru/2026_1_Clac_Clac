package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fixedUserLink = uuid.MustParse("11111111-1111-1111-1111-111111111111")

type stubSessionChecker struct {
	checkFn func(ctx context.Context, sessionID string) (uuid.UUID, error)
}

func (s *stubSessionChecker) CheckSession(ctx context.Context, sessionID string) (uuid.UUID, error) {
	return s.checkFn(ctx, sessionID)
}

func newAuthMiddleware(checkFn func(context.Context, string) (uuid.UUID, error)) func(http.Handler) http.Handler {
	logger := zerolog.Nop()
	stub := &stubSessionChecker{checkFn: checkFn}
	return AuthMiddleware(stub, &logger, 24*time.Hour)
}

func TestAuthMiddleware(t *testing.T) {
	okCheck := func(_ context.Context, _ string) (uuid.UUID, error) {
		return fixedUserLink, nil
	}
	failCheck := func(_ context.Context, _ string) (uuid.UUID, error) {
		return uuid.Nil, errors.New("session not found")
	}

	tests := []struct {
		name                string
		cookie              *http.Cookie
		checkFn             func(context.Context, string) (uuid.UUID, error)
		expectedStatus      int
		expectCtxLink       bool
		expectRefreshCookie bool
	}{
		{
			name:           "NoCookie",
			cookie:         nil,
			checkFn:        nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:                "Success",
			cookie:              &http.Cookie{Name: SessionIdKey, Value: "session-abc"},
			checkFn:             okCheck,
			expectedStatus:      http.StatusOK,
			expectCtxLink:       true,
			expectRefreshCookie: true,
		},
		{
			name:           "CheckSessionError",
			cookie:         &http.Cookie{Name: SessionIdKey, Value: "bad-session"},
			checkFn:        failCheck,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mw := newAuthMiddleware(tc.checkFn)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}
			res := httptest.NewRecorder()

			var gotLink uuid.UUID
			mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				val := r.Context().Value(UserContextLink{})
				if link, ok := val.(uuid.UUID); ok {
					gotLink = link
				}
				w.WriteHeader(http.StatusOK)
			})).ServeHTTP(res, req)

			assert.Equal(t, tc.expectedStatus, res.Code)

			if tc.expectCtxLink {
				assert.Equal(t, fixedUserLink, gotLink)
			}

			if tc.expectRefreshCookie {
				var found bool
				for _, c := range res.Result().Cookies() {
					if c.Name == SessionIdKey {
						found = true
						require.Equal(t, tc.cookie.Value, c.Value)
					}
				}
				assert.True(t, found, "session cookie must be refreshed in response")
			}
		})
	}
}
