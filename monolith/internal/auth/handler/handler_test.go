package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/auth/handler/dto"
	mockAuthSrv "github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/auth/handler/mock_auth_srv"
	vkOAuthMocks "github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/auth/handler/mock_vk_oauth"
	service "github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/auth/service"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/auth/service/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/middleware"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

type TestCase struct {
	Name               string
	Request            any
	ExpectedResponse   any
	ExpectedStatusCode int
	MockBehavior       func(m *mockAuthSrv.AuthService)
}

func newResponse(status string) api.Response {
	return api.Response{
		Status: status,
	}
}

func newOkResponse[T any](status string, data T) api.OkResponse[T] {
	return api.OkResponse[T]{
		Response: api.Response{
			Status: status,
		},
		Data: data,
	}
}

func newErrorResponse(code int, message string) api.ErrorResponse {
	return api.ErrorResponse{
		Response: api.Response{
			Status: api.StatusError,
		},
		Code:    code,
		Message: message,
	}
}

func TestLogInUserWithSchema(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Success login",
			Request: dto.LogInRequest{
				Email:    "test@mail.ru",
				Password: "123456789",
			},
			ExpectedResponse: newOkResponse(api.StatusOK, dto.UserInfoResponse{
				Link:        common.FixedUserUuiD,
				DisplayName: "Artem",
				Email:       "test@mail.ru",
			}),
			ExpectedStatusCode: http.StatusOK,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("LogIn", ctx, serviceDto.LogInUser{
					Email:    "test@mail.ru",
					Password: "123456789",
				}).Return(
					serviceDto.UserInfo{
						Link:        common.FixedUserUuiD,
						DisplayName: "Artem",
						Email:       "test@mail.ru",
					},
					common.FixedSessionID,
					nil,
				)
			},
		},
		{
			Name: "Wrong password or email",
			Request: dto.LogInRequest{
				Email:    "artem@mail.ru",
				Password: "wrong_password",
			},
			ExpectedResponse:   newErrorResponse(http.StatusUnauthorized, ErrWrongEmailOrPassword.Error()),
			ExpectedStatusCode: http.StatusUnauthorized,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("LogIn", ctx, serviceDto.LogInUser{
					Email:    "artem@mail.ru",
					Password: "wrong_password",
				}).Return(serviceDto.UserInfo{}, "", service.ErrorWrongPassword)
			},
		},
		{
			Name: "Size password smaller than 8",
			Request: dto.LogInRequest{
				Email:    "artem@mail.ru",
				Password: "123",
			},
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, ErrInvalidEmailOrPassword.Error()),
			ExpectedStatusCode: http.StatusBadRequest,
			MockBehavior:       nil,
		},
		{
			Name: "Size password biger than 128",
			Request: dto.LogInRequest{
				Email:    "artem@mail.ru",
				Password: "1231111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111",
			},
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, ErrInvalidEmailOrPassword.Error()),
			ExpectedStatusCode: http.StatusBadRequest,
			MockBehavior:       nil,
		},
		{
			Name: "Email hasn't @",
			Request: dto.LogInRequest{
				Email:    "testmail.ru",
				Password: "1234567",
			},
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, ErrInvalidEmailOrPassword.Error()),
			ExpectedStatusCode: http.StatusBadRequest,
			MockBehavior:       nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockLogInService := mockAuthSrv.NewAuthService(t)
			if test.MockBehavior != nil {
				test.MockBehavior(mockLogInService)
			}

			handler := NewHandler(mockLogInService, Config{MaxLenPassword: 128, MinLenPassword: 8})

			requestJson, err := json.Marshal(test.Request)
			require.NoError(t, err, "request marshal should not return error")

			requestReader := bytes.NewReader(requestJson)

			request := httptest.NewRequest(http.MethodPost, "/", requestReader)
			response := httptest.NewRecorder()

			handler.LogInUser(response, request)

			responseJson, err := json.Marshal(test.ExpectedResponse)
			require.NoError(t, err, "response marshal should not return error")

			assert.Equal(t, test.ExpectedStatusCode, response.Code, "incorrect status code")
			assert.Equal(t, string(responseJson), response.Body.String(), "incorrect body")

			if test.ExpectedStatusCode == http.StatusOK {
				res := response.Result()
				cookies := res.Cookies()

				var sessionCookie *http.Cookie
				for _, c := range cookies {
					if c.Name == "session_id" {
						sessionCookie = c
						break
					}
				}

				assert.NotNil(t, sessionCookie, "cookie wasn't found in response")
				assert.NotEmpty(t, sessionCookie.Value, "empty value session Cookie")
				assert.True(t, sessionCookie.HttpOnly, "flag HttpOnly must be true")
			}
		})
	}
}

