package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/dto"
	handlerCommon "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers/common"
	mockAuthUC "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers/mock_auth_use_case"
	mockCoolDownUC "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers/mock_cool_down_use_case"
	mockCSRFUC "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers/mock_csrf_use_case"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/middleware"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	fixedLink      = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	fixedSession   = "12345667"
	defaultAuthCfg = AuthConfig{MaxLenPassword: 128, MinLenPassword: 8, SessionLifetime: 24 * time.Hour}
)

func newAuthHandler(uc *mockAuthUC.AuthUseCase, cd *mockCoolDownUC.CoolDownUseCase, cs *mockCSRFUC.CSRFUseCase) *Auth {
	return NewAuthHandler(uc, cd, cs, defaultAuthCfg)
}

func newResponse(status string) api.Response {
	return api.Response{Status: status}
}

func newOkResponse[T any](status string, data T) api.OkResponse[T] {
	return api.OkResponse[T]{
		Response: api.Response{Status: status},
		Data:     data,
	}
}

func newErrorResponse(code int, message string) api.ErrorResponse {
	return api.ErrorResponse{
		Response: api.Response{Status: api.StatusError},
		Code:     code,
		Message:  message,
	}
}

func TestLogInUser(t *testing.T) {
	type TestCase struct {
		Name               string
		Request            any
		ExpectedResponse   any
		ExpectedStatusCode int
		MockBehavior       func(uc *mockAuthUC.AuthUseCase)
	}

	tests := []TestCase{
		{
			Name:               "SuccessLogin",
			Request:            dto.LogInRequest{Email: "test@mail.ru", Password: "pass12345"},
			ExpectedStatusCode: http.StatusOK,
			ExpectedResponse: newOkResponse(api.StatusOK, dto.UserInfoResponse{
				Link:  fixedLink,
				Email: "test@mail.ru",
			}),
			MockBehavior: func(uc *mockAuthUC.AuthUseCase) {
				uc.On("Login", mock.Anything, domain.Credentials{Email: "test@mail.ru", Password: "pass12345"}).
					Return(domain.UserInfo{Link: fixedLink, Email: "test@mail.ru"}, fixedSession, nil)
			},
		},
		{
			Name:               "WrongPassword",
			Request:            dto.LogInRequest{Email: "test@mail.ru", Password: "wrongpass"},
			ExpectedStatusCode: http.StatusUnauthorized,
			ExpectedResponse:   newErrorResponse(http.StatusUnauthorized, handlerCommon.ErrWrongEmailOrPassword.Error()),
			MockBehavior: func(uc *mockAuthUC.AuthUseCase) {
				uc.On("Login", mock.Anything, domain.Credentials{Email: "test@mail.ru", Password: "wrongpass"}).
					Return(domain.UserInfo{}, "", common.ErrorWrongCredentials)
			},
		},
		{
			Name:               "PasswordTooShort",
			Request:            dto.LogInRequest{Email: "test@mail.ru", Password: "123"},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, handlerCommon.ErrInvalidEmailOrPassword.Error()),
			MockBehavior:       nil,
		},
		{
			Name:               "InvalidEmail",
			Request:            dto.LogInRequest{Email: "notanemail", Password: "pass12345"},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, handlerCommon.ErrInvalidEmailOrPassword.Error()),
			MockBehavior:       nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			uc := mockAuthUC.NewAuthUseCase(t)
			cd := mockCoolDownUC.NewCoolDownUseCase(t)
			cs := mockCSRFUC.NewCSRFUseCase(t)

			if tc.MockBehavior != nil {
				tc.MockBehavior(uc)
			}

			h := newAuthHandler(uc, cd, cs)
			body, err := json.Marshal(tc.Request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
			rr := httptest.NewRecorder()
			h.LogInUser(rr, req)

			expectedJSON, err := json.Marshal(tc.ExpectedResponse)
			require.NoError(t, err)

			assert.Equal(t, tc.ExpectedStatusCode, rr.Code)
			assert.Equal(t, string(expectedJSON), rr.Body.String())

			if tc.ExpectedStatusCode == http.StatusOK {
				cookies := rr.Result().Cookies()
				var sessionCookie *http.Cookie
				for _, c := range cookies {
					if c.Name == middleware.SessiondIdKey {
						sessionCookie = c
					}
				}
				assert.NotNil(t, sessionCookie)
				assert.Equal(t, fixedSession, sessionCookie.Value)
			}
		})
	}
}

