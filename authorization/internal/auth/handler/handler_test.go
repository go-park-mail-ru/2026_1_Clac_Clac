package handler

import (
	"context"
	"errors"
	"testing"

	mockAuthSrv "github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/auth/handler/mock_auth_srv"
	mockVkOAuth "github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/auth/handler/mock_vk_oauth"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/common"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/auth/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// newHandler — хелпер для тестов, не использующих VK OAuth.
func newHandler(srv *mockAuthSrv.AuthService) *Handler {
	return NewHandler(srv, nil)
}

func TestNewHandler(t *testing.T) {
	t.Run("creates handler", func(t *testing.T) {
		mockSrv := mockAuthSrv.NewAuthService(t)
		h := NewHandler(mockSrv, nil)
		assert.NotNil(t, h)
	})
}

func TestHandlerCreateSession(t *testing.T) {
	userUUID := uuid.New()

	tests := []struct {
		nameTest     string
		req          *pb.CreateSessionRequest
		mockBehavior func(m *mockAuthSrv.AuthService)
		expectedID   string
		expectedCode codes.Code
	}{
		{
			nameTest: "Success create session",
			req:      &pb.CreateSessionRequest{UserLink: userUUID.String()},
			mockBehavior: func(m *mockAuthSrv.AuthService) {
				m.On("CreateSession", mock.Anything, userUUID).Return("session123", nil)
			},
			expectedID:   "session123",
			expectedCode: codes.OK,
		},
		{
			nameTest:     "Invalid UUID",
			req:          &pb.CreateSessionRequest{UserLink: "not-a-uuid"},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Service error",
			req:      &pb.CreateSessionRequest{UserLink: userUUID.String()},
			mockBehavior: func(m *mockAuthSrv.AuthService) {
				m.On("CreateSession", mock.Anything, userUUID).Return("", errors.New("internal"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockSrv := mockAuthSrv.NewAuthService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockSrv)
			}

			h := newHandler(mockSrv)
			resp, err := h.CreateSession(context.Background(), test.req)

			if test.expectedCode != codes.OK {
				assert.Nil(t, resp)
				s, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, test.expectedCode, s.Code())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedID, resp.SessionId)
			}
		})
	}
}