func TestLogInUserWithRawJSON(t *testing.T) {
	t.Run("Incorrect JSON", func(t *testing.T) {
		incorrectJson := `{"email":"test@mail.ru",,,}`
		requestBody := strings.NewReader(incorrectJson)

		expectedResponse := newErrorResponse(http.StatusBadRequest, ErrInvalidRequestSchema.Error())
		expectedBody, err := json.Marshal(expectedResponse)
		require.NoError(t, err, "response marshal should not return error")

		mockLogInService := mockAuthSrv.NewAuthService(t)
		handler := NewHandler(mockLogInService, Config{MaxLenPassword: 128, MinLenPassword: 8})

		req := httptest.NewRequest(http.MethodPost, "/", requestBody)
		res := httptest.NewRecorder()

		handler.LogInUser(res, req)

		assert.Equal(t, expectedResponse.Code, res.Code, "incorrect status code")
		assert.Equal(t, string(expectedBody), res.Body.String(), "incorrect body")
	})
}

func TestRegisterUserWithSchema(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Success registration",
			Request: dto.RegisterRequest{
				DisplayName:      "Artem",
				Password:         "12345678",
				RepeatedPassword: "12345678",
				Email:            "test@mail.ru",
			},
			ExpectedResponse: newOkResponse(api.StatusOK, dto.UserInfoResponse{
				Link:        common.FixedUserUuiD,
				DisplayName: "Artem",
				Email:       "test@mail.ru",
			}),
			ExpectedStatusCode: http.StatusCreated,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("Register", ctx, serviceDto.RegistrationUser{
					DisplayName: "Artem",
					Email:       "test@mail.ru",
					Password:    "12345678",
				}).Return(
					serviceDto.UserInfo{
						Link:        common.FixedUserUuiD,
						DisplayName: "Artem",
						Email:       "test@mail.ru",
					},
					common.FixedSessionID,
					nil,
				)
			},
		},
		{
			Name: "Passwords do not match",
			Request: dto.RegisterRequest{
				DisplayName:      "Artem",
				Password:         "12345678",
				RepeatedPassword: "65432178",
				Email:            "test@mail.ru",
			},
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, ErrInvalidEmailOrPassword.Error()),
			ExpectedStatusCode: http.StatusBadRequest,
			MockBehavior:       nil,
		},
		{
			Name: "Email is already existing",
			Request: dto.RegisterRequest{
				DisplayName:      "Artem",
				Password:         "123456789",
				RepeatedPassword: "123456789",
				Email:            "test@mail.ru",
			},
			ExpectedResponse:   newErrorResponse(http.StatusInternalServerError, ErrInternalServerError.Error()),
			ExpectedStatusCode: http.StatusInternalServerError,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("Register", ctx, serviceDto.RegistrationUser{
					DisplayName: "Artem",
					Email:       "test@mail.ru",
					Password:    "123456789",
				}).Return(
					serviceDto.UserInfo{},
					"",
					fmt.Errorf("repo.AddUser: user with this email alreday exists"),
				)
			},
		},
		{
			Name: "Incorrect symbol in password",
			Request: dto.RegisterRequest{
				DisplayName:      "Artem",
				Password:         "бобёр123",
				RepeatedPassword: "бобёр123",
				Email:            "test@mail.ru",
			},
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, ErrInvalidEmailOrPassword.Error()),
			ExpectedStatusCode: http.StatusBadRequest,
			MockBehavior:       nil,
		},
		{
			Name: "Incorrect symbol in email",
			Request: dto.RegisterRequest{
				DisplayName:      "Artem",
				Password:         "1234552323",
				RepeatedPassword: "1234552323",
				Email:            "бобёр@mail.ru",
			},
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, ErrInvalidEmailOrPassword.Error()),
			ExpectedStatusCode: http.StatusBadRequest,
			MockBehavior:       nil,
		},
		{
			Name: "Size password smaller than 6",
			Request: dto.RegisterRequest{
				DisplayName:      "Artem",
				Password:         "123",
				RepeatedPassword: "123",
				Email:            "test@mail.ru",
			},
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, ErrInvalidEmailOrPassword.Error()),
			ExpectedStatusCode: http.StatusBadRequest,
			MockBehavior:       nil,
		},
		{
			Name: "Email has 2 @",
			Request: dto.RegisterRequest{
				DisplayName:      "Artem",
				Password:         "123456789",
				RepeatedPassword: "123456789",
				Email:            "test@m@ail.ru",
			},
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, ErrInvalidEmailOrPassword.Error()),
			ExpectedStatusCode: http.StatusBadRequest,
			MockBehavior:       nil,
		},
		{
			Name: "Email hasn't @",
			Request: dto.RegisterRequest{
				DisplayName:      "Artem",
				Password:         "1234567",
				RepeatedPassword: "1234567",
				Email:            "testmail.ru",
			},
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, ErrInvalidEmailOrPassword.Error()),
			ExpectedStatusCode: http.StatusBadRequest,
			MockBehavior:       nil,
		},
		{
			Name: "Email has @.",
			Request: dto.RegisterRequest{
				DisplayName:      "Artem",
				Password:         "1234567",
				RepeatedPassword: "1234567",
				Email:            "test@.ru",
			},
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, ErrInvalidEmailOrPassword.Error()),
			ExpectedStatusCode: http.StatusBadRequest,
			MockBehavior:       nil,
		},
		{
			Name: "Error during hash password",
			Request: dto.RegisterRequest{
				DisplayName:      "Artem",
				Password:         "123456789",
				RepeatedPassword: "123456789",
				Email:            "test@mail.ru",
			},
			ExpectedResponse:   newErrorResponse(http.StatusInternalServerError, ErrInternalServerError.Error()),
			ExpectedStatusCode: http.StatusInternalServerError,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("Register", ctx, serviceDto.RegistrationUser{
					DisplayName: "Artem",
					Password:    "123456789",
					Email:       "test@mail.ru",
				}).Return(
					serviceDto.UserInfo{},
					"",
					fmt.Errorf("failed to create hash: error bcrypt"),
				)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockRegisterService := mockAuthSrv.NewAuthService(t)
			if test.MockBehavior != nil {
				test.MockBehavior(mockRegisterService)
			}

			handler := NewHandler(mockRegisterService, Config{MaxLenPassword: 128, MinLenPassword: 6})

			requestJson, err := json.Marshal(test.Request)
			require.NoError(t, err, "request marshal should not return error")

			requestReader := bytes.NewReader(requestJson)

			request := httptest.NewRequest(http.MethodPost, "/", requestReader)
			response := httptest.NewRecorder()

			handler.RegisterUser(response, request)

			responseJson, err := json.Marshal(test.ExpectedResponse)
			require.NoError(t, err, "response marshal should not return error")

			assert.Equal(t, test.ExpectedStatusCode, response.Code, "incorrect status code")
			assert.Equal(t, string(responseJson), response.Body.String(), "incorrect body")

			if test.ExpectedStatusCode == http.StatusCreated {
				res := response.Result()
				cookies := res.Cookies()

				var sessionCookie *http.Cookie
				for _, c := range cookies {
					if c.Name == "session_id" {
						sessionCookie = c
						break
					}
				}

				assert.NotNil(t, sessionCookie, "cookie wasn't found in response")
				assert.NotEmpty(t, sessionCookie.Value, "empty value session Cookie")
				assert.True(t, sessionCookie.HttpOnly, "flag HttpOnly must be true")
			}
		})
	}
}

