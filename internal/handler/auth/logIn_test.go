package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/handler/auth/mocks"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	service "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth"
	"github.com/stretchr/testify/assert"
)

func TestLogInUser(t *testing.T) {
	tests := []struct {
		nameTest           string
		jsonBody           string
		mockBehavior       func(m *mocks.AuthService)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			nameTest: "Success login",
			jsonBody: `{"email":"test@mail.ru","password":"123456"}`,
			mockBehavior: func(m *mocks.AuthService) {
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
			expectedStatusCode: http.StatusOK,
			expectedResponse:   "{\"message\":\"user was successfully logged in\",\"profile\":{\"id\":\"11111111-1111-1111-1111-111111111111\",\"display_name\":\"Artem\",\"email\":\"test@mail.ru\",\"boards\":null}}\n",
		},
		{
			nameTest:           "Incorrect JSON",
			jsonBody:           `{"email":"test@mail.ru",,,}`,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"decoding request is incorrect: invalid character ',' looking for beginning of object key string\"}\n",
		},
		{
			nameTest: "Wrong password or email",
			jsonBody: `{"email":"artem@mail.ru","password":"wrong_password"}`,
			mockBehavior: func(m *mocks.AuthService) {
				ctx := context.Background()
				m.On("LogIn", ctx, "artem@mail.ru", "wrong_password").Return(models.User{}, "", service.ErrorWrongPassword)
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   "{\"error\":\"wrong email or password\"}\n",
		},
		{
			nameTest:           "Size password smaller than 6",
			jsonBody:           `{"email":"artem@mail.ru","password":"123"}`,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"ValidatorRequestAuth: password must contain minimum 6\"}\n",
		},
		{
			nameTest:           "Email hasn't @",
			jsonBody:           `{"email":"testmail.ru","password":"1234567"}`,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"ValidatorRequestAuth: invalid email format\"}\n",
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockLogInService := mocks.NewAuthService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockLogInService)
			}

			handler := NewAuthHandler(mockLogInService)

			body := strings.NewReader(test.jsonBody)

			request := httptest.NewRequest(http.MethodPost, "/login", body)
			response := httptest.NewRecorder()
			handler.LogInUser(response, request)

			assert.Equal(t, test.expectedStatusCode, response.Code, "incorrect status code")
			assert.Equal(t, test.expectedResponse, response.Body.String(), "incorrect error message")

			if test.expectedStatusCode == http.StatusOK {
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
