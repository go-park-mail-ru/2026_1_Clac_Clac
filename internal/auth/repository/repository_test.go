package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/models"
	mockRedisEngine "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/repository/mock_redis_engine"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAddUser(t *testing.T) {
	fixedUUID := common.FixedUserUuiD

	tests := []struct {
		nameTest      string
		user          models.User
		mockSetup     func(mock pgxmock.PgxPoolIface, user models.User)
		expectedError bool
	}{
		{
			nameTest: "Success registration",
			user: models.User{
				Link: fixedUUID,
			},
			mockSetup: func(mock pgxmock.PgxPoolIface, user models.User) {
				query := `INSERT INTO "user" \(link, display_name, password_hash, email, avatar\)`

				mock.ExpectExec(query).
					WithArgs(user.Link, user.DisplayName, user.PasswordHash, user.Email, user.Avatar).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			test.mockSetup(mockPool, test.user)

			repoUsers := NewRepository(mockPool, nil)
			ctx := context.Background()

			err = repoUsers.AddUser(ctx, test.user)

			if test.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "not wait error")
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
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repoUsers := NewRepository(mockPool, nil)
			var errResult error
			ctx := context.Background()

			query := `INSERT INTO "user" \(link, display_name, password_hash, email, avatar\)`

			mockPool.ExpectExec(query).
				WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), test.emails[0], pgxmock.AnyArg()).
				WillReturnResult(pgxmock.NewResult("INSERT", 1))

			mockPool.ExpectExec(query).
				WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), test.emails[1], pgxmock.AnyArg()).
				WillReturnError(&pgconn.PgError{Code: common.CodeUniqError})

			for _, email := range test.emails {
				errResult = repoUsers.AddUser(ctx, models.User{Email: email})
			}

			assert.Equal(t, test.expectedError, errResult)

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "not wait error")
		})
	}
}

func TestAddSession(t *testing.T) {
	tests := []struct {
		nameTest     string
		session      dto.Session
		mockBehavior func(m *mockRedisEngine.RedisEngine, session dto.Session)
	}{
		{
			nameTest: "Success add session",
			session: dto.Session{
				SessionID: common.FixedSessionID,
				UserLink:  common.FixedUserUuiD,
				LifeTime:  24 * time.Hour,
			},
			mockBehavior: func(m *mockRedisEngine.RedisEngine, session dto.Session) {
				key := fmt.Sprintf("session:%s", session.SessionID)
				m.On("Set", mock.Anything, key, session.UserLink.String(), session.LifeTime).Return(redis.NewStatusResult("OK", nil))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			redisMock := mockRedisEngine.NewRedisEngine(t)
			if test.mockBehavior != nil {
				test.mockBehavior(redisMock, test.session)
			}

			repoUsers := NewRepository(nil, redisMock)
			ctx := context.Background()

			err := repoUsers.AddSession(ctx, test.session)
			assert.NoError(t, err)

			redisMock.AssertExpectations(t)
		})
	}
}

func TestDeleteSession(t *testing.T) {
	tests := []struct {
		nameTest     string
		sessionID    string
		mockBehavior func(m *mockRedisEngine.RedisEngine, sessionID string)
	}{
		{
			nameTest:  "Success delete session",
			sessionID: common.FixedSessionID,
			mockBehavior: func(m *mockRedisEngine.RedisEngine, sessionID string) {
				key := fmt.Sprintf("session:%s", sessionID)
				m.On("Del", mock.Anything, key).Return(redis.NewIntResult(1, nil))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			redisMock := mockRedisEngine.NewRedisEngine(t)
			if test.mockBehavior != nil {
				test.mockBehavior(redisMock, test.sessionID)
			}

			repoUsers := NewRepository(nil, redisMock)
			ctx := context.Background()

			err := repoUsers.DeleteSession(ctx, test.sessionID)
			assert.NoError(t, err)

			redisMock.AssertExpectations(t)
		})
	}
}

func TestGetUserIDBySession(t *testing.T) {
	tests := []struct {
		nameTest       string
		sessionID      string
		expectedUserID string
		mockBehavior   func(m *mockRedisEngine.RedisEngine, sessionID string, expectedUserID string)
	}{
		{
			nameTest:       "Success get user ID",
			sessionID:      common.FixedSessionID,
			expectedUserID: common.FixedUserUuiD.String(),
			mockBehavior: func(m *mockRedisEngine.RedisEngine, sessionID string, expectedUserID string) {
				key := fmt.Sprintf("session:%s", sessionID)
				m.On("Get", mock.Anything, key).Return(redis.NewStringResult(expectedUserID, nil))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			redisMock := mockRedisEngine.NewRedisEngine(t)
			test.mockBehavior(redisMock, test.sessionID, test.expectedUserID)

			repoUsers := NewRepository(nil, redisMock)
			ctx := context.Background()

			userID, err := repoUsers.GetUserIDBySession(ctx, test.sessionID)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedUserID, userID)

			redisMock.AssertExpectations(t)
		})
	}
}

