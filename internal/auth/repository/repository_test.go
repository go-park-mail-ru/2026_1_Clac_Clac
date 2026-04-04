package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/repository/dto"
	mockRedisEngine "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/repository/mock_redis_engine"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/redis/go-redis/v9"
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
				SessionKey: common.FixedSessionID,
				UserLink:   common.FixedUserUuiD,
				LifeTime:   24 * time.Hour,
			},
			mockBehavior: func(m *mockRedisEngine.RedisEngine, session dto.SessionEntity) {
				m.On("Set", mock.Anything, session.SessionKey, session.UserLink.String(), session.LifeTime).Return(redis.NewStatusResult("OK", nil))
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

func TestExtendSession(t *testing.T) {
	tests := []struct {
		nameTest     string
		sessionID    string
		timeExpires  time.Duration
		mockBehavior func(m *mockRedisEngine.RedisEngine, sessionID string, timeExpires time.Duration)
	}{
		{
			nameTest:    "Success extend session",
			sessionID:   common.FixedSessionID,
			timeExpires: time.Hour * 24,
			mockBehavior: func(m *mockRedisEngine.RedisEngine, sessionID string, timeExpires time.Duration) {
				key := fmt.Sprintf("session:%s", sessionID)

				successfulCmd := redis.NewBoolResult(true, nil)

				m.On("Expire", mock.Anything, key, timeExpires).Return(successfulCmd)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			redisMock := mockRedisEngine.NewRedisEngine(t)
			if test.mockBehavior != nil {
				test.mockBehavior(redisMock, test.sessionID, test.timeExpires)
			}

			repoUsers := NewRepository(nil, redisMock)
			err := repoUsers.ExtendSession(context.Background(), dto.ExtendedSession{
				Key:        fmt.Sprintf("session:%s", test.sessionID),
				Expiration: test.timeExpires,
			})

			assert.NoError(t, err, "not wait error")

			redisMock.AssertExpectations(t)
		})
	}
}

func TestExtendSessionError(t *testing.T) {
	tests := []struct {
		nameTest     string
		sessionID    string
		timeExpires  time.Duration
		mockBehavior func(m *mockRedisEngine.RedisEngine, sessionID string, timeExpires time.Duration)
	}{
		{
			nameTest:    "Error extend session",
			sessionID:   common.FixedSessionID,
			timeExpires: time.Hour * 24,
			mockBehavior: func(m *mockRedisEngine.RedisEngine, sessionID string, timeExpires time.Duration) {
				key := fmt.Sprintf("session:%s", sessionID)

				redisErr := errors.New("redis connection timeout")
				errorCmd := redis.NewBoolResult(false, redisErr)

				m.On("Expire", mock.Anything, key, timeExpires).Return(errorCmd)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			redisMock := mockRedisEngine.NewRedisEngine(t)
			if test.mockBehavior != nil {
				test.mockBehavior(redisMock, test.sessionID, test.timeExpires)
			}

			rep := NewRepository(nil, redisMock)
			err := rep.ExtendSession(context.Background(), dto.ExtendedSession{
				Key:        fmt.Sprintf("session:%s", test.sessionID),
				Expiration: test.timeExpires,
			})
			assert.Error(t, err, "wait error")

			redisMock.AssertExpectations(t)
		})
	}
}

func TestSetCooldown(t *testing.T) {
	defaultConfig := dto.CoolDownConfig{
		Key:        "cd:recovery_email:test@mail.ru",
		Expiration: 1 * time.Minute,
	}

	errRedis := errors.New("redis internal error")

	tests := []struct {
		nameTest        string
		config          dto.CoolDownConfig
		mockBehavior    func(m *mockRedisEngine.RedisEngine)
		expectedAllowed bool
		expectedTTL     time.Duration
		expectedError   error
	}{
		{
			nameTest: "Success set cooldown",
			config:   defaultConfig,
			mockBehavior: func(m *mockRedisEngine.RedisEngine) {
				ctx := context.Background()

				m.On("SetNX", ctx, defaultConfig.Key, "", defaultConfig.Expiration).Return(redis.NewBoolResult(true, nil))
			},
			expectedAllowed: true,
			expectedTTL:     0,
			expectedError:   nil,
		},
		{
			nameTest: "Cooldown already exists",
			config:   defaultConfig,
			mockBehavior: func(m *mockRedisEngine.RedisEngine) {
				ctx := context.Background()

				m.On("SetNX", ctx, defaultConfig.Key, "", defaultConfig.Expiration).Return(redis.NewBoolResult(false, nil))
				m.On("TTL", ctx, defaultConfig.Key).Return(redis.NewDurationResult(30*time.Second, nil))
			},
			expectedAllowed: false,
			expectedTTL:     30 * time.Second,
			expectedError:   nil,
		},
		{
			nameTest: "Cooldown exists but negative TTL",
			config:   defaultConfig,
			mockBehavior: func(m *mockRedisEngine.RedisEngine) {
				ctx := context.Background()

				m.On("SetNX", ctx, defaultConfig.Key, "", defaultConfig.Expiration).Return(redis.NewBoolResult(false, nil))
				m.On("TTL", ctx, defaultConfig.Key).Return(redis.NewDurationResult(-2*time.Second, nil))
			},
			expectedAllowed: false,
			expectedTTL:     0,
			expectedError:   nil,
		},
		{
			nameTest: "Error from SetNX",
			config:   defaultConfig,
			mockBehavior: func(m *mockRedisEngine.RedisEngine) {
				ctx := context.Background()

				m.On("SetNX", ctx, defaultConfig.Key, "", defaultConfig.Expiration).Return(redis.NewBoolResult(false, errRedis))
			},
			expectedAllowed: false,
			expectedTTL:     0,
			expectedError:   errRedis,
		},
		{
			nameTest: "Error from TTL",
			config:   defaultConfig,
			mockBehavior: func(m *mockRedisEngine.RedisEngine) {
				ctx := context.Background()

				m.On("SetNX", ctx, defaultConfig.Key, "", defaultConfig.Expiration).Return(redis.NewBoolResult(false, nil))
				m.On("TTL", ctx, defaultConfig.Key).Return(redis.NewDurationResult(0, errRedis))
			},
			expectedAllowed: false,
			expectedTTL:     0,
			expectedError:   errRedis,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRedis := mockRedisEngine.NewRedisEngine(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRedis)
			}

			rep := NewRepository(nil, mockRedis)

			isAllowed, ttl, err := rep.SetCooldown(context.Background(), test.config)

			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError, "incorrect error type")
			} else {
				assert.NoError(t, err, "expected no error")
			}

			assert.Equal(t, test.expectedAllowed, isAllowed, "incorrect isAllowed result")
			assert.Equal(t, test.expectedTTL, ttl, "incorrect TTL result")
		})
	}
}

