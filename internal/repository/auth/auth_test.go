package repository

import (
	"context"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAddUserError(t *testing.T) {
	tests := []struct {
		nameTest      string
		emails        []string
		expectedError error
	}{
		{
			nameTest:      "Email is already existing",
			emails:        []string{"bobr@mail.ru", "bobr@mail.ru"},
			expectedError: ErrorExistingUser,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			repoUsers := NewMapDB()

			var err error
			ctx := context.Background()

			for _, email := range test.emails {
				err = repoUsers.AddUser(ctx, models.User{Email: email})
			}

			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestAddUser(t *testing.T) {
	tests := []struct {
		nameTest         string
		emails           []string
		expectedDataBase map[string]models.User
	}{
		{
			nameTest: "Success registration",
			emails:   []string{"bobr@mail.ru"},
			expectedDataBase: map[string]models.User{
				"bobr@mail.ru": {
					Email: "bobr@mail.ru",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			repoUsers := NewMapDB()

			ctx := context.Background()

			for _, email := range test.emails {
				repoUsers.AddUser(ctx, models.User{Email: email})
			}

			assert.Equal(t, test.expectedDataBase, repoUsers.database)
		})
	}
}

func TestAddSeessionError(t *testing.T) {
	tests := []struct {
		nameTest      string
		userID        uuid.UUID
		sessionID     string
		expectedError error
	}{
		{
			nameTest:      "Colision session in database",
			userID:        common.FixedUuiD,
			sessionID:     common.FixedSessionID,
			expectedError: ErrorDetectingCollision,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			repoUsers := NewMapDB()

			ctx := context.Background()

			repoUsers.AddSession(ctx, test.userID, test.sessionID)
			err := repoUsers.AddSession(ctx, test.userID, test.sessionID)

			assert.Error(t, err, "expected error")
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestAddSeession(t *testing.T) {
	tests := []struct {
		nameTest       string
		userID         uuid.UUID
		sessionID      string
		expectedUserID uuid.UUID
	}{
		{
			nameTest:       "Success registration",
			userID:         common.FixedUuiD,
			sessionID:      common.FixedSessionID,
			expectedUserID: common.FixedUuiD,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			repoUsers := NewMapDB()

			ctx := context.Background()

			err := repoUsers.AddSession(ctx, test.userID, test.sessionID)
			assert.NoError(t, err, "not wait error")

			userID := repoUsers.sessions[test.sessionID].UserID

			assert.Equal(t, test.expectedUserID, userID)
		})
	}
}

func TestDeleteSession(t *testing.T) {
	tests := []struct {
		nameTest          string
		userID            uuid.UUID
		sessionID         string
		expectedSessionBD map[string]Session
	}{
		{
			nameTest:          "Success delete session",
			userID:            common.FixedUuiD,
			sessionID:         common.FixedSessionID,
			expectedSessionBD: map[string]Session{},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			repoUsers := NewMapDB()

			ctx := context.Background()

			err := repoUsers.AddSession(ctx, test.userID, test.sessionID)
			assert.NoError(t, err, "not wait error")
			err = repoUsers.DeleteSession(ctx, test.sessionID)
			assert.NoError(t, err, "not wait error")

			assert.Equal(t, test.expectedSessionBD, repoUsers.sessions)
		})
	}
}

func TestDeleteSessionError(t *testing.T) {
	tests := []struct {
		nameTest      string
		sessionID     string
		expectedError error
	}{
		{
			nameTest:      "Not existing seesion",
			sessionID:     common.FixedSessionID,
			expectedError: ErrorNotExistingSession,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			repoUsers := NewMapDB()

			ctx := context.Background()
			err := repoUsers.DeleteSession(ctx, test.sessionID)

			assert.Error(t, err, "expected error")
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestGetUserIDBySession(t *testing.T) {
	tests := []struct {
		nameTest  string
		sessionID string

		isExist   bool
		isExpired bool

		expectedUserID uuid.UUID
		expectedError  error
	}{
		{
			nameTest:       "Success get user ID",
			sessionID:      common.FixedSessionID,
			isExist:        true,
			isExpired:      false,
			expectedUserID: common.FixedUuiD,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			repoUsers := NewMapDB()
			ctx := context.Background()

			if test.isExist {
				expirationTime := time.Now().Add(1 * time.Hour)
				if test.isExpired {
					expirationTime = time.Now().Add(-1 * time.Hour)
				}

				repoUsers.sessions[test.sessionID] = Session{
					UserID:    common.FixedUuiD,
					ExpiresAt: expirationTime,
				}
			}

			userID, err := repoUsers.GetUserIDBySession(ctx, test.sessionID)
			assert.NoError(t, err, "not wait error")

			assert.Equal(t, test.expectedUserID, userID)
		})
	}
}

func TestGetUserIDBySessionError(t *testing.T) {
	tests := []struct {
		nameTest  string
		sessionID string

		isExist   bool
		isExpired bool

		expectedError error
	}{
		{
			nameTest:      "Error session not existing",
			sessionID:     common.FixedSessionID,
			isExist:       false,
			isExpired:     false,
			expectedError: ErrorNotExistingSession,
		},
		{
			nameTest:      "Error session expired",
			sessionID:     common.FixedSessionID,
			isExist:       true,
			isExpired:     true,
			expectedError: ErrorSeesionExpired,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			repoUsers := NewMapDB()
			ctx := context.Background()

			if test.isExist {
				expirationTime := time.Now().Add(1 * time.Hour)
				if test.isExpired {
					expirationTime = time.Now().Add(-1 * time.Hour)
				}

				repoUsers.sessions[test.sessionID] = Session{
					UserID:    common.FixedUuiD,
					ExpiresAt: expirationTime,
				}
			}

			_, err := repoUsers.GetUserIDBySession(ctx, test.sessionID)
			assert.Error(t, err, "expected error")
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestGetUserError(t *testing.T) {
	tests := []struct {
		nameTest      string
		email         string
		expectedError error
	}{
		{
			nameTest:      "Not existing user",
			email:         "bobr@mail.ru",
			expectedError: ErrorNonexistentUser,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			repoUsers := NewMapDB()

			ctx := context.Background()

			_, err := repoUsers.GetUser(ctx, test.email)

			assert.Equal(t, test.expectedError, err)
		})
	}
}
func TestGetUser(t *testing.T) {
	tests := []struct {
		nameTest     string
		email        string
		expectedUser models.User
	}{
		{
			nameTest: "Success get user",
			email:    "bobr@mail.ru",
			expectedUser: models.User{
				Email: "bobr@mail.ru",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			repoUsers := NewMapDB()

			ctx := context.Background()

			repoUsers.AddUser(ctx, models.User{Email: test.email})
			user, _ := repoUsers.GetUser(ctx, test.email)

			assert.Equal(t, test.expectedUser, user)
		})
	}
}