func TestRegisterUserWithRawJSON(t *testing.T) {
	t.Run("Incorrect JSON", func(t *testing.T) {
		incorrectJson := `{"display_name":"Artem",,,}`
		requestBody := strings.NewReader(incorrectJson)

		expectedResponse := newErrorResponse(http.StatusBadRequest, ErrInvalidRequestSchema.Error())
		expectedBody, err := json.Marshal(expectedResponse)
		require.NoError(t, err, "response marshal should not return error")

		mockRegisterService := mockAuthSrv.NewAuthService(t)
		handler := NewHandler(mockRegisterService, Config{MaxLenPassword: 128, MinLenPassword: 8})

		req := httptest.NewRequest(http.MethodPost, "/register", requestBody)
		res := httptest.NewRecorder()

		handler.RegisterUser(res, req)

		assert.Equal(t, expectedResponse.Code, res.Code, "incorrect status code")
		assert.Equal(t, string(expectedBody), res.Body.String(), "incorrect body")
	})
}

type LogOutTestCase struct {
	Name               string
	AddCookie          bool
	CookieValue        string
	ExpectedResponse   any
	ExpectedStatusCode int
	MockBehavior       func(m *mockAuthSrv.AuthService)
}

func TestLogOutUser(t *testing.T) {
	tests := []LogOutTestCase{
		{
			Name:               "Success logout",
			AddCookie:          true,
			CookieValue:        common.FixedSessionID,
			ExpectedResponse:   newResponse(api.StatusOK),
			ExpectedStatusCode: http.StatusOK,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("LogOut", ctx, common.FixedSessionID).Return(nil)
			},
		},
		{
			Name:               "Service error",
			AddCookie:          true,
			CookieValue:        common.FixedSessionID,
			ExpectedResponse:   newResponse(api.StatusOK),
			ExpectedStatusCode: http.StatusOK,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("LogOut", ctx, common.FixedSessionID).Return(fmt.Errorf("database down"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockAuthService := mockAuthSrv.NewAuthService(t)
			if test.MockBehavior != nil {
				test.MockBehavior(mockAuthService)
			}

			handler := NewHandler(mockAuthService, Config{})

			request := httptest.NewRequest(http.MethodPost, "/", nil)
			if test.AddCookie {
				request.AddCookie(&http.Cookie{
					Name:  "session_id",
					Value: test.CookieValue,
				})
			}
			response := httptest.NewRecorder()

			handler.LogOutUser(response, request)

			responseJson, err := json.Marshal(test.ExpectedResponse)
			require.NoError(t, err, "response marshal should not return error")

			assert.Equal(t, test.ExpectedStatusCode, response.Code, "incorrect status code")
			assert.Equal(t, string(responseJson), response.Body.String(), "incorrect body")

			if test.ExpectedStatusCode == http.StatusOK {
				res := response.Result()
				cookies := res.Cookies()

				var sessionCookie *http.Cookie
				for _, c := range cookies {
					if c.Name == "session_id" {
						sessionCookie = c
						break
					}
				}

				assert.NotNil(t, sessionCookie, "cookie wasn't found in response")
				assert.Empty(t, sessionCookie.Value, "cookie value must be empty")
			}
		})
	}
}