func TestLogInUserRawJSON(t *testing.T) {
	t.Run("InvalidJSON", func(t *testing.T) {
		uc := mockAuthUC.NewAuthUseCase(t)
		cd := mockCoolDownUC.NewCoolDownUseCase(t)
		cs := mockCSRFUC.NewCSRFUseCase(t)

		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{bad json`))
		rr := httptest.NewRecorder()
		newAuthHandler(uc, cd, cs).LogInUser(rr, req)

		expected, _ := json.Marshal(newErrorResponse(http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error()))
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, string(expected), rr.Body.String())
	})
}

func TestRegisterUser(t *testing.T) {
	type TestCase struct {
		Name               string
		Request            any
		ExpectedResponse   any
		ExpectedStatusCode int
		MockBehavior       func(uc *mockAuthUC.AuthUseCase)
	}

	tests := []TestCase{
		{
			Name: "SuccessRegister",
			Request: dto.RegisterRequest{
				DisplayName:      "Test User",
				Email:            "new@mail.ru",
				Password:         "pass12345",
				RepeatedPassword: "pass12345",
			},
			ExpectedStatusCode: http.StatusCreated,
			ExpectedResponse: newOkResponse(api.StatusOK, dto.UserInfoResponse{
				Link:        fixedLink,
				DisplayName: "Test User",
				Email:       "new@mail.ru",
			}),
			MockBehavior: func(uc *mockAuthUC.AuthUseCase) {
				uc.On("Register", mock.Anything, domain.NewCredentialsUser{
					DisplayName: "Test User",
					Email:       "new@mail.ru",
					Password:    "pass12345",
				}).Return(domain.UserInfo{Link: fixedLink, DisplayName: "Test User", Email: "new@mail.ru"}, fixedSession, nil)
			},
		},
		{
			Name: "PasswordsDoNotMatch",
			Request: dto.RegisterRequest{
				Email:            "new@mail.ru",
				Password:         "pass12345",
				RepeatedPassword: "pass99999",
			},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, handlerCommon.ErrInvalidEmailOrPassword.Error()),
			MockBehavior:       nil,
		},
		{
			Name: "UserAlreadyExists",
			Request: dto.RegisterRequest{
				Email:            "dup@mail.ru",
				Password:         "pass12345",
				RepeatedPassword: "pass12345",
			},
			ExpectedStatusCode: http.StatusConflict,
			ExpectedResponse:   newErrorResponse(http.StatusConflict, common.ErrorExistingUser.Error()),
			MockBehavior: func(uc *mockAuthUC.AuthUseCase) {
				uc.On("Register", mock.Anything, mock.Anything).
					Return(domain.UserInfo{}, "", common.ErrorExistingUser)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			uc := mockAuthUC.NewAuthUseCase(t)
			cd := mockCoolDownUC.NewCoolDownUseCase(t)
			cs := mockCSRFUC.NewCSRFUseCase(t)

			if tc.MockBehavior != nil {
				tc.MockBehavior(uc)
			}

			h := newAuthHandler(uc, cd, cs)
			body, err := json.Marshal(tc.Request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
			rr := httptest.NewRecorder()
			h.RegisterUser(rr, req)

			expectedJSON, err := json.Marshal(tc.ExpectedResponse)
			require.NoError(t, err)
			assert.Equal(t, tc.ExpectedStatusCode, rr.Code)
			assert.Equal(t, string(expectedJSON), rr.Body.String())
		})
	}
}

func TestLogOutUser(t *testing.T) {
	t.Run("WithSessionCookie", func(t *testing.T) {
		uc := mockAuthUC.NewAuthUseCase(t)
		cd := mockCoolDownUC.NewCoolDownUseCase(t)
		cs := mockCSRFUC.NewCSRFUseCase(t)

		uc.On("Logout", mock.Anything, fixedSession).Return(nil)

		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		req.AddCookie(&http.Cookie{Name: middleware.SessiondIdKey, Value: fixedSession})
		rr := httptest.NewRecorder()

		newAuthHandler(uc, cd, cs).LogOutUser(rr, req)

		expected, _ := json.Marshal(newResponse(api.StatusOK))
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, string(expected), rr.Body.String())
	})

	t.Run("WithoutCookie", func(t *testing.T) {
		uc := mockAuthUC.NewAuthUseCase(t)
		cd := mockCoolDownUC.NewCoolDownUseCase(t)
		cs := mockCSRFUC.NewCSRFUseCase(t)

		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		rr := httptest.NewRecorder()

		newAuthHandler(uc, cd, cs).LogOutUser(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestMeHandler(t *testing.T) {
	t.Run("Authorized", func(t *testing.T) {
		uc := mockAuthUC.NewAuthUseCase(t)
		cd := mockCoolDownUC.NewCoolDownUseCase(t)
		cs := mockCSRFUC.NewCSRFUseCase(t)

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, fixedLink)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		newAuthHandler(uc, cd, cs).MeHandler(rr, req)

		expected, _ := json.Marshal(newResponse(api.StatusOK))
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, string(expected), rr.Body.String())
	})

	t.Run("Unauthorized", func(t *testing.T) {
		uc := mockAuthUC.NewAuthUseCase(t)
		cd := mockCoolDownUC.NewCoolDownUseCase(t)
		cs := mockCSRFUC.NewCSRFUseCase(t)

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		rr := httptest.NewRecorder()

		newAuthHandler(uc, cd, cs).MeHandler(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestSendRecoveryEmail(t *testing.T) {
	type TestCase struct {
		Name               string
		Request            any
		ExpectedStatusCode int
		ExpectedResponse   any
		MockBehavior       func(uc *mockAuthUC.AuthUseCase, cd *mockCoolDownUC.CoolDownUseCase)
	}

	tests := []TestCase{
		{
			Name:               "SuccessSent",
			Request:            dto.PasswordRecoveryRequest{Email: "user@mail.ru"},
			ExpectedStatusCode: http.StatusOK,
			ExpectedResponse:   newResponse(api.StatusOK),
			MockBehavior: func(uc *mockAuthUC.AuthUseCase, cd *mockCoolDownUC.CoolDownUseCase) {
				cd.On("CheckCoolDown", mock.Anything, mock.Anything).
					Return(domain.CooldownResult{Allowed: true, WaitS: 0}, nil)
				uc.On("SendRecoveryCode", mock.Anything, "user@mail.ru").Return(nil)
			},
		},
		{
			Name:               "CoolDownActive",
			Request:            dto.PasswordRecoveryRequest{Email: "user@mail.ru"},
			ExpectedStatusCode: http.StatusTooManyRequests,
			MockBehavior: func(uc *mockAuthUC.AuthUseCase, cd *mockCoolDownUC.CoolDownUseCase) {
				cd.On("CheckCoolDown", mock.Anything, mock.Anything).
					Return(domain.CooldownResult{Allowed: false, WaitS: 30}, nil)
			},
		},
		{
			Name:               "EmailNotFound",
			Request:            dto.PasswordRecoveryRequest{Email: "notfound@mail.ru"},
			ExpectedStatusCode: http.StatusNotFound,
			ExpectedResponse:   newErrorResponse(http.StatusNotFound, handlerCommon.ErrUserDoesNotExists.Error()),
			MockBehavior: func(uc *mockAuthUC.AuthUseCase, cd *mockCoolDownUC.CoolDownUseCase) {
				cd.On("CheckCoolDown", mock.Anything, mock.Anything).
					Return(domain.CooldownResult{Allowed: true}, nil)
				uc.On("SendRecoveryCode", mock.Anything, "notfound@mail.ru").Return(common.ErrorNonexistentEmail)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			uc := mockAuthUC.NewAuthUseCase(t)
			cd := mockCoolDownUC.NewCoolDownUseCase(t)
			cs := mockCSRFUC.NewCSRFUseCase(t)

			tc.MockBehavior(uc, cd)

			body, err := json.Marshal(tc.Request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/forgot-password", bytes.NewReader(body))
			rr := httptest.NewRecorder()
			newAuthHandler(uc, cd, cs).SendRecoveryEmail(rr, req)

			assert.Equal(t, tc.ExpectedStatusCode, rr.Code)
			if tc.ExpectedResponse != nil {
				expectedJSON, err := json.Marshal(tc.ExpectedResponse)
				require.NoError(t, err)
				assert.Equal(t, string(expectedJSON), rr.Body.String())
			}
		})
	}
}

func TestCheckRecoveryCode(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		uc := mockAuthUC.NewAuthUseCase(t)
		cd := mockCoolDownUC.NewCoolDownUseCase(t)
		cs := mockCSRFUC.NewCSRFUseCase(t)

		uc.On("CheckRecoveryCode", mock.Anything, "123456").Return(nil)

		body, _ := json.Marshal(dto.RecoveryCodeRequest{Code: "123456"})
		req := httptest.NewRequest(http.MethodPost, "/check-code", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		newAuthHandler(uc, cd, cs).CheckRecoveryCode(rr, req)

		expected, _ := json.Marshal(newResponse(api.StatusOK))
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, string(expected), rr.Body.String())
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		uc := mockAuthUC.NewAuthUseCase(t)
		cd := mockCoolDownUC.NewCoolDownUseCase(t)
		cs := mockCSRFUC.NewCSRFUseCase(t)

		req := httptest.NewRequest(http.MethodPost, "/check-code", strings.NewReader("{bad}"))
		rr := httptest.NewRecorder()
		newAuthHandler(uc, cd, cs).CheckRecoveryCode(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestResetUserPassword(t *testing.T) {
	type TestCase struct {
		Name               string
		Request            any
		ExpectedStatusCode int
		ExpectedResponse   any
		MockBehavior       func(uc *mockAuthUC.AuthUseCase)
	}

	tests := []TestCase{
		{
			Name: "Success",
			Request: dto.NewPasswordRequest{
				TokenID:          "tok1",
				Password:         "newpass1",
				RepeatedPassword: "newpass1",
			},
			ExpectedStatusCode: http.StatusOK,
			ExpectedResponse:   newResponse(api.StatusOK),
			MockBehavior: func(uc *mockAuthUC.AuthUseCase) {
				uc.On("ResetPassword", mock.Anything, "tok1", "newpass1").Return(nil)
			},
		},
		{
			Name: "PasswordsDoNotMatch",
			Request: dto.NewPasswordRequest{
				TokenID:          "tok1",
				Password:         "newpass1",
				RepeatedPassword: "different",
			},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, handlerCommon.ErrInvalidEmailOrPassword.Error()),
			MockBehavior:       nil,
		},
		{
			Name: "TokenNotFound",
			Request: dto.NewPasswordRequest{
				TokenID:          "bad-token",
				Password:         "newpass1",
				RepeatedPassword: "newpass1",
			},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, common.ErrorResetTokenNotFound.Error()),
			MockBehavior: func(uc *mockAuthUC.AuthUseCase) {
				uc.On("ResetPassword", mock.Anything, "bad-token", "newpass1").Return(common.ErrorResetTokenNotFound)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			uc := mockAuthUC.NewAuthUseCase(t)
			cd := mockCoolDownUC.NewCoolDownUseCase(t)
			cs := mockCSRFUC.NewCSRFUseCase(t)

			if tc.MockBehavior != nil {
				tc.MockBehavior(uc)
			}

			body, err := json.Marshal(tc.Request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader(body))
			rr := httptest.NewRecorder()
			newAuthHandler(uc, cd, cs).ResetUserPassword(rr, req)

			assert.Equal(t, tc.ExpectedStatusCode, rr.Code)
			if tc.ExpectedResponse != nil {
				expectedJSON, err := json.Marshal(tc.ExpectedResponse)
				require.NoError(t, err)
				assert.Equal(t, string(expectedJSON), rr.Body.String())
			}
		})
	}
}

func TestSetCSRFCookieHandler(t *testing.T) {
	t.Run("WithSession", func(t *testing.T) {
		uc := mockAuthUC.NewAuthUseCase(t)
		cd := mockCoolDownUC.NewCoolDownUseCase(t)
		cs := mockCSRFUC.NewCSRFUseCase(t)

		expTime := time.Now().Add(time.Hour)
		cs.On("GetExpireTime", mock.Anything).Return(expTime)
		cs.On("Generate", mock.Anything, fixedSession, expTime.Unix()).Return("csrf-token-value", nil)

		req := httptest.NewRequest(http.MethodGet, "/csrf", nil)
		req.AddCookie(&http.Cookie{Name: middleware.SessiondIdKey, Value: fixedSession})
		rr := httptest.NewRecorder()

		newAuthHandler(uc, cd, cs).SetCSRFCookieHandler(rr, req)

		expected, _ := json.Marshal(newResponse(api.StatusOK))
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, string(expected), rr.Body.String())

		var csrfCookie *http.Cookie
		for _, c := range rr.Result().Cookies() {
			if c.Name == csrfCookieKey {
				csrfCookie = c
			}
		}
		assert.NotNil(t, csrfCookie)
		assert.Equal(t, "csrf-token-value", csrfCookie.Value)
	})

	t.Run("NoSession", func(t *testing.T) {
		uc := mockAuthUC.NewAuthUseCase(t)
		cd := mockCoolDownUC.NewCoolDownUseCase(t)
		cs := mockCSRFUC.NewCSRFUseCase(t)

		req := httptest.NewRequest(http.MethodGet, "/csrf", nil)
		rr := httptest.NewRecorder()
		newAuthHandler(uc, cd, cs).SetCSRFCookieHandler(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestVkOAuthCallback(t *testing.T) {
	t.Run("MissingCode", func(t *testing.T) {
		uc := mockAuthUC.NewAuthUseCase(t)
		cd := mockCoolDownUC.NewCoolDownUseCase(t)
		cs := mockCSRFUC.NewCSRFUseCase(t)

		req := httptest.NewRequest(http.MethodGet, "/oauth/vk", nil)
		rr := httptest.NewRecorder()
		newAuthHandler(uc, cd, cs).VkOAuthCallback(rr, req)

		assert.Equal(t, http.StatusFound, rr.Code)
	})

	t.Run("LoginWithVKSuccess", func(t *testing.T) {
		uc := mockAuthUC.NewAuthUseCase(t)
		cd := mockCoolDownUC.NewCoolDownUseCase(t)
		cs := mockCSRFUC.NewCSRFUseCase(t)

		uc.On("LoginWithVK", mock.Anything, "good-code").
			Return(domain.UserInfo{Link: fixedLink}, fixedSession, nil)

		req := httptest.NewRequest(http.MethodGet, "/oauth/vk?code=good-code", nil)
		rr := httptest.NewRecorder()
		newAuthHandler(uc, cd, cs).VkOAuthCallback(rr, req)

		assert.Equal(t, http.StatusFound, rr.Code)
	})

	t.Run("LoginWithVKError", func(t *testing.T) {
		uc := mockAuthUC.NewAuthUseCase(t)
		cd := mockCoolDownUC.NewCoolDownUseCase(t)
		cs := mockCSRFUC.NewCSRFUseCase(t)

		uc.On("LoginWithVK", mock.Anything, "bad-code").
			Return(domain.UserInfo{}, "", common.ErrorVKOAuthUnavailable)

		req := httptest.NewRequest(http.MethodGet, "/oauth/vk?code=bad-code", nil)
		rr := httptest.NewRecorder()
		newAuthHandler(uc, cd, cs).VkOAuthCallback(rr, req)

		assert.Equal(t, http.StatusFound, rr.Code)
	})
}

func TestRegisterUserInternalError(t *testing.T) {
	t.Run("InternalError", func(t *testing.T) {
		uc := mockAuthUC.NewAuthUseCase(t)
		cd := mockCoolDownUC.NewCoolDownUseCase(t)
		cs := mockCSRFUC.NewCSRFUseCase(t)

		uc.On("Register", mock.Anything, mock.Anything).
			Return(domain.UserInfo{}, "", common.ErrorNotNullValue)

		body, _ := json.Marshal(dto.RegisterRequest{
			Email: "t@mail.ru", Password: "pass12345", RepeatedPassword: "pass12345",
		})
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		newAuthHandler(uc, cd, cs).RegisterUser(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestSendRecoveryEmailCoolDownError(t *testing.T) {
	t.Run("CoolDownClientError", func(t *testing.T) {
		uc := mockAuthUC.NewAuthUseCase(t)
		cd := mockCoolDownUC.NewCoolDownUseCase(t)
		cs := mockCSRFUC.NewCSRFUseCase(t)

		cd.On("CheckCoolDown", mock.Anything, mock.Anything).
			Return(domain.CooldownResult{}, common.ErrorInvalidInput)

		body, _ := json.Marshal(dto.PasswordRecoveryRequest{Email: "t@mail.ru"})
		req := httptest.NewRequest(http.MethodPost, "/forgot-password", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		newAuthHandler(uc, cd, cs).SendRecoveryEmail(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestResetUserPasswordNullValue(t *testing.T) {
	t.Run("NullValue", func(t *testing.T) {
		uc := mockAuthUC.NewAuthUseCase(t)
		cd := mockCoolDownUC.NewCoolDownUseCase(t)
		cs := mockCSRFUC.NewCSRFUseCase(t)

		uc.On("ResetPassword", mock.Anything, "tok", "newpass1").Return(common.ErrorNotNullValue)

		body, _ := json.Marshal(dto.NewPasswordRequest{
			TokenID: "tok", Password: "newpass1", RepeatedPassword: "newpass1",
		})
		req := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		newAuthHandler(uc, cd, cs).ResetUserPassword(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestResetUserPasswordUserNotFound(t *testing.T) {
	t.Run("UserNotFound", func(t *testing.T) {
		uc := mockAuthUC.NewAuthUseCase(t)
		cd := mockCoolDownUC.NewCoolDownUseCase(t)
		cs := mockCSRFUC.NewCSRFUseCase(t)

		uc.On("ResetPassword", mock.Anything, "tok", "newpass1").Return(common.ErrorNonexistentUser)

		body, _ := json.Marshal(dto.NewPasswordRequest{
			TokenID: "tok", Password: "newpass1", RepeatedPassword: "newpass1",
		})
		req := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		newAuthHandler(uc, cd, cs).ResetUserPassword(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestSendRecoveryEmailServiceError(t *testing.T) {
	t.Run("CannotSendCode", func(t *testing.T) {
		uc := mockAuthUC.NewAuthUseCase(t)
		cd := mockCoolDownUC.NewCoolDownUseCase(t)
		cs := mockCSRFUC.NewCSRFUseCase(t)

		cd.On("CheckCoolDown", mock.Anything, mock.Anything).
			Return(domain.CooldownResult{Allowed: true}, nil)
		uc.On("SendRecoveryCode", mock.Anything, "user@mail.ru").Return(common.ErrorInvalidInput)

		body, _ := json.Marshal(dto.PasswordRecoveryRequest{Email: "user@mail.ru"})
		req := httptest.NewRequest(http.MethodPost, "/forgot-password", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		newAuthHandler(uc, cd, cs).SendRecoveryEmail(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestSetCSRFCookieHandlerGenerateError(t *testing.T) {
	t.Run("GenerateError", func(t *testing.T) {
		uc := mockAuthUC.NewAuthUseCase(t)
		cd := mockCoolDownUC.NewCoolDownUseCase(t)
		cs := mockCSRFUC.NewCSRFUseCase(t)

		expTime := fixedTime()
		cs.On("GetExpireTime", mock.Anything).Return(expTime)
		cs.On("Generate", mock.Anything, fixedSession, expTime.Unix()).Return("", common.ErrorInvalidInput)

		req := httptest.NewRequest(http.MethodGet, "/csrf", nil)
		req.AddCookie(&http.Cookie{Name: middleware.SessiondIdKey, Value: fixedSession})
		rr := httptest.NewRecorder()
		newAuthHandler(uc, cd, cs).SetCSRFCookieHandler(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func fixedTime() time.Time {
	return time.Now().Add(time.Hour)
}

func TestCheckRecoveryCodeError(t *testing.T) {
	t.Run("ServiceError", func(t *testing.T) {
		uc := mockAuthUC.NewAuthUseCase(t)
		cd := mockCoolDownUC.NewCoolDownUseCase(t)
		cs := mockCSRFUC.NewCSRFUseCase(t)

		uc.On("CheckRecoveryCode", mock.Anything, "bad").Return(common.ErrorInvalidInput)

		body, _ := json.Marshal(dto.RecoveryCodeRequest{Code: "bad"})
		req := httptest.NewRequest(http.MethodPost, "/check-code", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		newAuthHandler(uc, cd, cs).CheckRecoveryCode(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
