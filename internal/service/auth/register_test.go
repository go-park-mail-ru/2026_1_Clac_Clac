package service

import (
	"context"
	"fmt"
	"testing"

	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	repoRegistration "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/auth"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister(t *testing.T) {
	tests := []struct {
		nameTest          string
		display_name      string
		password          string
		email             string
		hasher            func(string) (string, error)
		checker           func(string, string) error
		generator         func() (string, error)
		mockBehavior      func(m *mocks.Database)
		expectedUser      models.User
		expectedSessionID string
	}{
		{
			nameTest:     "Success registration",
			display_name: "Artem",
			password:     "1234567",
			email:        "test@mail.ru",
			hasher:       spyHasher,
			generator:    spyGenerator,
			checker:      spyChecker,
			mockBehavior: func(m *mocks.Database) {
				ctx := context.Background()
				m.On("AddUser", ctx, mock.AnythingOfType("models.User")).Return(nil)
				m.On("AddSession", ctx, mock.AnythingOfType("uuid.UUID"), "sessionCLAC").Return(nil)
			},
			expectedUser: models.User{
				DisplayName:  "Artem",
				PasswordHash: "hash_1234567",
				Email:        "test@mail.ru",
			},
			expectedSessionID: "sessionCLAC",
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mocks.NewDatabase(t)
			ctx := context.Background()

			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			serviceRegistration := NewAuthService(mockRepo, test.hasher, test.checker, test.generator)

			user, sectionID, err := serviceRegistration.Register(ctx, test.display_name, test.password, test.email)
			test.expectedUser.ID = user.ID

			assert.NoError(t, err, "expected no error")
			assert.Equal(t, test.expectedSessionID, sectionID, "incorrect create sessionID")
			assert.Equal(t, test.expectedUser, user, "incorrect parse user")
		})
	}
}

func TestRegisterError(t *testing.T) {
	tests := []struct {
		nameTest     string
		display_name string
		password     string
		email        string
		hasher       func(string) (string, error)
		generator    func() (string, error)
		checker      func(string, string) error
		mockBehavior func(m *mocks.Database)

		expectedError error
	}{
		{
			nameTest:     "Email is already existing",
			display_name: "Artem",
			password:     "1234567",
			email:        "test@mail.ru",
			hasher:       spyHasher,
			generator:    spyGenerator,
			checker:      spyChecker,
			mockBehavior: func(m *mocks.Database) {
				m.On("AddUser", context.Background(), mock.AnythingOfType("models.User")).Return(repoRegistration.ErrorExistingUser)
			},
			expectedError: fmt.Errorf("repo.AddUser: %w", repoRegistration.ErrorExistingUser),
		},
		{
			nameTest:      "Error hash password",
			display_name:  "Artem",
			password:      "1234567",
			email:         "test@mail.ru",
			hasher:        spyHasherError,
			generator:     spyGenerator,
			checker:       spyChecker,
			mockBehavior:  func(m *mocks.Database) {},
			expectedError: fmt.Errorf("HashPassword: %w: %q", ErrorCreateHash, "error bcrypt"),
		},
		{
			nameTest:     "Error adding session",
			display_name: "Artem",
			password:     "1234567",
			email:        "test@mail.ru",
			hasher:       spyHasher,
			generator:    spyGenerator,
			checker:      spyChecker,
			mockBehavior: func(m *mocks.Database) {
				ctx := context.Background()
				m.On("AddUser", ctx, mock.AnythingOfType("models.User")).Return(nil)
				m.On("AddSession", ctx, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("string")).Return(repoRegistration.ErrorDetectingCollision)
			},
			expectedError: fmt.Errorf("repo.AddSession: %w", repoRegistration.ErrorDetectingCollision),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mocks.NewDatabase(t)

			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			ctx := context.Background()
			serviceRegistration := NewAuthService(mockRepo, test.hasher, test.checker, test.generator)

			_, _, err := serviceRegistration.Register(ctx, test.display_name, test.password, test.email)

			assert.EqualError(t, err, test.expectedError.Error(), "incorrect error message")
		})
	}
}
