package auth

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	mockAuthRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth/mock_auth_rep"
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
		mockBehavior      func(m *mockAuthRep.AuthRepository)
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
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				m.On("AddUser", ctx, mock.AnythingOfType("models.User")).Return(nil)
				m.On("AddSession", ctx, mock.AnythingOfType("dbConnection.Session")).Return(nil)
			},
			expectedUser: models.User{
				DisplayName:  "Artem",
				PasswordHash: "hash_1234567",
				Email:        "test@mail.ru",
				Boards:       make([]models.Board, 0),
			},
			expectedSessionID: "sessionCLAC",
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			ctx := context.Background()

			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			serviceRegistration := NewAuthService(mockRepo, nil, test.hasher, test.checker, test.generator, nil)

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
		mockBehavior func(m *mockAuthRep.AuthRepository)

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
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("AddUser", context.Background(), mock.AnythingOfType("models.User")).Return(common.ErrorExistingUser)
			},
			expectedError: fmt.Errorf("rep.AddUser: %w", common.ErrorExistingUser),
		},
		{
			nameTest:      "Error hash password",
			display_name:  "Artem",
			password:      "1234567",
			email:         "test@mail.ru",
			hasher:        spyHasherError,
			generator:     spyGenerator,
			checker:       spyChecker,
			mockBehavior:  nil,
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
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				m.On("AddUser", ctx, mock.AnythingOfType("models.User")).Return(nil)
				m.On("AddSession", ctx, mock.AnythingOfType("dbConnection.Session")).Return(common.ErrorDetectingSessionCollision)
			},
			expectedError: fmt.Errorf("rep.AddSession: %w", common.ErrorDetectingSessionCollision),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)

			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			ctx := context.Background()
			serviceRegistration := NewAuthService(mockRepo, nil, test.hasher, test.checker, test.generator, nil)

			_, _, err := serviceRegistration.Register(ctx, test.display_name, test.password, test.email)

			assert.EqualError(t, err, test.expectedError.Error(), "incorrect error message")
		})
	}
}
