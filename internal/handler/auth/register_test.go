package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/handler/auth/mocks"
	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestRegsisterUser(t *testing.T) {
	tests := []struct {
		nameTest           string
		jsonBody           string
		mockBehavior       func(m *mocks.AuthService)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			nameTest: "Success registration",
			jsonBody: `{"display_name":"Artem","password":"123456","repeated_password":"123456","email":"test@mail.ru","boards":null}`,
			mockBehavior: func(m *mocks.AuthService) {
				ctx := context.Background()
				m.On("Register", ctx, "Artem", "123456", "test@mail.ru").Return(models.User{
					ID:          common.FixedUserUuiD,
					DisplayName: "Artem",
					Email:       "test@mail.ru",
				},
					common.FixedSessionID,
					nil,
				)
			},
			expectedStatusCode: http.StatusCreated,
			expectedResponse:   "{\"message\":\"user was successsfully created\",\"profile\":{\"id\":\"11111111-1111-1111-1111-111111111111\",\"display_name\":\"Artem\",\"email\":\"test@mail.ru\",\"boards\":null}}\n",
		},
		{
			nameTest:           "Incorrect JSON",
			jsonBody:           `{"display_name":"Artem",,,}`,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"decoding request is incorrect: invalid character ',' looking for beginning of object key string\"}\n",
		},
		{
			nameTest:           "Passwords do not match",
			jsonBody:           `{"display_name":"Artem","password":"123456","repeated_password":"654321","email":"test@mail.ru"}`,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"ValidatorRequestAuth: passwords don't match\"}\n",
		},
		{
			nameTest: "Email is already existing",
			jsonBody: `{"display_name":"Artem","password":"123456","repeated_password":"123456","email":"test@mail.ru"}`,
			mockBehavior: func(m *mocks.AuthService) {
				ctx := context.Background()
				m.On("Register", ctx, "Artem", "123456", "test@mail.ru").Return(models.User{}, "", fmt.Errorf("repo.AddUser: user with this email alreday exists"))
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"cannot register user: repo.AddUser: user with this email alreday exists\"}\n",
		},
		{
			nameTest:           "Incorrect symbol in password",
			jsonBody:           `{"display_name":"Artem","password":"бобёр","repeated_password":"бобёр","email":"test@mail.ru"}`,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"ValidatorRequestAuth: allowed only a-z, A-Z, 0-9, and /?!@\"}\n",
		},
		{
			nameTest:           "Incorrect symbol in email",
			jsonBody:           `{"display_name":"Artem","password":"123455","repeated_password":"123455","email":"бобёр@mail.ru"}`,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"ValidatorRequestAuth: allowed only a-z, A-Z, 0-9, and /?!@\"}\n",
		},
		{
			nameTest:           "Size password smaller, then 6",
			jsonBody:           `{"display_name":"Artem","password":"123","repeated_password":"123","email":"test@mail.ru"}`,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"ValidatorRequestAuth: password must contain minimum 6\"}\n",
		},
		{
			nameTest:           "Email has 2 @",
			jsonBody:           `{"display_name":"Artem","password":"1234567","repeated_password":"1234567","email":"test@m@ail.ru"}`,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"ValidatorRequestAuth: invalid email format\"}\n",
		},
		{
			nameTest:           "Email has't @",
			jsonBody:           `{"display_name":"Artem","password":"1234567","repeated_password":"1234567","email":"testmail.ru"}`,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"ValidatorRequestAuth: invalid email format\"}\n",
		},
		{
			nameTest:           "Email has @.",
			jsonBody:           `{"display_name":"Artem","password":"1234567","repeated_password":"1234567","email":"test@.ru"}`,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"ValidatorRequestAuth: invalid email format\"}\n",
		},
		{
			nameTest: "Error during hash password",
			jsonBody: `{"display_name":"Artem","password":"123456","repeated_password":"123456","email":"test@mail.ru"}`,
			mockBehavior: func(m *mocks.AuthService) {
				ctx := context.Background()
				m.On("Register", ctx, "Artem", "123456", "test@mail.ru").Return(models.User{}, "", fmt.Errorf("failed to create hash: error bcrypt"))
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"cannot register user: failed to create hash: error bcrypt\"}\n",
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRegisterService := mocks.NewAuthService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRegisterService)
			}

			handler := NewAuthHandler(mockRegisterService)

			body := strings.NewReader(test.jsonBody)
			request := httptest.NewRequest(http.MethodPost, "/register", body)
			response := httptest.NewRecorder()

			handler.RegisterUser(response, request)

			assert.Equal(t, test.expectedStatusCode, response.Code, "incorrect status code")
			assert.Equal(t, test.expectedResponse, response.Body.String(), "incorrect error")

			if test.expectedStatusCode == http.StatusCreated {
				res := response.Result()
				cookies := res.Cookies()

				var sessionCookie *http.Cookie
				for _, c := range cookies {
					if c.Name == "session_id" {
						sessionCookie = c
						break
					}
				}

				assert.NotNil(t, sessionCookie, "cookie wasn`t find in responce")
				assert.NotEmpty(t, sessionCookie.Value, "empty value sesstion Cookie")
				assert.True(t, sessionCookie.HttpOnly, "flag HttpOnlly must be only true")
			}
		})
	}
}
