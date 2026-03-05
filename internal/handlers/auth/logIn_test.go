package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	common "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	service "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth"
	"github.com/stretchr/testify/assert"
)

type SpyLogInService struct {
	SpyWorkService func(ctx context.Context, email, password string) (models.User, string, error)
}

func (s *SpyLogInService) Login(ctx context.Context, email, password string) (models.User, string, error) {
	return s.SpyWorkService(ctx, email, password)
}

func TestLogInUser(t *testing.T) {
	tests := []struct {
		nameTest           string
		jsonBody           string
		funcWorkService    func(ctx context.Context, email, password string) (models.User, string, error)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			nameTest: "Success login",
			jsonBody: `{"email":"test@mail.ru","password":"123456"}`,
			funcWorkService: func(ctx context.Context, email, password string) (models.User, string, error) {
				user := models.User{
					ID:           common.FixedUuiD,
					DisplayName:  "Artem",
					PasswordHash: "hash_is_not_important_here",
					Email:        email,
				}
				return user, common.FixedSessionID, nil
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   "{\"message\":\"user was successfully logged in\",\"profile\":{\"id\":\"11111111-1111-1111-1111-111111111111\",\"display_name\":\"Artem\",\"email\":\"test@mail.ru\"}}\n",
		},
		{
			nameTest: "Incorrect JSON",
			jsonBody: `{"email":"test@mail.ru",,,}`,
			funcWorkService: func(ctx context.Context, email, password string) (models.User, string, error) {
				return models.User{}, "", nil
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"decoding request is incorrect: invalid character ',' looking for beginning of object key string\"}\n",
		},
		{
			nameTest: "Wrong password or email",
			jsonBody: `{"email":"artem@mail.ru","password":"wrong_password"}`,
			funcWorkService: func(ctx context.Context, email, password string) (models.User, string, error) {
				return models.User{}, "", service.ErrorWrongPassword
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   "{\"error\":\"wrong email or password\"}\n",
		},
		{
			nameTest: "Size password smaller than 6",
			jsonBody: `{"email":"artem@mail.ru","password":"123"}`,
			funcWorkService: func(ctx context.Context, email, password string) (models.User, string, error) {
				return models.User{}, "", ErrorLenPassword
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"password must contain minimum 6\"}\n",
		},
		{
			nameTest: "Email hasn't @",
			jsonBody: `{"email":"testmail.ru","password":"1234567"}`,
			funcWorkService: func(ctx context.Context, email, password string) (models.User, string, error) {
				return models.User{}, "", ErrorIncorrectEmail
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"invalid email format\"}\n",
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			logInService := &SpyLogInService{
				SpyWorkService: test.funcWorkService,
			}

			handler := NewLogInHandler(logInService)

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
