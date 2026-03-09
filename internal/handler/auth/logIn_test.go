package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	mockAuthSrv "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/handler/auth/mock_auth_srv"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	service "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				Password: "123456",
			},
			ExpectedResponse: newOkResponse(api.StatusOK, models.User{
				ID:          common.FixedUserUuiD,
				DisplayName: "Artem",
				Email:       "test@mail.ru",
			}),
			ExpectedStatusCode: http.StatusOK,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("LogIn", ctx, "test@mail.ru", "123456").Return(
					models.User{
						ID:          common.FixedUserUuiD,
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
			ExpectedResponse:   newErrorResponse(http.StatusUnauthorized, wrongEmailOrPassword),
			ExpectedStatusCode: http.StatusUnauthorized,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("LogIn", ctx, "artem@mail.ru", "wrong_password").Return(models.User{}, "", service.ErrorWrongPassword)
			},
		},
		{
			Name: "Size password smaller than 6",
			Request: api.LogInRequest{
				Email:    "artem@mail.ru",
				Password: "123",
			},
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, invalidEmailOrPassword),
			ExpectedStatusCode: http.StatusBadRequest,
			MockBehavior:       nil,
		},
		{
			Name: "Email hasn't @",
			Request: api.LogInRequest{
				Email:    "testmail.ru",
				Password: "1234567",
			},
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, invalidEmailOrPassword),
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

			handler := NewAuthHandler(mockLogInService)

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

		expectedResponse := newErrorResponse(http.StatusBadRequest, invalidDataMessage)
		expectedBody, err := json.Marshal(expectedResponse)
		require.NoError(t, err, "response marshal should not return error")

		mockLogInService := mockAuthSrv.NewAuthService(t)
		handler := NewAuthHandler(mockLogInService)

		req := httptest.NewRequest(http.MethodPost, "/", requestBody)
		res := httptest.NewRecorder()

		handler.LogInUser(res, req)

		assert.Equal(t, expectedResponse.Code, res.Code, "incorrect status code")
		assert.Equal(t, string(expectedBody), res.Body.String(), "incorrect body")
	})
}
