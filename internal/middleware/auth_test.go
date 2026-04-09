package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authSrv "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	mockSessionChecker "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware/mock_session_checker"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthMiddleware(t *testing.T) {
	t.Run("without error", func(t *testing.T) {
		mockAuth := mockSessionChecker.NewSessionCheker(t)
		mockAuth.On("GetUserLink", mock.Anything, mock.AnythingOfType("string")).Return(common.FixedUserUuiD, nil)
		mockAuth.On("RefreshSession", mock.Anything, mock.AnythingOfType("string")).Return(nil)

		protectedAuth := AuthMiddleware(mockAuth, zerolog.DefaultContextLogger, 1*time.Minute)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			val := r.Context().Value(UserContextLink{})
			assert.NotNil(t, val, "expected userID in context")
			assert.Equal(t, common.FixedUserUuiD, val, "different userIDs")

			w.WriteHeader(http.StatusOK)
		})

		testHandler := protectedAuth(handler)

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)

		req.AddCookie(&http.Cookie{
			Name:  authSrv.SessiondIdKey,
			Value: common.FixedSessionID,
		})

		testHandler.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "wait status 200")
	})
}

func TestAuthMiddlewareError(t *testing.T) {
	t.Run("not found user", func(t *testing.T) {
		mockAuth := mockSessionChecker.NewSessionCheker(t)
		mockAuth.On("GetUserLink", mock.Anything, mock.AnythingOfType("string")).Return(uuid.Nil, errors.New("session not found"))

		protectedAuth := AuthMiddleware(mockAuth, zerolog.DefaultContextLogger, 1*time.Millisecond)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			val := r.Context().Value(UserContextLink{})
			assert.NotNil(t, val, "expected userID in context")
			assert.Equal(t, common.FixedUserUuiD, val, "different userIDs")

			w.WriteHeader(http.StatusOK)
		})

		testHandler := protectedAuth(handler)

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)

		req.AddCookie(&http.Cookie{
			Name:  authSrv.SessiondIdKey,
			Value: common.FixedSessionID,
		})

		testHandler.ServeHTTP(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code, "wait error 401")
	})
}
