package service_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

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

func (m *mockRbacService) CheckPermissionOnAttachment(ctx context.Context, attachmentLink uuid.UUID, userLink uuid.UUID, action rbac.Action) error {
	args := m.Called(ctx, attachmentLink, userLink, action)
	return args.Error(0)
}

func (m *mockRbacService) InvalidateUserBoardRole(ctx context.Context, userLink, boardLink uuid.UUID) error {
	args := m.Called(ctx, userLink, boardLink)
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

			svc := service.NewService(mockRepo, mockPerm, nil, nil, nil)

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

			svc := service.NewService(mockRepo, mockPerm, nil, nil, nil)

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

			svc := service.NewService(mockRepo, mockPerm, nil, nil, nil)
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

			svc := service.NewService(mockRepo, mockPerm, nil, nil, nil)
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

			svc := service.NewService(mockRepo, mockPerm, nil, nil, nil)
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

			svc := service.NewService(mockRepo, mockPerm, nil, nil, nil)
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

	member1 := uuid.New()
	member2 := uuid.New()
	repoMembers := []repositoryDto.MemberEntry{
		{Link: member1, Role: rbac.Roles.Editor},
		{Link: member2, Role: rbac.Roles.Viewer},
	}

	expectedMembers := []dto.MemberInfo{
		{Link: member1, Role: rbac.Roles.Editor},
		{Link: member2, Role: rbac.Roles.Viewer},
	}

	tests := []struct {
		Name        string
		MockSetup   func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService)
		Expected    []dto.MemberInfo
		ExpectError bool
		ErrorIs     error
	}{
		{
			Name: "success get users",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.View).Return(nil).Once()
				mockRepo.On("GetUsersOfBoard", ctx, boardLink).Return(repoMembers, nil).Once()
			},
			Expected:    expectedMembers,
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
				mockRepo.On("GetUsersOfBoard", ctx, boardLink).Return([]repositoryDto.MemberEntry{}, fmt.Errorf("db error")).Once()
			},
			Expected:    []dto.MemberInfo{},
			ExpectError: true,
		},
		{
			Name: "board not found error",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.View).Return(nil).Once()
				mockRepo.On("GetUsersOfBoard", ctx, boardLink).Return([]repositoryDto.MemberEntry{}, common.ErrBoardNotFound).Once()
			},
			Expected:    []dto.MemberInfo{},
			ExpectError: true,
			ErrorIs:     common.ErrBoardNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockRepo := new(mocks.BoardRepository)
			mockPerm := new(mockRbacService)
			test.MockSetup(mockRepo, mockPerm)

			svc := service.NewService(mockRepo, mockPerm, nil, nil, nil)
			members, err := svc.GetUsersOfBoard(ctx, boardLink, userLink)

			if test.ExpectError {
				assert.Error(t, err)
				if test.ErrorIs != nil {
					assert.ErrorIs(t, err, test.ErrorIs)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.Expected, members)
			}
			mockRepo.AssertExpectations(t)
			mockPerm.AssertExpectations(t)
		})
	}
}

func TestCreateInvite(t *testing.T) {
	ctx := context.Background()
	boardLink := uuid.New()
	creatorLink := uuid.New()
	now := time.Now()

	repoEntry := repositoryDto.InviteEntry{
		InviteLink:  uuid.New(),
		BoardLink:   boardLink,
		UserLink:    nil,
		DefaultRole: rbac.Roles.Editor,
		ExpireTime:  &now,
		Status:      common.InviteStatuses.Active,
		CreatedAt:   now,
	}

	inviteInfo := dto.NewInviteInfo{
		BoardLink:   boardLink,
		UserLink:    nil,
		DefaultRole: rbac.Roles.Editor,
		ExpireTime:  &now,
	}

	repoNewInfo := repositoryDto.NewInviteInfo{
		BoardLink:   boardLink,
		UserLink:    nil,
		DefaultRole: rbac.Roles.Editor,
		ExpireTime:  &now,
	}

	tests := []struct {
		Name        string
		MockSetup   func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService)
		InviteInfo  dto.NewInviteInfo
		CreatorLink uuid.UUID
		ExpectError bool
		ErrorIs     error
	}{
		{
			Name: "success create invite",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, creatorLink, rbac.Actions.Invite).Return(nil).Once()
				mockRepo.On("CreateInvite", ctx, repoNewInfo).Return(repoEntry, nil).Once()
			},
			InviteInfo:  inviteInfo,
			CreatorLink: creatorLink,
			ExpectError: false,
		},
		{
			Name: "permission denied",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, creatorLink, rbac.Actions.Invite).Return(rbac.ErrActionDenied).Once()
			},
			InviteInfo:  inviteInfo,
			CreatorLink: creatorLink,
			ExpectError: true,
			ErrorIs:     rbac.ErrActionDenied,
		},
		{
			Name: "repo error",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, creatorLink, rbac.Actions.Invite).Return(nil).Once()
				mockRepo.On("CreateInvite", ctx, repoNewInfo).Return(repositoryDto.InviteEntry{}, fmt.Errorf("db error")).Once()
			},
			InviteInfo:  inviteInfo,
			CreatorLink: creatorLink,
			ExpectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockRepo := new(mocks.BoardRepository)
			mockPerm := new(mockRbacService)
			test.MockSetup(mockRepo, mockPerm)

			svc := service.NewService(mockRepo, mockPerm, nil, nil, nil)
			_, err := svc.CreateInvite(ctx, test.InviteInfo, test.CreatorLink)

			if test.ExpectError {
				assert.Error(t, err)
				if test.ErrorIs != nil {
					assert.ErrorIs(t, err, test.ErrorIs)
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
			mockPerm.AssertExpectations(t)
		})
	}
}

