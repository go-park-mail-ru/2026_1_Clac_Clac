package repository

import (
	"context"
	"testing"

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
		nameTest          string
		userID            uuid.UUID
		sessionID         string
		expectedSessionBD map[string]uuid.UUID
	}{
		{
			nameTest:  "Success registration",
			userID:    common.FixedUuiD,
			sessionID: common.FixedSessionID,
			expectedSessionBD: map[string]uuid.UUID{
				common.FixedSessionID: common.FixedUuiD,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			repoUsers := NewMapDB()

			ctx := context.Background()

			repoUsers.AddSession(ctx, test.userID, test.sessionID)

			assert.Equal(t, test.expectedSessionBD, repoUsers.sessions)
		})
	}
}
