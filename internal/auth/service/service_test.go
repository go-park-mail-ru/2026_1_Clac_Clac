package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/models"
	mockAuthRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service/mock_auth_rep"
	mockSender "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service/mock_sender"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	db "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/db"
	dbConnection "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/db"
	"github.com/google/uuid"
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
				m.On("AddSession", ctx, mock.AnythingOfType("db.Session")).Return(nil)
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

			serviceRegistration := NewService(mockRepo, nil, test.hasher, test.checker, test.generator, nil)

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
				m.On("AddSession", ctx, mock.AnythingOfType("db.Session")).Return(common.ErrorDetectingSessionCollision)
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
			serviceRegistration := NewService(mockRepo, nil, test.hasher, test.checker, test.generator, nil)

			_, _, err := serviceRegistration.Register(ctx, test.display_name, test.password, test.email)

			assert.EqualError(t, err, test.expectedError.Error(), "incorrect error message")
		})
	}
}

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

			serviceLogin := NewService(mockRepo, nil, test.hasher, test.checker, test.generator, nil)

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
			expectedError: fmt.Errorf("rep.CheckPassword: %w", ErrorWrongPassword),
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

			serviceLogin := NewService(mockRepo, nil, test.hasher, test.checker, test.generator, nil)

			_, _, err := serviceLogin.LogIn(ctx, test.email, test.password)

			assert.EqualError(t, err, test.expectedError.Error(), "incorrect error message")
		})
	}
}

func TestLogOut(t *testing.T) {
	tests := []struct {
		nameTest     string
		sessionID    string
		hasher       func(string) (string, error)
		checker      func(string, string) error
		generator    func() (string, error)
		mockBehavior func(m *mockAuthRep.AuthRepository)
	}{
		{
			nameTest:  "Success log out",
			sessionID: common.FixedSessionID,
			checker:   spyChecker,
			hasher:    spyHasher,
			generator: spyGenerator,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				m.On("DeleteSession", ctx, common.FixedSessionID).Return(nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			ctx := context.Background()

			serviceLogOut := NewService(mockRepo, nil, test.hasher, test.checker, test.generator, nil)

			err := serviceLogOut.LogOut(ctx, test.sessionID)
			assert.NoError(t, err, "not expected error")
		})
	}
}

func TestLogOutError(t *testing.T) {
	tests := []struct {
		nameTest      string
		sessionID     string
		hasher        func(string) (string, error)
		checker       func(string, string) error
		generator     func() (string, error)
		mockBehavior  func(m *mockAuthRep.AuthRepository)
		expectedError error
	}{
		{
			nameTest:  "Error session not found",
			sessionID: common.FixedSessionID,
			checker:   spyChecker,
			hasher:    spyHasher,
			generator: spyGenerator,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				m.On("DeleteSession", ctx, common.FixedSessionID).Return(common.ErrorNotExistingSession)
			},
			expectedError: fmt.Errorf("rep.DeleteSession: %w", common.ErrorNotExistingSession),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			ctx := context.Background()

			serviceLogOut := NewService(mockRepo, nil, test.hasher, test.checker, test.generator, nil)

			err := serviceLogOut.LogOut(ctx, test.sessionID)
			assert.Error(t, err, "expected error")
			assert.EqualError(t, test.expectedError, err.Error(), "incorrect error message")
		})
	}
}

