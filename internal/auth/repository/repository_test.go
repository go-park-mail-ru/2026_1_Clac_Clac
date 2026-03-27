package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/repository/dto"
	mockRedisEngine "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/repository/mock_redis_engine"
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
		user          dto.UserInitialize
		mockSetup     func(mock pgxmock.PgxPoolIface, user dto.UserInitialize)
		expectedError error
	}{
		{
			nameTest: "Success registration",
			user: dto.UserInitialize{
				Link:         fixedUUID,
				DisplayName:  "Bobr",
				PasswordHash: "hash123",
				Email:        "bobr@mail.ru",
			},
			mockSetup: func(mock pgxmock.PgxPoolIface, user dto.UserInitialize) {
				// Убрал avatar, так как в репозитории его нет в INSERT
				query := `INSERT INTO "user"\s+\(link, display_name, password_hash, email\)\s+VALUES\s+\(\$1, \$2, \$3, \$4\)`

				mock.ExpectExec(query).
					WithArgs(user.Link, user.DisplayName, user.PasswordHash, user.Email).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			if test.mockSetup != nil {
				test.mockSetup(mockPool, test.user)
			}

			repoUsers := NewRepository(mockPool, nil)
			ctx := context.Background()

			err = repoUsers.AddUser(ctx, test.user)

			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError)
			} else {
				assert.NoError(t, err)
			}

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "unfulfilled expectations")
		})
	}
}

func TestAddUserError(t *testing.T) {
	tests := []struct {
		nameTest      string
		user          dto.UserInitialize
		mockSetup     func(mock pgxmock.PgxPoolIface, user dto.UserInitialize)
		expectedError error
	}{
		{
			nameTest: "Email is already existing",
			user: dto.UserInitialize{
				Email: "bobr@mail.ru",
			},
			mockSetup: func(mock pgxmock.PgxPoolIface, user dto.UserInitialize) {
				query := `INSERT INTO "user"\s+\(link, display_name, password_hash, email\)\s+VALUES\s+\(\$1, \$2, \$3, \$4\)`

				mock.ExpectExec(query).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), user.Email).
					WillReturnError(&pgconn.PgError{Code: common.CodeUniqError})
			},
			expectedError: common.ErrorExistingUser,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			if test.mockSetup != nil {
				test.mockSetup(mockPool, test.user)
			}

			repoUsers := NewRepository(mockPool, nil)
			ctx := context.Background()

			err = repoUsers.AddUser(ctx, test.user)

			assert.ErrorIs(t, err, test.expectedError)

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "unfulfilled expectations")
		})
	}
}

func TestAddSession(t *testing.T) {
	tests := []struct {
		nameTest     string
		session      dto.SessionEntity
		mockBehavior func(m *mockRedisEngine.RedisEngine, session dto.SessionEntity)
	}{
		{
			nameTest: "Success add session",
			session: dto.SessionEntity{
				SessionID: common.FixedSessionID,
				UserLink:  common.FixedUserUuiD,
				LifeTime:  24 * time.Hour,
			},
			mockBehavior: func(m *mockRedisEngine.RedisEngine, session dto.SessionEntity) {
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
			if test.mockBehavior != nil {
				test.mockBehavior(redisMock, test.sessionID, test.expectedUserID)
			}

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
			assert.ErrorIs(t, err, test.expectedError) // Единый стиль ErrorIs

			redisMock.AssertExpectations(t)
		})
	}
}

