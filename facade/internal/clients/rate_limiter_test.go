package clients

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/rate_limiter/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockRateLimiterServiceClient struct {
	mock.Mock
}

func (m *mockRateLimiterServiceClient) CheckRateLimit(ctx context.Context, in *pb.CheckRateLimitRequest, opts ...grpc.CallOption) (*pb.CheckRateLimitResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.CheckRateLimitResponse), args.Error(1)
}

func (m *mockRateLimiterServiceClient) SetCooldown(ctx context.Context, in *pb.SetCooldownRequest, opts ...grpc.CallOption) (*pb.SetCooldownResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.SetCooldownResponse), args.Error(1)
}

func TestUpdateCountRequests(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name             string
		check            domain.RateLimitCheck
		mockResp         *pb.CheckRateLimitResponse
		mockErr          error
		expectedExceeded bool
		expectedErr      error
	}{
		{
			name: "success not exceeded",
			check: domain.RateLimitCheck{
				UserIP:  "192.168.1.1",
				Action:  "login",
				WindowS: 60,
				Limit:   10,
			},
			mockResp:         &pb.CheckRateLimitResponse{Exceeded: false},
			mockErr:          nil,
			expectedExceeded: false,
			expectedErr:      nil,
		},
		{
			name: "success exceeded",
			check: domain.RateLimitCheck{
				UserIP:  "192.168.1.1",
				Action:  "login",
				WindowS: 60,
				Limit:   10,
			},
			mockResp:         &pb.CheckRateLimitResponse{Exceeded: true},
			mockErr:          nil,
			expectedExceeded: true,
			expectedErr:      nil,
		},
		{
			name: "grpc error",
			check: domain.RateLimitCheck{
				UserIP:  "192.168.1.1",
				Action:  "login",
				WindowS: 60,
				Limit:   10,
			},
			mockResp:         nil,
			mockErr:          status.Error(codes.Unavailable, "service unavailable"),
			expectedExceeded: false,
			expectedErr:      common.ErrorVKOAuthUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockRateLimiterServiceClient)
			mc.On("CheckRateLimit", ctx, &pb.CheckRateLimitRequest{
				UserIp:  tt.check.UserIP,
				Action:  tt.check.Action,
				WindowS: tt.check.WindowS,
				Limit:   tt.check.Limit,
			}).Return(tt.mockResp, tt.mockErr)

			r := &RateLimiter{client: mc}
			exceeded, err := r.UpdateCountRequests(ctx, tt.check)

			assert.Equal(t, tt.expectedExceeded, exceeded)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSetCooldown(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		cooldown       domain.Cooldown
		mockResp       *pb.SetCooldownResponse
		mockErr        error
		expectedResult domain.CooldownResult
		expectedErr    error
	}{
		{
			name: "success allowed",
			cooldown: domain.Cooldown{
				Name:        "send_code",
				Email:       "user@example.com",
				ExpirationS: 300,
			},
			mockResp: &pb.SetCooldownResponse{
				Allowed: true,
				WaitS:   0,
			},
			mockErr: nil,
			expectedResult: domain.CooldownResult{
				Allowed: true,
				WaitS:   0,
			},
			expectedErr: nil,
		},
		{
			name: "success not allowed with wait",
			cooldown: domain.Cooldown{
				Name:        "send_code",
				Email:       "user@example.com",
				ExpirationS: 300,
			},
			mockResp: &pb.SetCooldownResponse{
				Allowed: false,
				WaitS:   120,
			},
			mockErr: nil,
			expectedResult: domain.CooldownResult{
				Allowed: false,
				WaitS:   120,
			},
			expectedErr: nil,
		},
		{
			name: "grpc error",
			cooldown: domain.Cooldown{
				Name:        "send_code",
				Email:       "user@example.com",
				ExpirationS: 300,
			},
			mockResp:       nil,
			mockErr:        status.Error(codes.Unavailable, "service unavailable"),
			expectedResult: domain.CooldownResult{},
			expectedErr:    common.ErrorVKOAuthUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockRateLimiterServiceClient)
			mc.On("SetCooldown", ctx, &pb.SetCooldownRequest{
				Name:        tt.cooldown.Name,
				Email:       tt.cooldown.Email,
				ExpirationS: tt.cooldown.ExpirationS,
			}).Return(tt.mockResp, tt.mockErr)

			r := &RateLimiter{client: mc}
			result, err := r.SetCooldown(ctx, tt.cooldown)

			assert.Equal(t, tt.expectedResult, result)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
