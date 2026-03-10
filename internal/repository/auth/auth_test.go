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

func TestAddSeession(t *testing.T) {
	tests := []struct {
		nameTest       string
		session        dbConnection.Session
		expectedUserID uuid.UUID
	}{
		{
			nameTest: "Success registration",
			session: dbConnection.Session{
				SessionID: common.FixedSessionID,
				UserID:    common.FixedUserUuiD,
			},
			expectedUserID: common.FixedUserUuiD,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			conectionDb := dbConnection.NewMapDatabse()
			repoUsers := NewAuthRepository(conectionDb)

			ctx := context.Background()

			err := repoUsers.AddSession(ctx, test.session)
			assert.NoError(t, err, "not wait error")

			userID := repoUsers.database.SessionsDB[test.session.SessionID].UserID

			assert.Equal(t, test.expectedUserID, userID)
		})
	}
}

func TestAddSeessionError(t *testing.T) {
	tests := []struct {
		nameTest      string
		session       dbConnection.Session
		expectedError error
	}{
		{
			nameTest: "Colision session in database",
			session: dbConnection.Session{
				SessionID: common.FixedSessionID,
				UserID:    common.FixedUserUuiD,
			},
			expectedError: common.ErrorDetectingSessionCollision,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			conectionDb := dbConnection.NewMapDatabse()
			repoUsers := NewAuthRepository(conectionDb)

			ctx := context.Background()

			repoUsers.AddSession(ctx, test.session)
			err := repoUsers.AddSession(ctx, test.session)

			assert.Error(t, err, "expected error")
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestDeleteSession(t *testing.T) {
	tests := []struct {
		nameTest          string
		session           dbConnection.Session
		expectedSessionBD map[string]dbConnection.Session
	}{
		{
			nameTest: "Success delete session",
			session: dbConnection.Session{
				SessionID: common.FixedSessionID,
				UserID:    common.FixedUserUuiD,
			},
			expectedSessionBD: map[string]dbConnection.Session{},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			conectionDb := dbConnection.NewMapDatabse()
			repoUsers := NewAuthRepository(conectionDb)

			ctx := context.Background()

			err := repoUsers.AddSession(ctx, test.session)
			assert.NoError(t, err, "not wait error")
			err = repoUsers.DeleteSession(ctx, test.session.SessionID)
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

func TestAddResetToken(t *testing.T) {
	tests := []struct {
		nameTest    string
		token       dbConnection.ResetToken
		expectedErr error
	}{
		{
			nameTest: "Success add reset token",
			token: dbConnection.ResetToken{
				ResetTokenID: "token-123",
				UserID:       uuid.New(),
				ExpiresAt:    time.Now().Add(15 * time.Minute),
			},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			conectionDb := dbConnection.NewMapDatabse()
			repoAuth := NewAuthRepository(conectionDb)

			ctx := context.Background()
			err := repoAuth.AddResetToken(ctx, test.token)

			assert.NoError(t, err)

			_, exist := conectionDb.ResetTokensDB[test.token.ResetTokenID]
			assert.True(t, exist)
		})
	}
}

func TestAddResetTokenError(t *testing.T) {
	tests := []struct {
		nameTest      string
		token         dbConnection.ResetToken
		expectedError error
	}{
		{
			nameTest: "Error token collision",
			token: dbConnection.ResetToken{
				ResetTokenID: common.FixedResetTokenID,
			},
			expectedError: common.ErrorDetectingTokenCollision,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			conectionDb := dbConnection.NewMapDatabse()
			repoAuth := NewAuthRepository(conectionDb)

			conectionDb.ResetTokensDB[common.FixedResetTokenID] = dbConnection.ResetToken{
				ResetTokenID: common.FixedResetTokenID,
			}

			ctx := context.Background()
			err := repoAuth.AddResetToken(ctx, test.token)

			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestGetResetToken(t *testing.T) {
	targetUserID := uuid.New()

	tests := []struct {
		nameTest           string
		tokenID            string
		expectedResetToken dbConnection.ResetToken
	}{
		{
			nameTest: "Success get reset token",
			tokenID:  common.FixedResetTokenID,
			expectedResetToken: dbConnection.ResetToken{
				ResetTokenID: common.FixedResetTokenID,
				UserID:       targetUserID,
				ExpiresAt:    time.Now().Add(15 * time.Minute),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			conectionDb := dbConnection.NewMapDatabse()
			repoAuth := NewAuthRepository(conectionDb)
			ctx := context.Background()

			conectionDb.ResetTokensDB[common.FixedResetTokenID] = dbConnection.ResetToken{
				ResetTokenID: common.FixedResetTokenID,
				UserID:       targetUserID,
				ExpiresAt:    time.Now().Add(15 * time.Minute),
			}

			token, err := repoAuth.GetResetToken(ctx, test.tokenID)

			assert.NoError(t, err)
			assert.Equal(t, test.expectedResetToken.UserID, token.UserID)
		})
	}
}

func TestGetResetTokenError(t *testing.T) {
	tests := []struct {
		nameTest      string
		tokenID       string
		expectedError error
	}{
		{
			nameTest:      "Not existing token",
			tokenID:       "unknown-token",
			expectedError: common.ErrorNotExistingResetToken,
		},
		{
			nameTest:      "Token expired",
			tokenID:       common.FixedResetTokenID,
			expectedError: common.ErrorResetTokenExpired,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			conectionDb := dbConnection.NewMapDatabse()
			repoAuth := NewAuthRepository(conectionDb)
			ctx := context.Background()

			conectionDb.ResetTokensDB[common.FixedResetTokenID] = dbConnection.ResetToken{
				ResetTokenID: common.FixedResetTokenID,
				UserID:       uuid.New(),
				ExpiresAt:    time.Now().Add(-1 * time.Hour),
			}

			_, err := repoAuth.GetResetToken(ctx, test.tokenID)

			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestDeleteResetToken(t *testing.T) {
	tests := []struct {
		nameTest string
		tokenID  string
	}{
		{
			nameTest: "Success delete token",
			tokenID:  common.FixedResetTokenID,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			conectionDb := dbConnection.NewMapDatabse()
			repoAuth := NewAuthRepository(conectionDb)

			conectionDb.ResetTokensDB[common.FixedResetTokenID] = dbConnection.ResetToken{
				ResetTokenID: common.FixedResetTokenID,
			}

			ctx := context.Background()
			err := repoAuth.DeleteResetToken(ctx, test.tokenID)

			assert.NoError(t, err)
			_, exist := conectionDb.ResetTokensDB[test.tokenID]
			assert.False(t, exist)
		})
	}
}

func TestDeleteResetTokenError(t *testing.T) {
	tests := []struct {
		nameTest      string
		tokenID       string
		expectedError error
	}{
		{
			nameTest:      "Error delete not existing token",
			tokenID:       "unknown-token",
			expectedError: common.ErrorNotExistingResetToken,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			conectionDb := dbConnection.NewMapDatabse()
			repoAuth := NewAuthRepository(conectionDb)

			ctx := context.Background()
			err := repoAuth.DeleteResetToken(ctx, test.tokenID)

			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestUpdatePassword(t *testing.T) {
	targetUserID := uuid.New()
	newHash := "1234"

	tests := []struct {
		nameTest        string
		userID          uuid.UUID
		newPasswordHash string
	}{
		{
			nameTest:        "Success update password",
			userID:          targetUserID,
			newPasswordHash: newHash,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			conectionDb := dbConnection.NewMapDatabse()
			repoAuth := NewAuthRepository(conectionDb)

			conectionDb.UsersDB[targetUserID] = models.User{
				ID:           targetUserID,
				PasswordHash: "123",
			}

			ctx := context.Background()
			err := repoAuth.UpdatePassword(ctx, test.userID, test.newPasswordHash)

			assert.NoError(t, err)

			updatedUser := conectionDb.UsersDB[test.userID]
			assert.Equal(t, test.newPasswordHash, updatedUser.PasswordHash)
		})
	}
}

func TestUpdatePasswordError(t *testing.T) {
	tests := []struct {
		nameTest        string
		userID          uuid.UUID
		newPasswordHash string
		expectedError   error
	}{
		{
			nameTest:        "Error user not found",
			userID:          uuid.New(),
			newPasswordHash: "new-hash",
			expectedError:   common.ErrorNonexistentUser,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			conectionDb := dbConnection.NewMapDatabse()
			repoAuth := NewAuthRepository(conectionDb)

			ctx := context.Background()
			err := repoAuth.UpdatePassword(ctx, test.userID, test.newPasswordHash)

			assert.Equal(t, test.expectedError, err)
		})
	}
}
