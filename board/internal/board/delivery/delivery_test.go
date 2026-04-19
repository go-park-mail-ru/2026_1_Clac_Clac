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
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/service/dto"
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
			name:      "Success get boards",
			req:       &pb.GetBoardsRequest{UserLink: userLink.String()},
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
				m.On("DeleteBoard", mock.Anything, boardLink, userLink).Return(common.ErrActionDenied).Once()
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
				m.On("UpdateBoard", mock.Anything, mock.Anything, userLink).Return(common.ErrActionDenied).Once()
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
				m.On("GetBoard", mock.Anything, boardLink, userLink).Return(serviceDto.BoardInfo{}, common.ErrActionDenied).Once()
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
	usersLinks := []uuid.UUID{uuid.New(), uuid.New()}

	tests := []struct {
		name         string
		req          *pb.GetMembersRequest
		setupMock    func(m *mocks.BoardService)
		expectedCode codes.Code
	}{
		{
			name: "Success get members",
			req:  &pb.GetMembersRequest{BoardLink: boardLink.String()},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetUsersOfBoard", mock.Anything, boardLink).Return(usersLinks, nil).Once()
			},
			expectedCode: codes.OK,
		},
		{
			name:         "Error invalid board link",
			req:          &pb.GetMembersRequest{BoardLink: "bad-uuid"},
			setupMock:    func(m *mocks.BoardService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Error board not found",
			req:  &pb.GetMembersRequest{BoardLink: boardLink.String()},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetUsersOfBoard", mock.Anything, boardLink).Return(nil, common.ErrBoardNotFound).Once()
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Error internal server",
			req:  &pb.GetMembersRequest{BoardLink: boardLink.String()},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetUsersOfBoard", mock.Anything, boardLink).Return(nil, errors.New("db error")).Once()
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