func TestAcceptInvite(t *testing.T) {
	ctx := context.Background()
	inviteLink := uuid.New()
	boardLink := uuid.New()
	userLink := uuid.New()
	targetUserLink := uuid.New()

	tests := []struct {
		Name        string
		MockSetup   func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService)
		InviteLink  uuid.UUID
		UserLink    uuid.UUID
		ExpectError bool
		ErrorIs     error
	}{
		{
			Name: "success accept public invite",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockRepo.On("GetInviteByLink", ctx, inviteLink).Return(repositoryDto.InviteEntry{
					InviteLink:  inviteLink,
					BoardLink:   boardLink,
					UserLink:    nil,
					DefaultRole: rbac.Roles.Editor,
					ExpireTime:  nil,
					Status:      common.InviteStatuses.Active,
				}, nil).Once()
				mockRepo.On("AddMemberToBoard", ctx, boardLink, userLink, rbac.Roles.Editor).Return(nil).Once()
				mockPerm.On("InvalidateUserBoardRole", ctx, userLink, boardLink).Return(nil).Once()
			},
			InviteLink:  inviteLink,
			UserLink:    userLink,
			ExpectError: false,
		},
		{
			Name: "success accept personal invite and close it",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockRepo.On("GetInviteByLink", ctx, inviteLink).Return(repositoryDto.InviteEntry{
					InviteLink:  inviteLink,
					BoardLink:   boardLink,
					UserLink:    &targetUserLink,
					DefaultRole: rbac.Roles.Viewer,
					ExpireTime:  nil,
					Status:      common.InviteStatuses.Active,
				}, nil).Once()
				mockRepo.On("AddMemberToBoard", ctx, boardLink, targetUserLink, rbac.Roles.Viewer).Return(nil).Once()
				mockPerm.On("InvalidateUserBoardRole", ctx, targetUserLink, boardLink).Return(nil).Once()
				mockRepo.On("CloseInvite", ctx, inviteLink).Return(nil).Once()
			},
			InviteLink:  inviteLink,
			UserLink:    targetUserLink,
			ExpectError: false,
		},
		{
			Name: "invite not found",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockRepo.On("GetInviteByLink", ctx, inviteLink).Return(repositoryDto.InviteEntry{}, common.ErrInviteNotFound).Once()
			},
			InviteLink:  inviteLink,
			UserLink:    userLink,
			ExpectError: true,
			ErrorIs:     common.ErrInviteNotFound,
		},
		{
			Name: "invite closed",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockRepo.On("GetInviteByLink", ctx, inviteLink).Return(repositoryDto.InviteEntry{
					InviteLink:  inviteLink,
					BoardLink:   boardLink,
					DefaultRole: rbac.Roles.Editor,
					Status:      common.InviteStatuses.Closed,
				}, nil).Once()
			},
			InviteLink:  inviteLink,
			UserLink:    userLink,
			ExpectError: true,
			ErrorIs:     common.ErrInviteClosed,
		},
		{
			Name: "invite expired",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				past := time.Now().Add(-time.Hour)
				mockRepo.On("GetInviteByLink", ctx, inviteLink).Return(repositoryDto.InviteEntry{
					InviteLink:  inviteLink,
					BoardLink:   boardLink,
					DefaultRole: rbac.Roles.Editor,
					Status:      common.InviteStatuses.Active,
					ExpireTime:  &past,
				}, nil).Once()
				mockRepo.On("CloseInvite", ctx, inviteLink).Return(nil).Once()
			},
			InviteLink:  inviteLink,
			UserLink:    userLink,
			ExpectError: true,
			ErrorIs:     common.ErrInviteExpired,
		},
		{
			Name: "invite not for this user",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockRepo.On("GetInviteByLink", ctx, inviteLink).Return(repositoryDto.InviteEntry{
					InviteLink:  inviteLink,
					BoardLink:   boardLink,
					UserLink:    &targetUserLink,
					DefaultRole: rbac.Roles.Editor,
					Status:      common.InviteStatuses.Active,
				}, nil).Once()
			},
			InviteLink:  inviteLink,
			UserLink:    userLink,
			ExpectError: true,
			ErrorIs:     common.ErrInviteNotForUser,
		},
		{
			Name: "user already member",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockRepo.On("GetInviteByLink", ctx, inviteLink).Return(repositoryDto.InviteEntry{
					InviteLink:  inviteLink,
					BoardLink:   boardLink,
					UserLink:    nil,
					DefaultRole: rbac.Roles.Editor,
					Status:      common.InviteStatuses.Active,
				}, nil).Once()
				mockRepo.On("AddMemberToBoard", ctx, boardLink, userLink, rbac.Roles.Editor).Return(common.ErrUserAlreadyMember).Once()
			},
			InviteLink:  inviteLink,
			UserLink:    userLink,
			ExpectError: true,
			ErrorIs:     common.ErrUserAlreadyMember,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockRepo := new(mocks.BoardRepository)
			mockPerm := new(mockRbacService)
			test.MockSetup(mockRepo, mockPerm)

			svc := service.NewService(mockRepo, mockPerm, nil, nil, nil)
			_, err := svc.AcceptInvite(ctx, test.InviteLink, test.UserLink)

			if test.ExpectError {
				assert.Error(t, err)
				if test.ErrorIs != nil {
					assert.ErrorIs(t, err, test.ErrorIs)
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
			mockPerm.AssertExpectations(t)
		})
	}
}

