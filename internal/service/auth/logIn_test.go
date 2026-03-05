package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	repository "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type SpyLogInRepository struct {
	SpyAddUser    func(ctx context.Context, user models.User) error
	SpyAddSession func(ctx context.Context, userID uuid.UUID, sessionID string) error
	SpyGetUser    func(ctx context.Context, email string) (models.User, error)
}

func (s *SpyLogInRepository) AddUser(ctx context.Context, user models.User) error {
	return s.SpyAddUser(ctx, user)
}

func (s *SpyLogInRepository) AddSession(ctx context.Context, userID uuid.UUID, sessionID string) error {
	return s.SpyAddSession(ctx, userID, sessionID)
}

func (s *SpyLogInRepository) GetUser(ctx context.Context, email string) (models.User, error) {
	return s.SpyGetUser(ctx, email)
}

func TestCheckPassword(t *testing.T) {
	newPassword := "newPassword"
	hashNewPassword, err := HashPassword(newPassword)
	assert.NoError(t, err, "expected no error while creating password hash")

	inputPassword := "newPassword"
	err = CheckPassword(inputPassword, hashNewPassword)

	assert.Nil(t, err, "expected passwords must be same")
}

func TestCheckPasswordError(t *testing.T) {
	newPassword := "newPassword"
	hashNewPassword, err := HashPassword(newPassword)
	assert.NoError(t, err, "expected no error while creating password hash")

	inputPassword := "inputPassword"
	err = CheckPassword(inputPassword, hashNewPassword)
	assert.EqualError(t, err, ErrorWrongPassword.Error(), "expected error for wrong password")
}

func TestLogin(t *testing.T) {
	tests := []struct {
		nameTest          string
		email             string
		password          string
		getUser           func(context.Context, string) (models.User, error)
		addSession        func(context.Context, uuid.UUID, string) error
		checker           func(string, string) error
		generator         func() (string, error)
		expectedSessionID string
		expectedUser      models.User
		expectedError     error
	}{
		{
			nameTest: "Success login",
			email:    "bobr@mail.ru",
			password: "12345",
			getUser: func(ctx context.Context, email string) (models.User, error) {
				return models.User{
					DisplayName:  "Artem",
					PasswordHash: "12345",
					Email:        "bobr@mail.ru",
				}, nil
			},
			addSession: func(ctx context.Context, userID uuid.UUID, sessionID string) error {
				return nil
			},
			checker:   SpyChecker,
			generator: SpyGenerator,
			expectedUser: models.User{
				DisplayName:  "Artem",
				PasswordHash: "12345",
				Email:        "bobr@mail.ru",
			},
			expectedSessionID: "sessionCLAC",
			expectedError:     nil,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			rep := SpyLogInRepository{
				SpyGetUser:    test.getUser,
				SpyAddSession: test.addSession,
			}

			serviceLogin := NewLogInService(&rep, test.checker, test.generator)

			ctx := context.Background()

			user, sessionID, err := serviceLogin.Login(ctx, test.email, test.password)

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
		getUser       func(context.Context, string) (models.User, error)
		addSession    func(context.Context, uuid.UUID, string) error
		checker       func(string, string) error
		generator     func() (string, error)
		expectedError error
	}{
		{
			nameTest: "Error user not found",
			email:    "bobr@mail.ru",
			password: "12345",
			getUser: func(ctx context.Context, email string) (models.User, error) {
				return models.User{}, repository.ErrorNonexistentUser
			},
			addSession:    func(ctx context.Context, userID uuid.UUID, sessionID string) error { return nil },
			checker:       SpyChecker,
			generator:     SpyGenerator,
			expectedError: fmt.Errorf("repo.GetUser: %w", repository.ErrorNonexistentUser),
		},
		{
			nameTest: "Error wrong password",
			email:    "test@mail.ru",
			password: "wrong_password",
			getUser: func(ctx context.Context, email string) (models.User, error) {
				return models.User{
					PasswordHash: "1234",
				}, nil
			},
			addSession:    func(ctx context.Context, userID uuid.UUID, sessionID string) error { return nil },
			checker:       SpyChecker,
			generator:     SpyGenerator,
			expectedError: fmt.Errorf("repo.CheckPassword: %w", ErrorWrongPassword),
		},
		{
			nameTest: "Error adding session to DB",
			email:    "test@mail.ru",
			password: "12345",
			getUser: func(ctx context.Context, email string) (models.User, error) {
				return models.User{
					PasswordHash: "12345",
				}, nil
			},
			addSession: func(ctx context.Context, userID uuid.UUID, sessionID string) error {
				return repository.ErrorDetectingCollision
			},
			checker:       SpyChecker,
			generator:     SpyGenerator,
			expectedError: fmt.Errorf("repo.AddSession: %w", repository.ErrorDetectingCollision),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			rep := SpyLogInRepository{
				SpyGetUser:    test.getUser,
				SpyAddSession: test.addSession,
			}

			serviceLogin := NewLogInService(&rep, test.checker, test.generator)

			ctx := context.Background()

			_, _, err := serviceLogin.Login(ctx, test.email, test.password)

			assert.EqualError(t, err, test.expectedError.Error(), "incorrect error message")
		})
	}
}
