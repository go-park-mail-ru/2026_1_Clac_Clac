package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	common "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	service "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth"
	"github.com/stretchr/testify/assert"
)

type SpyRegistrationService struct {
	SpyWorkService func(ctx context.Context, name, password, email string) (models.User, string, error)
}

func (s *SpyRegistrationService) Register(ctx context.Context, name, password, email string) (models.User, string, error) {
	return s.SpyWorkService(ctx, name, password, email)
}

func TestRegsisterUser(t *testing.T) {
	tests := []struct {
		nameTest           string
		jsonBody           string
		funcWorkService    func(ctx context.Context, name, password, email string) (models.User, string, error)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			nameTest: "Success registration",
			jsonBody: `{"display_name":"Artem","password":"123456","email":"test@mail.ru"}`,
			funcWorkService: func(ctx context.Context, name, password, email string) (models.User, string, error) {
				user := models.User{
					ID:           common.FixedUuiD,
					DisplayName:  name,
					PasswordHash: password,
					Email:        email,
				}

				return user, common.FixedSessionID, nil
			},
			expectedStatusCode: http.StatusCreated,
			expectedResponse:   "{\"message\":\"user was successsfully created\",\"profile\":{\"id\":\"11111111-1111-1111-1111-111111111111\",\"display_name\":\"Artem\",\"email\":\"test@mail.ru\"}}\n",
		},
		{
			nameTest: "Incorrect JSON",
			jsonBody: `{"display_name":"Artem",,,}`,
			funcWorkService: func(ctx context.Context, name, password, email string) (models.User, string, error) {
				return models.User{}, "", nil
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"decoding request is incorrect: invalid character ',' looking for beginning of object key string\"}\n",
		},
		{
			nameTest: "Email is already existing",
			jsonBody: `{"display_name":"Artem","password":"123456","email":"test@mail.ru"}`,
			funcWorkService: func(ctx context.Context, name, password, email string) (models.User, string, error) {
				return models.User{}, "", fmt.Errorf("repo.AddUser: user with this email alreday exists")
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"cannot register user: repo.AddUser: user with this email alreday exists\"}\n",
		},
		{
			nameTest: "Incorrect symbol in password",
			jsonBody: `{"display_name":"Artem","password":"бобёр","email":"test@mail.ru"}`,
			funcWorkService: func(ctx context.Context, name, password, email string) (models.User, string, error) {
				return models.User{}, "", ErrorIncorrectSymbol
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"allowed only a-z, A-Z, 0-9, and /?!@\"}\n",
		},
		{
			nameTest: "Incorrect symbol in email",
			jsonBody: `{"display_name":"Artem","password":"123455","email":"бобёр@mail.ru"}`,
			funcWorkService: func(ctx context.Context, name, password, email string) (models.User, string, error) {
				return models.User{}, "", ErrorIncorrectSymbol
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"allowed only a-z, A-Z, 0-9, and /?!@\"}\n",
		},
		{
			nameTest: "Size password smaller, then 6",
			jsonBody: `{"display_name":"Artem","password":"123","email":"test@mail.ru"}`,
			funcWorkService: func(ctx context.Context, name, password, email string) (models.User, string, error) {
				return models.User{}, "", ErrorLenPassword
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"password must contain minimum 6\"}\n",
		},
		{
			nameTest: "Email has 2 @",
			jsonBody: `{"display_name":"Artem","password":"1234567","email":"test@m@ail.ru"}`,
			funcWorkService: func(ctx context.Context, name, password, email string) (models.User, string, error) {
				return models.User{}, "", ErrorIncorrectEmail
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"invalid email format\"}\n",
		},
		{
			nameTest: "Email has`t @",
			jsonBody: `{"display_name":"Artem","password":"1234567","email":"testmail.ru"}`,
			funcWorkService: func(ctx context.Context, name, password, email string) (models.User, string, error) {
				return models.User{}, "", ErrorIncorrectEmail
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"invalid email format\"}\n",
		},
		{
			nameTest: "Email has @.",
			jsonBody: `{"display_name":"Artem","password":"1234567","email":"test@.ru"}`,
			funcWorkService: func(ctx context.Context, name, password, email string) (models.User, string, error) {
				return models.User{}, "", ErrorIncorrectEmail
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"invalid email format\"}\n",
		},
		{
			nameTest: "Error during hash password",
			jsonBody: `{"display_name":"Artem","password":"123456","email":"test@mail.ru"}`,
			funcWorkService: func(ctx context.Context, name, password, email string) (models.User, string, error) {
				return models.User{}, "", fmt.Errorf("%w: error bcrypt", service.ErrorCreateHash)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"cannot register user: failed to create hash: error bcrypt\"}\n",
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			registrationService := SpyRegistrationService{
				SpyWorkService: test.funcWorkService,
			}

			handler := NewRegisterHandler(&registrationService)

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