func TestCloseInvite(t *testing.T) {
	ctx := context.Background()
	inviteLink := uuid.New()
	boardLink := uuid.New()
	userLink := uuid.New()

	tests := []struct {
		Name        string
		MockSetup   func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService)
		ExpectError bool
		ErrorIs     error
	}{
		{
			Name: "success close invite",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockRepo.On("GetInviteByLink", ctx, inviteLink).Return(repositoryDto.InviteEntry{
					InviteLink: inviteLink,
					BoardLink:  boardLink,
				}, nil).Once()
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.Invite).Return(nil).Once()
				mockRepo.On("CloseInvite", ctx, inviteLink).Return(nil).Once()
			},
			ExpectError: false,
		},
		{
			Name: "invite not found",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockRepo.On("GetInviteByLink", ctx, inviteLink).Return(repositoryDto.InviteEntry{}, common.ErrInviteNotFound).Once()
			},
			ExpectError: true,
			ErrorIs:     common.ErrInviteNotFound,
		},
		{
			Name: "permission denied",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockRepo.On("GetInviteByLink", ctx, inviteLink).Return(repositoryDto.InviteEntry{
					InviteLink: inviteLink,
					BoardLink:  boardLink,
				}, nil).Once()
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.Invite).Return(rbac.ErrActionDenied).Once()
			},
			ExpectError: true,
			ErrorIs:     rbac.ErrActionDenied,
		},
		{
			Name: "repo close error",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockRepo.On("GetInviteByLink", ctx, inviteLink).Return(repositoryDto.InviteEntry{
					InviteLink: inviteLink,
					BoardLink:  boardLink,
				}, nil).Once()
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.Invite).Return(nil).Once()
				mockRepo.On("CloseInvite", ctx, inviteLink).Return(fmt.Errorf("db error")).Once()
			},
			ExpectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockRepo := new(mocks.BoardRepository)
			mockPerm := new(mockRbacService)
			test.MockSetup(mockRepo, mockPerm)

			svc := service.NewService(mockRepo, mockPerm, nil, nil, nil)
			err := svc.CloseInvite(ctx, inviteLink, userLink)

			if test.ExpectError {
				assert.Error(t, err)
				if test.ErrorIs != nil {
					assert.ErrorIs(t, err, test.ErrorIs)
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
			mockPerm.AssertExpectations(t)
		})
	}
}

