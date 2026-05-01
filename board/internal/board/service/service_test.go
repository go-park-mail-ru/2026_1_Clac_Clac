package service_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/common"
	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/service"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/service/dto"
	mocks "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/service/mock_board_rep"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/boardRbac"
)

type mockRbacService struct {
	mock.Mock
}

func (m *mockRbacService) CheckPermissionOnBoard(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID, action rbac.Action) error {
	args := m.Called(ctx, boardLink, userLink, action)
	return args.Error(0)
}

func (m *mockRbacService) CheckPermissionOnSection(ctx context.Context, sectionLink uuid.UUID, userLink uuid.UUID, action rbac.Action) error {
	args := m.Called(ctx, sectionLink, userLink, action)
	return args.Error(0)
}

func (m *mockRbacService) CheckPermissionOnCard(ctx context.Context, cardLink uuid.UUID, userLink uuid.UUID, action rbac.Action) error {
	args := m.Called(ctx, cardLink, userLink, action)
	return args.Error(0)
}

func (m *mockRbacService) CheckPermissionOnComment(ctx context.Context, commentLink uuid.UUID, userLink uuid.UUID, action rbac.Action) error {
	args := m.Called(ctx, commentLink, userLink, action)
	return args.Error(0)
}

func (m *mockRbacService) CheckPermissionOnSubtask(ctx context.Context, subtaskLink uuid.UUID, userLink uuid.UUID, action rbac.Action) error {
	args := m.Called(ctx, subtaskLink, userLink, action)
	return args.Error(0)
}

func TestGetBoards(t *testing.T) {
	userLink := uuid.New()
	ctx := context.Background()

	repoEntries := []repositoryDto.BoardEntry{
		{Link: uuid.New(), Name: "Board 1"},
		{Link: uuid.New(), Name: "Board 2"},
	}

	expectedBoards := []dto.BoardInfo{
		dto.BoardInfoFromEntry(repoEntries[0]),
		dto.BoardInfoFromEntry(repoEntries[1]),
	}

	tests := []struct {
		Name           string
		UserLink       uuid.UUID
		MockSetup      func(mockRepo *mocks.BoardRepository)
		ExpectedBoards []dto.BoardInfo
		ExpectError    bool
	}{
		{
			Name:     "success get boards",
			UserLink: userLink,
			MockSetup: func(mockRepo *mocks.BoardRepository) {
				mockRepo.On("GetBoards", ctx, userLink).Return(repoEntries, nil).Once()
			},
			ExpectedBoards: expectedBoards,
			ExpectError:    false,
		},
		{
			Name:     "error on get boards",
			UserLink: userLink,
			MockSetup: func(mockRepo *mocks.BoardRepository) {
				mockRepo.On("GetBoards", ctx, userLink).Return(nil, fmt.Errorf("db error")).Once()
			},
			ExpectedBoards: []dto.BoardInfo{},
			ExpectError:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockRepo := new(mocks.BoardRepository)
			mockPerm := new(mockRbacService)
			test.MockSetup(mockRepo)

			svc := service.NewService(mockRepo, mockPerm)

			boards, err := svc.GetBoards(ctx, test.UserLink)

			if test.ExpectError {
				assert.Error(t, err)
				assert.Equal(t, []dto.BoardInfo{}, boards)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.ExpectedBoards, boards)
			}

			mockRepo.AssertExpectations(t)
			mockPerm.AssertExpectations(t)
		})
	}
}

