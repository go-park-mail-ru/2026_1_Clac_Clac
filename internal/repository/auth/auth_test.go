package auth

import (
	"context"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	dbConnection "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/db_connection"
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
			expectedError: common.ErrorExistingUser,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			conectionDb := dbConnection.NewMapDatabse()
			repoUsers := NewAuthRepository(conectionDb)

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
		IDs              []uuid.UUID
		expectedDataBase map[uuid.UUID]models.User
	}{
		{
			nameTest: "Success registration",
			IDs:      []uuid.UUID{common.FixedUserUuiD},
			expectedDataBase: map[uuid.UUID]models.User{
				common.FixedUserUuiD: {
					ID:     common.FixedUserUuiD,
					Boards: make([]models.Board, 0),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			conectionDb := dbConnection.NewMapDatabse()
			repoUsers := NewAuthRepository(conectionDb)

			ctx := context.Background()

			for _, id := range test.IDs {
				repoUsers.AddUser(ctx, models.User{ID: id, Boards: make([]models.Board, 0)})
			}

			assert.Equal(t, test.expectedDataBase, repoUsers.database.UsersDB)
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
			userID:        common.FixedUserUuiD,
			sessionID:     common.FixedSessionID,
			expectedError: common.ErrorDetectingCollision,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			conectionDb := dbConnection.NewMapDatabse()
			repoUsers := NewAuthRepository(conectionDb)

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
			userID:         common.FixedUserUuiD,
			sessionID:      common.FixedSessionID,
			expectedUserID: common.FixedUserUuiD,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			conectionDb := dbConnection.NewMapDatabse()
			repoUsers := NewAuthRepository(conectionDb)

			ctx := context.Background()

			err := repoUsers.AddSession(ctx, test.userID, test.sessionID)
			assert.NoError(t, err, "not wait error")

			userID := repoUsers.database.SessionsDB[test.sessionID].UserID

			assert.Equal(t, test.expectedUserID, userID)
		})
	}
}

func TestDeleteSession(t *testing.T) {
	tests := []struct {
		nameTest          string
		userID            uuid.UUID
		sessionID         string
		expectedSessionBD map[string]dbConnection.Session
	}{
		{
			nameTest:          "Success delete session",
			userID:            common.FixedUserUuiD,
			sessionID:         common.FixedSessionID,
			expectedSessionBD: map[string]dbConnection.Session{},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			conectionDb := dbConnection.NewMapDatabse()
			repoUsers := NewAuthRepository(conectionDb)

			ctx := context.Background()

			err := repoUsers.AddSession(ctx, test.userID, test.sessionID)
			assert.NoError(t, err, "not wait error")
			err = repoUsers.DeleteSession(ctx, test.sessionID)
			assert.NoError(t, err, "not wait error")

			assert.Equal(t, test.expectedSessionBD, repoUsers.database.SessionsDB)
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
			expectedError: common.ErrorNotExistingSession,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			conectionDb := dbConnection.NewMapDatabse()
			repoUsers := NewAuthRepository(conectionDb)

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
			expectedUserID: common.FixedUserUuiD,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			conectionDb := dbConnection.NewMapDatabse()
			repoUsers := NewAuthRepository(conectionDb)

			ctx := context.Background()

			if test.isExist {
				expirationTime := time.Now().Add(1 * time.Hour)
				if test.isExpired {
					expirationTime = time.Now().Add(-1 * time.Hour)
				}

				repoUsers.database.SessionsDB[test.sessionID] = dbConnection.Session{
					UserID:    common.FixedUserUuiD,
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
			expectedError: common.ErrorNotExistingSession,
		},
		{
			nameTest:      "Error session expired",
			sessionID:     common.FixedSessionID,
			isExist:       true,
			isExpired:     true,
			expectedError: common.ErrorSeesionExpired,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			conectionDb := dbConnection.NewMapDatabse()
			repoUsers := NewAuthRepository(conectionDb)
			ctx := context.Background()

			if test.isExist {
				expirationTime := time.Now().Add(1 * time.Hour)
				if test.isExpired {
					expirationTime = time.Now().Add(-1 * time.Hour)
				}

				repoUsers.database.SessionsDB[test.sessionID] = dbConnection.Session{
					UserID:    common.FixedUserUuiD,
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
			expectedError: common.ErrorNonexistentUser,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			conectionDb := dbConnection.NewMapDatabse()
			repoUsers := NewAuthRepository(conectionDb)

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
			conectionDb := dbConnection.NewMapDatabse()
			repoUsers := NewAuthRepository(conectionDb)

			ctx := context.Background()

			repoUsers.AddUser(ctx, models.User{Email: "bobr@mail.ru"})
			user, _ := repoUsers.GetUser(ctx, test.email)

			assert.Equal(t, test.expectedUser, user)
		})
	}
}