func TestUpdateMemberRoleService(t *testing.T) {
	ctx := context.Background()
	boardLink := uuid.New()
	userLink := uuid.New()
	callerLink := uuid.New()

	tests := []struct {
		Name        string
		MockSetup   func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService)
		ExpectError bool
		ErrorIs     error
	}{
		{
			Name: "success",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, callerLink, rbac.Actions.Invite).Return(nil).Once()
				mockRepo.On("UpdateMemberRole", ctx, boardLink, userLink, rbac.Roles.Editor).Return(nil).Once()
				mockPerm.On("InvalidateUserBoardRole", ctx, userLink, boardLink).Return(nil).Once()
			},
			ExpectError: false,
		},
		{
			Name: "permission denied",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, callerLink, rbac.Actions.Invite).Return(rbac.ErrActionDenied).Once()
			},
			ExpectError: true,
			ErrorIs:     rbac.ErrActionDenied,
		},
		{
			Name: "board not found",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, callerLink, rbac.Actions.Invite).Return(nil).Once()
				mockRepo.On("UpdateMemberRole", ctx, boardLink, userLink, rbac.Roles.Editor).Return(common.ErrBoardNotFound).Once()
			},
			ExpectError: true,
			ErrorIs:     common.ErrBoardNotFound,
		},
		{
			Name: "user not found",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, callerLink, rbac.Actions.Invite).Return(nil).Once()
				mockRepo.On("UpdateMemberRole", ctx, boardLink, userLink, rbac.Roles.Editor).Return(common.ErrUserNotFound).Once()
			},
			ExpectError: true,
			ErrorIs:     common.ErrUserNotFound,
		},
		{
			Name: "self role change",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
			},
			ExpectError: true,
			ErrorIs:     common.ErrSelfRoleChange,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockRepo := new(mocks.BoardRepository)
			mockPerm := new(mockRbacService)
			test.MockSetup(mockRepo, mockPerm)

			svc := service.NewService(mockRepo, mockPerm, nil, nil, nil)
			var err error
			if test.Name == "self role change" {
				err = svc.UpdateMemberRole(ctx, boardLink, callerLink, rbac.Roles.Editor, callerLink)
			} else {
				err = svc.UpdateMemberRole(ctx, boardLink, userLink, rbac.Roles.Editor, callerLink)
			}

			if test.ExpectError {
				assert.Error(t, err)
				if test.ErrorIs != nil {
					assert.ErrorIs(t, err, test.ErrorIs)
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
			mockPerm.AssertExpectations(t)
		})
	}
}

func TestRemoveMemberFromBoardService(t *testing.T) {
	ctx := context.Background()
	boardLink := uuid.New()
	userLink := uuid.New()
	callerLink := uuid.New()
	creatorLink := uuid.New()

	tests := []struct {
		Name        string
		MockSetup   func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService)
		BoardLink   uuid.UUID
		TargetLink  uuid.UUID
		CallerLink  uuid.UUID
		ExpectError bool
		ErrorIs     error
	}{
		{
			Name: "success remove other user",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, callerLink, rbac.Actions.Invite).Return(nil).Once()
				mockRepo.On("RemoveMemberFromBoard", ctx, boardLink, userLink).Return(nil).Once()
				mockPerm.On("InvalidateUserBoardRole", ctx, userLink, boardLink).Return(nil).Once()
			},
			BoardLink:   boardLink,
			TargetLink:  userLink,
			CallerLink:  callerLink,
			ExpectError: false,
		},
		{
			Name: "success self exit (non-creator)",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, callerLink, rbac.Actions.Delete).Return(rbac.ErrActionDenied).Once()
				mockRepo.On("RemoveMemberFromBoard", ctx, boardLink, callerLink).Return(nil).Once()
				mockPerm.On("InvalidateUserBoardRole", ctx, callerLink, boardLink).Return(nil).Once()
			},
			BoardLink:   boardLink,
			TargetLink:  callerLink,
			CallerLink:  callerLink,
			ExpectError: false,
		},
		{
			Name: "permission denied removing other",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, callerLink, rbac.Actions.Invite).Return(rbac.ErrActionDenied).Once()
			},
			BoardLink:   boardLink,
			TargetLink:  userLink,
			CallerLink:  callerLink,
			ExpectError: true,
			ErrorIs:     rbac.ErrActionDenied,
		},
		{
			Name: "board not found",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, callerLink, rbac.Actions.Invite).Return(nil).Once()
				mockRepo.On("RemoveMemberFromBoard", ctx, boardLink, userLink).Return(common.ErrBoardNotFound).Once()
			},
			BoardLink:   boardLink,
			TargetLink:  userLink,
			CallerLink:  callerLink,
			ExpectError: true,
			ErrorIs:     common.ErrBoardNotFound,
		},
		{
			Name: "user not found",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, callerLink, rbac.Actions.Invite).Return(nil).Once()
				mockRepo.On("RemoveMemberFromBoard", ctx, boardLink, userLink).Return(common.ErrUserNotFound).Once()
			},
			BoardLink:   boardLink,
			TargetLink:  userLink,
			CallerLink:  callerLink,
			ExpectError: true,
			ErrorIs:     common.ErrUserNotFound,
		},
		{
			Name: "creator cannot leave",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, creatorLink, rbac.Actions.Delete).Return(nil).Once()
			},
			BoardLink:   boardLink,
			TargetLink:  creatorLink,
			CallerLink:  creatorLink,
			ExpectError: true,
			ErrorIs:     common.ErrCreatorCannotLeave,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockRepo := new(mocks.BoardRepository)
			mockPerm := new(mockRbacService)
			test.MockSetup(mockRepo, mockPerm)

			svc := service.NewService(mockRepo, mockPerm, nil, nil, nil)
			err := svc.RemoveMemberFromBoard(ctx, test.BoardLink, test.TargetLink, test.CallerLink)

			if test.ExpectError {
				assert.Error(t, err)
				if test.ErrorIs != nil {
					assert.ErrorIs(t, err, test.ErrorIs)
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
			mockPerm.AssertExpectations(t)
		})
	}
}

