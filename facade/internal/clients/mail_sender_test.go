package clients

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/mail_sender/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockMailSenderServiceClient struct {
	mock.Mock
}

func (m *mockMailSenderServiceClient) SendRecoveryCode(ctx context.Context, in *pb.SendRecoveryCodeRequest, opts ...grpc.CallOption) (*pb.SendRecoveryCodeResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.SendRecoveryCodeResponse), args.Error(1)
}

func (m *mockMailSenderServiceClient) CheckRecoveryCode(ctx context.Context, in *pb.CheckRecoveryCodeRequest, opts ...grpc.CallOption) (*pb.CheckRecoveryCodeResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.CheckRecoveryCodeResponse), args.Error(1)
}

func (m *mockMailSenderServiceClient) ExchangeTokenForUser(ctx context.Context, in *pb.ExchangeTokenRequest, opts ...grpc.CallOption) (*pb.ExchangeTokenResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ExchangeTokenResponse), args.Error(1)
}

func TestSendRecoveryCode(t *testing.T) {
	ctx := context.Background()
	validUUID := uuid.New()

	tests := []struct {
		name         string
		recoveryInfo domain.RecoveryCode
		mockResp     *pb.SendRecoveryCodeResponse
		mockErr      error
		expectedErr  error
	}{
		{
			name: "success",
			recoveryInfo: domain.RecoveryCode{
				UserLink: validUUID,
				Email:    "user@example.com",
			},
			mockResp:    &pb.SendRecoveryCodeResponse{},
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name: "grpc error",
			recoveryInfo: domain.RecoveryCode{
				UserLink: validUUID,
				Email:    "user@example.com",
			},
			mockResp:    nil,
			mockErr:     status.Error(codes.NotFound, "email not found"),
			expectedErr: common.ErrorNonexistentEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockMailSenderServiceClient)
			mc.On("SendRecoveryCode", ctx, &pb.SendRecoveryCodeRequest{
				Email:    tt.recoveryInfo.Email,
				UserLink: tt.recoveryInfo.UserLink.String(),
			}).Return(tt.mockResp, tt.mockErr)

			ms := &MailSender{client: mc}
			err := ms.SendRecoveryCode(ctx, tt.recoveryInfo)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckRecoveryCode(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		check       domain.RecoveryCodeCheck
		mockResp    *pb.CheckRecoveryCodeResponse
		mockErr     error
		expectedErr error
	}{
		{
			name:        "success",
			check:       domain.RecoveryCodeCheck{Code: "123456"},
			mockResp:    &pb.CheckRecoveryCodeResponse{},
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "grpc error",
			check:       domain.RecoveryCodeCheck{Code: "000000"},
			mockResp:    nil,
			mockErr:     status.Error(codes.NotFound, "reset token expired"),
			expectedErr: common.ErrorResetTokenNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockMailSenderServiceClient)
			mc.On("CheckRecoveryCode", ctx, &pb.CheckRecoveryCodeRequest{
				Code: tt.check.Code,
			}).Return(tt.mockResp, tt.mockErr)

			ms := &MailSender{client: mc}
			err := ms.CheckRecoveryCode(ctx, tt.check)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExchangeTokenForUser(t *testing.T) {
	ctx := context.Background()
	validUUID := uuid.New()

	tests := []struct {
		name         string
		resetToken   domain.ResetToken
		mockResp     *pb.ExchangeTokenResponse
		mockErr      error
		expectedLink uuid.UUID
		expectedErr  error
	}{
		{
			name:         "success",
			resetToken:   domain.ResetToken{Token: "reset-token-xyz"},
			mockResp:     &pb.ExchangeTokenResponse{UserLink: validUUID.String()},
			mockErr:      nil,
			expectedLink: validUUID,
			expectedErr:  nil,
		},
		{
			name:         "grpc error",
			resetToken:   domain.ResetToken{Token: "expired-token"},
			mockResp:     nil,
			mockErr:      status.Error(codes.NotFound, "reset token not found"),
			expectedLink: uuid.Nil,
			expectedErr:  common.ErrorResetTokenNotFound,
		},
		{
			name:         "invalid uuid in response",
			resetToken:   domain.ResetToken{Token: "reset-token-xyz"},
			mockResp:     &pb.ExchangeTokenResponse{UserLink: "bad-uuid"},
			mockErr:      nil,
			expectedLink: uuid.Nil,
			expectedErr:  common.ErrorParseLink,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockMailSenderServiceClient)
			mc.On("ExchangeTokenForUser", ctx, &pb.ExchangeTokenRequest{
				ResetToken: tt.resetToken.Token,
			}).Return(tt.mockResp, tt.mockErr)

			ms := &MailSender{client: mc}
			link, err := ms.ExchangeTokenForUser(ctx, tt.resetToken)

			assert.Equal(t, tt.expectedLink, link)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
