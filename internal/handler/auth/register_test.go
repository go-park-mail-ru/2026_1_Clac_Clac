package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	mockAuthSrv "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/handler/auth/mock_auth_srv"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
				ID:          common.FixedUserUuiD,
				DisplayName: "Artem",
				Email:       "test@mail.ru",
			}),
			ExpectedStatusCode: http.StatusCreated,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("Register", ctx, "Artem", "12345678", "test@mail.ru").Return(
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
			Name: "Passwords do not match",
			Request: api.RegisterRequest{
				DisplayName:      "Artem",
				Password:         "12345678",
				RepeatedPassword: "65432178",
				Email:            "test@mail.ru",
			},
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, invalidEmailOrPassword),
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
			ExpectedResponse:   newErrorResponse(http.StatusInternalServerError, somethingWentWrong),
			ExpectedStatusCode: http.StatusInternalServerError,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("Register", ctx, "Artem", "123456789", "test@mail.ru").Return(
					models.User{},
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
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, invalidEmailOrPassword),
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
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, invalidEmailOrPassword),
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
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, invalidEmailOrPassword),
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
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, invalidEmailOrPassword),
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
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, invalidEmailOrPassword),
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
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, invalidEmailOrPassword),
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
			ExpectedResponse:   newErrorResponse(http.StatusInternalServerError, somethingWentWrong),
			ExpectedStatusCode: http.StatusInternalServerError,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("Register", ctx, "Artem", "123456789", "test@mail.ru").Return(
					models.User{},
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

			handler := NewAuthHandler(mockRegisterService)

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

		expectedResponse := newErrorResponse(http.StatusBadRequest, invalidDataMessage)
		expectedBody, err := json.Marshal(expectedResponse)
		require.NoError(t, err, "response marshal should not return error")

		mockRegisterService := mockAuthSrv.NewAuthService(t)
		handler := NewAuthHandler(mockRegisterService)

		req := httptest.NewRequest(http.MethodPost, "/register", requestBody)
		res := httptest.NewRecorder()

		handler.RegisterUser(res, req)

		assert.Equal(t, expectedResponse.Code, res.Code, "incorrect status code")
		assert.Equal(t, string(expectedBody), res.Body.String(), "incorrect body")
	})
}
