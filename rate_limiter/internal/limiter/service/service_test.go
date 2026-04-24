package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/rate_limiter/internal/limiter/repository/dto"
	mockLimiterRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/rate_limiter/internal/limiter/service/mock_limiter_rep"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/rate_limiter/internal/limiter/service/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewService(t *testing.T) {
	t.Run("creates service", func(t *testing.T) {
		rep := mockLimiterRep.NewLimiterRepository(t)
		svc := NewService(rep)
		assert.NotNil(t, svc)
	})
}

func TestCheckRateLimit(t *testing.T) {
	defaultConfig := dto.RateLimiterConfig{
		UserIP: "192.168.1.1",
		Action: "login",
		Window: 1 * time.Minute,
		Limit:  5,
	}

	repoConfig := repositoryDto.RateLimiterConfig{
		UserIP: defaultConfig.UserIP,
		Action: defaultConfig.Action,
		Window: defaultConfig.Window,
	}

	tests := []struct {
		nameTest        string
		config          dto.RateLimiterConfig
		mockBehavior    func(m *mockLimiterRep.LimiterRepository)
		expectedExceeded bool
		expectedErr     string
	}{
		{
			nameTest: "Not exceeded",
			config:   defaultConfig,
			mockBehavior: func(m *mockLimiterRep.LimiterRepository) {
				m.On("CheckLimit", mock.Anything, repoConfig).Return(int64(3), nil)
			},
			expectedExceeded: false,
		},
		{
			nameTest: "Limit exceeded",
			config:   defaultConfig,
			mockBehavior: func(m *mockLimiterRep.LimiterRepository) {
				m.On("CheckLimit", mock.Anything, repoConfig).Return(int64(6), nil)
			},
			expectedExceeded: true,
		},
		{
			nameTest: "Repository error",
			config:   defaultConfig,
			mockBehavior: func(m *mockLimiterRep.LimiterRepository) {
				m.On("CheckLimit", mock.Anything, repoConfig).Return(int64(0), errors.New("redis error"))
			},
			expectedErr: fmt.Errorf("rep.CheckLimit: %w", errors.New("redis error")).Error(),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			rep := mockLimiterRep.NewLimiterRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(rep)
			}

			svc := NewService(rep)
			exceeded, err := svc.CheckRateLimit(context.Background(), test.config)

			if test.expectedErr != "" {
				assert.EqualError(t, err, test.expectedErr)
				assert.False(t, exceeded)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedExceeded, exceeded)
			}
		})
	}
}

func TestSetCooldown(t *testing.T) {
	defaultConfig := dto.CooldownConfig{
		Name:       "login",
		Email:      "test@mail.ru",
		Expiration: 1 * time.Minute,
	}

	expectedKey := "cd:login:test@mail.ru"

	tests := []struct {
		nameTest        string
		config          dto.CooldownConfig
		mockBehavior    func(m *mockLimiterRep.LimiterRepository)
		expectedAllowed bool
		expectedWait    time.Duration
		expectedErr     string
	}{
		{
			nameTest: "Cooldown allowed",
			config:   defaultConfig,
			mockBehavior: func(m *mockLimiterRep.LimiterRepository) {
				m.On("SetCooldown", mock.Anything, repositoryDto.CooldownConfig{
					Key:        expectedKey,
					Expiration: defaultConfig.Expiration,
				}).Return(true, time.Duration(0), nil)
			},
			expectedAllowed: true,
			expectedWait:    0,
		},
		{
			nameTest: "Cooldown active",
			config:   defaultConfig,
			mockBehavior: func(m *mockLimiterRep.LimiterRepository) {
				m.On("SetCooldown", mock.Anything, repositoryDto.CooldownConfig{
					Key:        expectedKey,
					Expiration: defaultConfig.Expiration,
				}).Return(false, 30*time.Second, nil)
			},
			expectedAllowed: false,
			expectedWait:    30 * time.Second,
		},
		{
			nameTest: "Repository error",
			config:   defaultConfig,
			mockBehavior: func(m *mockLimiterRep.LimiterRepository) {
				m.On("SetCooldown", mock.Anything, mock.Anything).
					Return(false, time.Duration(0), errors.New("redis error"))
			},
			expectedErr: fmt.Errorf("rep.SetCooldown: %w", errors.New("redis error")).Error(),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			rep := mockLimiterRep.NewLimiterRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(rep)
			}

			svc := NewService(rep)
			allowed, wait, err := svc.SetCooldown(context.Background(), test.config)

			if test.expectedErr != "" {
				assert.EqualError(t, err, test.expectedErr)
				assert.False(t, allowed)
				assert.Zero(t, wait)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedAllowed, allowed)
				assert.Equal(t, test.expectedWait, wait)
			}
		})
	}
}