func TestCanView(t *testing.T) {
	ctx := context.Background()
	boardLink := uuid.New()
	userLink := uuid.New()

	tests := []struct {
		Name        string
		MockSetup   func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService)
		ExpectError bool
		ErrorIs     error
	}{
		{
			Name: "success can view",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.View).Return(nil).Once()
			},
			ExpectError: false,
		},
		{
			Name: "permission denied",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.View).Return(rbac.ErrActionDenied).Once()
			},
			ExpectError: true,
			ErrorIs:     rbac.ErrActionDenied,
		},
		{
			Name: "permission checker error",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.View).Return(fmt.Errorf("db error")).Once()
			},
			ExpectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockRepo := new(mocks.BoardRepository)
			mockPerm := new(mockRbacService)
			test.MockSetup(mockRepo, mockPerm)

			svc := service.NewService(mockRepo, mockPerm, nil, nil, nil)
			err := svc.CanView(ctx, boardLink, userLink)

			if test.ExpectError {
				assert.Error(t, err)
				if test.ErrorIs != nil {
					assert.ErrorIs(t, err, test.ErrorIs)
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
			mockPerm.AssertExpectations(t)
		})
	}
}

func TestGetActiveInvitesService(t *testing.T) {
	ctx := context.Background()
	boardLink := uuid.New()
	userLink := uuid.New()

	inviteEntry := repositoryDto.InviteEntry{
		InviteLink:  uuid.New(),
		BoardLink:   boardLink,
		DefaultRole: rbac.Roles.Editor,
		Status:      common.InviteStatuses.Active,
	}

	tests := []struct {
		Name        string
		MockSetup   func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService)
		ExpectError bool
		ErrorIs     error
	}{
		{
			Name: "success",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.Invite).Return(nil).Once()
				mockRepo.On("GetActiveInvitesByBoard", ctx, boardLink).Return([]repositoryDto.InviteEntry{inviteEntry}, nil).Once()
			},
			ExpectError: false,
		},
		{
			Name: "permission denied",
			MockSetup: func(mockRepo *mocks.BoardRepository, mockPerm *mockRbacService) {
				mockPerm.On("CheckPermissionOnBoard", ctx, boardLink, userLink, rbac.Actions.Invite).Return(rbac.ErrActionDenied).Once()
			},
			ExpectError: true,
			ErrorIs:     rbac.ErrActionDenied,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockRepo := new(mocks.BoardRepository)
			mockPerm := new(mockRbacService)
			test.MockSetup(mockRepo, mockPerm)

			svc := service.NewService(mockRepo, mockPerm, nil, nil, nil)
			_, err := svc.GetActiveInvites(ctx, boardLink, userLink)

			if test.ExpectError {
				assert.Error(t, err)
				if test.ErrorIs != nil {
					assert.ErrorIs(t, err, test.ErrorIs)
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
			mockPerm.AssertExpectations(t)
		})
	}
}
