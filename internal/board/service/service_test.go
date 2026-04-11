package service_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/common"
	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/service"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/service/dto"
	mocks "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/service/mock_board_rep"
)

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
			test.MockSetup(mockRepo)

			svc := service.NewService(mockRepo)

			boards, err := svc.GetBoards(ctx, test.UserLink)

			if test.ExpectError {
				assert.Error(t, err)
				assert.Equal(t, []dto.BoardInfo{}, boards)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.ExpectedBoards, boards)
			}

			mockRepo.AssertExpectations(t)
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
			test.MockSetup(mockRepo)

			svc := service.NewService(mockRepo)

			board, err := svc.CreateBoard(ctx, test.BoardInfo, test.AuthorLink)

			if test.ExpectError {
				assert.Error(t, err)
				assert.Equal(t, dto.BoardInfo{}, board)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.ExpectedBoard, board)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeleteBoard(t *testing.T) {
	boardLink := uuid.New()
	userLink := uuid.New()
	ctx := context.Background()

	tests := []struct {
		Name        string
		MockSetup   func(mockRepo *mocks.BoardRepository)
		ExpectError error
	}{
		{
			Name: "success delete board",
			MockSetup: func(mockRepo *mocks.BoardRepository) {
				mockRepo.On("GetUserRoleOnBoard", ctx, userLink, boardLink).Return(common.Roles.Creator, nil).Once()
				mockRepo.On("DeleteBoard", ctx, boardLink).Return(nil).Once()
			},
			ExpectError: nil,
		},
		{
			Name: "error permission denied",
			MockSetup: func(mockRepo *mocks.BoardRepository) {
				mockRepo.On("GetUserRoleOnBoard", ctx, userLink, boardLink).Return(common.Roles.Viewer, nil).Once()
			},
			ExpectError: common.ErrActionDenied,
		},
		{
			Name: "error on delete board",
			MockSetup: func(mockRepo *mocks.BoardRepository) {
				mockRepo.On("GetUserRoleOnBoard", ctx, userLink, boardLink).Return(common.Roles.Creator, nil).Once()
				mockRepo.On("DeleteBoard", ctx, boardLink).Return(fmt.Errorf("db error")).Once()
			},
			ExpectError: fmt.Errorf("db error"),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockRepo := new(mocks.BoardRepository)
			test.MockSetup(mockRepo)

			svc := service.NewService(mockRepo)
			err := svc.DeleteBoard(ctx, boardLink, userLink)

			if test.ExpectError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
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
		MockSetup   func(mockRepo *mocks.BoardRepository)
		ExpectError bool
	}{
		{
			Name: "success update",
			MockSetup: func(mockRepo *mocks.BoardRepository) {
				mockRepo.On("GetUserRoleOnBoard", ctx, userLink, boardInfo.Link).Return(common.Roles.Editor, nil).Once()
				mockRepo.On("UpdateBoard", ctx, repoBoardInfo).Return(nil).Once()
			},
			ExpectError: false,
		},
		{
			Name: "access denied",
			MockSetup: func(mockRepo *mocks.BoardRepository) {
				mockRepo.On("GetUserRoleOnBoard", ctx, userLink, boardInfo.Link).Return(common.Roles.Viewer, nil).Once()
			},
			ExpectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockRepo := new(mocks.BoardRepository)
			test.MockSetup(mockRepo)

			svc := service.NewService(mockRepo)
			err := svc.UpdateBoard(ctx, boardInfo, userLink)

			if test.ExpectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCheckPermission(t *testing.T) {
	ctx := context.Background()
	uLink := uuid.New()
	bLink := uuid.New()

	tests := []struct {
		Name      string
		Action    common.Action
		MockSetup func(m *mocks.BoardRepository)
		WantErr   error
	}{
		{
			Name:   "allowed action",
			Action: common.Actions.Edit,
			MockSetup: func(m *mocks.BoardRepository) {
				m.On("GetUserRoleOnBoard", ctx, uLink, bLink).Return(common.Roles.Editor, nil).Once()
			},
			WantErr: nil,
		},
		{
			Name:   "denied action",
			Action: common.Actions.Delete,
			MockSetup: func(m *mocks.BoardRepository) {
				m.On("GetUserRoleOnBoard", ctx, uLink, bLink).Return(common.Roles.Editor, nil).Once()
			},
			WantErr: common.ErrActionDenied,
		},
		{
			Name:   "repo error",
			Action: common.Actions.Edit,
			MockSetup: func(m *mocks.BoardRepository) {
				m.On("GetUserRoleOnBoard", ctx, uLink, bLink).Return(common.Role(""), fmt.Errorf("fail")).Once()
			},
			WantErr: fmt.Errorf("fail"),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockRepo := new(mocks.BoardRepository)
			test.MockSetup(mockRepo)
			svc := service.NewService(mockRepo)

			err := svc.CheckPermission(ctx, bLink, uLink, test.Action)
			if test.WantErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
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
		MockSetup     func(mockRepo *mocks.BoardRepository)
		ExpectedBoard dto.BoardInfo
		ExpectError   bool
		ErrorIs       error
	}{
		{
			Name: "success get board",
			MockSetup: func(mockRepo *mocks.BoardRepository) {
				mockRepo.On("GetUserRoleOnBoard", ctx, userLink, boardLink).Return(common.Roles.Viewer, nil).Once()
				mockRepo.On("GetBoard", ctx, boardLink).Return(repoEntry, nil).Once()
			},
			ExpectedBoard: expectedBoard,
			ExpectError:   false,
		},
		{
			Name: "permission denied",
			MockSetup: func(mockRepo *mocks.BoardRepository) {
				mockRepo.On("GetUserRoleOnBoard", ctx, userLink, boardLink).Return(common.Role("none"), nil).Once()
			},
			ExpectedBoard: dto.BoardInfo{},
			ExpectError:   true,
			ErrorIs:       common.ErrActionDenied,
		},
		{
			Name: "board not found",
			MockSetup: func(mockRepo *mocks.BoardRepository) {
				mockRepo.On("GetUserRoleOnBoard", ctx, userLink, boardLink).Return(common.Roles.Viewer, nil).Once()
				mockRepo.On("GetBoard", ctx, boardLink).Return(repositoryDto.BoardEntry{}, common.ErrBoardNotFound).Once()
			},
			ExpectedBoard: dto.BoardInfo{},
			ExpectError:   true,
			ErrorIs:       common.ErrBoardNotFound,
		},
		{
			Name: "repo get board error",
			MockSetup: func(mockRepo *mocks.BoardRepository) {
				mockRepo.On("GetUserRoleOnBoard", ctx, userLink, boardLink).Return(common.Roles.Viewer, nil).Once()
				mockRepo.On("GetBoard", ctx, boardLink).Return(repositoryDto.BoardEntry{}, fmt.Errorf("db error")).Once()
			},
			ExpectedBoard: dto.BoardInfo{},
			ExpectError:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockRepo := new(mocks.BoardRepository)
			test.MockSetup(mockRepo)

			svc := service.NewService(mockRepo)
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
		MockSetup   func(mockRepo *mocks.BoardRepository)
		ExpectedKey string
		ExpectError bool
		ErrorIs     error
	}{
		{
			Name: "success update background",
			MockSetup: func(mockRepo *mocks.BoardRepository) {
				mockRepo.On("GetUserRoleOnBoard", ctx, userLink, boardLink).Return(common.Roles.Editor, nil).Once()
				mockRepo.On("UploadBackground", ctx, file, mock.AnythingOfType("string"), contentType).Return(expectedKey, nil).Once()
				mockRepo.On("UpdateBackground", ctx, expectedKey, boardLink).Return(nil).Once()
			},
			ExpectedKey: expectedKey,
			ExpectError: false,
		},
		{
			Name: "permission denied",
			MockSetup: func(mockRepo *mocks.BoardRepository) {
				mockRepo.On("GetUserRoleOnBoard", ctx, userLink, boardLink).Return(common.Roles.Viewer, nil).Once()
			},
			ExpectedKey: "",
			ExpectError: true,
			ErrorIs:     common.ErrActionDenied,
		},
		{
			Name: "upload failed",
			MockSetup: func(mockRepo *mocks.BoardRepository) {
				mockRepo.On("GetUserRoleOnBoard", ctx, userLink, boardLink).Return(common.Roles.Editor, nil).Once()
				mockRepo.On("UploadBackground", ctx, file, mock.AnythingOfType("string"), contentType).Return("", fmt.Errorf("s3 timeout")).Once()
			},
			ExpectedKey: "",
			ExpectError: true,
		},
		{
			Name: "update board background db error (not found)",
			MockSetup: func(mockRepo *mocks.BoardRepository) {
				mockRepo.On("GetUserRoleOnBoard", ctx, userLink, boardLink).Return(common.Roles.Editor, nil).Once()
				mockRepo.On("UploadBackground", ctx, file, mock.AnythingOfType("string"), contentType).Return(expectedKey, nil).Once()
				mockRepo.On("UpdateBackground", ctx, expectedKey, boardLink).Return(common.ErrBoardNotFound).Once()
			},
			ExpectedKey: "",
			ExpectError: true,
			ErrorIs:     common.ErrBoardNotFound,
		},
		{
			Name: "update board background db error (generic)",
			MockSetup: func(mockRepo *mocks.BoardRepository) {
				mockRepo.On("GetUserRoleOnBoard", ctx, userLink, boardLink).Return(common.Roles.Editor, nil).Once()
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
			test.MockSetup(mockRepo)

			svc := service.NewService(mockRepo)
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
		})
	}
}

func TestGetUsersOfBoard(t *testing.T) {
	ctx := context.Background()
	boardLink := uuid.New()
	expectedUsers := []uuid.UUID{uuid.New(), uuid.New()}

	tests := []struct {
		Name        string
		MockSetup   func(mockRepo *mocks.BoardRepository)
		Expected    []uuid.UUID
		ExpectError bool
		ErrorIs     error
	}{
		{
			Name: "success get users",
			MockSetup: func(mockRepo *mocks.BoardRepository) {
				mockRepo.On("GetUsersOfBoard", ctx, boardLink).Return(expectedUsers, nil).Once()
			},
			Expected:    expectedUsers,
			ExpectError: false,
		},
		{
			Name: "repo error",
			MockSetup: func(mockRepo *mocks.BoardRepository) {
				mockRepo.On("GetUsersOfBoard", ctx, boardLink).Return([]uuid.UUID{}, fmt.Errorf("db error")).Once()
			},
			Expected:    []uuid.UUID{},
			ExpectError: true,
		},
		{
			Name: "board not found error",
			MockSetup: func(mockRepo *mocks.BoardRepository) {
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
			test.MockSetup(mockRepo)

			svc := service.NewService(mockRepo)
			users, err := svc.GetUsersOfBoard(ctx, boardLink)

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
		})
	}
}