func TestMeHandler(t *testing.T) {
	handler := NewHandler(new(mockAuthSrv.AuthService), Config{})

	tests := []struct {
		Name           string
		SetupRequest   func(req *http.Request) *http.Request
		ExpectedStatus int
	}{
		{
			Name: "success",
			SetupRequest: func(req *http.Request) *http.Request {
				userID := uuid.New()
				ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, userID)
				return req.WithContext(ctx)
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name: "unauthorized no context value",
			SetupRequest: func(req *http.Request) *http.Request {
				return req
			},
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name: "unauthorized wrong type",
			SetupRequest: func(req *http.Request) *http.Request {
				ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, "invalid-uuid-string")
				return req.WithContext(ctx)
			},
			ExpectedStatus: http.StatusUnauthorized,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
			req = test.SetupRequest(req)
			res := httptest.NewRecorder()

			handler.MeHandler(res, req)

			assert.Equal(t, test.ExpectedStatus, res.Code)
		})
	}
}

func TestSendRecoveryEmail(t *testing.T) {
	defaultConf := serviceDto.CoolDownConfig{
		Name:       nameCoolDown,
		Email:      "bobr@mail.ru",
		Expiration: 1 * time.Minute,
	}

	tests := []TestCase{
		{
			Name: "Success send email",
			Request: dto.PasswordRecoveryRequest{
				Email: "bobr@mail.ru",
			},
			ExpectedResponse:   newResponse(api.StatusOK),
			ExpectedStatusCode: http.StatusOK,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("SendRecoveryCode", ctx, "bobr@mail.ru").Return(nil)
				m.On("CheckCoolDown", ctx, defaultConf).Return(true, time.Duration(0), nil)
			},
		},
		{
			Name: "Service error",
			Request: dto.PasswordRecoveryRequest{
				Email: "bobr@mail.ru",
			},
			ExpectedResponse:   newErrorResponse(http.StatusInternalServerError, ErrCannotSendRecoveryCode.Error()),
			ExpectedStatusCode: http.StatusInternalServerError,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("SendRecoveryCode", ctx, "bobr@mail.ru").Return(errors.New("some internal error"))
				m.On("CheckCoolDown", ctx, defaultConf).Return(true, time.Duration(0), nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockSrv := mockAuthSrv.NewAuthService(t)
			if test.MockBehavior != nil {
				test.MockBehavior(mockSrv)
			}

			handler := NewHandler(mockSrv, Config{})

			requestJson, err := json.Marshal(test.Request)
			require.NoError(t, err, "request marshal should not return error")

			requestReader := bytes.NewReader(requestJson)
			request := httptest.NewRequest(http.MethodPost, "/forgot-password", requestReader)
			response := httptest.NewRecorder()

			handler.SendRecoveryEmail(response, request)

			responseJson, err := json.Marshal(test.ExpectedResponse)
			require.NoError(t, err, "response marshal should not return error")

			assert.Equal(t, test.ExpectedStatusCode, response.Code, "incorrect status code")
			assert.Equal(t, string(responseJson), response.Body.String(), "incorrect body")
		})
	}
}

func TestSendRecoveryEmailWithRawJSON(t *testing.T) {
	t.Run("Incorrect JSON", func(t *testing.T) {
		incorrectJson := `{"email":"test@mail.ru",,,}`
		requestBody := strings.NewReader(incorrectJson)

		expectedResponse := newErrorResponse(http.StatusBadRequest, ErrInvalidRequestSchema.Error())
		expectedBody, err := json.Marshal(expectedResponse)
		require.NoError(t, err)

		mockSrv := mockAuthSrv.NewAuthService(t)
		handler := NewHandler(mockSrv, Config{})

		req := httptest.NewRequest(http.MethodPost, "/forgot-password", requestBody)
		res := httptest.NewRecorder()

		handler.SendRecoveryEmail(res, req)

		assert.Equal(t, expectedResponse.Code, res.Code, "incorrect status code")
		assert.Equal(t, string(expectedBody), res.Body.String(), "incorrect body")
	})
}

