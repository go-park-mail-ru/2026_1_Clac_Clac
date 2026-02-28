package service

import (
	"context"
	"fmt"
	"testing"

	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	repoRegistration "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type SpyRegistrationRepository struct {
	SpyAddUser    func(ctx context.Context, user models.User) error
	SpyAddSession func(ctx context.Context, userID uuid.UUID, sessionID string) error
}

func (s *SpyRegistrationRepository) AddUser(ctx context.Context, user models.User) error {
	return s.SpyAddUser(ctx, user)
}

func (s *SpyRegistrationRepository) AddSession(ctx context.Context, userID uuid.UUID, sessionID string) error {
	return s.SpyAddSession(ctx, userID, sessionID)
}

func SpyHasherError(password string) (string, error) {
	return "", fmt.Errorf("%w: %q", ErrorCreateHash, "error bcrypt")
}

func SpyHasher(password string) (string, error) {
	return "hash_" + password, nil
}

func SpyGenerator() (string, error) {
	return "sessionCLAC", nil
}

func TestHashPassword(t *testing.T) {
	password := "my_secret_password"

	hash1, err := HashPassword(password)
	assert.NoError(t, err, "expected no error while hashing")
	assert.NotEmpty(t, hash1, "hash should not be empty")

	hash2, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash2)
	assert.NotEqual(t, hash1, hash2, "bcrypt should generate unique hashes for the same password")
}

func TestGenerateSessionID(t *testing.T) {
	id1, err := GenerateSessionID()
	assert.NoError(t, err, "expected no error while generating session ID")
	assert.Equal(t, 64, len(id1), "hex encoded array should be 64 characters long")

	id2, err := GenerateSessionID()
	assert.NoError(t, err)
	assert.NotEqual(t, id1, id2, "generated sessionID should be unique")
}

func TestRegister(t *testing.T) {
	tests := []struct {
		nameTest          string
		name              string
		password          string
		email             string
		addUser           func(context.Context, models.User) error
		addSession        func(context.Context, uuid.UUID, string) error
		hasher            func(string) (string, error)
		generator         func() (string, error)
		expectedUser      models.User
		expectedSessionID string
		expectedError     error
	}{
		{
			nameTest: "Success registration",
			name:     "Artem",
			password: "1234567",
			email:    "test@mail.ru",
			addUser: func(ctx context.Context, user models.User) error {
				return nil
			},
			addSession: func(ctx context.Context, userID uuid.UUID, sessionID string) error {
				return nil
			},
			hasher:    SpyHasher,
			generator: SpyGenerator,
			expectedUser: models.User{
				Name:     "Artem",
				Password: "hash_1234567",
				Email:    "test@mail.ru",
				Boards:   make([]models.Board, 0),
			},
			expectedSessionID: "sessionCLAC",
			expectedError:     nil,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			rep := SpyRegistrationRepository{
				SpyAddUser:    test.addUser,
				SpyAddSession: test.addSession,
			}

			serviceRegistration := CreateRegistrationService(&rep, test.hasher, test.generator)

			ctx := context.Background()

			user, sectionID, err := serviceRegistration.Register(ctx, test.name, test.password, test.email)
			test.expectedUser.ID = user.ID

			assert.NoError(t, err, "expected no error while generating session ID")
			assert.Equal(t, test.expectedSessionID, sectionID, "incorrect create sessionID")
			assert.Equal(t, test.expectedUser, user, "incorrect parse user")
		})
	}
}

func TestRegisterError(t *testing.T) {
	tests := []struct {
		nameTest          string
		name              string
		password          string
		email             string
		addUser           func(context.Context, models.User) error
		addSession        func(context.Context, uuid.UUID, string) error
		hasher            func(string) (string, error)
		generator         func() (string, error)
		expectedUser      models.User
		expectedSessionID string
		expectedError     error
	}{
		{
			nameTest: "Email is already existing",
			name:     "Artem",
			password: "1233456",
			email:    "test@mail.ru",
			addUser: func(ctx context.Context, user models.User) error {
				return repoRegistration.ErrorExistingUser
			},
			addSession: func(ctx context.Context, userID uuid.UUID, sessionID string) error {
				return nil
			},
			hasher:        SpyHasher,
			generator:     SpyGenerator,
			expectedError: fmt.Errorf("repo.AddUser: %w", repoRegistration.ErrorExistingUser),
		},
		{
			nameTest: "Error hash password",
			name:     "Artem",
			password: "1234567",
			email:    "test@mail.ru",
			addUser: func(ctx context.Context, user models.User) error {
				return nil
			},
			addSession: func(ctx context.Context, userID uuid.UUID, sessionID string) error {
				return nil
			},
			hasher:        SpyHasherError,
			expectedError: fmt.Errorf("HashPassword: %w: %q", ErrorCreateHash, "error bcrypt"),
		},
		{
			nameTest: "Error adding session",
			name:     "Artem",
			password: "1234567",
			email:    "test@mail.ru",
			addUser: func(ctx context.Context, user models.User) error {
				return nil
			},
			addSession: func(ctx context.Context, userID uuid.UUID, sessionID string) error {
				return repoRegistration.ErrorDetectingCollision
			},
			hasher:        SpyHasher,
			generator:     SpyGenerator,
			expectedError: fmt.Errorf("repo.AddSession: %w", repoRegistration.ErrorDetectingCollision),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			rep := SpyRegistrationRepository{
				SpyAddUser:    test.addUser,
				SpyAddSession: test.addSession,
			}

			serviceRegistration := CreateRegistrationService(&rep, test.hasher, test.generator)

			ctx := context.Background()

			_, _, err := serviceRegistration.Register(ctx, test.name, test.password, test.email)

			assert.EqualError(t, err, test.expectedError.Error(), "incorrect error message")
		})
	}
}
