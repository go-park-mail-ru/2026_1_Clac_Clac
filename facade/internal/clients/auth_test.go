package clients

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/auth/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockAuthServiceClient struct {
	mock.Mock
}

func (m *mockAuthServiceClient) CreateSession(ctx context.Context, in *pb.CreateSessionRequest, opts ...grpc.CallOption) (*pb.CreateSessionResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.CreateSessionResponse), args.Error(1)
}

func (m *mockAuthServiceClient) GetUserLink(ctx context.Context, in *pb.GetUserLinkRequest, opts ...grpc.CallOption) (*pb.GetUserLinkResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.GetUserLinkResponse), args.Error(1)
}

func (m *mockAuthServiceClient) DeleteSession(ctx context.Context, in *pb.DeleteSessionRequest, opts ...grpc.CallOption) (*pb.DeleteSessionResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.DeleteSessionResponse), args.Error(1)
}

func (m *mockAuthServiceClient) ExtendSession(ctx context.Context, in *pb.ExtendSessionRequest, opts ...grpc.CallOption) (*pb.ExtendSessionResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ExtendSessionResponse), args.Error(1)
}

func (m *mockAuthServiceClient) ExchangeVKCode(ctx context.Context, in *pb.ExchangeVKCodeRequest, opts ...grpc.CallOption) (*pb.ExchangeVKCodeResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ExchangeVKCodeResponse), args.Error(1)
}

func TestAuthCreateSession(t *testing.T) {
	ctx := context.Background()
	validUUID := uuid.New()

	tests := []struct {
		name        string
		userLink    uuid.UUID
		mockResp    *pb.CreateSessionResponse
		mockErr     error
		expectedID  string
		expectedErr error
	}{
		{
			name:        "success",
			userLink:    validUUID,
			mockResp:    &pb.CreateSessionResponse{SessionId: "session-abc"},
			mockErr:     nil,
			expectedID:  "session-abc",
			expectedErr: nil,
		},
		{
			name:        "grpc error",
			userLink:    validUUID,
			mockResp:    nil,
			mockErr:     status.Error(codes.AlreadyExists, "exists"),
			expectedID:  "",
			expectedErr: common.ErrorExistingUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockAuthServiceClient)
			mc.On("CreateSession", ctx, &pb.CreateSessionRequest{UserLink: tt.userLink.String()}).
				Return(tt.mockResp, tt.mockErr)

			a := &Auth{client: mc}
			id, err := a.CreateSession(ctx, tt.userLink)

			assert.Equal(t, tt.expectedID, id)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthCheckSession(t *testing.T) {
	ctx := context.Background()
	validUUID := uuid.New()

	tests := []struct {
		name         string
		sessionID    string
		mockResp     *pb.GetUserLinkResponse
		mockErr      error
		expectedLink uuid.UUID
		expectedErr  error
	}{
		{
			name:         "success",
			sessionID:    "session-123",
			mockResp:     &pb.GetUserLinkResponse{UserLink: validUUID.String()},
			mockErr:      nil,
			expectedLink: validUUID,
			expectedErr:  nil,
		},
		{
			name:         "grpc error",
			sessionID:    "bad-session",
			mockResp:     nil,
			mockErr:      status.Error(codes.NotFound, "session not found"),
			expectedLink: uuid.Nil,
			expectedErr:  common.ErrorSessionNotFound,
		},
		{
			name:         "invalid uuid in response",
			sessionID:    "session-123",
			mockResp:     &pb.GetUserLinkResponse{UserLink: "not-a-uuid"},
			mockErr:      nil,
			expectedLink: uuid.Nil,
			expectedErr:  common.ErrorParseLink,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockAuthServiceClient)
			mc.On("GetUserLink", ctx, &pb.GetUserLinkRequest{SessionId: tt.sessionID}).
				Return(tt.mockResp, tt.mockErr)

			a := &Auth{client: mc}
			link, err := a.CheckSession(ctx, tt.sessionID)

			assert.Equal(t, tt.expectedLink, link)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteSession(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		sessionID   string
		mockResp    *pb.DeleteSessionResponse
		mockErr     error
		expectedErr error
	}{
		{
			name:        "success",
			sessionID:   "session-123",
			mockResp:    &pb.DeleteSessionResponse{},
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "grpc error",
			sessionID:   "session-123",
			mockResp:    nil,
			mockErr:     status.Error(codes.NotFound, "session not found"),
			expectedErr: common.ErrorSessionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockAuthServiceClient)
			mc.On("DeleteSession", ctx, &pb.DeleteSessionRequest{SessionId: tt.sessionID}).
				Return(tt.mockResp, tt.mockErr)

			a := &Auth{client: mc}
			err := a.DeleteSession(ctx, tt.sessionID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRefreshSession(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		sessionID   string
		mockResp    *pb.ExtendSessionResponse
		mockErr     error
		expectedErr error
	}{
		{
			name:        "success",
			sessionID:   "session-123",
			mockResp:    &pb.ExtendSessionResponse{},
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "grpc error",
			sessionID:   "session-123",
			mockResp:    nil,
			mockErr:     status.Error(codes.NotFound, "session not found"),
			expectedErr: common.ErrorSessionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockAuthServiceClient)
			mc.On("ExtendSession", ctx, &pb.ExtendSessionRequest{SessionId: tt.sessionID}).
				Return(tt.mockResp, tt.mockErr)

			a := &Auth{client: mc}
			err := a.RefreshSession(ctx, tt.sessionID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExchangeVKCode(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		code          string
		mockResp      *pb.ExchangeVKCodeResponse
		mockErr       error
		expectedToken string
		expectedEmail string
		expectedErr   error
	}{
		{
			name:          "success",
			code:          "vk-code-123",
			mockResp:      &pb.ExchangeVKCodeResponse{AccessToken: "token-xyz", Email: "user@example.com"},
			mockErr:       nil,
			expectedToken: "token-xyz",
			expectedEmail: "user@example.com",
			expectedErr:   nil,
		},
		{
			name:          "grpc error",
			code:          "bad-code",
			mockResp:      nil,
			mockErr:       status.Error(codes.Unavailable, "vk unavailable"),
			expectedToken: "",
			expectedEmail: "",
			expectedErr:   common.ErrorServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockAuthServiceClient)
			mc.On("ExchangeVKCode", ctx, &pb.ExchangeVKCodeRequest{Code: tt.code}).
				Return(tt.mockResp, tt.mockErr)

			a := &Auth{client: mc}
			token, email, err := a.ExchangeVKCode(ctx, tt.code)

			assert.Equal(t, tt.expectedToken, token)
			assert.Equal(t, tt.expectedEmail, email)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
