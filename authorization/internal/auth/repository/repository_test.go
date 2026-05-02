package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/auth/repository/dto"
	mockRedisEngine "github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/auth/repository/mock_redis_engine"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/common"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewRepository(t *testing.T) {
	t.Run("creates repository", func(t *testing.T) {
		redisMock := mockRedisEngine.NewRedisEngine(t)
		rep := NewRepository(redisMock)
		assert.NotNil(t, rep)
	})
}

func TestAddSession(t *testing.T) {
	tests := []struct {
		nameTest     string
		session      dto.SessionEntity
		mockBehavior func(m *mockRedisEngine.RedisEngine, session dto.SessionEntity)
		expectedErr  error
	}{
		{
			nameTest: "Success add session",
			session: dto.SessionEntity{
				SessionKey: "session:abc123",
				UserLink:   uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				LifeTime:   24 * time.Hour,
			},
			mockBehavior: func(m *mockRedisEngine.RedisEngine, session dto.SessionEntity) {
				m.On("Set", mock.Anything, session.SessionKey, session.UserLink.String(), session.LifeTime).
					Return(redis.NewStatusResult("OK", nil))
			},
		},
		{
			nameTest: "Error redis Set fails",
			session: dto.SessionEntity{
				SessionKey: "session:abc123",
				UserLink:   uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				LifeTime:   24 * time.Hour,
			},
			mockBehavior: func(m *mockRedisEngine.RedisEngine, session dto.SessionEntity) {
				m.On("Set", mock.Anything, session.SessionKey, session.UserLink.String(), session.LifeTime).
					Return(redis.NewStatusResult("", errors.New("connection refused")))
			},
			expectedErr: fmt.Errorf("redisClient.Set: %w", errors.New("connection refused")),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			redisMock := mockRedisEngine.NewRedisEngine(t)
			if test.mockBehavior != nil {
				test.mockBehavior(redisMock, test.session)
			}

			rep := NewRepository(redisMock)
			err := rep.AddSession(context.Background(), test.session)

			if test.expectedErr != nil {
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetUserLink(t *testing.T) {
	expectedLink := "00000000-0000-0000-0000-000000000001"

	tests := []struct {
		nameTest     string
		sessionKey   string
		mockBehavior func(m *mockRedisEngine.RedisEngine)
		expectedLink string
		expectedErr  error
	}{
		{
			nameTest:   "Success get user link",
			sessionKey: "session:abc123",
			mockBehavior: func(m *mockRedisEngine.RedisEngine) {
				m.On("Get", mock.Anything, "session:abc123").
					Return(redis.NewStringResult(expectedLink, nil))
			},
			expectedLink: expectedLink,
		},
		{
			nameTest:   "Session not found",
			sessionKey: "session:abc123",
			mockBehavior: func(m *mockRedisEngine.RedisEngine) {
				m.On("Get", mock.Anything, "session:abc123").
					Return(redis.NewStringResult("", redis.Nil))
			},
			expectedErr: common.ErrorNotExistingSession,
		},
		{
			nameTest:   "Redis error",
			sessionKey: "session:abc123",
			mockBehavior: func(m *mockRedisEngine.RedisEngine) {
				m.On("Get", mock.Anything, "session:abc123").
					Return(redis.NewStringResult("", errors.New("connection refused")))
			},
			expectedErr: fmt.Errorf("redisClient.Get: %w", errors.New("connection refused")),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			redisMock := mockRedisEngine.NewRedisEngine(t)
			if test.mockBehavior != nil {
				test.mockBehavior(redisMock)
			}

			rep := NewRepository(redisMock)
			link, err := rep.GetUserLink(context.Background(), test.sessionKey)

			if test.expectedErr != nil {
				if errors.Is(test.expectedErr, common.ErrorNotExistingSession) {
					assert.ErrorIs(t, err, common.ErrorNotExistingSession)
				} else {
					assert.EqualError(t, err, test.expectedErr.Error())
				}
				assert.Empty(t, link)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedLink, link)
			}
		})
	}
}

func TestRepositoryDeleteSession(t *testing.T) {
	tests := []struct {
		nameTest     string
		sessionKey   string
		mockBehavior func(m *mockRedisEngine.RedisEngine)
		expectedErr  error
	}{
		{
			nameTest:   "Success delete session",
			sessionKey: "session:abc123",
			mockBehavior: func(m *mockRedisEngine.RedisEngine) {
				m.On("Del", mock.Anything, "session:abc123").
					Return(redis.NewIntResult(1, nil))
			},
		},
		{
			nameTest:   "Redis error",
			sessionKey: "session:abc123",
			mockBehavior: func(m *mockRedisEngine.RedisEngine) {
				m.On("Del", mock.Anything, "session:abc123").
					Return(redis.NewIntResult(0, errors.New("connection refused")))
			},
			expectedErr: fmt.Errorf("redisClient.Del: %w", errors.New("connection refused")),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			redisMock := mockRedisEngine.NewRedisEngine(t)
			if test.mockBehavior != nil {
				test.mockBehavior(redisMock)
			}

			rep := NewRepository(redisMock)
			err := rep.DeleteSession(context.Background(), test.sessionKey)

			if test.expectedErr != nil {
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExtendSession(t *testing.T) {
	tests := []struct {
		nameTest     string
		session      dto.ExtendedSession
		mockBehavior func(m *mockRedisEngine.RedisEngine)
		expectedErr  error
	}{
		{
			nameTest: "Success extend session",
			session: dto.ExtendedSession{
				SessionKey: "session:abc123",
				Expiration: 24 * time.Hour,
			},
			mockBehavior: func(m *mockRedisEngine.RedisEngine) {
				m.On("Expire", mock.Anything, "session:abc123", 24*time.Hour).
					Return(redis.NewBoolResult(true, nil))
			},
		},
		{
			nameTest: "Session not found",
			session: dto.ExtendedSession{
				SessionKey: "session:abc123",
				Expiration: 24 * time.Hour,
			},
			mockBehavior: func(m *mockRedisEngine.RedisEngine) {
				m.On("Expire", mock.Anything, "session:abc123", 24*time.Hour).
					Return(redis.NewBoolResult(false, nil))
			},
			expectedErr: common.ErrorNotExistingSession,
		},
		{
			nameTest: "Redis error",
			session: dto.ExtendedSession{
				SessionKey: "session:abc123",
				Expiration: 24 * time.Hour,
			},
			mockBehavior: func(m *mockRedisEngine.RedisEngine) {
				m.On("Expire", mock.Anything, "session:abc123", 24*time.Hour).
					Return(redis.NewBoolResult(false, errors.New("connection refused")))
			},
			expectedErr: fmt.Errorf("redisClient.Expire: %w", errors.New("connection refused")),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			redisMock := mockRedisEngine.NewRedisEngine(t)
			if test.mockBehavior != nil {
				test.mockBehavior(redisMock)
			}

			rep := NewRepository(redisMock)
			err := rep.ExtendSession(context.Background(), test.session)

			if test.expectedErr != nil {
				if errors.Is(test.expectedErr, common.ErrorNotExistingSession) {
					assert.ErrorIs(t, err, common.ErrorNotExistingSession)
				} else {
					assert.EqualError(t, err, test.expectedErr.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
