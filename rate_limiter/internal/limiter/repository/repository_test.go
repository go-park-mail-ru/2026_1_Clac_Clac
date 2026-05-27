package repository

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/rate_limiter/internal/limiter/repository/dto"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
		mr.Close()
	})

	return mr, client
}

func TestNewRepository(t *testing.T) {
	t.Run("creates repository", func(t *testing.T) {
		_, client := newTestClient(t)
		rep := NewRepository(client)
		assert.NotNil(t, rep)
	})
}

func TestCheckLimit(t *testing.T) {
	t.Run("first request returns 1", func(t *testing.T) {
		_, client := newTestClient(t)
		rep := NewRepository(client)

		config := dto.RateLimiterConfig{
			UserIP: "192.168.1.1",
			Action: "login",
			Window: 1 * time.Minute,
		}

		count, err := rep.CheckLimit(context.Background(), config)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("second request returns 2", func(t *testing.T) {
		_, client := newTestClient(t)
		rep := NewRepository(client)

		config := dto.RateLimiterConfig{
			UserIP: "192.168.1.1",
			Action: "login",
			Window: 1 * time.Minute,
		}

		_, _ = rep.CheckLimit(context.Background(), config)
		count, err := rep.CheckLimit(context.Background(), config)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("different actions are counted separately", func(t *testing.T) {
		_, client := newTestClient(t)
		rep := NewRepository(client)

		configLogin := dto.RateLimiterConfig{
			UserIP: "192.168.1.1",
			Action: "login",
			Window: 1 * time.Minute,
		}
		configRegister := dto.RateLimiterConfig{
			UserIP: "192.168.1.1",
			Action: "register",
			Window: 1 * time.Minute,
		}

		count1, err := rep.CheckLimit(context.Background(), configLogin)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count1)

		count2, err := rep.CheckLimit(context.Background(), configRegister)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count2)
	})
}

func TestSetCooldown(t *testing.T) {
	t.Run("first call allows", func(t *testing.T) {
		_, client := newTestClient(t)
		rep := NewRepository(client)

		config := dto.CooldownConfig{
			Key:        "cd:login:test@mail.ru",
			Expiration: 1 * time.Minute,
		}

		allowed, wait, err := rep.SetCooldown(context.Background(), config)
		assert.NoError(t, err)
		assert.True(t, allowed)
		assert.Zero(t, wait)
	})

	t.Run("second call returns not allowed with TTL", func(t *testing.T) {
		mr, client := newTestClient(t)
		rep := NewRepository(client)

		config := dto.CooldownConfig{
			Key:        "cd:login:test@mail.ru",
			Expiration: 1 * time.Minute,
		}

		_, _, _ = rep.SetCooldown(context.Background(), config)

		mr.FastForward(10 * time.Second)

		allowed, wait, err := rep.SetCooldown(context.Background(), config)
		assert.NoError(t, err)
		assert.False(t, allowed)
		assert.Greater(t, wait, time.Duration(0))
	})

	t.Run("after expiration allows again", func(t *testing.T) {
		mr, client := newTestClient(t)
		rep := NewRepository(client)

		config := dto.CooldownConfig{
			Key:        "cd:login:test@mail.ru",
			Expiration: 1 * time.Second,
		}

		_, _, _ = rep.SetCooldown(context.Background(), config)

		mr.FastForward(2 * time.Second)

		allowed, wait, err := rep.SetCooldown(context.Background(), config)
		assert.NoError(t, err)
		assert.True(t, allowed)
		assert.Zero(t, wait)
	})
}
