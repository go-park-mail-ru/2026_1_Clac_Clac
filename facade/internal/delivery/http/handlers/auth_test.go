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
	MaxLenPassword:    128,
	MinLenPassword:    8,
	MaxLenNameUser:    128,
	SessionLifetime:   24 * time.Hour,
	VKOAuthRedirectTo: "http://localhost/oauth",
}

var fixedLink = uuid.New()

func newAuthHandler(auth AuthUsecase, user UserUsecase) *Auth {
	return NewAuthHandler(auth, user, defaultAuthCfg)
}

func TestMeHandler(t *testing.T) {
	expectedUser := domain.FullInfoUser{
		UserLink:    fixedLink,
		Email:       "test@mail.ru",
		DisplayName: "Test User",
		Description: "Test Description",
		AvatarURL:   "http://example.com/avatar.jpg",
	}

	tests := []struct {
		name               string
		setContext         bool
		mockBehavior       func(userMock *mockUserUC.UserUsecase)
		expectedStatusCode int
		expectedContains   string
	}{
		{
			name:       "Success",
			setContext: true,
			mockBehavior: func(userMock *mockUserUC.UserUsecase) {
				userMock.On("GetProfile", mock.Anything, fixedLink).Return(expectedUser, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedContains:   "test@mail.ru",
		},
		{
			name:               "Unauthorized (No Context)",
			setContext:         false,
			mockBehavior:       func(userMock *mockUserUC.UserUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
			expectedContains:   handlerCommon.ErrUserNotAuthorized.Error(),
		},
		{
			name:       "Unauthorized (User NotFound)",
			setContext: true,
			mockBehavior: func(userMock *mockUserUC.UserUsecase) {
				userMock.On("GetProfile", mock.Anything, fixedLink).Return(domain.FullInfoUser{}, common.ErrorNonexistentUser)
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedContains:   handlerCommon.ErrUserNotAuthorized.Error(),
		},
		{
			name:       "Internal Server Error",
			setContext: true,
			mockBehavior: func(userMock *mockUserUC.UserUsecase) {
				userMock.On("GetProfile", mock.Anything, fixedLink).Return(domain.FullInfoUser{}, assert.AnError)
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedContains:   handlerCommon.ErrInternalServerError.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			userMock := mockUserUC.NewUserUsecase(t)
			tc.mockBehavior(userMock)

			req := httptest.NewRequest(http.MethodGet, "/me", nil)
			if tc.setContext {
				ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, fixedLink)
				req = req.WithContext(ctx)
			}
			rr := httptest.NewRecorder()

			newAuthHandler(nil, userMock).MeHandler(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
			if tc.expectedContains != "" {
				assert.Contains(t, rr.Body.String(), tc.expectedContains)
			} else {
				assert.NotEmpty(t, rr.Body.String())
			}
		})
	}
}

func TestLogInUser(t *testing.T) {
	reqData := dto.LogInRequest{Email: "test@mail.ru", Password: "Password123"}
	expectedUser := domain.FullInfoUser{UserLink: fixedLink, Email: reqData.Email, DisplayName: "Test User"}

	tests := []struct {
		name               string
		requestBody        any
		mockBehavior       func(authMock *mockAuthUC.AuthUsecase, userMock *mockUserUC.UserUsecase)
		expectedStatusCode int
		expectCookie       bool
	}{
		{
			name:        "Success",
			requestBody: reqData,
			mockBehavior: func(authMock *mockAuthUC.AuthUsecase, userMock *mockUserUC.UserUsecase) {
				userMock.On("GetUser", mock.Anything, domain.Credentials{
					Email:    reqData.Email,
					Password: reqData.Password,
				}).Return(expectedUser, nil)
				authMock.On("CreateSession", mock.Anything, fixedLink).Return("session_token", nil)
			},
			expectedStatusCode: http.StatusOK,
			expectCookie:       true,
		},
		{
			name:               "InvalidJSON",
			requestBody:        "{bad json}",
			mockBehavior:       func(authMock *mockAuthUC.AuthUsecase, userMock *mockUserUC.UserUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
			expectCookie:       false,
		},
		{
			name:        "WrongCredentials",
			requestBody: reqData,
			mockBehavior: func(authMock *mockAuthUC.AuthUsecase, userMock *mockUserUC.UserUsecase) {
				userMock.On("GetUser", mock.Anything, mock.Anything).Return(domain.FullInfoUser{}, common.ErrorWrongCredentials)
			},
			expectedStatusCode: http.StatusNotFound,
			expectCookie:       false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			authMock := mockAuthUC.NewAuthUsecase(t)
			userMock := mockUserUC.NewUserUsecase(t)
			tc.mockBehavior(authMock, userMock)

			var bodyReader *bytes.Reader
			if strBody, ok := tc.requestBody.(string); ok {
				bodyReader = bytes.NewReader([]byte(strBody))
			} else {
				b, _ := json.Marshal(tc.requestBody)
				bodyReader = bytes.NewReader(b)
			}

			req := httptest.NewRequest(http.MethodPost, "/login", bodyReader)
			rr := httptest.NewRecorder()

			newAuthHandler(authMock, userMock).LogInUser(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)

			cookies := rr.Result().Cookies()
			if tc.expectCookie {
				require.Len(t, cookies, 1)
				assert.Equal(t, middleware.SessiondIdKey, cookies[0].Name)
				assert.Equal(t, "session_token", cookies[0].Value)
			} else {
				assert.Len(t, cookies, 0)
			}
		})
	}
}

func TestRegisterUser(t *testing.T) {
	reqData := dto.RegisterRequest{
		Email:            "test@mail.ru",
		Password:         "Password123",
		RepeatedPassword: "Password123",
		DisplayName:      "Test User",
	}
	expectedUser := domain.FullInfoUser{UserLink: fixedLink, Email: reqData.Email, DisplayName: reqData.DisplayName}

	tests := []struct {
		name               string
		requestBody        any
		mockBehavior       func(authMock *mockAuthUC.AuthUsecase, userMock *mockUserUC.UserUsecase)
		expectedStatusCode int
		expectCookie       bool
	}{
		{
			name:        "Success",
			requestBody: reqData,
			mockBehavior: func(authMock *mockAuthUC.AuthUsecase, userMock *mockUserUC.UserUsecase) {
				userMock.On("CreateUser", mock.Anything, mock.Anything).Return(expectedUser, nil)
				authMock.On("CreateSession", mock.Anything, fixedLink).Return("session_token", nil)
			},
			expectedStatusCode: http.StatusCreated,
			expectCookie:       true,
		},
		{
			name:        "UserAlreadyExists",
			requestBody: reqData,
			mockBehavior: func(authMock *mockAuthUC.AuthUsecase, userMock *mockUserUC.UserUsecase) {
				userMock.On("CreateUser", mock.Anything, mock.Anything).Return(domain.FullInfoUser{}, common.ErrorExistingUser)
			},
			expectedStatusCode: http.StatusConflict,
			expectCookie:       false,
		},
		{
			name: "PasswordsMismatch",
			requestBody: dto.RegisterRequest{
				Email:            "test@mail.ru",
				Password:         "Password123",
				RepeatedPassword: "Password321",
				DisplayName:      "Test User",
			},
			mockBehavior:       func(authMock *mockAuthUC.AuthUsecase, userMock *mockUserUC.UserUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
			expectCookie:       false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			authMock := mockAuthUC.NewAuthUsecase(t)
			userMock := mockUserUC.NewUserUsecase(t)
			tc.mockBehavior(authMock, userMock)

			b, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(b))
			rr := httptest.NewRecorder()

			newAuthHandler(authMock, userMock).RegisterUser(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)

			cookies := rr.Result().Cookies()
			if tc.expectCookie {
				require.Len(t, cookies, 1)
				assert.Equal(t, middleware.SessiondIdKey, cookies[0].Name)
			} else {
				assert.Len(t, cookies, 0)
			}
		})
	}
}

func TestLogOutUser(t *testing.T) {
	tests := []struct {
		name         string
		cookie       *http.Cookie
		mockBehavior func(authMock *mockAuthUC.AuthUsecase)
	}{
		{
			name:   "SuccessWithCookie",
			cookie: &http.Cookie{Name: middleware.SessiondIdKey, Value: "valid_session"},
			mockBehavior: func(authMock *mockAuthUC.AuthUsecase) {
				authMock.On("DeleteSession", mock.Anything, "valid_session").Return(nil)
			},
		},
		{
			name:         "WithoutCookie",
			cookie:       nil,
			mockBehavior: func(authMock *mockAuthUC.AuthUsecase) {},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			authMock := mockAuthUC.NewAuthUsecase(t)
			tc.mockBehavior(authMock)

			req := httptest.NewRequest(http.MethodPost, "/logout", nil)
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}
			rr := httptest.NewRecorder()

			newAuthHandler(authMock, nil).LogOutUser(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)

			cookies := rr.Result().Cookies()
			require.Len(t, cookies, 1)
			assert.Equal(t, middleware.SessiondIdKey, cookies[0].Name)
			assert.True(t, cookies[0].MaxAge < 0 || cookies[0].Expires.Before(time.Now()))
		})
	}
}

func TestVkOAuthCallback(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		mockBehavior func(authMock *mockAuthUC.AuthUsecase, userMock *mockUserUC.UserUsecase)
		expectCookie bool
	}{
		{
			name: "Success",
			url:  "/oauth/vk?code=valid_code",
			mockBehavior: func(authMock *mockAuthUC.AuthUsecase, userMock *mockUserUC.UserUsecase) {
				authMock.On("ExchangeVKCode", mock.Anything, "valid_code").Return("vk_token", "test@vk.com", nil)
				userMock.On("ProcessUserWithVK", mock.Anything, "vk_token", "test@vk.com").Return(fixedLink, nil)
				authMock.On("CreateSession", mock.Anything, fixedLink).Return("session_token", nil)
			},
			expectCookie: true,
		},
		{
			name:         "EmptyCode",
			url:          "/oauth/vk",
			mockBehavior: func(authMock *mockAuthUC.AuthUsecase, userMock *mockUserUC.UserUsecase) {},
			expectCookie: false,
		},
		{
			name: "ExchangeVKCodeFailed",
			url:  "/oauth/vk?code=invalid_code",
			mockBehavior: func(authMock *mockAuthUC.AuthUsecase, userMock *mockUserUC.UserUsecase) {
				authMock.On("ExchangeVKCode", mock.Anything, "invalid_code").Return("", "", common.ErrorVKOAuthUnavailable)
			},
			expectCookie: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			authMock := mockAuthUC.NewAuthUsecase(t)
			userMock := mockUserUC.NewUserUsecase(t)
			tc.mockBehavior(authMock, userMock)

			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			rr := httptest.NewRecorder()

			newAuthHandler(authMock, userMock).VkOAuthCallback(rr, req)

			cookies := rr.Result().Cookies()
			if tc.expectCookie {
				require.Len(t, cookies, 1)
				assert.Equal(t, middleware.SessiondIdKey, cookies[0].Name)
				assert.Equal(t, "session_token", cookies[0].Value)
			} else {
				assert.Len(t, cookies, 0)
			}
		})
	}
}
