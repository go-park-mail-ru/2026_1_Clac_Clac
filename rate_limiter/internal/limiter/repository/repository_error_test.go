package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/rate_limiter/internal/limiter/repository/dto"
	mockRedisEngine "github.com/go-park-mail-ru/2026_1_Clac_Clac/rate_limiter/internal/limiter/repository/mock_redis_engine"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCheckLimitPipeExecError(t *testing.T) {
	t.Run("pipeline exec error closes server", func(t *testing.T) {
		mr, err := miniredis.Run()
		require.NoError(t, err)

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		rep := NewRepository(client)

		mr.Close()
		_ = client.Close()

		config := dto.RateLimiterConfig{
			UserIP: "192.168.1.1",
			Action: "login",
			Window: 1 * time.Minute,
		}

		count, err := rep.CheckLimit(context.Background(), config)
		assert.Error(t, err)
		assert.Zero(t, count)
	})
}

func TestSetCooldownErrors(t *testing.T) {
	t.Run("SetNX error", func(t *testing.T) {
		redisMock := mockRedisEngine.NewRedisEngine(t)
		redisMock.On("SetNX", mock.Anything, "cd:test", "", 1*time.Minute).
			Return(redis.NewBoolResult(false, errors.New("connection refused")))

		rep := NewRepository(redisMock)
		allowed, wait, err := rep.SetCooldown(context.Background(), dto.CooldownConfig{
			Key:        "cd:test",
			Expiration: 1 * time.Minute,
		})

		assert.EqualError(t, err, "redisClient.SetNX: connection refused")
		assert.False(t, allowed)
		assert.Zero(t, wait)
	})

	t.Run("TTL error", func(t *testing.T) {
		redisMock := mockRedisEngine.NewRedisEngine(t)
		redisMock.On("SetNX", mock.Anything, "cd:test", "", 1*time.Minute).
			Return(redis.NewBoolResult(false, nil))
		redisMock.On("TTL", mock.Anything, "cd:test").
			Return(redis.NewDurationResult(0, errors.New("redis error")))

		rep := NewRepository(redisMock)
		allowed, wait, err := rep.SetCooldown(context.Background(), dto.CooldownConfig{
			Key:        "cd:test",
			Expiration: 1 * time.Minute,
		})

		assert.EqualError(t, err, "redisClient.TTL: redis error")
		assert.False(t, allowed)
		assert.Zero(t, wait)
	})

	t.Run("TTL negative clamped to zero", func(t *testing.T) {
		redisMock := mockRedisEngine.NewRedisEngine(t)
		redisMock.On("SetNX", mock.Anything, "cd:test", "", 1*time.Minute).
			Return(redis.NewBoolResult(false, nil))
		redisMock.On("TTL", mock.Anything, "cd:test").
			Return(redis.NewDurationResult(-1, nil))

		rep := NewRepository(redisMock)
		allowed, wait, err := rep.SetCooldown(context.Background(), dto.CooldownConfig{
			Key:        "cd:test",
			Expiration: 1 * time.Minute,
		})

		assert.NoError(t, err)
		assert.False(t, allowed)
		assert.Zero(t, wait)
	})
}
