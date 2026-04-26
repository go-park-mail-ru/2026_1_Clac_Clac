package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/contracts/rate_limiter"
	mockServiceLimiter "github.com/go-park-mail-ru/2026_1_Clac_Clac/rate_limiter/internal/limiter/handler/mock_service_limiter"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/rate_limiter/internal/limiter/service/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewHandler(t *testing.T) {
	t.Run("creates handler", func(t *testing.T) {
		mockSrv := mockServiceLimiter.NewServiceLimiter(t)
		h := NewHandler(mockSrv)
		assert.NotNil(t, h)
	})
}

func TestHandlerCheckRateLimit(t *testing.T) {
	tests := []struct {
		nameTest         string
		req              *pb.CheckRateLimitRequest
		mockBehavior     func(m *mockServiceLimiter.ServiceLimiter)
		expectedExceeded bool
		expectedCode     codes.Code
	}{
		{
			nameTest: "Success not exceeded",
			req: &pb.CheckRateLimitRequest{
				UserIp:  "192.168.1.1",
				Action:  "login",
				WindowS: 60,
				Limit:   5,
			},
			mockBehavior: func(m *mockServiceLimiter.ServiceLimiter) {
				m.On("CheckRateLimit", mock.Anything, serviceDto.RateLimiterConfig{
					UserIP: "192.168.1.1",
					Action: "login",
					Window: 60 * time.Second,
					Limit:  5,
				}).Return(false, nil)
			},
			expectedExceeded: false,
			expectedCode:     codes.OK,
		},
		{
			nameTest: "Success exceeded",
			req: &pb.CheckRateLimitRequest{
				UserIp:  "192.168.1.1",
				Action:  "login",
				WindowS: 60,
				Limit:   5,
			},
			mockBehavior: func(m *mockServiceLimiter.ServiceLimiter) {
				m.On("CheckRateLimit", mock.Anything, mock.Anything).Return(true, nil)
			},
			expectedExceeded: true,
			expectedCode:     codes.OK,
		},
		{
			nameTest: "Invalid input: empty user_ip",
			req: &pb.CheckRateLimitRequest{
				UserIp:  "",
				Action:  "login",
				WindowS: 60,
				Limit:   5,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Invalid input: empty action",
			req: &pb.CheckRateLimitRequest{
				UserIp:  "192.168.1.1",
				Action:  "",
				WindowS: 60,
				Limit:   5,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Invalid input: window_ms <= 0",
			req: &pb.CheckRateLimitRequest{
				UserIp:  "192.168.1.1",
				Action:  "login",
				WindowS: 0, Limit: 5,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Invalid input: limit <= 0",
			req: &pb.CheckRateLimitRequest{
				UserIp:  "192.168.1.1",
				Action:  "login",
				WindowS: 60,
				Limit:   0,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Service error",
			req: &pb.CheckRateLimitRequest{
				UserIp:  "192.168.1.1",
				Action:  "login",
				WindowS: 60,
				Limit:   5,
			},
			mockBehavior: func(m *mockServiceLimiter.ServiceLimiter) {
				m.On("CheckRateLimit", mock.Anything, mock.Anything).Return(false, errors.New("internal"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockSrv := mockServiceLimiter.NewServiceLimiter(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockSrv)
			}

			h := NewHandler(mockSrv)
			resp, err := h.CheckRateLimit(context.Background(), test.req)

			if test.expectedCode != codes.OK {
				assert.Nil(t, resp)
				s, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, test.expectedCode, s.Code())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedExceeded, resp.Exceeded)
			}
		})
	}
}

func TestHandlerSetCooldown(t *testing.T) {
	tests := []struct {
		nameTest        string
		req             *pb.SetCooldownRequest
		mockBehavior    func(m *mockServiceLimiter.ServiceLimiter)
		expectedAllowed bool
		expectedWaitMs  int64
		expectedCode    codes.Code
	}{
		{
			nameTest: "Success cooldown allowed",
			req: &pb.SetCooldownRequest{
				Name:        "login",
				Email:       "test@mail.ru",
				ExpirationS: 60,
			},
			mockBehavior: func(m *mockServiceLimiter.ServiceLimiter) {
				m.On("SetCooldown", mock.Anything, serviceDto.CooldownConfig{
					Name:       "login",
					Email:      "test@mail.ru",
					Expiration: 60 * time.Second,
				}).Return(true, time.Duration(0), nil)
			},
			expectedAllowed: true,
			expectedWaitMs:  0,
			expectedCode:    codes.OK,
		},
		{
			nameTest: "Cooldown active",
			req: &pb.SetCooldownRequest{
				Name:        "login",
				Email:       "test@mail.ru",
				ExpirationS: 60,
			},
			mockBehavior: func(m *mockServiceLimiter.ServiceLimiter) {
				m.On("SetCooldown", mock.Anything, mock.Anything).Return(false, 30*time.Second, nil)
			},
			expectedAllowed: false,
			expectedWaitMs:  30000,
			expectedCode:    codes.OK,
		},
		{
			nameTest: "Invalid input: empty name",
			req: &pb.SetCooldownRequest{
				Name:        "",
				Email:       "test@mail.ru",
				ExpirationS: 60,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Invalid input: empty email",
			req: &pb.SetCooldownRequest{
				Name:        "login",
				Email:       "",
				ExpirationS: 60,
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Invalid input: expiration <= 0",
			req: &pb.SetCooldownRequest{
				Name:        "login",
				Email:       "test@mail.ru",
				ExpirationS: 0},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Service error",
			req: &pb.SetCooldownRequest{
				Name:        "login",
				Email:       "test@mail.ru",
				ExpirationS: 60,
			},
			mockBehavior: func(m *mockServiceLimiter.ServiceLimiter) {
				m.On("SetCooldown", mock.Anything, mock.Anything).Return(false, time.Duration(0), errors.New("internal"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockSrv := mockServiceLimiter.NewServiceLimiter(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockSrv)
			}

			h := NewHandler(mockSrv)
			resp, err := h.SetCooldown(context.Background(), test.req)

			if test.expectedCode != codes.OK {
				assert.Nil(t, resp)
				s, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, test.expectedCode, s.Code())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedAllowed, resp.Allowed)
				assert.Equal(t, test.expectedWaitMs, resp.WaitS)
			}
		})
	}
}
