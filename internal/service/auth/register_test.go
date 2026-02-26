package service

import (
	"context"
	"fmt"
	"testing"

	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	repoRegistration "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/auth"
	"github.com/stretchr/testify/assert"
)

type SpyRegistrationRepository struct {
	SpyAddUser func(ctx context.Context, user models.User) error
}

func (s *SpyRegistrationRepository) AddUser(ctx context.Context, user models.User) error {
	return s.SpyAddUser(ctx, user)
}

func SpyHasherError(password string) (string, error) {
	return "", fmt.Errorf("%w: %q", ErrorCreateHash, "error bcrypt")
}

func SpyHasher(password string) (string, error) {
	return "hash_" + password, nil
}

func TestRegister(t *testing.T) {
	tests := []struct {
		nameTest      string
		name          string
		surname       string
		password      string
		email         string
		addUser       func(ctx context.Context, user models.User) error
		hasher        func(string) (string, error)
		expectedUser  models.User
		expectedError error
	}{
		{
			nameTest: "Success registration",
			name:     "Artem",
			surname:  "Busygin",
			password: "1234567",
			email:    "test@mail.ru",
			addUser: func(ctx context.Context, user models.User) error {
				return nil
			},
			hasher: SpyHasher,
			expectedUser: models.User{
				Name:     "Artem",
				Surname:  "Busygin",
				Password: "hash_1234567",
				Email:    "test@mail.ru",
				Boards:   make([]models.Board, 0),
			},
			expectedError: nil,
		},
		{
			nameTest: "Success with empty optional fields (edge case)",
			name:     "",
			surname:  "",
			password: "123",
			email:    "test@mail.ru",
			addUser: func(ctx context.Context, user models.User) error {
				return nil
			},
			hasher: SpyHasher,
			expectedUser: models.User{
				Name:     "",
				Surname:  "",
				Password: "hash_123",
				Email:    "test@mail.ru",
				Boards:   make([]models.Board, 0),
			},
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			rep := SpyRegistrationRepository{
				SpyAddUser: test.addUser,
			}

			serviceRegistration := CreateRegistrationService(&rep, test.hasher)

			ctx := context.Background()

			user, err := serviceRegistration.Register(ctx, test.name, test.surname, test.password, test.email)
			test.expectedUser.ID = user.ID

			assert.Equal(t, test.expectedUser, user, "incorrect parse user")
			assert.NoError(t, err, "expected no error, but got one")
		})
	}
}

func TestRegisterError(t *testing.T) {
	tests := []struct {
		nameTest      string
		name          string
		surname       string
		password      string
		email         string
		addUser       func(ctx context.Context, user models.User) error
		hasher        func(string) (string, error)
		expectedError error
	}{
		{
			nameTest: "Email is already existing",
			name:     "Artem",
			surname:  "Busygin",
			password: "1233456",
			email:    "test@mail.ru",
			addUser: func(ctx context.Context, user models.User) error {
				return repoRegistration.ErrorExistingEmail
			},
			hasher:        SpyHasher,
			expectedError: fmt.Errorf("repo.AddUser: %w", repoRegistration.ErrorExistingEmail),
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
			hasher:        SpyHasherError,
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

			_, err := serviceRegistration.Register(ctx, test.name, test.surname, test.password, test.email)

			assert.EqualError(t, err, test.expectedError.Error(), "incorrect error message")
		})
	}
}