func TestAddResetToken(t *testing.T) {
	tests := []struct {
		nameTest     string
		token        dto.ResetTokenEntity
		mockBehavior func(m *mockRedisEngine.RedisEngine, token dto.ResetTokenEntity)
	}{
		{
			nameTest: "Success add reset token",
			token: dto.ResetTokenEntity{
				ResetTokenID: "token-123",
				UserLink:     uuid.New(),
				LifeTime:     15 * time.Minute,
			},
			mockBehavior: func(m *mockRedisEngine.RedisEngine, token dto.ResetTokenEntity) {
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
		mockBehavior  func(m *mockRedisEngine.RedisEngine, tokenID string)
	}{
		{
			nameTest:      "Not existing token",
			tokenID:       "unknown-token",
			expectedError: common.ErrorNotExistingResetToken,
			mockBehavior: func(m *mockRedisEngine.RedisEngine, tokenID string) {
				key := fmt.Sprintf("reset_token:%s", tokenID)
				m.On("Get", mock.Anything, key).Return(redis.NewStringResult("", redis.Nil))
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

			_, err := repoAuth.GetUserLinkByResetToken(ctx, test.tokenID)

			assert.ErrorIs(t, err, test.expectedError)

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
		expectedUser dto.UserEntity
	}{
		{
			nameTest: "Success get user",
			email:    "bobr@mail.ru",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				query := `SELECT link, display_name, password_hash, email, avatar\s+FROM "user"\s+WHERE email = \$1`
				rows := pgxmock.NewRows([]string{"link", "display_name", "password_hash", "email", "avatar"}).
					AddRow(common.FixedUserUuiD, "Bobr", "hash", "bobr@mail.ru", "avatar.jpg")

				mock.ExpectQuery(query).
					WithArgs("bobr@mail.ru").
					WillReturnRows(rows)
			},
			expectedUser: dto.UserEntity{
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

			if test.mockSetup != nil {
				test.mockSetup(mockPool)
			}

			repoUsers := NewRepository(mockPool, nil)
			ctx := context.Background()

			user, err := repoUsers.GetUser(ctx, test.email)

			assert.NoError(t, err)
			assert.Equal(t, test.expectedUser, user)

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "unfulfilled expectations")
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
				query := `SELECT link, display_name, password_hash, email, avatar\s+FROM "user"\s+WHERE email = \$1`

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

			if test.mockSetup != nil {
				test.mockSetup(mockPool)
			}

			repoUsers := NewRepository(mockPool, nil)
			ctx := context.Background()

			_, err = repoUsers.GetUser(ctx, test.email)

			assert.ErrorIs(t, err, test.expectedError)

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "unfulfilled expectations")
		})
	}
}

func TestGetUserLink(t *testing.T) {
	tests := []struct {
		nameTest         string
		email            string
		mockSetUp        func(mock pgxmock.PgxPoolIface)
		expectedUserLink uuid.UUID
	}{
		{
			nameTest: "Success get user link",
			email:    "bobr@mail.ru",
			mockSetUp: func(mock pgxmock.PgxPoolIface) {
				query := `SELECT link\s+FROM "user"\s+WHERE email = \$1`
				row := pgxmock.NewRows([]string{"link"}).AddRow(common.FixedUserUuiD)

				mock.ExpectQuery(query).WithArgs("bobr@mail.ru").WillReturnRows(row)
			},
			expectedUserLink: common.FixedUserUuiD,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			if test.mockSetUp != nil {
				test.mockSetUp(mockPool)
			}

			rep := NewRepository(mockPool, nil)

			ctx := context.Background()
			userLink, err := rep.GetUserLink(ctx, test.email)

			assert.NoError(t, err)
			assert.Equal(t, test.expectedUserLink, userLink)

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "unfulfilled expectations")
		})
	}
}

func TestGetUserLinkError(t *testing.T) {
	tests := []struct {
		nameTest      string
		email         string
		mockSetUp     func(mock pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			nameTest: "Error user not found",
			email:    "bobr@mail.ru",
			mockSetUp: func(mock pgxmock.PgxPoolIface) {
				query := `SELECT link\s+FROM "user"\s+WHERE email = \$1`

				mock.ExpectQuery(query).WithArgs("bobr@mail.ru").WillReturnError(pgx.ErrNoRows)
			},
			expectedError: common.ErrorNonexistentEmail,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			if test.mockSetUp != nil {
				test.mockSetUp(mockPool)
			}

			rep := NewRepository(mockPool, nil)

			ctx := context.Background()
			_, err = rep.GetUserLink(ctx, test.email)

			assert.ErrorIs(t, err, test.expectedError)

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "unfulfilled expectations")
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
				query := `UPDATE "users"\s+SET password_hash = \$1,\s+updated_at = NOW\(\)\s+WHERE link = \$2`

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

			if test.mockSetup != nil {
				test.mockSetup(mockPool, test.userID, test.newPasswordHash)
			}

			repoAuth := NewRepository(mockPool, nil)
			ctx := context.Background()

			err = repoAuth.UpdatePassword(ctx, test.userID, test.newPasswordHash)

			assert.NoError(t, err)

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "unfulfilled expectations")
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
				query := `UPDATE "users"\s+SET password_hash = \$1,\s+updated_at = NOW\(\)\s+WHERE link = \$2`

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

			if test.mockSetup != nil {
				test.mockSetup(mockPool, test.userID, test.newPasswordHash)
			}

			repoAuth := NewRepository(mockPool, nil)
			ctx := context.Background()

			err = repoAuth.UpdatePassword(ctx, test.userID, test.newPasswordHash)

			assert.ErrorIs(t, err, test.expectedError)

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err, "unfulfilled expectations")
		})
	}
}
