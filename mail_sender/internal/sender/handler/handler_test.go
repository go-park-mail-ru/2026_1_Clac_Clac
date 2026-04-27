package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/mail_sender/internal/common"
	mockServiceSender "github.com/go-park-mail-ru/2026_1_Clac_Clac/mail_sender/internal/sender/handler/mock_service_sender"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/contracts/mail_sender"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewHandler(t *testing.T) {
	t.Run("creates handler", func(t *testing.T) {
		mockSrv := mockServiceSender.NewServiceSender(t)
		h := NewHandler(mockSrv)
		assert.NotNil(t, h)
	})
}

func TestHandlerSendRecoveryCode(t *testing.T) {
	userUUID := uuid.New()

	tests := []struct {
		nameTest     string
		req          *pb.SendRecoveryCodeRequest
		mockBehavior func(m *mockServiceSender.ServiceSender)
		expectedCode codes.Code
	}{
		{
			nameTest: "Success send recovery code",
			req:      &pb.SendRecoveryCodeRequest{UserLink: userUUID.String(), Email: "test@mail.ru"},
			mockBehavior: func(m *mockServiceSender.ServiceSender) {
				m.On("SendRecoveryCode", mock.Anything, userUUID, "test@mail.ru").Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			nameTest:     "Invalid user link UUID",
			req:          &pb.SendRecoveryCodeRequest{UserLink: "not-a-uuid", Email: "test@mail.ru"},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Service error",
			req:      &pb.SendRecoveryCodeRequest{UserLink: userUUID.String(), Email: "test@mail.ru"},
			mockBehavior: func(m *mockServiceSender.ServiceSender) {
				m.On("SendRecoveryCode", mock.Anything, userUUID, "test@mail.ru").Return(errors.New("internal"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockSrv := mockServiceSender.NewServiceSender(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockSrv)
			}

			h := NewHandler(mockSrv)
			resp, err := h.SendRecoveryCode(context.Background(), test.req)

			if test.expectedCode != codes.OK {
				assert.Nil(t, resp)
				s, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, test.expectedCode, s.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestHandlerCheckRecoveryCode(t *testing.T) {
	tests := []struct {
		nameTest     string
		req          *pb.CheckRecoveryCodeRequest
		mockBehavior func(m *mockServiceSender.ServiceSender)
		expectedCode codes.Code
	}{
		{
			nameTest: "Success check recovery code",
			req:      &pb.CheckRecoveryCodeRequest{Code: "123456"},
			mockBehavior: func(m *mockServiceSender.ServiceSender) {
				m.On("CheckRecoveryCode", mock.Anything, "123456").Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			nameTest: "Token not found",
			req:      &pb.CheckRecoveryCodeRequest{Code: "123456"},
			mockBehavior: func(m *mockServiceSender.ServiceSender) {
				m.On("CheckRecoveryCode", mock.Anything, "123456").Return(common.ErrorNotExistingResetToken)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Service internal error",
			req:      &pb.CheckRecoveryCodeRequest{Code: "123456"},
			mockBehavior: func(m *mockServiceSender.ServiceSender) {
				m.On("CheckRecoveryCode", mock.Anything, "123456").Return(errors.New("internal"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockSrv := mockServiceSender.NewServiceSender(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockSrv)
			}

			h := NewHandler(mockSrv)
			resp, err := h.CheckRecoveryCode(context.Background(), test.req)

			if test.expectedCode != codes.OK {
				assert.Nil(t, resp)
				s, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, test.expectedCode, s.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestHandlerExchangeTokenForUser(t *testing.T) {
	expectedLink := uuid.New().String()

	tests := []struct {
		nameTest     string
		req          *pb.ExchangeTokenRequest
		mockBehavior func(m *mockServiceSender.ServiceSender)
		expectedLink string
		expectedCode codes.Code
	}{
		{
			nameTest: "Success exchange token",
			req:      &pb.ExchangeTokenRequest{ResetToken: "token123"},
			mockBehavior: func(m *mockServiceSender.ServiceSender) {
				m.On("GetUserLink", mock.Anything, "token123").Return(expectedLink, nil)
			},
			expectedLink: expectedLink,
			expectedCode: codes.OK,
		},
		{
			nameTest: "Token not found",
			req:      &pb.ExchangeTokenRequest{ResetToken: "token123"},
			mockBehavior: func(m *mockServiceSender.ServiceSender) {
				m.On("GetUserLink", mock.Anything, "token123").Return("", common.ErrorNotExistingResetToken)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Service internal error",
			req:      &pb.ExchangeTokenRequest{ResetToken: "token123"},
			mockBehavior: func(m *mockServiceSender.ServiceSender) {
				m.On("GetUserLink", mock.Anything, "token123").Return("", errors.New("internal"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockSrv := mockServiceSender.NewServiceSender(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockSrv)
			}

			h := NewHandler(mockSrv)
			resp, err := h.ExchangeTokenForUser(context.Background(), test.req)

			if test.expectedCode != codes.OK {
				assert.Nil(t, resp)
				s, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, test.expectedCode, s.Code())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedLink, resp.UserLink)
			}
		})
	}
}
