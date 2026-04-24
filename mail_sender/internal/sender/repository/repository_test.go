package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/mail_sender/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/mail_sender/internal/sender/repository/dto"
	mockRedisEngine "github.com/go-park-mail-ru/2026_1_Clac_Clac/mail_sender/internal/sender/repository/mock_redis_engine"
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

func TestAddResetToken(t *testing.T) {
	userUUID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	tests := []struct {
		nameTest     string
		token        dto.ResetTokenEntity
		mockBehavior func(m *mockRedisEngine.RedisEngine)
		expectedErr  error
	}{
		{
			nameTest: "Success add reset token",
			token: dto.ResetTokenEntity{
				ResetTokenKey: "reset_token:123456",
				UserLink:      userUUID,
				LifeTime:      15 * time.Minute,
			},
			mockBehavior: func(m *mockRedisEngine.RedisEngine) {
				m.On("Set", mock.Anything, "reset_token:123456", userUUID.String(), 15*time.Minute).
					Return(redis.NewStatusResult("OK", nil))
			},
		},
		{
			nameTest: "Redis Set error",
			token: dto.ResetTokenEntity{
				ResetTokenKey: "reset_token:123456",
				UserLink:      userUUID,
				LifeTime:      15 * time.Minute,
			},
			mockBehavior: func(m *mockRedisEngine.RedisEngine) {
				m.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(redis.NewStatusResult("", errors.New("connection refused")))
			},
			expectedErr: fmt.Errorf("client.Set: %w", errors.New("connection refused")),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			redisMock := mockRedisEngine.NewRedisEngine(t)
			if test.mockBehavior != nil {
				test.mockBehavior(redisMock)
			}

			rep := NewRepository(redisMock)
			err := rep.AddResetToken(context.Background(), test.token)

			if test.expectedErr != nil {
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetUserLinkByResetToken(t *testing.T) {
	expectedLink := "00000000-0000-0000-0000-000000000001"

	tests := []struct {
		nameTest     string
		tokenKey     string
		mockBehavior func(m *mockRedisEngine.RedisEngine)
		expectedLink string
		expectedErr  error
	}{
		{
			nameTest: "Success get user link",
			tokenKey: "reset_token:123456",
			mockBehavior: func(m *mockRedisEngine.RedisEngine) {
				m.On("Get", mock.Anything, "reset_token:123456").
					Return(redis.NewStringResult(expectedLink, nil))
			},
			expectedLink: expectedLink,
		},
		{
			nameTest: "Token not found",
			tokenKey: "reset_token:123456",
			mockBehavior: func(m *mockRedisEngine.RedisEngine) {
				m.On("Get", mock.Anything, "reset_token:123456").
					Return(redis.NewStringResult("", redis.Nil))
			},
			expectedErr: common.ErrorNotExistingResetToken,
		},
		{
			nameTest: "Redis error",
			tokenKey: "reset_token:123456",
			mockBehavior: func(m *mockRedisEngine.RedisEngine) {
				m.On("Get", mock.Anything, "reset_token:123456").
					Return(redis.NewStringResult("", errors.New("connection refused")))
			},
			expectedErr: fmt.Errorf("client.Get: %w", errors.New("connection refused")),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			redisMock := mockRedisEngine.NewRedisEngine(t)
			if test.mockBehavior != nil {
				test.mockBehavior(redisMock)
			}

			rep := NewRepository(redisMock)
			link, err := rep.GetUserLinkByResetToken(context.Background(), test.tokenKey)

			if test.expectedErr != nil {
				if errors.Is(test.expectedErr, common.ErrorNotExistingResetToken) {
					assert.ErrorIs(t, err, common.ErrorNotExistingResetToken)
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

func TestDeleteResetToken(t *testing.T) {
	tests := []struct {
		nameTest     string
		tokenKey     string
		mockBehavior func(m *mockRedisEngine.RedisEngine)
		expectedErr  error
	}{
		{
			nameTest: "Success delete reset token",
			tokenKey: "reset_token:123456",
			mockBehavior: func(m *mockRedisEngine.RedisEngine) {
				m.On("Del", mock.Anything, "reset_token:123456").
					Return(redis.NewIntResult(1, nil))
			},
		},
		{
			nameTest: "Redis Del error",
			tokenKey: "reset_token:123456",
			mockBehavior: func(m *mockRedisEngine.RedisEngine) {
				m.On("Del", mock.Anything, "reset_token:123456").
					Return(redis.NewIntResult(0, errors.New("connection refused")))
			},
			expectedErr: fmt.Errorf("client.Del: %w", errors.New("connection refused")),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			redisMock := mockRedisEngine.NewRedisEngine(t)
			if test.mockBehavior != nil {
				test.mockBehavior(redisMock)
			}

			rep := NewRepository(redisMock)
			err := rep.DeleteResetToken(context.Background(), test.tokenKey)

			if test.expectedErr != nil {
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
