package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	repository "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/auth"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	tests := []struct {
		nameTest          string
		email             string
		password          string
		hasher            func(string) (string, error)
		checker           func(string, string) error
		generator         func() (string, error)
		mockBehavior      func(m *mocks.Database)
		expectedSessionID string
		expectedUser      models.User
	}{
		{
			nameTest:  "Success login",
			email:     "bobr@mail.ru",
			password:  "12345",
			checker:   spyChecker,
			hasher:    spyHasher,
			generator: spyGenerator,
			mockBehavior: func(m *mocks.Database) {
				ctx := context.Background()
				userFromDB := models.User{
					ID:           uuid.New(),
					DisplayName:  "Artem",
					PasswordHash: "12345",
					Email:        "bobr@mail.ru",
				}
				m.On("GetUser", ctx, "bobr@mail.ru").Return(userFromDB, nil)
				m.On("AddSession", ctx, userFromDB.ID, "sessionCLAC").Return(nil)
			},
			expectedUser: models.User{
				DisplayName:  "Artem",
				PasswordHash: "12345",
				Email:        "bobr@mail.ru",
			},
			expectedSessionID: "sessionCLAC",
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mocks.NewDatabase(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			ctx := context.Background()

			serviceLogin := NewAuthService(mockRepo, test.hasher, test.checker, test.generator)

			user, sessionID, err := serviceLogin.LogIn(ctx, test.email, test.password)

			test.expectedUser.ID = user.ID

			assert.NoError(t, err, "expected no error")
			assert.Equal(t, test.expectedUser, user, "incorrect parsed user")
			assert.Equal(t, test.expectedSessionID, sessionID, "expected same sessionID")
		})
	}
}

func TestLoginError(t *testing.T) {
	tests := []struct {
		nameTest      string
		email         string
		password      string
		checker       func(string, string) error
		hasher        func(string) (string, error)
		generator     func() (string, error)
		mockBehavior  func(m *mocks.Database)
		expectedError error
	}{
		{
			nameTest:  "Error user not found",
			email:     "bobr@mail.ru",
			password:  "12345",
			checker:   spyChecker,
			hasher:    spyHasher,
			generator: spyGenerator,
			mockBehavior: func(m *mocks.Database) {
				m.On("GetUser", context.Background(), "bobr@mail.ru").Return(models.User{}, repository.ErrorNonexistentUser)
			},
			expectedError: fmt.Errorf("repo.GetUser: %w", repository.ErrorNonexistentUser),
		},
		{
			nameTest:  "Error wrong password",
			email:     "test@mail.ru",
			password:  "wrong_password",
			checker:   spyChecker,
			hasher:    spyHasher,
			generator: spyGenerator,
			mockBehavior: func(m *mocks.Database) {
				m.On("GetUser", context.Background(), "test@mail.ru").Return(models.User{
					PasswordHash: "1234",
				}, nil)
			},
			expectedError: fmt.Errorf("repo.CheckPassword: %w", ErrorWrongPassword),
		},
		{
			nameTest:  "Error adding session to DB",
			email:     "test@mail.ru",
			password:  "12345",
			checker:   spyChecker,
			hasher:    spyHasher,
			generator: spyGenerator,
			mockBehavior: func(m *mocks.Database) {
				ctx := context.Background()
				userFromDB := models.User{
					ID:           uuid.New(),
					PasswordHash: "12345",
				}
				m.On("GetUser", ctx, "test@mail.ru").Return(userFromDB, nil)
				m.On("AddSession", ctx, userFromDB.ID, "sessionCLAC").Return(repository.ErrorDetectingCollision)
			},
			expectedError: fmt.Errorf("repo.AddSession: %w", repository.ErrorDetectingCollision),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mocks.NewDatabase(t)
			ctx := context.Background()

			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			serviceLogin := NewAuthService(mockRepo, test.hasher, test.checker, test.generator)

			_, _, err := serviceLogin.LogIn(ctx, test.email, test.password)

			assert.EqualError(t, err, test.expectedError.Error(), "incorrect error message")
		})
	}
}
