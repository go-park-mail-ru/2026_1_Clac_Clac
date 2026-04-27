package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	mockRateClient "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/usecase/mock_rate_limiter_client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckCoolDown(t *testing.T) {
	cooldown := domain.Cooldown{Name: "recovery_email", Email: "u@mail.ru", ExpirationMs: 60}

	t.Run("Allowed", func(t *testing.T) {
		m := mockRateClient.NewRateLimiterClient(t)
		expected := domain.CooldownResult{Allowed: true, WaitS: 0}
		m.On("SetCooldown", context.Background(), cooldown).Return(expected, nil)

		result, err := NewCoolDown(m).CheckCoolDown(context.Background(), cooldown)
		require.NoError(t, err)
		assert.True(t, result.Allowed)
	})

	t.Run("NotAllowed", func(t *testing.T) {
		m := mockRateClient.NewRateLimiterClient(t)
		expected := domain.CooldownResult{Allowed: false, WaitS: 45}
		m.On("SetCooldown", context.Background(), cooldown).Return(expected, nil)

		result, err := NewCoolDown(m).CheckCoolDown(context.Background(), cooldown)
		require.NoError(t, err)
		assert.False(t, result.Allowed)
		assert.EqualValues(t, 45, result.WaitS)
	})

	t.Run("ClientError", func(t *testing.T) {
		m := mockRateClient.NewRateLimiterClient(t)
		m.On("SetCooldown", context.Background(), cooldown).Return(domain.CooldownResult{}, errors.New("redis unavailable"))

		_, err := NewCoolDown(m).CheckCoolDown(context.Background(), cooldown)
		require.Error(t, err)
	})
}