func TestGetUserIDBySessionError(t *testing.T) {
	tests := []struct {
		nameTest      string
		sessionID     string
		expectedError error
		mockBehavior  func(m *mockRedisEngine.RedisEngine, sessionID string)
	}{
		{
			nameTest:      "Error session not existing",
			sessionID:     common.FixedSessionID,
			expectedError: common.ErrorNotExistingSession,
			mockBehavior: func(m *mockRedisEngine.RedisEngine, sessionID string) {
				key := fmt.Sprintf("session:%s", sessionID)
				m.On("Get", mock.Anything, key).Return(redis.NewStringResult("", redis.Nil))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			redisMock := mockRedisEngine.NewRedisEngine(t)
			if test.mockBehavior != nil {
				test.mockBehavior(redisMock, test.sessionID)
			}

			repoUsers := NewRepository(nil, redisMock)
			ctx := context.Background()

			_, err := repoUsers.GetUserIDBySession(ctx, test.sessionID)
			assert.Equal(t, test.expectedError, err)

			redisMock.AssertExpectations(t)
		})
	}
}

func TestAddResetToken(t *testing.T) {
	tests := []struct {
		nameTest     string
		token        dto.ResetToken
		mockBehavior func(m *mockRedisEngine.RedisEngine, token dto.ResetToken)
	}{
		{
			nameTest: "Success add reset token",
			token: dto.ResetToken{
				ResetTokenID: "token-123",
				UserLink:     uuid.New(),
				LifeTime:     15 * time.Minute,
			},
			mockBehavior: func(m *mockRedisEngine.RedisEngine, token dto.ResetToken) {
				key := fmt.Sprintf("reset_token:%s", token.ResetTokenID)
				m.On("Set", mock.Anything, key, token.UserLink.String(), token.LifeTime).Return(redis.NewStatusResult("OK", nil))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			redisMock := mockRedisEngine.NewRedisEngine(t)
			if test.mockBehavior != nil {
				test.mockBehavior(redisMock, test.token)
			}

			repoAuth := NewRepository(nil, redisMock)
			ctx := context.Background()

			err := repoAuth.AddResetToken(ctx, test.token)
			assert.NoError(t, err)

			redisMock.AssertExpectations(t)
		})
	}
}

func TestGetUserLinkByResetToken(t *testing.T) {
	targetUserID := uuid.New()

	tests := []struct {
		nameTest       string
		tokenID        string
		expectedUserID string
		mockBehavior   func(m *mockRedisEngine.RedisEngine, tokenID string, userID string)
	}{
		{
			nameTest:       "Success get reset token",
			tokenID:        common.FixedResetTokenID,
			expectedUserID: targetUserID.String(),
			mockBehavior: func(m *mockRedisEngine.RedisEngine, tokenID string, userID string) {
				key := fmt.Sprintf("reset_token:%s", tokenID)
				m.On("Get", mock.Anything, key).Return(redis.NewStringResult(userID, nil))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			redisMock := mockRedisEngine.NewRedisEngine(t)
			if test.mockBehavior != nil {
				test.mockBehavior(redisMock, test.tokenID, test.expectedUserID)
			}

			repoAuth := NewRepository(nil, redisMock)
			ctx := context.Background()

			token, err := repoAuth.GetUserLinkByResetToken(ctx, test.tokenID)

			assert.NoError(t, err)
			assert.Equal(t, test.expectedUserID, token)

			redisMock.AssertExpectations(t)
		})
	}
}

func TestGetUserLinkByResetTokenError(t *testing.T) {
	tests := []struct {
		nameTest      string
		tokenID       string
		expectedError error
		mockBehaviar  func(m *mockRedisEngine.RedisEngine, tokenID string)
	}{
		{
			nameTest:      "Not existing token",
			tokenID:       "unknown-token",
			expectedError: common.ErrorNotExistingResetToken,
			mockBehaviar: func(m *mockRedisEngine.RedisEngine, tokenID string) {
				key := fmt.Sprintf("reset_token:%s", tokenID)
				m.On("Get", mock.Anything, key).Return(redis.NewStringResult("", redis.Nil))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			redisMock := mockRedisEngine.NewRedisEngine(t)
			if test.mockBehaviar != nil {
				test.mockBehaviar(redisMock, test.tokenID)
			}

			repoAuth := NewRepository(nil, redisMock)
			ctx := context.Background()

			_, err := repoAuth.GetUserLinkByResetToken(ctx, test.tokenID)

			assert.Equal(t, test.expectedError, err)

			redisMock.AssertExpectations(t)
		})
	}
}