func TestSendRecoveryCode(t *testing.T) {
	targetEmail := "test@mail.ru"

	tests := []struct {
		nameTest      string
		email         string
		generator     func() (string, error)
		mockBehavior  func(m *mockAuthRep.AuthRepository)
		senderMock    func(m *mockSender.SenderLetters)
		expectedError error
	}{
		{
			nameTest:  "Success delivery code",
			email:     targetEmail,
			generator: func() (string, error) { return "123456", nil },
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("GetUser", mock.Anything, targetEmail).Return(models.User{ID: common.FixedUserUuiD}, nil)
				m.On("AddResetToken", mock.Anything, mock.AnythingOfType("db.ResetToken")).Return(nil)
			},
			senderMock: func(m *mockSender.SenderLetters) {
				m.On("SendLetter", targetEmail, "Code to create a new password", mock.AnythingOfType("string")).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest:  "Error user not found",
			email:     "testing@mail.ru",
			generator: func() (string, error) { return "123456", nil },
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("GetUser", mock.Anything, "testing@mail.ru").Return(models.User{}, common.ErrorNonexistentUser)
			},
			senderMock:    func(m *mockSender.SenderLetters) {},
			expectedError: fmt.Errorf("rep.GetUser: %w", common.ErrorNonexistentUser),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			mockMail := mockSender.NewSenderLetters(t)

			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}
			if test.senderMock != nil {
				test.senderMock(mockMail)
			}

			service := NewService(mockRepo, mockMail, nil, nil, test.generator, test.generator)

			err := service.SendRecoveryCode(context.Background(), test.email)

			time.Sleep(10 * time.Millisecond)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckCode(t *testing.T) {
	validToken := "123456"

	tests := []struct {
		nameTest      string
		tokenID       string
		mockBehavior  func(m *mockAuthRep.AuthRepository)
		expectedError error
	}{
		{
			nameTest: "Success check code",
			tokenID:  validToken,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				token := dbConnection.ResetToken{
					ResetTokenID: validToken,
					ExpiresAt:    time.Now().Add(15 * time.Minute),
				}
				m.On("GetResetToken", mock.Anything, validToken).Return(token, nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error code expired",
			tokenID:  validToken,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				token := dbConnection.ResetToken{
					ResetTokenID: validToken,
					ExpiresAt:    time.Now().Add(-1 * time.Minute),
				}
				m.On("GetResetToken", mock.Anything, validToken).Return(token, nil)
				m.On("DeleteResetToken", mock.Anything, validToken).Return(nil)
			},
			expectedError: common.ErrorResetTokenExpired,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			service := NewService(mockRepo, nil, nil, nil, nil, nil)
			err := service.CheckRecoveryCode(context.Background(), test.tokenID)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResetPassword(t *testing.T) {
	tests := []struct {
		nameTest     string
		tokenID      string
		newPassword  string
		hasher       func(string) (string, error)
		mockBehavior func(m *mockAuthRep.AuthRepository)
	}{
		{
			nameTest:    "Success reset password",
			tokenID:     common.FixedResetTokenID,
			newPassword: "new_password",
			hasher:      spyHasher,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				validToken := db.ResetToken{
					ResetTokenID: common.FixedResetTokenID,
					UserID:       common.FixedUserUuiD,
					ExpiresAt:    time.Now().Add(15 * time.Minute),
				}

				m.On("GetResetToken", ctx, common.FixedResetTokenID).Return(validToken, nil)
				m.On("UpdatePassword", ctx, common.FixedUserUuiD, "hash_new_password").Return(nil)
				m.On("DeleteResetToken", ctx, common.FixedResetTokenID).Return(nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			ctx := context.Background()

			serviceAuth := NewService(mockRepo, nil, test.hasher, nil, nil, nil)

			err := serviceAuth.ResetPassword(ctx, test.tokenID, test.newPassword)

			assert.NoError(t, err, "expected no error")
		})
	}
}

func TestResetPasswordError(t *testing.T) {
	targetUserID := uuid.New()

	tests := []struct {
		nameTest      string
		tokenID       string
		newPassword   string
		hasher        func(string) (string, error)
		mockBehavior  func(m *mockAuthRep.AuthRepository)
		expectedError error
	}{
		{
			nameTest:    "Error token expired",
			tokenID:     common.FixedResetTokenID,
			newPassword: "new_password",
			hasher:      spyHasher,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				m.On("GetResetToken", ctx, common.FixedResetTokenID).Return(db.ResetToken{}, common.ErrorResetTokenExpired)
			},
			expectedError: fmt.Errorf("rep.GetResetToken: %w", common.ErrorResetTokenExpired),
		},
		{
			nameTest:    "Error hasher fails",
			tokenID:     common.FixedResetTokenID,
			newPassword: "new_password",
			hasher:      spyHasherError,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				validToken := db.ResetToken{
					ResetTokenID: common.FixedResetTokenID,
					UserID:       targetUserID,
					ExpiresAt:    time.Now().Add(15 * time.Minute),
				}
				m.On("GetResetToken", ctx, common.FixedResetTokenID).Return(validToken, nil)
			},
			expectedError: errors.New("hasher: failed to create hash: \"error bcrypt\""),
		},
		{
			nameTest:    "Error update password in DB",
			tokenID:     common.FixedResetTokenID,
			newPassword: "new_password",
			hasher:      spyHasher,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				validToken := db.ResetToken{
					ResetTokenID: common.FixedResetTokenID,
					UserID:       targetUserID,
					ExpiresAt:    time.Now().Add(15 * time.Minute),
				}

				m.On("GetResetToken", ctx, common.FixedResetTokenID).Return(validToken, nil)
				m.On("UpdatePassword", ctx, targetUserID, "hash_new_password").
					Return(errors.New("db connection lost"))
			},
			expectedError: fmt.Errorf("rep.UpdatePassword: %w", errors.New("db connection lost")),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			ctx := context.Background()

			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			serviceAuth := NewService(mockRepo, nil, test.hasher, nil, nil, nil)

			err := serviceAuth.ResetPassword(ctx, test.tokenID, test.newPassword)

			assert.EqualError(t, err, test.expectedError.Error(), "incorrect error message")
		})
	}
}