func TestHandlerGetUserLink(t *testing.T) {
	tests := []struct {
		nameTest     string
		req          *pb.GetUserLinkRequest
		mockBehavior func(m *mockAuthSrv.AuthService)
		expectedLink string
		expectedCode codes.Code
	}{
		{
			nameTest: "Success get user link",
			req:      &pb.GetUserLinkRequest{SessionId: "session123"},
			mockBehavior: func(m *mockAuthSrv.AuthService) {
				m.On("GetUserLink", mock.Anything, "session123").Return("user-link-uuid", nil)
			},
			expectedLink: "user-link-uuid",
			expectedCode: codes.OK,
		},
		{
			nameTest: "Session not found",
			req:      &pb.GetUserLinkRequest{SessionId: "session123"},
			mockBehavior: func(m *mockAuthSrv.AuthService) {
				m.On("GetUserLink", mock.Anything, "session123").Return("", common.ErrorNotExistingSession)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Service internal error",
			req:      &pb.GetUserLinkRequest{SessionId: "session123"},
			mockBehavior: func(m *mockAuthSrv.AuthService) {
				m.On("GetUserLink", mock.Anything, "session123").Return("", errors.New("internal"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockSrv := mockAuthSrv.NewAuthService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockSrv)
			}

			h := newHandler(mockSrv)
			resp, err := h.GetUserLink(context.Background(), test.req)

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

func TestHandlerDeleteSession(t *testing.T) {
	tests := []struct {
		nameTest     string
		req          *pb.DeleteSessionRequest
		mockBehavior func(m *mockAuthSrv.AuthService)
		expectedCode codes.Code
	}{
		{
			nameTest: "Success delete session",
			req:      &pb.DeleteSessionRequest{SessionId: "session123"},
			mockBehavior: func(m *mockAuthSrv.AuthService) {
				m.On("DeleteSession", mock.Anything, "session123").Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			nameTest: "Service error",
			req:      &pb.DeleteSessionRequest{SessionId: "session123"},
			mockBehavior: func(m *mockAuthSrv.AuthService) {
				m.On("DeleteSession", mock.Anything, "session123").Return(errors.New("internal"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockSrv := mockAuthSrv.NewAuthService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockSrv)
			}

			h := newHandler(mockSrv)
			resp, err := h.DeleteSession(context.Background(), test.req)

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

func TestHandlerExtendSession(t *testing.T) {
	tests := []struct {
		nameTest     string
		req          *pb.ExtendSessionRequest
		mockBehavior func(m *mockAuthSrv.AuthService)
		expectedCode codes.Code
	}{
		{
			nameTest: "Success extend session",
			req:      &pb.ExtendSessionRequest{SessionId: "session123"},
			mockBehavior: func(m *mockAuthSrv.AuthService) {
				m.On("ExtendSession", mock.Anything, "session123").Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			nameTest: "Session not found",
			req:      &pb.ExtendSessionRequest{SessionId: "session123"},
			mockBehavior: func(m *mockAuthSrv.AuthService) {
				m.On("ExtendSession", mock.Anything, "session123").Return(common.ErrorNotExistingSession)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Service internal error",
			req:      &pb.ExtendSessionRequest{SessionId: "session123"},
			mockBehavior: func(m *mockAuthSrv.AuthService) {
				m.On("ExtendSession", mock.Anything, "session123").Return(errors.New("internal"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockSrv := mockAuthSrv.NewAuthService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockSrv)
			}

			h := newHandler(mockSrv)
			resp, err := h.ExtendSession(context.Background(), test.req)

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

func TestHandlerExchangeVKCode(t *testing.T) {
	tokenWithEmail := &oauth2.Token{
		AccessToken: "vk_access_token_123",
	}
	tokenWithEmail = tokenWithEmail.WithExtra(map[string]interface{}{
		"email": "user@vk.com",
	})

	tokenEmptyEmail := &oauth2.Token{
		AccessToken: "vk_access_token_456",
	}
	tokenEmptyEmail = tokenEmptyEmail.WithExtra(map[string]interface{}{
		"email": "",
	})

	tests := []struct {
		nameTest      string
		req           *pb.ExchangeVKCodeRequest
		mockBehavior  func(vk *mockVkOAuth.VkOAuth)
		expectedCode  codes.Code
		expectedToken string
		expectedEmail string
	}{
		{
			nameTest: "Success exchange code",
			req:      &pb.ExchangeVKCodeRequest{Code: "vk_code_abc"},
			mockBehavior: func(vk *mockVkOAuth.VkOAuth) {
				vk.On("Exchange", mock.Anything, "vk_code_abc").Return(tokenWithEmail, nil)
			},
			expectedCode:  codes.OK,
			expectedToken: "vk_access_token_123",
			expectedEmail: "user@vk.com",
		},
		{
			nameTest: "Error exchange fails",
			req:      &pb.ExchangeVKCodeRequest{Code: "bad_code"},
			mockBehavior: func(vk *mockVkOAuth.VkOAuth) {
				vk.On("Exchange", mock.Anything, "bad_code").Return((*oauth2.Token)(nil), errors.New("vk exchange error"))
			},
			expectedCode: codes.Unavailable,
		},
		{
			nameTest: "Error no email in token",
			req:      &pb.ExchangeVKCodeRequest{Code: "vk_code_noemail"},
			mockBehavior: func(vk *mockVkOAuth.VkOAuth) {
				tokenNoEmail := &oauth2.Token{AccessToken: "tok"}
				vk.On("Exchange", mock.Anything, "vk_code_noemail").Return(tokenNoEmail, nil)
			},
			expectedCode: codes.Unavailable,
		},
		{
			nameTest: "Error empty email string in token",
			req:      &pb.ExchangeVKCodeRequest{Code: "vk_code_emptyemail"},
			mockBehavior: func(vk *mockVkOAuth.VkOAuth) {
				vk.On("Exchange", mock.Anything, "vk_code_emptyemail").Return(tokenEmptyEmail, nil)
			},
			expectedCode: codes.Unavailable,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockSrv := mockAuthSrv.NewAuthService(t)
			mockVK := mockVkOAuth.NewVkOAuth(t)

			if test.mockBehavior != nil {
				test.mockBehavior(mockVK)
			}

			h := NewHandler(mockSrv, mockVK)
			resp, err := h.ExchangeVKCode(context.Background(), test.req)

			if test.expectedCode != codes.OK {
				assert.Nil(t, resp)
				s, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, test.expectedCode, s.Code())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedToken, resp.AccessToken)
				assert.Equal(t, test.expectedEmail, resp.Email)
			}
		})
	}
}