func TestCreateBoard(t *testing.T) {
	authorLink := uuid.New()
	ctx := context.Background()

	boardInfo := dto.NewBoardInfo{
		Name:        "Nexus Core",
		Description: "Main board",
	}

	repoBoardInfo := dto.ToNewBoardInfo(boardInfo)

	repoEntry := repositoryDto.BoardEntry{
		Link:        uuid.New(),
		Name:        "Nexus Core",
		Description: "Main board",
	}

	expectedBoard := dto.BoardInfoFromEntry(repoEntry)

	tests := []struct {
		Name          string
		BoardInfo     dto.NewBoardInfo
		AuthorLink    uuid.UUID
		MockSetup     func(mockRepo *mocks.BoardRepository)
		ExpectedBoard dto.BoardInfo
		ExpectError   bool
	}{
		{
			Name:       "success create board",
			BoardInfo:  boardInfo,
			AuthorLink: authorLink,
			MockSetup: func(mockRepo *mocks.BoardRepository) {
				mockRepo.On("CreateBoard", ctx, repoBoardInfo, authorLink).Return(repoEntry, nil).Once()
			},
			ExpectedBoard: expectedBoard,
			ExpectError:   false,
		},
		{
			Name:       "error on create board",
			BoardInfo:  boardInfo,
			AuthorLink: authorLink,
			MockSetup: func(mockRepo *mocks.BoardRepository) {
				mockRepo.On("CreateBoard", ctx, repoBoardInfo, authorLink).Return(repositoryDto.BoardEntry{}, fmt.Errorf("db error")).Once()
			},
			ExpectedBoard: dto.BoardInfo{},
			ExpectError:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockRepo := new(mocks.BoardRepository)
			mockPerm := new(mockRbacService)
			test.MockSetup(mockRepo)

			svc := service.NewService(mockRepo, mockPerm)

			board, err := svc.CreateBoard(ctx, test.BoardInfo, test.AuthorLink)

			if test.ExpectError {
				assert.Error(t, err)
				assert.Equal(t, dto.BoardInfo{}, board)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.ExpectedBoard, board)
			}

			mockRepo.AssertExpectations(t)
			mockPerm.AssertExpectations(t)
		})
	}
}

func TestDeleteBoard(t *testing.T) {
	boardLink := uuid.New()
	userLink := uuid.New()
	ctx := context.Background()

	tests := []struct {
		Name        string
		MockSetup   func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService)
		ExpectError error
	}{
		{
			Name: "success delete board",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.Delete).Return(nil).Once()
				mockRepo.On("DeleteBoard", ctx, boardLink).Return(nil).Once()
			},
			ExpectError: nil,
		},
		{
			Name: "error permission denied",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.Delete).Return(rbac.ErrActionDenied).Once()
			},
			ExpectError: rbac.ErrActionDenied,
		},
		{
			Name: "error on delete board",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.Delete).Return(nil).Once()
				mockRepo.On("DeleteBoard", ctx, boardLink).Return(fmt.Errorf("db error")).Once()
			},
			ExpectError: fmt.Errorf("db error"),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockRepo := new(mocks.BoardRepository)
			mockPerm := new(mockRbacService)
			test.MockSetup(mockRepo, mockPerm)

			svc := service.NewService(mockRepo, mockPerm)
			err := svc.DeleteBoard(ctx, boardLink, userLink)

			if test.ExpectError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
			mockPerm.AssertExpectations(t)
		})
	}
}

func TestUpdateBoard(t *testing.T) {
	ctx := context.Background()
	userLink := uuid.New()
	boardInfo := dto.UpdateBoardInfo{
		Link: uuid.New(),
		Name: "Updated Name",
	}
	repoBoardInfo := dto.ToUpdateBoardInfo(boardInfo)

	tests := []struct {
		Name        string
		MockSetup   func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService)
		ExpectError bool
	}{
		{
			Name: "success update",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardInfo.Link, userLink, rbac.Actions.Edit).Return(nil).Once()
				mockRepo.On("UpdateBoard", ctx, repoBoardInfo).Return(nil).Once()
			},
			ExpectError: false,
		},
		{
			Name: "access denied",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardInfo.Link, userLink, rbac.Actions.Edit).Return(rbac.ErrActionDenied).Once()
			},
			ExpectError: true,
		},
		{
			Name: "repo error",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardInfo.Link, userLink, rbac.Actions.Edit).Return(nil).Once()
				mockRepo.On("UpdateBoard", ctx, repoBoardInfo).Return(fmt.Errorf("db error")).Once()
			},
			ExpectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockRepo := new(mocks.BoardRepository)
			mockPerm := new(mockRbacService)
			test.MockSetup(mockRepo, mockPerm)

			svc := service.NewService(mockRepo, mockPerm)
			err := svc.UpdateBoard(ctx, boardInfo, userLink)

			if test.ExpectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
			mockPerm.AssertExpectations(t)
		})
	}
}