func TestResetUserPassword(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Success reset password",
			Request: dto.NewPasswordRequest{
				TokenID:          "valid-token-123",
				Password:         "new_password",
				RepeatedPassword: "new_password",
			},
			ExpectedResponse:   newResponse(api.StatusOK),
			ExpectedStatusCode: http.StatusOK,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("ResetPassword", ctx, "valid-token-123", "new_password").Return(nil)
			},
		},
		{
			Name: "Validation failed",
			Request: dto.NewPasswordRequest{
				TokenID:          "valid-token-123",
				Password:         "new_secure_password",
				RepeatedPassword: "different_password",
			},
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, ErrInvalidEmailOrPassword.Error()),
			ExpectedStatusCode: http.StatusBadRequest,
			MockBehavior:       nil,
		},
		{
			Name: "Service error",
			Request: dto.NewPasswordRequest{
				TokenID:          "expired-token",
				Password:         "new_secure_password",
				RepeatedPassword: "new_secure_password",
			},
			ExpectedResponse:   newErrorResponse(http.StatusInternalServerError, ErrCannotResetPassword.Error()),
			ExpectedStatusCode: http.StatusInternalServerError,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("ResetPassword", ctx, "expired-token", "new_secure_password").Return(errors.New("token expired"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockSrv := mockAuthSrv.NewAuthService(t)
			if test.MockBehavior != nil {
				test.MockBehavior(mockSrv)
			}

			handler := NewHandler(mockSrv, Config{MaxLenPassword: 128, MinLenPassword: 8})

			requestJson, err := json.Marshal(test.Request)
			require.NoError(t, err, "request marshal should not return error")

			requestReader := bytes.NewReader(requestJson)
			request := httptest.NewRequest(http.MethodPost, "/reset-password", requestReader)
			response := httptest.NewRecorder()

			handler.ResetUserPassword(response, request)

			responseJson, err := json.Marshal(test.ExpectedResponse)
			require.NoError(t, err, "response marshal should not return error")

			assert.Equal(t, test.ExpectedStatusCode, response.Code, "incorrect status code")
			assert.Equal(t, string(responseJson), response.Body.String(), "incorrect body")
		})
	}
}

func TestResetUserPasswordWithRawJSON(t *testing.T) {
	t.Run("Incorrect JSON", func(t *testing.T) {
		incorrectJson := `{"password":"123", "repeat"`
		requestBody := strings.NewReader(incorrectJson)

		expectedResponse := newErrorResponse(http.StatusBadRequest, ErrInvalidRequestSchema.Error())
		expectedBody, err := json.Marshal(expectedResponse)
		require.NoError(t, err)

		mockSrv := mockAuthSrv.NewAuthService(t)
		handler := NewHandler(mockSrv, Config{MaxLenPassword: 128, MinLenPassword: 8})

		req := httptest.NewRequest(http.MethodPost, "/reset-password", requestBody)
		res := httptest.NewRecorder()

		handler.ResetUserPassword(res, req)

		assert.Equal(t, expectedResponse.Code, res.Code)
		assert.Equal(t, string(expectedBody), res.Body.String())
	})
}

func TestCheckRecoveryCode(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Success check code",
			Request: dto.RecoveryCodeRequest{
				Code: "123456",
			},
			ExpectedResponse:   newResponse(api.StatusOK),
			ExpectedStatusCode: http.StatusOK,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("CheckRecoveryCode", ctx, "123456").Return(nil)
			},
		},
		{
			Name: "Wrong or expired code",
			Request: dto.RecoveryCodeRequest{
				Code: "000000",
			},
			ExpectedResponse:   newErrorResponse(http.StatusInternalServerError, ErrInternalServerError.Error()),
			ExpectedStatusCode: http.StatusInternalServerError,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("CheckRecoveryCode", ctx, "000000").Return(errors.New("invalid code"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockSrv := mockAuthSrv.NewAuthService(t)
			if test.MockBehavior != nil {
				test.MockBehavior(mockSrv)
			}

			handler := NewHandler(mockSrv, Config{})

			requestJson, err := json.Marshal(test.Request)
			require.NoError(t, err, "request marshal should not return error")

			requestReader := bytes.NewReader(requestJson)
			request := httptest.NewRequest(http.MethodPost, "/verify-code", requestReader)
			response := httptest.NewRecorder()

			handler.CheckRecoveryCode(response, request)

			responseJson, err := json.Marshal(test.ExpectedResponse)
			require.NoError(t, err, "response marshal should not return error")

			assert.Equal(t, test.ExpectedStatusCode, response.Code, "incorrect status code")
			assert.Equal(t, string(responseJson), response.Body.String(), "incorrect body")
		})
	}
}

func TestCheckRecoveryCodeWithRawJSON(t *testing.T) {
	t.Run("Incorrect JSON", func(t *testing.T) {
		incorrectJson := `{"code":123456`
		requestBody := strings.NewReader(incorrectJson)

		expectedResponse := newErrorResponse(http.StatusBadRequest, ErrInvalidRequestSchema.Error())
		expectedBody, err := json.Marshal(expectedResponse)
		require.NoError(t, err)

		mockSrv := mockAuthSrv.NewAuthService(t)
		handler := NewHandler(mockSrv, Config{})

		req := httptest.NewRequest(http.MethodPost, "/verify-code", requestBody)
		res := httptest.NewRecorder()

		handler.CheckRecoveryCode(res, req)

		assert.Equal(t, expectedResponse.Code, res.Code)
		assert.Equal(t, string(expectedBody), res.Body.String())
	})
}

type mockTransport struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req)
}

