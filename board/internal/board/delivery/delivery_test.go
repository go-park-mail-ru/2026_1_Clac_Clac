package delivery_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/common"
	handler "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/delivery"
	mocks "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/delivery/mock_board_srv"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/service"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/service/dto"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/boardRbac"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/board/v1"
)

var testConf = handler.Config{
	MaxBackgroundSize:          10 << 20,
	MultipartBackgroundFileKey: "background",
}

func grpcCode(err error) codes.Code {
	return status.Code(err)
}

func TestGetBoards(t *testing.T) {
	userLink := uuid.New()
	boardsInfo := []serviceDto.BoardInfo{
		{Link: uuid.New(), Name: "Board 1", CreatedAt: time.Now()},
		{Link: uuid.New(), Name: "Board 2", CreatedAt: time.Now()},
	}

	tests := []struct {
		name         string
		req          *pb.GetBoardsRequest
		setupMock    func(m *mocks.BoardService)
		expectedCode codes.Code
	}{
		{
			name: "Success get boards",
			req:  &pb.GetBoardsRequest{UserLink: userLink.String()},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetBoards", mock.Anything, userLink).Return(boardsInfo, nil).Once()
			},
			expectedCode: codes.OK,
		},
		{
			name:         "Error invalid user link",
			req:          &pb.GetBoardsRequest{UserLink: "not-a-uuid"},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Error internal server",
			req:  &pb.GetBoardsRequest{UserLink: userLink.String()},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetBoards", mock.Anything, userLink).Return(nil, errors.New("db error")).Once()
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			_, err := h.GetBoards(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestCreateBoard(t *testing.T) {
	userLink := uuid.New()
	boardInfo := serviceDto.BoardInfo{Link: uuid.New(), Name: "New Board"}

	tests := []struct {
		name         string
		req          *pb.CreateBoardRequest
		setupMock    func(m *mocks.BoardService)
		expectedCode codes.Code
	}{
		{
			name: "Success create board",
			req:  &pb.CreateBoardRequest{UserLink: userLink.String(), Name: "New Board", Description: "Desc"},
			setupMock: func(m *mocks.BoardService) {
				m.On("CreateBoard", mock.Anything, mock.Anything, userLink).
					Return(boardInfo, nil).Once()
			},
			expectedCode: codes.OK,
		},
		{
			name:         "Error invalid user link",
			req:          &pb.CreateBoardRequest{UserLink: "bad-uuid", Name: "Board"},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Error missing required field",
			req:  &pb.CreateBoardRequest{UserLink: userLink.String(), Name: "Board"},
			setupMock: func(m *mocks.BoardService) {
				m.On("CreateBoard", mock.Anything, mock.Anything, userLink).
					Return(serviceDto.BoardInfo{}, common.ErrNotNullValue).Once()
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Error invalid board data",
			req:  &pb.CreateBoardRequest{UserLink: userLink.String(), Name: "Board"},
			setupMock: func(m *mocks.BoardService) {
				m.On("CreateBoard", mock.Anything, mock.Anything, userLink).
					Return(serviceDto.BoardInfo{}, common.ErrInvalidBoardData).Once()
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Error invalid board reference",
			req:  &pb.CreateBoardRequest{UserLink: userLink.String(), Name: "Board"},
			setupMock: func(m *mocks.BoardService) {
				m.On("CreateBoard", mock.Anything, mock.Anything, userLink).
					Return(serviceDto.BoardInfo{}, common.ErrInvalidBoardReference).Once()
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Error user already member",
			req:  &pb.CreateBoardRequest{UserLink: userLink.String(), Name: "Board"},
			setupMock: func(m *mocks.BoardService) {
				m.On("CreateBoard", mock.Anything, mock.Anything, userLink).
					Return(serviceDto.BoardInfo{}, common.ErrUserAlreadyMember).Once()
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Error internal server",
			req:  &pb.CreateBoardRequest{UserLink: userLink.String(), Name: "Board"},
			setupMock: func(m *mocks.BoardService) {
				m.On("CreateBoard", mock.Anything, mock.Anything, userLink).
					Return(serviceDto.BoardInfo{}, errors.New("db error")).Once()
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			_, err := h.CreateBoard(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestDeleteBoard(t *testing.T) {
	userLink := uuid.New()
	boardLink := uuid.New()

	tests := []struct {
		name         string
		req          *pb.DeleteBoardRequest
		setupMock    func(m *mocks.BoardService)
		expectedCode codes.Code
	}{
		{
			name: "Success delete board",
			req:  &pb.DeleteBoardRequest{UserLink: userLink.String(), BoardLink: boardLink.String()},
			setupMock: func(m *mocks.BoardService) {
				m.On("DeleteBoard", mock.Anything, boardLink, userLink).Return(nil).Once()
			},
			expectedCode: codes.OK,
		},
		{
			name:         "Error invalid user link",
			req:          &pb.DeleteBoardRequest{UserLink: "bad-uuid", BoardLink: boardLink.String()},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "Error invalid board link",
			req:          &pb.DeleteBoardRequest{UserLink: userLink.String(), BoardLink: "bad-uuid"},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Error forbidden",
			req:  &pb.DeleteBoardRequest{UserLink: userLink.String(), BoardLink: boardLink.String()},
			setupMock: func(m *mocks.BoardService) {
				m.On("DeleteBoard", mock.Anything, boardLink, userLink).Return(rbac.ErrActionDenied).Once()
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "Error not found",
			req:  &pb.DeleteBoardRequest{UserLink: userLink.String(), BoardLink: boardLink.String()},
			setupMock: func(m *mocks.BoardService) {
				m.On("DeleteBoard", mock.Anything, boardLink, userLink).Return(common.ErrBoardNotFound).Once()
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Error internal server",
			req:  &pb.DeleteBoardRequest{UserLink: userLink.String(), BoardLink: boardLink.String()},
			setupMock: func(m *mocks.BoardService) {
				m.On("DeleteBoard", mock.Anything, boardLink, userLink).Return(errors.New("db crash")).Once()
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			_, err := h.DeleteBoard(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestUpdateBoard(t *testing.T) {
	userLink := uuid.New()
	boardLink := uuid.New()

	tests := []struct {
		name         string
		req          *pb.UpdateBoardRequest
		setupMock    func(m *mocks.BoardService)
		expectedCode codes.Code
	}{
		{
			name: "Success update board",
			req:  &pb.UpdateBoardRequest{UserLink: userLink.String(), BoardLink: boardLink.String(), Name: "Updated Name"},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateBoard", mock.Anything, mock.Anything, userLink).Return(nil).Once()
			},
			expectedCode: codes.OK,
		},
		{
			name:         "Error invalid user link",
			req:          &pb.UpdateBoardRequest{UserLink: "bad-uuid", BoardLink: boardLink.String()},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "Error invalid board link",
			req:          &pb.UpdateBoardRequest{UserLink: userLink.String(), BoardLink: "bad-uuid"},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Error forbidden",
			req:  &pb.UpdateBoardRequest{UserLink: userLink.String(), BoardLink: boardLink.String(), Name: "Name"},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateBoard", mock.Anything, mock.Anything, userLink).Return(rbac.ErrActionDenied).Once()
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "Error not found",
			req:  &pb.UpdateBoardRequest{UserLink: userLink.String(), BoardLink: boardLink.String(), Name: "Name"},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateBoard", mock.Anything, mock.Anything, userLink).Return(common.ErrBoardNotFound).Once()
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Error invalid board data",
			req:  &pb.UpdateBoardRequest{UserLink: userLink.String(), BoardLink: boardLink.String(), Name: "Name"},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateBoard", mock.Anything, mock.Anything, userLink).
					Return(common.ErrInvalidBoardData).Once()
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Error invalid board reference",
			req:  &pb.UpdateBoardRequest{UserLink: userLink.String(), BoardLink: boardLink.String(), Name: "Name"},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateBoard", mock.Anything, mock.Anything, userLink).
					Return(common.ErrInvalidBoardReference).Once()
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Error internal server",
			req:  &pb.UpdateBoardRequest{UserLink: userLink.String(), BoardLink: boardLink.String(), Name: "Name"},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateBoard", mock.Anything, mock.Anything, userLink).
					Return(errors.New("some unexpected error")).Once()
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			_, err := h.UpdateBoard(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestGetBoard(t *testing.T) {
	userLink := uuid.New()
	boardLink := uuid.New()
	boardInfo := serviceDto.BoardInfo{Link: boardLink, Name: "Target Board", CreatedAt: time.Now()}

	tests := []struct {
		name         string
		req          *pb.GetBoardRequest
		setupMock    func(m *mocks.BoardService)
		expectedCode codes.Code
	}{
		{
			name: "Success get board",
			req:  &pb.GetBoardRequest{UserLink: userLink.String(), BoardLink: boardLink.String()},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetBoard", mock.Anything, boardLink, userLink).Return(boardInfo, nil).Once()
			},
			expectedCode: codes.OK,
		},
		{
			name:         "Error invalid user link",
			req:          &pb.GetBoardRequest{UserLink: "bad-uuid", BoardLink: boardLink.String()},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "Error invalid board link",
			req:          &pb.GetBoardRequest{UserLink: userLink.String(), BoardLink: "bad-uuid"},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Error forbidden",
			req:  &pb.GetBoardRequest{UserLink: userLink.String(), BoardLink: boardLink.String()},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetBoard", mock.Anything, boardLink, userLink).Return(serviceDto.BoardInfo{}, rbac.ErrActionDenied).Once()
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "Error not found",
			req:  &pb.GetBoardRequest{UserLink: userLink.String(), BoardLink: boardLink.String()},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetBoard", mock.Anything, boardLink, userLink).Return(serviceDto.BoardInfo{}, common.ErrBoardNotFound).Once()
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Error internal server",
			req:  &pb.GetBoardRequest{UserLink: userLink.String(), BoardLink: boardLink.String()},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetBoard", mock.Anything, boardLink, userLink).Return(serviceDto.BoardInfo{}, errors.New("db error")).Once()
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			_, err := h.GetBoard(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestUploadBackground(t *testing.T) {
	userLink := uuid.New()
	boardLink := uuid.New()
	pngImage := append([]byte("\x89PNG\r\n\x1a\n"), []byte("dummy image content")...)

	tests := []struct {
		name         string
		req          *pb.UploadBackgroundRequest
		setupMock    func(m *mocks.BoardService)
		expectedCode codes.Code
	}{
		{
			name: "Success upload background",
			req: &pb.UploadBackgroundRequest{
				UserLink:  userLink.String(),
				BoardLink: boardLink.String(),
				Image:     pngImage,
				Filename:  "bg.png",
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateBackground", mock.Anything, mock.Anything, "image/png", ".png", boardLink, userLink).
					Return("backgrounds/bg.png", nil).Once()
			},
			expectedCode: codes.OK,
		},
		{
			name: "Error invalid user link",
			req: &pb.UploadBackgroundRequest{
				UserLink:  "bad-uuid",
				BoardLink: boardLink.String(),
				Image:     pngImage,
				Filename:  "bg.png",
			},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Error invalid board link",
			req: &pb.UploadBackgroundRequest{
				UserLink:  userLink.String(),
				BoardLink: "bad-uuid",
				Image:     pngImage,
				Filename:  "bg.png",
			},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Error invalid content type",
			req: &pb.UploadBackgroundRequest{
				UserLink:  userLink.String(),
				BoardLink: boardLink.String(),
				Image:     []byte("just a regular text string"),
				Filename:  "text.txt",
			},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Error board not found",
			req: &pb.UploadBackgroundRequest{
				UserLink:  userLink.String(),
				BoardLink: boardLink.String(),
				Image:     pngImage,
				Filename:  "bg.png",
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateBackground", mock.Anything, mock.Anything, "image/png", ".png", boardLink, userLink).
					Return("", common.ErrBoardNotFound).Once()
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Error internal service",
			req: &pb.UploadBackgroundRequest{
				UserLink:  userLink.String(),
				BoardLink: boardLink.String(),
				Image:     pngImage,
				Filename:  "bg.png",
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateBackground", mock.Anything, mock.Anything, "image/png", ".png", boardLink, userLink).
					Return("", errors.New("s3 upload failed")).Once()
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			_, err := h.UploadBackground(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestGetMembers(t *testing.T) {
	boardLink := uuid.New()
	userLink := uuid.New()
	member1 := uuid.New()
	member2 := uuid.New()
	members := []serviceDto.MemberInfo{
		{Link: member1, Role: rbac.Roles.Editor},
		{Link: member2, Role: rbac.Roles.Viewer},
	}

	tests := []struct {
		name         string
		req          *pb.GetMembersRequest
		setupMock    func(m *mocks.BoardService)
		expectedCode codes.Code
	}{
		{
			name: "Success get members",
			req:  &pb.GetMembersRequest{BoardLink: boardLink.String(), UserLink: userLink.String()},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetUsersOfBoard", mock.Anything, boardLink, userLink).Return(members, nil).Once()
			},
			expectedCode: codes.OK,
		},
		{
			name:         "Error invalid board link",
			req:          &pb.GetMembersRequest{BoardLink: "bad-uuid", UserLink: userLink.String()},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "Error invalid user link",
			req:          &pb.GetMembersRequest{BoardLink: boardLink.String(), UserLink: "bad-uuid"},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Error board not found",
			req:  &pb.GetMembersRequest{BoardLink: boardLink.String(), UserLink: userLink.String()},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetUsersOfBoard", mock.Anything, boardLink, userLink).Return(nil, common.ErrBoardNotFound).Once()
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Error permission denied",
			req:  &pb.GetMembersRequest{BoardLink: boardLink.String(), UserLink: userLink.String()},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetUsersOfBoard", mock.Anything, boardLink, userLink).Return(nil, rbac.ErrActionDenied).Once()
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "Error internal server",
			req:  &pb.GetMembersRequest{BoardLink: boardLink.String(), UserLink: userLink.String()},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetUsersOfBoard", mock.Anything, boardLink, userLink).Return(nil, errors.New("db error")).Once()
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			_, err := h.GetMembers(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestCreateInviteHandler(t *testing.T) {
	userLink := uuid.New()
	boardLink := uuid.New()

	inviteInfo := serviceDto.InviteInfo{
		InviteLink:  uuid.New(),
		BoardLink:   boardLink,
		TargetUser:  nil,
		DefaultRole: rbac.Roles.Editor,
		Status:      common.InviteStatuses.Active,
		CreatedAt:   time.Now(),
	}

	tests := []struct {
		name         string
		req          *pb.CreateInviteRequest
		setupMock    func(m *mocks.BoardService)
		expectedCode codes.Code
	}{
		{
			name: "Success create invite",
			req: &pb.CreateInviteRequest{
				UserLink:      userLink.String(),
				BoardLink:     boardLink.String(),
				DefaultRole:   "editor",
				ExpireSeconds: 86400,
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("CreateInvite", mock.Anything, mock.AnythingOfType("dto.NewInviteInfo"), userLink).Return(inviteInfo, nil).Once()
			},
			expectedCode: codes.OK,
		},
		{
			name: "Invalid user link",
			req: &pb.CreateInviteRequest{
				UserLink:    "not-a-uuid",
				BoardLink:   boardLink.String(),
				DefaultRole: "editor",
			},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Invalid board link",
			req: &pb.CreateInviteRequest{
				UserLink:    userLink.String(),
				BoardLink:   "not-a-uuid",
				DefaultRole: "editor",
			},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Empty role",
			req: &pb.CreateInviteRequest{
				UserLink:    userLink.String(),
				BoardLink:   boardLink.String(),
				DefaultRole: "",
			},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Invalid role",
			req: &pb.CreateInviteRequest{
				UserLink:    userLink.String(),
				BoardLink:   boardLink.String(),
				DefaultRole: "superadmin",
			},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Permission denied",
			req: &pb.CreateInviteRequest{
				UserLink:    userLink.String(),
				BoardLink:   boardLink.String(),
				DefaultRole: "editor",
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("CreateInvite", mock.Anything, mock.AnythingOfType("dto.NewInviteInfo"), userLink).Return(serviceDto.InviteInfo{}, rbac.ErrActionDenied).Once()
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "Internal server error",
			req: &pb.CreateInviteRequest{
				UserLink:    userLink.String(),
				BoardLink:   boardLink.String(),
				DefaultRole: "editor",
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("CreateInvite", mock.Anything, mock.AnythingOfType("dto.NewInviteInfo"), userLink).Return(serviceDto.InviteInfo{}, errors.New("db error")).Once()
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			_, err := h.CreateInvite(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestAcceptInviteHandler(t *testing.T) {
	inviteLink := uuid.New()
	userLink := uuid.New()
	boardLink := uuid.New()

	inviteInfo := serviceDto.InviteInfo{
		InviteLink:  inviteLink,
		BoardLink:   boardLink,
		DefaultRole: rbac.Roles.Editor,
		Status:      common.InviteStatuses.Active,
	}

	tests := []struct {
		name         string
		req          *pb.AcceptInviteRequest
		setupMock    func(m *mocks.BoardService)
		expectedCode codes.Code
	}{
		{
			name: "Success accept invite",
			req: &pb.AcceptInviteRequest{
				InviteLink: inviteLink.String(),
				UserLink:   userLink.String(),
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("AcceptInvite", mock.Anything, inviteLink, userLink).Return(inviteInfo, nil).Once()
			},
			expectedCode: codes.OK,
		},
		{
			name: "Invalid invite link",
			req: &pb.AcceptInviteRequest{
				InviteLink: "not-a-uuid",
				UserLink:   userLink.String(),
			},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Invalid user link",
			req: &pb.AcceptInviteRequest{
				InviteLink: inviteLink.String(),
				UserLink:   "not-a-uuid",
			},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Invite not found",
			req: &pb.AcceptInviteRequest{
				InviteLink: inviteLink.String(),
				UserLink:   userLink.String(),
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("AcceptInvite", mock.Anything, inviteLink, userLink).Return(serviceDto.InviteInfo{}, common.ErrInviteNotFound).Once()
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Invite closed",
			req: &pb.AcceptInviteRequest{
				InviteLink: inviteLink.String(),
				UserLink:   userLink.String(),
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("AcceptInvite", mock.Anything, inviteLink, userLink).Return(serviceDto.InviteInfo{}, common.ErrInviteClosed).Once()
			},
			expectedCode: codes.FailedPrecondition,
		},
		{
			name: "Invite expired",
			req: &pb.AcceptInviteRequest{
				InviteLink: inviteLink.String(),
				UserLink:   userLink.String(),
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("AcceptInvite", mock.Anything, inviteLink, userLink).Return(serviceDto.InviteInfo{}, common.ErrInviteExpired).Once()
			},
			expectedCode: codes.FailedPrecondition,
		},
		{
			name: "Not for this user",
			req: &pb.AcceptInviteRequest{
				InviteLink: inviteLink.String(),
				UserLink:   userLink.String(),
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("AcceptInvite", mock.Anything, inviteLink, userLink).Return(serviceDto.InviteInfo{}, common.ErrInviteNotForUser).Once()
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "User already member",
			req: &pb.AcceptInviteRequest{
				InviteLink: inviteLink.String(),
				UserLink:   userLink.String(),
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("AcceptInvite", mock.Anything, inviteLink, userLink).Return(serviceDto.InviteInfo{}, common.ErrUserAlreadyMember).Once()
			},
			expectedCode: codes.AlreadyExists,
		},
		{
			name: "Internal server error",
			req: &pb.AcceptInviteRequest{
				InviteLink: inviteLink.String(),
				UserLink:   userLink.String(),
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("AcceptInvite", mock.Anything, inviteLink, userLink).Return(serviceDto.InviteInfo{}, errors.New("db error")).Once()
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			_, err := h.AcceptInvite(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestCloseInviteHandler(t *testing.T) {
	inviteLink := uuid.New()
	userLink := uuid.New()

	tests := []struct {
		name         string
		req          *pb.CloseInviteRequest
		setupMock    func(m *mocks.BoardService)
		expectedCode codes.Code
	}{
		{
			name: "Success close invite",
			req: &pb.CloseInviteRequest{
				UserLink:   userLink.String(),
				InviteLink: inviteLink.String(),
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("CloseInvite", mock.Anything, inviteLink, userLink).Return(nil).Once()
			},
			expectedCode: codes.OK,
		},
		{
			name: "Invalid user link",
			req: &pb.CloseInviteRequest{
				UserLink:   "not-a-uuid",
				InviteLink: inviteLink.String(),
			},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Invalid invite link",
			req: &pb.CloseInviteRequest{
				UserLink:   userLink.String(),
				InviteLink: "not-a-uuid",
			},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Permission denied",
			req: &pb.CloseInviteRequest{
				UserLink:   userLink.String(),
				InviteLink: inviteLink.String(),
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("CloseInvite", mock.Anything, inviteLink, userLink).Return(rbac.ErrActionDenied).Once()
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "Invite not found",
			req: &pb.CloseInviteRequest{
				UserLink:   userLink.String(),
				InviteLink: inviteLink.String(),
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("CloseInvite", mock.Anything, inviteLink, userLink).Return(common.ErrInviteNotFound).Once()
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Internal server error",
			req: &pb.CloseInviteRequest{
				UserLink:   userLink.String(),
				InviteLink: inviteLink.String(),
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("CloseInvite", mock.Anything, inviteLink, userLink).Return(errors.New("db error")).Once()
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			_, err := h.CloseInvite(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestUpdateMemberRoleHandler(t *testing.T) {
	userLink := uuid.New()
	boardLink := uuid.New()
	targetLink := uuid.New()

	tests := []struct {
		name         string
		req          *pb.UpdateMemberRoleRequest
		setupMock    func(m *mocks.BoardService)
		expectedCode codes.Code
	}{
		{
			name: "Success",
			req: &pb.UpdateMemberRoleRequest{
				UserLink:       userLink.String(),
				BoardLink:      boardLink.String(),
				TargetUserLink: targetLink.String(),
				NewRole:        "editor",
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateMemberRole", mock.Anything, boardLink, targetLink, rbac.Roles.Editor, userLink).Return(nil).Once()
			},
			expectedCode: codes.OK,
		},
		{
			name: "Permission denied",
			req: &pb.UpdateMemberRoleRequest{
				UserLink:       userLink.String(),
				BoardLink:      boardLink.String(),
				TargetUserLink: targetLink.String(),
				NewRole:        "editor",
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateMemberRole", mock.Anything, boardLink, targetLink, rbac.Roles.Editor, userLink).Return(rbac.ErrActionDenied).Once()
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "Invalid user link",
			req: &pb.UpdateMemberRoleRequest{
				UserLink:       "not-a-uuid",
				BoardLink:      boardLink.String(),
				TargetUserLink: targetLink.String(),
				NewRole:        "editor",
			},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "User not found",
			req: &pb.UpdateMemberRoleRequest{
				UserLink:       userLink.String(),
				BoardLink:      boardLink.String(),
				TargetUserLink: targetLink.String(),
				NewRole:        "editor",
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateMemberRole", mock.Anything, boardLink, targetLink, rbac.Roles.Editor, userLink).Return(common.ErrUserNotFound).Once()
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Self role change",
			req: &pb.UpdateMemberRoleRequest{
				UserLink:       userLink.String(),
				BoardLink:      boardLink.String(),
				TargetUserLink: userLink.String(),
				NewRole:        "editor",
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateMemberRole", mock.Anything, boardLink, userLink, rbac.Roles.Editor, userLink).Return(common.ErrSelfRoleChange).Once()
			},
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			_, err := h.UpdateMemberRole(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestRemoveMemberFromBoardHandler(t *testing.T) {
	userLink := uuid.New()
	boardLink := uuid.New()
	targetLink := uuid.New()

	tests := []struct {
		name         string
		req          *pb.RemoveMemberFromBoardRequest
		setupMock    func(m *mocks.BoardService)
		expectedCode codes.Code
	}{
		{
			name: "Success",
			req: &pb.RemoveMemberFromBoardRequest{
				UserLink:       userLink.String(),
				BoardLink:      boardLink.String(),
				TargetUserLink: targetLink.String(),
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("RemoveMemberFromBoard", mock.Anything, boardLink, targetLink, userLink).Return(nil).Once()
			},
			expectedCode: codes.OK,
		},
		{
			name: "Permission denied",
			req: &pb.RemoveMemberFromBoardRequest{
				UserLink:       userLink.String(),
				BoardLink:      boardLink.String(),
				TargetUserLink: targetLink.String(),
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("RemoveMemberFromBoard", mock.Anything, boardLink, targetLink, userLink).Return(rbac.ErrActionDenied).Once()
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "User not found",
			req: &pb.RemoveMemberFromBoardRequest{
				UserLink:       userLink.String(),
				BoardLink:      boardLink.String(),
				TargetUserLink: targetLink.String(),
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("RemoveMemberFromBoard", mock.Anything, boardLink, targetLink, userLink).Return(common.ErrUserNotFound).Once()
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Creator cannot leave",
			req: &pb.RemoveMemberFromBoardRequest{
				UserLink:       userLink.String(),
				BoardLink:      boardLink.String(),
				TargetUserLink: userLink.String(),
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("RemoveMemberFromBoard", mock.Anything, boardLink, userLink, userLink).Return(common.ErrCreatorCannotLeave).Once()
			},
			expectedCode: codes.PermissionDenied,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			_, err := h.RemoveMemberFromBoard(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestCanViewHandler(t *testing.T) {
	userLink := uuid.New()
	boardLink := uuid.New()

	tests := []struct {
		name         string
		req          *pb.CanViewRequest
		setupMock    func(m *mocks.BoardService)
		expectedCode codes.Code
	}{
		{
			name: "Success can view",
			req:  &pb.CanViewRequest{UserLink: userLink.String(), BoardLink: boardLink.String()},
			setupMock: func(m *mocks.BoardService) {
				m.On("CanView", mock.Anything, boardLink, userLink).Return(nil).Once()
			},
			expectedCode: codes.OK,
		},
		{
			name:         "Error invalid user link",
			req:          &pb.CanViewRequest{UserLink: "bad-uuid", BoardLink: boardLink.String()},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "Error invalid board link",
			req:          &pb.CanViewRequest{UserLink: userLink.String(), BoardLink: "bad-uuid"},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Error permission denied",
			req:  &pb.CanViewRequest{UserLink: userLink.String(), BoardLink: boardLink.String()},
			setupMock: func(m *mocks.BoardService) {
				m.On("CanView", mock.Anything, boardLink, userLink).Return(rbac.ErrActionDenied).Once()
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "Error internal server",
			req:  &pb.CanViewRequest{UserLink: userLink.String(), BoardLink: boardLink.String()},
			setupMock: func(m *mocks.BoardService) {
				m.On("CanView", mock.Anything, boardLink, userLink).Return(errors.New("db error")).Once()
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			_, err := h.CanView(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestGetActivePollHandler(t *testing.T) {
	userLink := uuid.New()
	boardLink := uuid.New()

	adminLink := uuid.New()
	cardLink := uuid.New()
	invitedUser := uuid.New()
	points := 5
	poll := &service.Poll{
		BoardLink:  boardLink,
		AdminLink:  adminLink,
		CurrentIdx: 0,
		Tasks: []service.PollTask{
			{
				CardLink: cardLink,
				Title:    "Task 1",
				Votes:    map[uuid.UUID]*int{invitedUser: &points},
			},
		},
		Invitees: []uuid.UUID{invitedUser},
	}

	tests := []struct {
		name         string
		req          *pb.GetActivePollRequest
		setupMock    func(m *mocks.BoardService)
		expectedCode codes.Code
	}{
		{
			name: "Success",
			req: &pb.GetActivePollRequest{
				UserLink:  userLink.String(),
				BoardLink: boardLink.String(),
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetActivePoll", mock.Anything, boardLink, userLink).Return(poll, nil).Once()
			},
			expectedCode: codes.OK,
		},
		{
			name: "Error_InvalidUserLink",
			req: &pb.GetActivePollRequest{
				UserLink:  "bad-uuid",
				BoardLink: boardLink.String(),
			},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Error_InvalidBoardLink",
			req: &pb.GetActivePollRequest{
				UserLink:  userLink.String(),
				BoardLink: "bad-uuid",
			},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Error_PermissionDenied",
			req: &pb.GetActivePollRequest{
				UserLink:  userLink.String(),
				BoardLink: boardLink.String(),
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetActivePoll", mock.Anything, boardLink, userLink).Return(nil, rbac.ErrActionDenied).Once()
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "Error_PollNotFound",
			req: &pb.GetActivePollRequest{
				UserLink:  userLink.String(),
				BoardLink: boardLink.String(),
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetActivePoll", mock.Anything, boardLink, userLink).Return(nil, common.ErrPollNotFound).Once()
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Error_InternalServerError",
			req: &pb.GetActivePollRequest{
				UserLink:  userLink.String(),
				BoardLink: boardLink.String(),
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetActivePoll", mock.Anything, boardLink, userLink).Return(nil, errors.New("some error")).Once()
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			resp, err := h.GetActivePoll(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			mockSrv.AssertExpectations(t)

			if test.expectedCode == codes.OK {
				assert.NotNil(t, resp)
				assert.Equal(t, adminLink.String(), resp.AdminLink)
				assert.Equal(t, int32(0), resp.CurrentIdx)
				assert.Len(t, resp.Tasks, 1)
				assert.Equal(t, cardLink.String(), resp.Tasks[0].CardLink)
				assert.Equal(t, "Task 1", resp.Tasks[0].Title)
				assert.Len(t, resp.Tasks[0].Votes, 1)
				assert.Equal(t, invitedUser.String(), resp.Tasks[0].Votes[0].UserLink)
				assert.Equal(t, int32(5), resp.Tasks[0].Votes[0].Points)
				assert.Len(t, resp.Invitees, 1)
				assert.Equal(t, invitedUser.String(), resp.Invitees[0])
			}
		})
	}
}

func TestGetActiveInvitesHandler(t *testing.T) {
	userLink := uuid.New()
	boardLink := uuid.New()

	serviceInvites := []serviceDto.InviteInfo{
		{InviteLink: uuid.New(), BoardLink: boardLink, DefaultRole: rbac.Roles.Editor, Status: common.InviteStatuses.Active, CreatedAt: time.Now()},
	}

	tests := []struct {
		name         string
		req          *pb.GetActiveInvitesRequest
		setupMock    func(m *mocks.BoardService)
		expectedCode codes.Code
	}{
		{
			name: "Success",
			req: &pb.GetActiveInvitesRequest{
				UserLink:  userLink.String(),
				BoardLink: boardLink.String(),
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetActiveInvites", mock.Anything, boardLink, userLink).Return(serviceInvites, nil).Once()
			},
			expectedCode: codes.OK,
		},
		{
			name: "Permission denied",
			req: &pb.GetActiveInvitesRequest{
				UserLink:  userLink.String(),
				BoardLink: boardLink.String(),
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetActiveInvites", mock.Anything, boardLink, userLink).Return(nil, rbac.ErrActionDenied).Once()
			},
			expectedCode: codes.PermissionDenied,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			_, err := h.GetActiveInvites(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			mockSrv.AssertExpectations(t)
		})
	}
}