func TestGetBoard(t *testing.T) {
	ctx := context.Background()
	userLink := uuid.New()
	boardLink := uuid.New()

	repoEntry := repositoryDto.BoardEntry{
		Link: boardLink,
		Name: "Test Board",
	}
	expectedBoard := dto.BoardInfoFromEntry(repoEntry)

	tests := []struct {
		Name          string
		MockSetup     func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService)
		ExpectedBoard dto.BoardInfo
		ExpectError   bool
		ErrorIs       error
	}{
		{
			Name: "success get board",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.View).Return(nil).Once()
				mockRepo.On("GetBoard", ctx, boardLink).Return(repoEntry, nil).Once()
			},
			ExpectedBoard: expectedBoard,
			ExpectError:   false,
		},
		{
			Name: "permission denied",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.View).Return(rbac.ErrActionDenied).Once()
			},
			ExpectedBoard: dto.BoardInfo{},
			ExpectError:   true,
			ErrorIs:       rbac.ErrActionDenied,
		},
		{
			Name: "board not found",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.View).Return(nil).Once()
				mockRepo.On("GetBoard", ctx, boardLink).Return(repositoryDto.BoardEntry{}, common.ErrBoardNotFound).Once()
			},
			ExpectedBoard: dto.BoardInfo{},
			ExpectError:   true,
			ErrorIs:       common.ErrBoardNotFound,
		},
		{
			Name: "repo get board error",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.View).Return(nil).Once()
				mockRepo.On("GetBoard", ctx, boardLink).Return(repositoryDto.BoardEntry{}, fmt.Errorf("db error")).Once()
			},
			ExpectedBoard: dto.BoardInfo{},
			ExpectError:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockRepo := new(mocks.BoardRepository)
			mockPerm := new(mockRbacService)
			test.MockSetup(mockRepo, mockPerm)

			svc := service.NewService(mockRepo, mockPerm)
			board, err := svc.GetBoard(ctx, boardLink, userLink)

			if test.ExpectError {
				assert.Error(t, err)
				if test.ErrorIs != nil {
					assert.ErrorIs(t, err, test.ErrorIs)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.ExpectedBoard, board)
			}
			mockRepo.AssertExpectations(t)
			mockPerm.AssertExpectations(t)
		})
	}
}

