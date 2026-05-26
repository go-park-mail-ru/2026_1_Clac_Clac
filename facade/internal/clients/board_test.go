package clients

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/board/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockBoardServiceClient struct {
	mock.Mock
}

func (m *mockBoardServiceClient) GetBoards(ctx context.Context, in *pb.GetBoardsRequest, opts ...grpc.CallOption) (*pb.GetBoardsResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.GetBoardsResponse), args.Error(1)
}

func (m *mockBoardServiceClient) GetBoard(ctx context.Context, in *pb.GetBoardRequest, opts ...grpc.CallOption) (*pb.GetBoardResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.GetBoardResponse), args.Error(1)
}

func (m *mockBoardServiceClient) CreateBoard(ctx context.Context, in *pb.CreateBoardRequest, opts ...grpc.CallOption) (*pb.CreateBoardResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.CreateBoardResponse), args.Error(1)
}

func (m *mockBoardServiceClient) DeleteBoard(ctx context.Context, in *pb.DeleteBoardRequest, opts ...grpc.CallOption) (*pb.DeleteBoardResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.DeleteBoardResponse), args.Error(1)
}

func (m *mockBoardServiceClient) UpdateBoard(ctx context.Context, in *pb.UpdateBoardRequest, opts ...grpc.CallOption) (*pb.UpdateBoardResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.UpdateBoardResponse), args.Error(1)
}

func (m *mockBoardServiceClient) UploadBackground(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[pb.UploadBackgroundRequest, pb.UploadBackgroundResponse], error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(grpc.ClientStreamingClient[pb.UploadBackgroundRequest, pb.UploadBackgroundResponse]), args.Error(1)
}

func (m *mockBoardServiceClient) GetMembers(ctx context.Context, in *pb.GetMembersRequest, opts ...grpc.CallOption) (*pb.GetMembersResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.GetMembersResponse), args.Error(1)
}

func (m *mockBoardServiceClient) CreateInvite(ctx context.Context, in *pb.CreateInviteRequest, opts ...grpc.CallOption) (*pb.CreateInviteResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.CreateInviteResponse), args.Error(1)
}

func (m *mockBoardServiceClient) AcceptInvite(ctx context.Context, in *pb.AcceptInviteRequest, opts ...grpc.CallOption) (*pb.AcceptInviteResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.AcceptInviteResponse), args.Error(1)
}

func (m *mockBoardServiceClient) CloseInvite(ctx context.Context, in *pb.CloseInviteRequest, opts ...grpc.CallOption) (*pb.CloseInviteResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.CloseInviteResponse), args.Error(1)
}

func (m *mockBoardServiceClient) UpdateMemberRole(ctx context.Context, in *pb.UpdateMemberRoleRequest, opts ...grpc.CallOption) (*pb.UpdateMemberRoleResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.UpdateMemberRoleResponse), args.Error(1)
}

func (m *mockBoardServiceClient) RemoveMemberFromBoard(ctx context.Context, in *pb.RemoveMemberFromBoardRequest, opts ...grpc.CallOption) (*pb.RemoveMemberFromBoardResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.RemoveMemberFromBoardResponse), args.Error(1)
}

func (m *mockBoardServiceClient) GetActiveInvites(ctx context.Context, in *pb.GetActiveInvitesRequest, opts ...grpc.CallOption) (*pb.GetActiveInvitesResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.GetActiveInvitesResponse), args.Error(1)
}

func (m *mockBoardServiceClient) CanView(ctx context.Context, in *pb.CanViewRequest, opts ...grpc.CallOption) (*pb.CanViewResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.CanViewResponse), args.Error(1)
}

func (m *mockBoardServiceClient) CreatePoll(ctx context.Context, in *pb.CreatePollRequest, opts ...grpc.CallOption) (*pb.CreatePollResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.CreatePollResponse), args.Error(1)
}

func (m *mockBoardServiceClient) DeletePoll(ctx context.Context, in *pb.DeletePollRequest, opts ...grpc.CallOption) (*pb.DeletePollResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.DeletePollResponse), args.Error(1)
}

func (m *mockBoardServiceClient) NextPollCard(ctx context.Context, in *pb.NextPollCardRequest, opts ...grpc.CallOption) (*pb.NextPollCardResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.NextPollCardResponse), args.Error(1)
}

func (m *mockBoardServiceClient) VotePoll(ctx context.Context, in *pb.VotePollRequest, opts ...grpc.CallOption) (*pb.VotePollResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.VotePollResponse), args.Error(1)
}

func (m *mockBoardServiceClient) GetActivePoll(ctx context.Context, in *pb.GetActivePollRequest, opts ...grpc.CallOption) (*pb.GetActivePollResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.GetActivePollResponse), args.Error(1)
}

