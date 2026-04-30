package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/dto"
	handlerCommon "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers/common"
	mockAuthUC "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers/mock_auth_use_case"
	mockUserUC "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers/mock_user_use_case"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/middleware"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var defaultAuthCfg = AuthConfig{
	MaxLenPassword:        128,
	MinLenPassword:        8,
	SessionLifetime:       24 * time.Hour,
	VKOAuthRedirectTo:     "http://localhost/oauth",
	CoolDownExpirationSec: 60,
}

var fixedLink = uuid.New()

func newAuthHandler(auth AuthUsecase, user UserUsecase) *Auth {
	return NewAuthHandler(auth, user, defaultAuthCfg)
}

func TestMeHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, fixedLink)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		newAuthHandler(nil, nil).MeHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NotEmpty(t, rr.Body.String())
	})

	t.Run("Unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		rr := httptest.NewRecorder()

		newAuthHandler(nil, nil).MeHandler(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), handlerCommon.ErrUserNotAuthorized.Error()) // Проверяем наличие текста ошибки
	})
}

func TestLogInUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		authMock := mockAuthUC.NewAuthUsecase(t)
		userMock := mockUserUC.NewUserUsecase(t)

		reqData := dto.LogInRequest{Email: "test@mail.ru", Password: "Password123"}
		expectedUser := domain.FullInfoUser{UserLink: fixedLink, Email: reqData.Email, DisplayName: "Test User"}

		userMock.On("GetUser", mock.Anything, domain.Credentials{
			Email:    reqData.Email,
			Password: reqData.Password,
		}).Return(expectedUser, nil)

		authMock.On("CreateSession", mock.Anything, fixedLink).Return("session_token", nil)

		body, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		newAuthHandler(authMock, userMock).LogInUser(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		cookies := rr.Result().Cookies()
		require.Len(t, cookies, 1)
		assert.Equal(t, middleware.SessiondIdKey, cookies[0].Name)
		assert.Equal(t, "session_token", cookies[0].Value)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte("{bad json}")))
		rr := httptest.NewRecorder()

		newAuthHandler(nil, nil).LogInUser(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("WrongCredentials", func(t *testing.T) {
		userMock := mockUserUC.NewUserUsecase(t)
		reqData := dto.LogInRequest{Email: "test@mail.ru", Password: "Password123"}

		userMock.On("GetUser", mock.Anything, mock.Anything).Return(domain.FullInfoUser{}, common.ErrorWrongCredentials)

		body, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		newAuthHandler(nil, userMock).LogInUser(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestRegisterUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		authMock := mockAuthUC.NewAuthUsecase(t)
		userMock := mockUserUC.NewUserUsecase(t)

		reqData := dto.RegisterRequest{
			Email:            "test@mail.ru",
			Password:         "Password123",
			RepeatedPassword: "Password123",
			DisplayName:      "Test User",
		}
		expectedUser := domain.FullInfoUser{UserLink: fixedLink, Email: reqData.Email, DisplayName: reqData.DisplayName}

		userMock.On("CreateUser", mock.Anything, mock.Anything).Return(expectedUser, nil)
		authMock.On("CreateSession", mock.Anything, fixedLink).Return("session_token", nil)

		body, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		newAuthHandler(authMock, userMock).RegisterUser(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		cookies := rr.Result().Cookies()
		require.Len(t, cookies, 1)
		assert.Equal(t, middleware.SessiondIdKey, cookies[0].Name)
	})

	t.Run("UserAlreadyExists", func(t *testing.T) {
		userMock := mockUserUC.NewUserUsecase(t)
		reqData := dto.RegisterRequest{
			Email:            "test@mail.ru",
			Password:         "Password123",
			RepeatedPassword: "Password123",
			DisplayName:      "Test User",
		}

		userMock.On("CreateUser", mock.Anything, mock.Anything).Return(domain.FullInfoUser{}, common.ErrorExistingUser)

		body, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		newAuthHandler(nil, userMock).RegisterUser(rr, req)
		assert.Equal(t, http.StatusConflict, rr.Code)
	})

	t.Run("PasswordsMismatch", func(t *testing.T) {
		reqData := dto.RegisterRequest{
			Email:            "test@mail.ru",
			Password:         "Password123",
			RepeatedPassword: "Password321",
			DisplayName:      "Test User",
		}

		body, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		newAuthHandler(nil, nil).RegisterUser(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestLogOutUser(t *testing.T) {
	t.Run("SuccessWithCookie", func(t *testing.T) {
		authMock := mockAuthUC.NewAuthUsecase(t)
		authMock.On("DeleteSession", mock.Anything, "valid_session").Return(nil)

		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		req.AddCookie(&http.Cookie{Name: middleware.SessiondIdKey, Value: "valid_session"})
		rr := httptest.NewRecorder()

		newAuthHandler(authMock, nil).LogOutUser(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		cookies := rr.Result().Cookies()
		require.Len(t, cookies, 1)
		// Проверяем, что кука инвалидирована (max-age < 0)
		assert.Equal(t, middleware.SessiondIdKey, cookies[0].Name)
		assert.True(t, cookies[0].MaxAge < 0 || cookies[0].Expires.Before(time.Now()))
	})

	t.Run("WithoutCookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		rr := httptest.NewRecorder()

		newAuthHandler(nil, nil).LogOutUser(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestVkOAuthCallback(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		authMock := mockAuthUC.NewAuthUsecase(t)
		userMock := mockUserUC.NewUserUsecase(t)

		authMock.On("ExchangeVKCode", mock.Anything, "valid_code").Return("vk_token", "test@vk.com", nil)
		userMock.On("ProcessUserWithVK", mock.Anything, "vk_token", "test@vk.com").Return(fixedLink, nil)
		authMock.On("CreateSession", mock.Anything, fixedLink).Return("session_token", nil)

		req := httptest.NewRequest(http.MethodGet, "/oauth/vk?code=valid_code", nil)
		rr := httptest.NewRecorder()

		newAuthHandler(authMock, userMock).VkOAuthCallback(rr, req)

		// Проверяем, что кука была установлена перед редиректом
		cookies := rr.Result().Cookies()
		require.Len(t, cookies, 1)
		assert.Equal(t, middleware.SessiondIdKey, cookies[0].Name)
		assert.Equal(t, "session_token", cookies[0].Value)
	})

	t.Run("EmptyCode", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/oauth/vk", nil)
		rr := httptest.NewRecorder()

		newAuthHandler(nil, nil).VkOAuthCallback(rr, req)

		// Проверяем вызов внутренней ошибки (Redirect должен отработать, куки нет)
		cookies := rr.Result().Cookies()
		assert.Len(t, cookies, 0)
	})

	t.Run("ExchangeVKCodeFailed", func(t *testing.T) {
		authMock := mockAuthUC.NewAuthUsecase(t)

		authMock.On("ExchangeVKCode", mock.Anything, "invalid_code").Return("", "", common.ErrorVKOAuthUnavailable)

		req := httptest.NewRequest(http.MethodGet, "/oauth/vk?code=invalid_code", nil)
		rr := httptest.NewRecorder()

		newAuthHandler(authMock, nil).VkOAuthCallback(rr, req)

		cookies := rr.Result().Cookies()
		assert.Len(t, cookies, 0)
	})
}