func TestUpdateBackground(t *testing.T) {
	ctx := context.Background()
	userLink := uuid.New()
	boardLink := uuid.New()
	file := bytes.NewReader([]byte("dummy image content"))
	contentType := "image/png"
	extension := ".png"
	expectedKey := "s3/background/path.png"

	tests := []struct {
		Name        string
		MockSetup   func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService)
		ExpectedKey string
		ExpectError bool
		ErrorIs     error
	}{
		{
			Name: "success update background",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.Edit).Return(nil).Once()
				mockRepo.On("UploadBackground", ctx, file, mock.AnythingOfType("string"), contentType).Return(expectedKey, nil).Once()
				mockRepo.On("UpdateBackground", ctx, expectedKey, boardLink).Return(nil).Once()
			},
			ExpectedKey: expectedKey,
			ExpectError: false,
		},
		{
			Name: "permission denied",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.Edit).Return(rbac.ErrActionDenied).Once()
			},
			ExpectedKey: "",
			ExpectError: true,
			ErrorIs:     rbac.ErrActionDenied,
		},
		{
			Name: "upload failed",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.Edit).Return(nil).Once()
				mockRepo.On("UploadBackground", ctx, file, mock.AnythingOfType("string"), contentType).Return("", fmt.Errorf("s3 timeout")).Once()
			},
			ExpectedKey: "",
			ExpectError: true,
		},
		{
			Name: "update board background db error (not found)",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.Edit).Return(nil).Once()
				mockRepo.On("UploadBackground", ctx, file, mock.AnythingOfType("string"), contentType).Return(expectedKey, nil).Once()
				mockRepo.On("UpdateBackground", ctx, expectedKey, boardLink).Return(common.ErrBoardNotFound).Once()
			},
			ExpectedKey: "",
			ExpectError: true,
			ErrorIs:     common.ErrBoardNotFound,
		},
		{
			Name: "update board background db error (generic)",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.Edit).Return(nil).Once()
				mockRepo.On("UploadBackground", ctx, file, mock.AnythingOfType("string"), contentType).Return(expectedKey, nil).Once()
				mockRepo.On("UpdateBackground", ctx, expectedKey, boardLink).Return(fmt.Errorf("db crash")).Once()
			},
			ExpectedKey: "",
			ExpectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockRepo := new(mocks.BoardRepository)
			mockPerm := new(mockRbacService)
			test.MockSetup(mockRepo, mockPerm)

			svc := service.NewService(mockRepo, mockPerm)
			key, err := svc.UpdateBackground(ctx, file, contentType, extension, boardLink, userLink)

			if test.ExpectError {
				assert.Error(t, err)
				if test.ErrorIs != nil {
					assert.ErrorIs(t, err, test.ErrorIs)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.ExpectedKey, key)
			}
			mockRepo.AssertExpectations(t)
			mockPerm.AssertExpectations(t)
		})
	}
}

func TestGetUsersOfBoard(t *testing.T) {
	ctx := context.Background()
	userLink := uuid.New()
	boardLink := uuid.New()
	expectedUsers := []uuid.UUID{uuid.New(), uuid.New()}

	tests := []struct {
		Name        string
		MockSetup   func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService)
		Expected    []uuid.UUID
		ExpectError bool
		ErrorIs     error
	}{
		{
			Name: "success get users",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.View).Return(nil).Once()
				mockRepo.On("GetUsersOfBoard", ctx, boardLink).Return(expectedUsers, nil).Once()
			},
			Expected:    expectedUsers,
			ExpectError: false,
		},
		{
			Name: "permission denied",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.View).Return(rbac.ErrActionDenied).Once()
			},
			Expected:    nil,
			ExpectError: true,
			ErrorIs:     rbac.ErrActionDenied,
		},
		{
			Name: "repo error",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.View).Return(nil).Once()
				mockRepo.On("GetUsersOfBoard", ctx, boardLink).Return([]uuid.UUID{}, fmt.Errorf("db error")).Once()
			},
			Expected:    []uuid.UUID{},
			ExpectError: true,
		},
		{
			Name: "board not found error",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.View).Return(nil).Once()
				mockRepo.On("GetUsersOfBoard", ctx, boardLink).Return([]uuid.UUID{}, common.ErrBoardNotFound).Once()
			},
			Expected:    []uuid.UUID{},
			ExpectError: true,
			ErrorIs:     common.ErrBoardNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockRepo := new(mocks.BoardRepository)
			mockPerm := new(mockRbacService)
			test.MockSetup(mockRepo, mockPerm)

			svc := service.NewService(mockRepo, mockPerm)
			users, err := svc.GetUsersOfBoard(ctx, boardLink, userLink)

			if test.ExpectError {
				assert.Error(t, err)
				if test.ErrorIs != nil {
					assert.ErrorIs(t, err, test.ErrorIs)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.Expected, users)
			}
			mockRepo.AssertExpectations(t)
			mockPerm.AssertExpectations(t)
		})
	}
}