func TestBoardCreateInvite(t *testing.T) {
	userLink := uuid.New()
	boardLink := uuid.New()
	targetUser := uuid.New()
	inviteLink := uuid.New()

	tests := []struct {
		name        string
		setupMock   func(m *mockBoardServiceClient)
		request     domain.CreateInviteRequest
		expectError bool
		errorIs     error
	}{
		{
			name: "Success public invite",
			setupMock: func(m *mockBoardServiceClient) {
				resp := &pb.CreateInviteResponse{
					InviteLink:  inviteLink.String(),
					BoardLink:   boardLink.String(),
					DefaultRole: "editor",
					Status:      "active",
					CreatedAt:   100,
				}
				m.On("CreateInvite", mock.Anything, mock.Anything).Return(resp, nil)
			},
			request: domain.CreateInviteRequest{
				UserLink:    userLink,
				BoardLink:   boardLink,
				DefaultRole: "editor",
			},
			expectError: false,
		},
		{
			name: "Success personal invite with target user",
			setupMock: func(m *mockBoardServiceClient) {
				target := targetUser.String()
				resp := &pb.CreateInviteResponse{
					InviteLink:     inviteLink.String(),
					BoardLink:      boardLink.String(),
					TargetUserLink: &target,
					DefaultRole:    "viewer",
					Status:         "active",
					CreatedAt:      100,
				}
				m.On("CreateInvite", mock.Anything, mock.Anything).Return(resp, nil)
			},
			request: domain.CreateInviteRequest{
				UserLink:       userLink,
				BoardLink:      boardLink,
				TargetUserLink: &targetUser,
				DefaultRole:    "viewer",
			},
			expectError: false,
		},
		{
			name: "Board not found",
			setupMock: func(m *mockBoardServiceClient) {
				m.On("CreateInvite", mock.Anything, mock.Anything).Return(nil, status.Error(codes.NotFound, "board not found"))
			},
			request: domain.CreateInviteRequest{
				UserLink:    userLink,
				BoardLink:   boardLink,
				DefaultRole: "editor",
			},
			expectError: true,
			errorIs:     common.ErrorBoardNotFound,
		},
		{
			name: "Permission denied",
			setupMock: func(m *mockBoardServiceClient) {
				m.On("CreateInvite", mock.Anything, mock.Anything).Return(nil, status.Error(codes.PermissionDenied, "action denied"))
			},
			request: domain.CreateInviteRequest{
				UserLink:    userLink,
				BoardLink:   boardLink,
				DefaultRole: "editor",
			},
			expectError: true,
			errorIs:     common.ErrorBoardPermissionDenied,
		},
		{
			name: "gRPC error",
			setupMock: func(m *mockBoardServiceClient) {
				m.On("CreateInvite", mock.Anything, mock.Anything).Return(nil, status.Error(codes.Internal, "internal"))
			},
			request: domain.CreateInviteRequest{
				UserLink:    userLink,
				BoardLink:   boardLink,
				DefaultRole: "editor",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockBoardServiceClient)
			tt.setupMock(m)

			client := &Board{client: m}

			_, err := client.CreateInvite(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorIs != nil {
					assert.ErrorIs(t, err, tt.errorIs)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBoardAcceptInvite(t *testing.T) {
	userLink := uuid.New()
	inviteLink := uuid.New()
	boardLink := uuid.New()

	tests := []struct {
		name        string
		setupMock   func(m *mockBoardServiceClient)
		request     domain.AcceptInviteRequest
		expectError bool
		errorIs     error
	}{
		{
			name: "Success",
			setupMock: func(m *mockBoardServiceClient) {
				resp := &pb.AcceptInviteResponse{
					BoardLink: boardLink.String(),
					UserLink:  userLink.String(),
					Role:      "editor",
				}
				m.On("AcceptInvite", mock.Anything, mock.Anything).Return(resp, nil)
			},
			request: domain.AcceptInviteRequest{
				InviteLink: inviteLink.String(),
				UserLink:   userLink,
			},
			expectError: false,
		},
		{
			name: "Invite not found",
			setupMock: func(m *mockBoardServiceClient) {
				m.On("AcceptInvite", mock.Anything, mock.Anything).Return(nil, status.Error(codes.NotFound, "invite not found"))
			},
			request: domain.AcceptInviteRequest{
				InviteLink: inviteLink.String(),
				UserLink:   userLink,
			},
			expectError: true,
			errorIs:     common.ErrorInviteNotFound,
		},
		{
			name: "Already member",
			setupMock: func(m *mockBoardServiceClient) {
				m.On("AcceptInvite", mock.Anything, mock.Anything).Return(nil, status.Error(codes.AlreadyExists, "already a member"))
			},
			request: domain.AcceptInviteRequest{
				InviteLink: inviteLink.String(),
				UserLink:   userLink,
			},
			expectError: true,
			errorIs:     common.ErrorUserAlreadyMember,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockBoardServiceClient)
			tt.setupMock(m)

			client := &Board{client: m}

			_, _, err := client.AcceptInvite(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorIs != nil {
					assert.ErrorIs(t, err, tt.errorIs)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBoardCloseInvite(t *testing.T) {
	userLink := uuid.New()
	inviteLink := uuid.New()

	tests := []struct {
		name        string
		setupMock   func(m *mockBoardServiceClient)
		request     domain.CloseInviteRequest
		expectError bool
		errorIs     error
	}{
		{
			name: "Success",
			setupMock: func(m *mockBoardServiceClient) {
				m.On("CloseInvite", mock.Anything, mock.Anything).Return(&pb.CloseInviteResponse{}, nil)
			},
			request: domain.CloseInviteRequest{
				UserLink:   userLink,
				InviteLink: inviteLink.String(),
			},
			expectError: false,
		},
		{
			name: "Invite not found",
			setupMock: func(m *mockBoardServiceClient) {
				m.On("CloseInvite", mock.Anything, mock.Anything).Return(nil, status.Error(codes.NotFound, "invite not found"))
			},
			request: domain.CloseInviteRequest{
				UserLink:   userLink,
				InviteLink: inviteLink.String(),
			},
			expectError: true,
			errorIs:     common.ErrorInviteNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockBoardServiceClient)
			tt.setupMock(m)

			client := &Board{client: m}

			err := client.CloseInvite(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorIs != nil {
					assert.ErrorIs(t, err, tt.errorIs)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
