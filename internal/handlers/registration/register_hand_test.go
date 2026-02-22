package registration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	repository "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository"
	service "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service"
	"github.com/stretchr/testify/assert"
)

type SpyRegistrationService struct {
	SpyWorkService func(ctx context.Context, name, surname, password, email string) error
}

func (s *SpyRegistrationService) Register(ctx context.Context, name, surname, password, email string) error {
	return s.SpyWorkService(ctx, name, surname, password, email)
}

func TestRegisterUser(t *testing.T) {
	tests := []struct {
		nameTest           string
		jsonBody           string
		funcWorkService    func(ctx context.Context, name, surname, password, email string) error
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			nameTest: "Success registration",
			jsonBody: `{"name":"Artem","surname":"Busygin","password":"123456","email":"test@mail.ru"}`,
			funcWorkService: func(ctx context.Context, name, surname, password, email string) error {
				return nil
			},
			expectedStatusCode: http.StatusCreated,
			expectedResponse:   `{"message":"user was successsfully created"}` + "\n",
		},
		{
			nameTest: "Incorrect JSON",
			jsonBody: `{"name":"Artem",,,}`,
			funcWorkService: func(ctx context.Context, name, surname, password, email string) error {
				return nil
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "invalid character ',' looking for beginning of object key string\n",
		},
		{
			nameTest: "Email is already existing",
			jsonBody: `{"name":"Artem","surname":"Busygin","password":"123456","email":"test@mail.ru"}`,
			funcWorkService: func(ctx context.Context, name, surname, password, email string) error {
				return fmt.Errorf("AddUser: %w", repository.ErrorExistingEmail)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "AddUser: user with this email alreday exists\n",
		},
		{
			nameTest: "Incorrect symbol",
			jsonBody: `{"name":"Артём","surname":"Busygin","password":"123","email":"test@mail.ru"}`,
			funcWorkService: func(ctx context.Context, name, surname, password, email string) error {
				return service.ErrorIncorrectSymbol
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "invalid character: allowed only a-z, A-Z, 0-9, and /?!@\n",
		},
		{
			nameTest: "Incorrect count @",
			jsonBody: `{"name":"Artem","surname":"Busygin","password":"123456","email":"test@@mail.ru"}`,
			funcWorkService: func(ctx context.Context, name, surname, password, email string) error {
				return service.ErrorCountAtSignEmail
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "must use only one @ in email\n",
		},
		{
			nameTest: "Error during hash password",
			jsonBody: `{"name":"Artem","surname":"Busygin","password":"123456","email":"test@mail.ru"}`,
			funcWorkService: func(ctx context.Context, name, surname, password, email string) error {
				return fmt.Errorf("AddUser: %w: %s", service.ErrorCreateHash, "error bcrypt")
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "AddUser: failed to create hash: error bcrypt\n",
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
