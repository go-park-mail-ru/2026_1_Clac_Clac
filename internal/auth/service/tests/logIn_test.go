package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/models"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"

	service "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service"
	mockAuthRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service/tests/mock_auth_rep"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogin(t *testing.T) {
	tests := []struct {
		id                uuid.UUID
		nameTest          string
		email             string
		password          string
		hasher            func(string) (string, error)
		checker           func(string, string) error
		generator         func() (string, error)
		mockBehavior      func(m *mockAuthRep.AuthRepository)
		expectedSessionID string
		expectedUser      models.User
	}{
		{
			id:        common.FixedUserUuiD,
			nameTest:  "Success login",
			email:     "bobr@mail.ru",
			password:  "12345",
			checker:   spyChecker,
			hasher:    spyHasher,
			generator: spyGenerator,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				userFromDB := models.User{
					ID:           common.FixedUserUuiD,
					DisplayName:  "Artem",
					PasswordHash: "12345",
					Email:        "bobr@mail.ru",
				}
				m.On("GetUser", ctx, "bobr@mail.ru").Return(userFromDB, nil)
				m.On("AddSession", ctx, mock.AnythingOfType("db.Session")).Return(nil)
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
			mockRepo := mockAuthRep.NewAuthRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			ctx := context.Background()

			serviceLogin := service.NewService(mockRepo, nil, test.hasher, test.checker, test.generator, nil)

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
		id            uuid.UUID
		email         string
		password      string
		checker       func(string, string) error
		hasher        func(string) (string, error)
		generator     func() (string, error)
		mockBehavior  func(m *mockAuthRep.AuthRepository)
		expectedError error
	}{
		{
			nameTest:  "Error user not found",
			id:        common.FixedUserUuiD,
			email:     "bobr@mail.ru",
			password:  "12345",
			checker:   spyChecker,
			hasher:    spyHasher,
			generator: spyGenerator,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("GetUser", context.Background(), "bobr@mail.ru").Return(models.User{}, common.ErrorNonexistentUser)
			},
			expectedError: fmt.Errorf("rep.GetUser: %w", common.ErrorNonexistentUser),
		},
		{
			nameTest:  "Error wrong password",
			id:        common.FixedUserUuiD,
			email:     "test@mail.ru",
			password:  "wrong_password",
			checker:   spyChecker,
			hasher:    spyHasher,
			generator: spyGenerator,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("GetUser", context.Background(), "test@mail.ru").Return(models.User{
					PasswordHash: "1234",
				}, nil)
			},
			expectedError: fmt.Errorf("rep.CheckPassword: %w", service.ErrorWrongPassword),
		},
		{
			nameTest:  "Error adding session to DB",
			id:        common.FixedUserUuiD,
			email:     "test@mail.ru",
			password:  "12345",
			checker:   spyChecker,
			hasher:    spyHasher,
			generator: spyGenerator,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				userFromDB := models.User{
					ID:           uuid.New(),
					PasswordHash: "12345",
				}
				m.On("GetUser", ctx, "test@mail.ru").Return(userFromDB, nil)
				m.On("AddSession", ctx, mock.AnythingOfType("db.Session")).Return(common.ErrorDetectingSessionCollision)
			},
			expectedError: fmt.Errorf("rep.AddSession: %w", common.ErrorDetectingSessionCollision),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			ctx := context.Background()

			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			serviceLogin := service.NewService(mockRepo, nil, test.hasher, test.checker, test.generator, nil)

			_, _, err := serviceLogin.LogIn(ctx, test.email, test.password)

			assert.EqualError(t, err, test.expectedError.Error(), "incorrect error message")
		})
	}
}