func TestCheckLimit(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	defaultConfig := dto.RateLimiterConfig{
		UserIP: "1.1.1.1",
		Action: "login",
		Window: 1 * time.Minute,
	}

	tests := []struct {
		nameTest      string
		configLimiter dto.RateLimiterConfig
		expectedValue int64
		expectedError error
	}{
		{
			nameTest:      "Success exec request",
			configLimiter: defaultConfig,
			expectedValue: 1,
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			rep := NewRepository(nil, client)

			val, err := rep.CheckLimit(context.Background(), test.configLimiter)

			assert.NoError(t, err, "not expected error")
			assert.Equal(t, test.expectedValue, val, fmt.Sprintf("expected %d, got %d", test.expectedValue, val))

			val2, _ := rep.CheckLimit(context.Background(), test.configLimiter)
			assert.Equal(t, int64(2), val2, "second call should increment counter")
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
				m.On("Del", mock.Anything, sessionID).Return(redis.NewIntResult(1, nil))
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
				m.On("Get", mock.Anything, sessionID).Return(redis.NewStringResult(expectedUserID, nil))
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
				m.On("Get", mock.Anything, sessionID).Return(redis.NewStringResult("", redis.Nil))
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
			assert.ErrorIs(t, err, test.expectedError)

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
				ResetTokenKey: "token-123",
				UserLink:      uuid.New(),
				LifeTime:      15 * time.Minute,
			},
			mockBehavior: func(m *mockRedisEngine.RedisEngine, token dto.ResetTokenEntity) {
				m.On("Set", mock.Anything, token.ResetTokenKey, token.UserLink.String(), token.LifeTime).Return(redis.NewStatusResult("OK", nil))
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
				m.On("Get", mock.Anything, tokenID).Return(redis.NewStringResult(userID, nil))
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
				m.On("Get", mock.Anything, tokenID).Return(redis.NewStringResult("", redis.Nil))
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
				m.On("Del", mock.Anything, tokenID).Return(redis.NewIntResult(1, nil))
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
				query := `UPDATE "user"\s+SET password_hash = \$1,\s+updated_at = NOW\(\)\s+WHERE link = \$2`

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
				query := `UPDATE "user"\s+SET password_hash = \$1,\s+updated_at = NOW\(\)\s+WHERE link = \$2`

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
