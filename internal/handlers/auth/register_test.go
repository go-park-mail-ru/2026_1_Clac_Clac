package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	service "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type SpyRegistrationService struct {
	SpyWorkService func(ctx context.Context, name, surname, password, email string) (models.User, error)
}

func (s *SpyRegistrationService) Register(ctx context.Context, name, surname, password, email string) (models.User, error) {
	return s.SpyWorkService(ctx, name, surname, password, email)
}

var (
	fixedUuiD = uuid.MustParse("11111111-1111-1111-1111-111111111111")
)

func TestRegisterUser(t *testing.T) {
	tests := []struct {
		nameTest           string
		jsonBody           string
		funcWorkService    func(ctx context.Context, name, surname, password, email string) (models.User, error)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			nameTest: "Success registration",
			jsonBody: `{"name":"Artem","surname":"Busygin","password":"123456","email":"test@mail.ru"}`,
			funcWorkService: func(ctx context.Context, name, surname, password, email string) (models.User, error) {
				user := models.User{
					ID:       fixedUuiD,
					Name:     name,
					Surname:  surname,
					Password: password,
					Email:    email,
					Boards:   make([]models.Board, 0),
				}

				return user, nil
			},
			expectedStatusCode: http.StatusCreated,
			expectedResponse:   "{\"message\":\"user was successsfully created\",\"profile\":{\"id\":\"11111111-1111-1111-1111-111111111111\",\"name\":\"Artem\",\"surname\":\"Busygin\",\"password\":\"123456\",\"email\":\"test@mail.ru\",\"boards\":[]}}\n",
		},
		{
			nameTest: "Incorrect JSON",
			jsonBody: `{"name":"Artem",,,}`,
			funcWorkService: func(ctx context.Context, name, surname, password, email string) (models.User, error) {
				return models.User{}, nil
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "decoding request is incorrect: invalid character ',' looking for beginning of object key string\n",
		},
		{
			nameTest: "Email is already existing",
			jsonBody: `{"name":"Artem","surname":"Busygin","password":"123456","email":"test@mail.ru"}`,
			funcWorkService: func(ctx context.Context, name, surname, password, email string) (models.User, error) {
				return models.User{}, fmt.Errorf("repo.AddUser: user with this email alreday exists")
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "cannot register user: repo.AddUser: user with this email alreday exists\n",
		},
		{
			nameTest: "Incorrect symbol in surname",
			jsonBody: `{"name":"Artem","surname":"Бусыгин","password":"123343","email":"test@mail.ru"}`,
			funcWorkService: func(ctx context.Context, name, surname, password, email string) (models.User, error) {
				return models.User{}, ErrorIncorrectSymbol
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "allowed only a-z, A-Z, 0-9, and /?!@\n",
		},
		{
			nameTest: "Incorrect symbol in password",
			jsonBody: `{"name":"Artem","surname":"Busygin","password":"бобёр","email":"test@mail.ru"}`,
			funcWorkService: func(ctx context.Context, name, surname, password, email string) (models.User, error) {
				return models.User{}, ErrorIncorrectSymbol
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "allowed only a-z, A-Z, 0-9, and /?!@\n",
		},
		{
			nameTest: "Incorrect symbol in email",
			jsonBody: `{"name":"Artem","surname":"Busygin","password":"123455","email":"бобёр@mail.ru"}`,
			funcWorkService: func(ctx context.Context, name, surname, password, email string) (models.User, error) {
				return models.User{}, ErrorIncorrectSymbol
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "allowed only a-z, A-Z, 0-9, and /?!@\n",
		},
		{
			nameTest: "Size password smaller, then 6",
			jsonBody: `{"name":"Artem","surname":"Busygin","password":"123","email":"test@mail.ru"}`,
			funcWorkService: func(ctx context.Context, name, surname, password, email string) (models.User, error) {
				return models.User{}, ErrorLenPassword
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "password must contain minimum 6\n",
		},
		{
			nameTest: "Email has 2 @",
			jsonBody: `{"name":"Artem","surname":"Busygin","password":"1234567","email":"test@m@ail.ru"}`,
			funcWorkService: func(ctx context.Context, name, surname, password, email string) (models.User, error) {
				return models.User{}, ErrorCountAtSignEmail
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "must use only one @ in email\n",
		},
		{
			nameTest: "Email has`t @",
			jsonBody: `{"name":"Artem","surname":"Busygin","password":"1234567","email":"testmail.ru"}`,
			funcWorkService: func(ctx context.Context, name, surname, password, email string) (models.User, error) {
				return models.User{}, ErrorCountAtSignEmail
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "must use only one @ in email\n",
		},
		{
			nameTest: "Error during hash password",
			jsonBody: `{"name":"Artem","surname":"Busygin","password":"123456","email":"test@mail.ru"}`,
			funcWorkService: func(ctx context.Context, name, surname, password, email string) (models.User, error) {
				return models.User{}, fmt.Errorf("%w: error bcrypt", service.ErrorCreateHash)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "cannot register user: failed to create hash: error bcrypt\n",
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			registrationService := SpyRegistrationService{
				SpyWorkService: test.funcWorkService,
			}

			handler := CreatedRegisterHandler(&registrationService)

			body := strings.NewReader(test.jsonBody)
			request := httptest.NewRequest(http.MethodPost, "/register", body)
			response := httptest.NewRecorder()

			handler.RegisterUser(response, request)

			assert.Equal(t, test.expectedStatusCode, response.Code, "incorrect status code")
			assert.Equal(t, test.expectedResponse, response.Body.String(), "incorrect error")
		})
	}
}
