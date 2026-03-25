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

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/handler/dto"
	mockAuthSrv "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/handler/mock_auth_srv"
	vkOAuthMocks "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/handler/mock_vk_oauth"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/models"
	service "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
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
			Request: api.LogInRequest{
				Email:    "test@mail.ru",
				Password: "123456789",
			},
			ExpectedResponse: newOkResponse(api.StatusOK, models.User{
				Link:        common.FixedUserUuiD,
				DisplayName: "Artem",
				Email:       "test@mail.ru",
			}),
			ExpectedStatusCode: http.StatusOK,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("LogIn", ctx, dto.LoginInfoRequest{
					Email:    "test@mail.ru",
					Password: "123456789",
				}).Return(
					serviceDto.UserInfoResponse{
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
			Request: api.LogInRequest{
				Email:    "artem@mail.ru",
				Password: "wrong_password",
			},
			ExpectedResponse:   newErrorResponse(http.StatusUnauthorized, ErrWrongEmailOrPassword.Error()),
			ExpectedStatusCode: http.StatusUnauthorized,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("LogIn", ctx, dto.LoginInfoRequest{
					Email:    "artem@mail.ru",
					Password: "wrong_password",
				}).Return(serviceDto.UserInfoResponse{}, "", service.ErrorWrongPassword)
			},
		},
		{
			Name: "Size password smaller than 8",
			Request: api.LogInRequest{
				Email:    "artem@mail.ru",
				Password: "123",
			},
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, ErrInvalidEmailOrPassword.Error()),
			ExpectedStatusCode: http.StatusBadRequest,
			MockBehavior:       nil,
		},
		{
			Name: "Size password biger than 128",
			Request: api.LogInRequest{
				Email:    "artem@mail.ru",
				Password: "123111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111",
			},
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, ErrInvalidEmailOrPassword.Error()),
			ExpectedStatusCode: http.StatusBadRequest,
			MockBehavior:       nil,
		},
		{
			Name: "Email hasn't @",
			Request: api.LogInRequest{
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

			handler := NewHandler(mockLogInService)

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
		handler := NewHandler(mockLogInService)

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
			Request: api.RegisterRequest{
				DisplayName:      "Artem",
				Password:         "12345678",
				RepeatedPassword: "12345678",
				Email:            "test@mail.ru",
			},
			ExpectedResponse: newOkResponse(api.StatusOK, models.User{
				Link:        common.FixedUserUuiD,
				DisplayName: "Artem",
				Email:       "test@mail.ru",
			}),
			ExpectedStatusCode: http.StatusCreated,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("Register", ctx, dto.RegistraionInfoRequest{
					Name:     "Artem",
					Email:    "test@mail.ru",
					Password: "12345678",
				}).Return(
					serviceDto.UserInfoResponse{
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
			Request: api.RegisterRequest{
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
			Request: api.RegisterRequest{
				DisplayName:      "Artem",
				Password:         "123456789",
				RepeatedPassword: "123456789",
				Email:            "test@mail.ru",
			},
			ExpectedResponse:   newErrorResponse(http.StatusInternalServerError, ErrInternalServerError.Error()),
			ExpectedStatusCode: http.StatusInternalServerError,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("Register", ctx, dto.RegistraionInfoRequest{
					Name:     "Artem",
					Email:    "test@mail.ru",
					Password: "123456789",
				}).Return(
					serviceDto.UserInfoResponse{},
					"",
					fmt.Errorf("repo.AddUser: user with this email alreday exists"),
				)
			},
		},
		{
			Name: "Incorrect symbol in password",
			Request: api.RegisterRequest{
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
			Request: api.RegisterRequest{
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
			Request: api.RegisterRequest{
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
			Request: api.RegisterRequest{
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
			Request: api.RegisterRequest{
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
			Request: api.RegisterRequest{
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
			Request: api.RegisterRequest{
				DisplayName:      "Artem",
				Password:         "123456789",
				RepeatedPassword: "123456789",
				Email:            "test@mail.ru",
			},
			ExpectedResponse:   newErrorResponse(http.StatusInternalServerError, ErrInternalServerError.Error()),
			ExpectedStatusCode: http.StatusInternalServerError,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("Register", ctx, dto.RegistraionInfoRequest{
					Name:     "Artem",
					Password: "123456789",
					Email:    "test@mail.ru",
				}).Return(
					serviceDto.UserInfoResponse{},
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

			handler := NewHandler(mockRegisterService)

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
		handler := NewHandler(mockRegisterService)

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

			handler := NewHandler(mockAuthService)

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
	handler := &AuthHandler{}

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
	tests := []TestCase{
		{
			Name: "Success send email",
			Request: api.PasswordRecoveryRequest{
				Email: "test@mail.ru",
			},
			ExpectedResponse:   newResponse(api.StatusOK),
			ExpectedStatusCode: http.StatusOK,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("SendRecoveryCode", ctx, "test@mail.ru").Return(nil)
			},
		},
		{
			Name: "Service error",
			Request: api.PasswordRecoveryRequest{
				Email: "notfound@mail.ru",
			},
			ExpectedResponse:   newErrorResponse(http.StatusInternalServerError, ErrCannotSendRecoveryCode.Error()),
			ExpectedStatusCode: http.StatusInternalServerError,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("SendRecoveryCode", ctx, "notfound@mail.ru").Return(errors.New("some internal error"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockSrv := mockAuthSrv.NewAuthService(t)
			if test.MockBehavior != nil {
				test.MockBehavior(mockSrv)
			}

			handler := NewHandler(mockSrv)

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
		handler := NewHandler(mockSrv)

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
			Request: api.NewPasswordRequest{
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
			Request: api.NewPasswordRequest{
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
			Request: api.NewPasswordRequest{
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

			handler := NewHandler(mockSrv)

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
		handler := NewHandler(mockSrv)

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
			Request: api.RecoveryCodeRequest{
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
			Request: api.RecoveryCodeRequest{
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

			handler := NewHandler(mockSrv)

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
		handler := NewHandler(mockSrv)

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

func TestVkOAuthCallbackExistingUser(t *testing.T) {
	mockVkOAuth := new(vkOAuthMocks.VkOAuth)
	mockAuthService := new(mockAuthSrv.AuthService)

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

	mockVkOAuth.On("Exchange", mock.Anything, "valid_code").Return(testToken, nil)
	mockVkOAuth.On("Client", mock.Anything, testToken).Return(mockClient)

	mockAuthService.On("GetUserByEmail", mock.Anything, testEmail).
		Return(models.User{Link: common.FixedUserUuiD, Email: testEmail}, nil)

	mockAuthService.On("CreateSessionForUser", mock.Anything, mock.AnythingOfType("uuid.UUID")).
		Return("fake-session-id", nil)

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
				a.On("EnsureUserByEmail", mock.Anything, mock.AnythingOfType("dto.UserInfo")).
					Return(models.User{ID: common.FixedUserUuiD, Email: testUserEmail}, nil)

				a.On("SaveRefreshTokenFroUser", mock.Anything, mock.AnythingOfType("dto.UserInfo"), mock.Anything).
					Return(nil)

// func TestVkOAuthCallbackNewUserRegistration(t *testing.T) {
// 	mockVkOAuth := new(vkOAuthMocks.VkOAuth)
// 	mockAuthService := new(authServiceMocks.AuthService)

// 	handler := &AuthHandler{Srv: mockAuthService}

// 	conf := &config.VkOAuth{APIMethod: "https://api.vk.com/method/users.get?access_token=%s"}
// 	redirectTo := "/"
// 	testEmail := "new@example.com"
// 	testToken := &oauth2.Token{AccessToken: "fake-token"}
// 	testToken = testToken.WithExtra(map[string]any{"email": testEmail})

// 	mockClient := &http.Client{
// 		Transport: &mockTransport{
// 			RoundTripFunc: func(req *http.Request) *http.Response {
// 				vkResp := api.VkAPIUsersData{
// 					Response: []api.VkAPIUserData{{FirstName: "Alice"}},
// 				}
// 				body, _ := json.Marshal(vkResp)
// 				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer(body))}
// 			},
// 		},
// 	}

// 	mockVkOAuth.On("Exchange", mock.Anything, "valid_code").Return(testToken, nil)
// 	mockVkOAuth.On("Client", mock.Anything, testToken).Return(mockClient)

// 	mockAuthService.On("GetUserByEmail", mock.Anything, testEmail).
// 		Return(models.User{}, common.ErrorNonexistentUser)

// 	mockAuthService.On("Register", mock.Anything, dto.RegistraionInfoRequest{
// 		Name:     "Alice",
// 		Password: "g1TNzmfXeEAw5-qarEXPmY7q6dddWSuhyMl789RtUoY=",
// 		Email:    testEmail,
// 	}).
// 		Return(models.User{Link: common.FixedUserUuiD}, "new-session-id", nil)

// 	req := httptest.NewRequest(http.MethodGet, "/callback?code=valid_code", nil)
// 	rr := httptest.NewRecorder()

// 	httpHandler := handler.VkOAuthCallback(conf, redirectTo, mockVkOAuth)
// 	httpHandler(rr, req)

// 	assert.Equal(t, http.StatusFound, rr.Code)
// 	assert.Equal(t, "/?code=200&message=success", rr.Header().Get("Location"))

// 	mockVkOAuth.AssertExpectations(t)
// 	mockAuthService.AssertExpectations(t)
// }

func TestVkOAuthCallbackExchangeError(t *testing.T) {
	mockVkOAuth := new(vkOAuthMocks.VkOAuth)
	mockAuthService := new(mockAuthSrv.AuthService)

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