func TestVkOAuthCallback(t *testing.T) {
	const codeParam = "code"
	const messageParam = "message"
	const redirectTo = "/"
	const testUserEmail = "user@example.com"
	vkOAuthConf := &config.VkOAuth{APIMethod: "https://api.vk.com/method/users.get?access_token=%s"}

	successMockClient := &http.Client{
		Transport: &mockTransport{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				vkResp := api.VkAPIUsersData{
					Response: []api.VkAPIUserData{{FirstName: "Ivan"}},
				}
				body, _ := json.Marshal(vkResp)

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBuffer(body)),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	errorMockClient := &http.Client{
		Transport: &mockTransport{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("network error")
			},
		},
	}

	invalidSchemeMockClient := &http.Client{
		Transport: &mockTransport{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				body := []byte("invalid json")

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBuffer(body)),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	emptyMockClinet := &http.Client{
		Transport: &mockTransport{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				body := []byte(`{"response":[]}`)

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBuffer(body)),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	tests := []struct {
		Name                    string
		ExpectedCode            int
		ExpectedMessage         string
		ExpectError             bool
		OAuthCode               string
		VkOAuthMockBehavior     func(*vkOAuthMocks.VkOAuth)
		AuthServiceMockBehavior func(*mockAuthSrv.AuthService)
	}{
		{
			Name:            "no error",
			ExpectedCode:    http.StatusOK,
			ExpectedMessage: "success",
			ExpectError:     false,
			OAuthCode:       "valid_code",
			VkOAuthMockBehavior: func(v *vkOAuthMocks.VkOAuth) {
				token := &oauth2.Token{AccessToken: "fake-token"}
				token = token.WithExtra(map[string]any{"email": testUserEmail})

				v.On("Exchange", mock.Anything, "valid_code").Return(token, nil)
				v.On("Client", mock.Anything, token).Return(successMockClient)
			},
			AuthServiceMockBehavior: func(a *mockAuthSrv.AuthService) {
				a.On("EnsureUserByEmail", mock.Anything, mock.AnythingOfType("dto.RegistrationUser")).
					Return(serviceDto.UserInfo{Link: common.FixedUserUuiD, Email: testUserEmail}, nil)

				a.On("SaveRefreshTokenFroUser", mock.Anything, mock.AnythingOfType("dto.UserInfo"), mock.Anything).
					Return(nil)

				a.On("CreateSessionForUser", mock.Anything, mock.Anything).
					Return("fake-session-id", nil)
			},
		},
		{
			Name:                    "empty code",
			ExpectedCode:            http.StatusBadRequest,
			ExpectedMessage:         ErrOAuthCodeEmpty.Error(),
			ExpectError:             true,
			OAuthCode:               "",
			VkOAuthMockBehavior:     nil,
			AuthServiceMockBehavior: nil,
		},
		{
			Name:            "exchange error",
			ExpectedCode:    http.StatusBadGateway,
			ExpectedMessage: ErrOAuthExchangeFailed.Error(),
			ExpectError:     true,
			OAuthCode:       "invalid_code",
			VkOAuthMockBehavior: func(v *vkOAuthMocks.VkOAuth) {
				v.On("Exchange", mock.Anything, "invalid_code").
					Return((*oauth2.Token)(nil), errors.New("exchange failed"))
			},
			AuthServiceMockBehavior: nil,
		},
		{
			Name:            "no email provided",
			ExpectedCode:    http.StatusBadGateway,
			ExpectedMessage: ErrOAuthNoEmailProvided.Error(),
			ExpectError:     true,
			OAuthCode:       "valid_code",
			VkOAuthMockBehavior: func(v *vkOAuthMocks.VkOAuth) {
				token := &oauth2.Token{AccessToken: "fake-token"}
				v.On("Exchange", mock.Anything, "valid_code").Return(token, nil)
			},
			AuthServiceMockBehavior: nil,
		},
		{
			Name:            "not string email",
			ExpectedCode:    http.StatusBadGateway,
			ExpectedMessage: ErrOAuthNoEmailProvided.Error(),
			ExpectError:     true,
			OAuthCode:       "valid_code",
			VkOAuthMockBehavior: func(v *vkOAuthMocks.VkOAuth) {
				token := &oauth2.Token{AccessToken: "fake-token"}
				token = token.WithExtra(map[string]any{"email": 123456})
				v.On("Exchange", mock.Anything, "valid_code").Return(token, nil)
			},
			AuthServiceMockBehavior: nil,
		},
		{
			Name:            "invalid email",
			ExpectedCode:    http.StatusBadGateway,
			ExpectedMessage: ErrOAuthInvalidEmail.Error(),
			ExpectError:     true,
			OAuthCode:       "valid_code",
			VkOAuthMockBehavior: func(v *vkOAuthMocks.VkOAuth) {
				token := &oauth2.Token{AccessToken: "fake-token"}
				token = token.WithExtra(map[string]any{"email": "testmain.ru"})
				v.On("Exchange", mock.Anything, "valid_code").Return(token, nil)
			},
			AuthServiceMockBehavior: nil,
		},
		{
			Name:            "vk api client error",
			ExpectedCode:    http.StatusBadGateway,
			ExpectedMessage: ErrOAuthCannotRequestUserData.Error(),
			ExpectError:     true,
			OAuthCode:       "valid_code",
			VkOAuthMockBehavior: func(v *vkOAuthMocks.VkOAuth) {
				token := &oauth2.Token{AccessToken: "fake-token"}
				token = token.WithExtra(map[string]any{"email": testUserEmail})

				v.On("Exchange", mock.Anything, "valid_code").Return(token, nil)
				v.On("Client", mock.Anything, token).Return(errorMockClient)
			},
			AuthServiceMockBehavior: nil,
		},
		{
			Name:            "response with invalid scheme from vk api",
			ExpectedCode:    http.StatusInternalServerError,
			ExpectedMessage: ErrOAuthInternalServerError.Error(),
			ExpectError:     true,
			OAuthCode:       "valid_code",
			VkOAuthMockBehavior: func(v *vkOAuthMocks.VkOAuth) {
				token := &oauth2.Token{AccessToken: "fake-token"}
				token = token.WithExtra(map[string]any{"email": testUserEmail})

				v.On("Exchange", mock.Anything, "valid_code").Return(token, nil)
				v.On("Client", mock.Anything, token).Return(invalidSchemeMockClient)
			},
			AuthServiceMockBehavior: nil,
		},
		{
			Name:            "empty response from vk api",
			ExpectedCode:    http.StatusInternalServerError,
			ExpectedMessage: ErrOAuthEmptyUserData.Error(),
			ExpectError:     true,
			OAuthCode:       "valid_code",
			VkOAuthMockBehavior: func(v *vkOAuthMocks.VkOAuth) {
				token := &oauth2.Token{AccessToken: "fake-token"}
				token = token.WithExtra(map[string]any{"email": testUserEmail})

				v.On("Exchange", mock.Anything, "valid_code").Return(token, nil)
				v.On("Client", mock.Anything, token).Return(emptyMockClinet)
			},
			AuthServiceMockBehavior: nil,
		},
		{
			Name:            "ensure user error",
			ExpectedCode:    http.StatusInternalServerError,
			ExpectedMessage: ErrOAuthInternalServerError.Error(),
			ExpectError:     true,
			OAuthCode:       "valid_code",
			VkOAuthMockBehavior: func(v *vkOAuthMocks.VkOAuth) {
				token := &oauth2.Token{AccessToken: "fake-token"}
				token = token.WithExtra(map[string]any{"email": testUserEmail})

				v.On("Exchange", mock.Anything, "valid_code").Return(token, nil)
				v.On("Client", mock.Anything, token).Return(successMockClient)
			},
			AuthServiceMockBehavior: func(a *mockAuthSrv.AuthService) {
				a.On("EnsureUserByEmail", mock.Anything, mock.AnythingOfType("dto.RegistrationUser")).
					Return(serviceDto.UserInfo{}, errors.New("cannot create user"))
			},
		},
		{
			Name:            "save refresh token error",
			ExpectedCode:    http.StatusInternalServerError,
			ExpectedMessage: ErrOAuthInternalServerError.Error(),
			ExpectError:     true,
			OAuthCode:       "valid_code",
			VkOAuthMockBehavior: func(v *vkOAuthMocks.VkOAuth) {
				token := &oauth2.Token{AccessToken: "fake-token"}
				token = token.WithExtra(map[string]any{"email": testUserEmail})

				v.On("Exchange", mock.Anything, "valid_code").Return(token, nil)
				v.On("Client", mock.Anything, token).Return(successMockClient)
			},
			AuthServiceMockBehavior: func(a *mockAuthSrv.AuthService) {
				a.On("EnsureUserByEmail", mock.Anything, mock.AnythingOfType("dto.RegistrationUser")).
					Return(serviceDto.UserInfo{Link: common.FixedUserUuiD, Email: testUserEmail}, nil)

				a.On("SaveRefreshTokenFroUser", mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("cannot save refresh token"))
			},
		},
		{
			Name:            "create session error",
			ExpectedCode:    http.StatusInternalServerError,
			ExpectedMessage: ErrOAuthInternalServerError.Error(),
			ExpectError:     true,
			OAuthCode:       "valid_code",
			VkOAuthMockBehavior: func(v *vkOAuthMocks.VkOAuth) {
				token := &oauth2.Token{AccessToken: "fake-token"}
				token = token.WithExtra(map[string]any{"email": testUserEmail})

				v.On("Exchange", mock.Anything, "valid_code").Return(token, nil)
				v.On("Client", mock.Anything, token).Return(successMockClient)
			},
			AuthServiceMockBehavior: func(a *mockAuthSrv.AuthService) {
				a.On("EnsureUserByEmail", mock.Anything, mock.AnythingOfType("dto.RegistrationUser")).
					Return(serviceDto.UserInfo{Link: common.FixedUserUuiD, Email: testUserEmail}, nil)

				a.On("SaveRefreshTokenFroUser", mock.Anything, mock.AnythingOfType("dto.UserInfo"), mock.Anything).
					Return(nil)

				a.On("CreateSessionForUser", mock.Anything, mock.Anything).
					Return("", errors.New("cannot create session"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockVkOAuth := new(vkOAuthMocks.VkOAuth)
			if test.VkOAuthMockBehavior != nil {
				test.VkOAuthMockBehavior(mockVkOAuth)
			}

			mockAuthService := new(mockAuthSrv.AuthService)
			if test.AuthServiceMockBehavior != nil {
				test.AuthServiceMockBehavior(mockAuthService)
			}

			handler := NewHandler(mockAuthService, Config{})

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/callback?code=%s", test.OAuthCode), nil)
			res := httptest.NewRecorder()

			callbackHandler := handler.VkOAuthCallback(vkOAuthConf, redirectTo, mockVkOAuth)
			callbackHandler(res, req)

			r := res.Result()

			require.Equal(t, http.StatusFound, r.StatusCode, "http code must be 302")

			location, err := r.Location()
			require.NoError(t, err, "location must be provided")

			q := location.Query()

			require.True(t, q.Has(codeParam), "must be code query param")
			assert.Equal(t, strconv.Itoa(test.ExpectedCode), q.Get(codeParam), "codes must be equal")

			require.True(t, q.Has(messageParam), "must be message query param")
			assert.Equal(t, test.ExpectedMessage, q.Get(messageParam), "messages must be equal")

			if !test.ExpectError {
				cookies := r.Cookies()
				require.NotEmpty(t, cookies)
				assert.Equal(t, "fake-session-id", cookies[0].Value)
			}

			mockVkOAuth.AssertExpectations(t)
			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestSetCSRFCookieHandler(t *testing.T) {
	const csrfCookieKey = "csrf_token"

	fixedTime := time.Now().Add(time.Hour).Truncate(time.Second)
	validSessionID := "session-123"

	tests := []struct {
		Name          string
		SessionCookie *http.Cookie
		ExpectedCode  int
		ExpectedToken string
		MockBehavior  func(*mockAuthSrv.AuthService)
	}{
		{
			Name:          "success set cookie",
			SessionCookie: &http.Cookie{Name: service.SessiondIdKey, Value: validSessionID},
			ExpectedCode:  http.StatusOK,
			ExpectedToken: "signed-hmac-token",
			MockBehavior: func(a *mockAuthSrv.AuthService) {
				a.On("GetCSRFTokenExpireTime", mock.Anything).Return(fixedTime, nil)
				a.On("GenerateCSRFToken", mock.Anything, validSessionID, fixedTime.Unix()).
					Return("signed-hmac-token", nil)
			},
		},
		{
			Name:          "no session cookie - Unauthorized",
			SessionCookie: nil,
			ExpectedCode:  http.StatusUnauthorized,
			MockBehavior:  nil,
		},
		{
			Name:          "error getting expire time",
			SessionCookie: &http.Cookie{Name: service.SessiondIdKey, Value: validSessionID},
			ExpectedCode:  http.StatusInternalServerError,
			MockBehavior: func(a *mockAuthSrv.AuthService) {
				a.On("GetCSRFTokenExpireTime", mock.Anything).
					Return(time.Time{}, errors.New("db error"))
			},
		},
		{
			Name:          "error generating token",
			SessionCookie: &http.Cookie{Name: service.SessiondIdKey, Value: validSessionID},
			ExpectedCode:  http.StatusInternalServerError,
			MockBehavior: func(a *mockAuthSrv.AuthService) {
				a.On("GetCSRFTokenExpireTime", mock.Anything).Return(fixedTime, nil)
				a.On("GenerateCSRFToken", mock.Anything, validSessionID, fixedTime.Unix()).
					Return("", errors.New("crypto error"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockAuthService := new(mockAuthSrv.AuthService)
			if test.MockBehavior != nil {
				test.MockBehavior(mockAuthService)
			}

			handler := NewHandler(mockAuthService, Config{})

			res := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			if test.SessionCookie != nil {
				req.AddCookie(test.SessionCookie)
			}

			handler.SetCSRFCookieHandler(res, req)

			assert.Equal(t, test.ExpectedCode, res.Code)

			cookies := res.Result().Cookies()
			var csrfCookie *http.Cookie
			for _, c := range cookies {
				if c.Name == csrfCookieKey {
					csrfCookie = c
					break
				}
			}

			if test.ExpectedCode == http.StatusOK {
				require.NotNil(t, csrfCookie, "CSRF cookie must be present")
				assert.Equal(t, test.ExpectedToken, csrfCookie.Value)
				assert.True(t, csrfCookie.Secure)
				assert.False(t, csrfCookie.HttpOnly)
				assert.Equal(t, "/", csrfCookie.Path)
				assert.Equal(t, http.SameSiteLaxMode, csrfCookie.SameSite)
				assert.Equal(t, fixedTime.Unix(), csrfCookie.Expires.Unix())
			} else {
				assert.Nil(t, csrfCookie, "CSRF cookie should not be set on error")
			}

			mockAuthService.AssertExpectations(t)
		})
	}
}
