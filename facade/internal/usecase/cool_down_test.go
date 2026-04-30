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
	cooldown := domain.Cooldown{Name: "recovery_email", Email: "u@mail.ru", ExpirationS: 60}

	tests := []struct {
		name         string
		mockBehavior func(m *mockRateClient.RateLimiterClient)
		expectedRes  domain.CooldownResult
		expectError  bool
	}{
		{
			name: "Allowed",
			mockBehavior: func(m *mockRateClient.RateLimiterClient) {
				expected := domain.CooldownResult{Allowed: true, WaitS: 0}
				m.On("SetCooldown", context.Background(), cooldown).Return(expected, nil)
			},
			expectedRes: domain.CooldownResult{Allowed: true, WaitS: 0},
			expectError: false,
		},
		{
			name: "NotAllowed",
			mockBehavior: func(m *mockRateClient.RateLimiterClient) {
				expected := domain.CooldownResult{Allowed: false, WaitS: 45}
				m.On("SetCooldown", context.Background(), cooldown).Return(expected, nil)
			},
			expectedRes: domain.CooldownResult{Allowed: false, WaitS: 45},
			expectError: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockRateClient.RateLimiterClient) {
				m.On("SetCooldown", context.Background(), cooldown).Return(domain.CooldownResult{}, errors.New("redis unavailable"))
			},
			expectedRes: domain.CooldownResult{},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockRateClient.NewRateLimiterClient(t)
			tc.mockBehavior(m)

			result, err := NewCoolDown(m).CheckCoolDown(context.Background(), cooldown)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedRes, result)
			}
		})
	}
}
