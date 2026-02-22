package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository"
	"github.com/stretchr/testify/assert"
)

type SpyRegistrationRepository struct {
	SpyAddUser func(ctx context.Context, user models.User) error
}

func (s *SpyRegistrationRepository) AddUser(ctx context.Context, user models.User) error {
	return s.SpyAddUser(ctx, user)
}

func SpyHasher(password string) ([]byte, error) {
	return []byte{}, fmt.Errorf("%w: %q", ErrorCreateHash, "error bcrypt")
}

func TestRegister(t *testing.T) {
	tests := []struct {
		nameTest      string
		name          string
		surname       string
		password      string
		email         string
		addUser       func(ctx context.Context, user models.User) error
		hasher        func(string) ([]byte, error)
		expectedError error
	}{
		{
			nameTest: "Success registration",
			name:     "Artem",
			surname:  "Busygin",
			password: "1233456",
			email:    "test@mail.ru",
			addUser: func(ctx context.Context, user models.User) error {
				return nil
			},
			hasher:        HashPassword,
			expectedError: nil,
		},
		{
			nameTest: "Email is already existing",
			name:     "Artem",
			surname:  "Busygin",
			password: "1233456",
			email:    "test@mail.ru",
			addUser: func(ctx context.Context, user models.User) error {
				return repository.ErrorExistingEmail
			},
			hasher:        HashPassword,
			expectedError: fmt.Errorf("AddUser: %w", repository.ErrorExistingEmail),
		},
		{
			nameTest: "Incorrect symbol in name",
			name:     "Артём",
			surname:  "Busygin",
			password: "12334343",
			email:    "test@mail.ru",
			addUser: func(ctx context.Context, user models.User) error {
				return nil
			},
			hasher:        HashPassword,
			expectedError: ErrorIncorrectSymbol,
		},
		{
			nameTest: "Incorrect symbol in surname",
			name:     "Artem",
			surname:  "Бусыгин",
			password: "123343",
			email:    "test@mail.ru",
			addUser: func(ctx context.Context, user models.User) error {
				return nil
			},
			hasher:        HashPassword,
			expectedError: ErrorIncorrectSymbol,
		},
		{
			nameTest: "Incorrect symbol in password",
			name:     "Artem",
			surname:  "Busygin",
			password: "бобёр",
			email:    "test@mail.ru",
			addUser: func(ctx context.Context, user models.User) error {
				return nil
			},
			hasher:        HashPassword,
			expectedError: ErrorIncorrectSymbol,
		},
		{
			nameTest: "Incorrect symbol in email",
			name:     "Artem",
			surname:  "Busygin",
			password: "123455",
			email:    "бобёр@mail.ru",
			addUser: func(ctx context.Context, user models.User) error {
				return nil
			},
			hasher:        HashPassword,
			expectedError: ErrorIncorrectSymbol,
		},
		{
			nameTest: "Size password smaller, then 6",
			name:     "Artem",
			surname:  "Busygin",
			password: "123",
			email:    "test@mail.ru",
			addUser: func(ctx context.Context, user models.User) error {
				return nil
			},
			hasher:        HashPassword,
			expectedError: ErrorLenPassword,
		},
		{
			nameTest: "Email has 2 @",
			name:     "Artem",
			surname:  "Busygin",
			password: "1234567",
			email:    "test@m@ail.ru",
			addUser: func(ctx context.Context, user models.User) error {
				return nil
			},
			hasher:        HashPassword,
			expectedError: ErrorCountAtSignEmail,
		},
		{
			nameTest: "Email has`t @",
			name:     "Artem",
			surname:  "Busygin",
			password: "1234567",
			email:    "testmail.ru",
			addUser: func(ctx context.Context, user models.User) error {
				return nil
			},
			hasher:        HashPassword,
			expectedError: ErrorCountAtSignEmail,
		},
		{
			nameTest: "Error hash password",
			name:     "Artem",
			surname:  "Busygin",
			password: "1234567",
			email:    "test@mail.ru",
			addUser: func(ctx context.Context, user models.User) error {
				return nil
			},
			hasher:        SpyHasher,
			expectedError: fmt.Errorf("HashPassword: %w: %q", ErrorCreateHash, "error bcrypt"),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			rep := SpyRegistrationRepository{
				SpyAddUser: test.addUser,
			}

			serviceRegistration := CreateRegistrationService(&rep, test.hasher)

			ctx := context.Background()

			err := serviceRegistration.Register(ctx, test.name, test.surname, test.password, test.email)

			if test.expectedError == nil {
				assert.NoError(t, err, "expected no error, but got one")
			} else {
				if assert.Error(t, err, "expected an error, but got nil") {
					assert.Equal(t, test.expectedError.Error(), err.Error(), "incorrect error message")
				}
			}
		})
	}
}