func TestDeleteResetToken(t *testing.T) {
	tests := []struct {
		nameTest     string
		tokenID      string
		mockBehavior func(m *mockRedisEngine.RedisEngine, tokenID string)
	}{
		{
			nameTest: "Success delete token",
			tokenID:  common.FixedResetTokenID,
			mockBehavior: func(m *mockRedisEngine.RedisEngine, tokenID string) {
				key := fmt.Sprintf("reset_token:%s", tokenID)
				m.On("Del", mock.Anything, key).Return(redis.NewIntResult(1, nil))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			redisMock := mockRedisEngine.NewRedisEngine(t)
			if test.mockBehavior != nil {
				test.mockBehavior(redisMock, test.tokenID)
			}

			repoAuth := NewRepository(nil, redisMock)
			ctx := context.Background()

			err := repoAuth.DeleteResetToken(ctx, test.tokenID)

			assert.NoError(t, err)

			redisMock.AssertExpectations(t)
		})
	}
}

func TestGetUser(t *testing.T) {
	tests := []struct {
		nameTest     string
		email        string
		mockSetup    func(mock pgxmock.PgxPoolIface)
		expectedUser models.User
	}{
		{
			nameTest: "Success get user",
			email:    "bobr@mail.ru",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				query := `SELECT link, display_name, password_hash, email, avatar FROM "user" WHERE email = \$1`
				rows := pgxmock.NewRows([]string{"link", "display_name", "password_hash", "email", "avatar"}).
					AddRow(common.FixedUserUuiD, "Bobr", "hash", "bobr@mail.ru", "avatar.jpg")

				mock.ExpectQuery(query).
					WithArgs("bobr@mail.ru").
					WillReturnRows(rows)
			},
			expectedUser: models.User{
				Link:         common.FixedUserUuiD,
				DisplayName:  "Bobr",
				PasswordHash: "hash",
				Email:        "bobr@mail.ru",
				Avatar:       "avatar.jpg",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			test.mockSetup(mockPool)

			repoUsers := NewRepository(mockPool, nil)
			ctx := context.Background()

			user, err := repoUsers.GetUser(ctx, test.email)

			assert.NoError(t, err)
			assert.Equal(t, test.expectedUser, user)

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "not wait error")
		})
	}
}

func TestGetUserError(t *testing.T) {
	tests := []struct {
		nameTest      string
		email         string
		mockSetup     func(mock pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			nameTest: "Not existing user",
			email:    "bobr@mail.ru",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				query := `SELECT link, display_name, password_hash, email, avatar FROM "user" WHERE email = \$1`

				mock.ExpectQuery(query).
					WithArgs("bobr@mail.ru").
					WillReturnError(pgx.ErrNoRows)
			},
			expectedError: common.ErrorNonexistentEmail,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			test.mockSetup(mockPool)

			repoUsers := NewRepository(mockPool, nil)
			ctx := context.Background()

			_, err = repoUsers.GetUser(ctx, test.email)

			assert.Equal(t, test.expectedError, err)

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "wait error")
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
		mockSetup       func(mock pgxmock.PgxPoolIface, userID uuid.UUID, hash string)
	}{
		{
			nameTest:        "Success update password",
			userID:          targetUserID,
			newPasswordHash: newHash,
			mockSetup: func(mock pgxmock.PgxPoolIface, userID uuid.UUID, hash string) {
				query := `UPDATE "users" SET password_hash = \$1, updated_at = NOW\(\) WHERE link = \$2`

				mock.ExpectExec(query).
					WithArgs(hash, userID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			test.mockSetup(mockPool, test.userID, test.newPasswordHash)

			repoAuth := NewRepository(mockPool, nil)
			ctx := context.Background()

			err = repoAuth.UpdatePassword(ctx, test.userID, test.newPasswordHash)

			assert.NoError(t, err)

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "not wait error")
		})
	}
}

func TestUpdatePasswordError(t *testing.T) {
	tests := []struct {
		nameTest        string
		userID          uuid.UUID
		newPasswordHash string
		expectedError   error
		mockSetup       func(mock pgxmock.PgxPoolIface, userID uuid.UUID, hash string)
	}{
		{
			nameTest:        "Error user not found",
			userID:          uuid.New(),
			newPasswordHash: "newhash",
			expectedError:   common.ErrorNonexistentUser,
			mockSetup: func(mock pgxmock.PgxPoolIface, userID uuid.UUID, hash string) {
				query := `UPDATE "users" SET password_hash = \$1, updated_at = NOW\(\) WHERE link = \$2`

				mock.ExpectExec(query).
					WithArgs(hash, userID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			test.mockSetup(mockPool, test.userID, test.newPasswordHash)

			repoAuth := NewRepository(mockPool, nil)
			ctx := context.Background()

			err = repoAuth.UpdatePassword(ctx, test.userID, test.newPasswordHash)

			assert.Equal(t, test.expectedError, err)

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "wait error")
		})
	}
}
