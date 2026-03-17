package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	authSrv "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	mockSessionOutput "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware/mock_session_checker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthMiddleware(t *testing.T) {
	mockAuth := mockSessionOutput.NewSessionCheker(t)
	mockAuth.On("GetUserID", mock.Anything, mock.AnythingOfType("string")).Return(common.FixedUserUuiD, nil)

	protectedAuth := AuthMiddleware(mockAuth)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val := r.Context().Value(UserIDKey{})
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
}
