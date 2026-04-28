package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/common"
	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/service/dto"
	mockCardRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/service/mock_card_rep"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/boardRbac"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- MOCK RBAC SERVICE ---

type MockRbacService struct {
	mock.Mock
}

func (m *MockRbacService) CheckPermissionOnBoard(ctx context.Context, itemLink uuid.UUID, userLink uuid.UUID, action rbac.Action) error {
	args := m.Called(ctx, itemLink, userLink, action)
	return args.Error(0)
}

func (m *MockRbacService) CheckPermissionOnSection(ctx context.Context, itemLink uuid.UUID, userLink uuid.UUID, action rbac.Action) error {
	args := m.Called(ctx, itemLink, userLink, action)
	return args.Error(0)
}

func (m *MockRbacService) CheckPermissionOnCard(ctx context.Context, itemLink uuid.UUID, userLink uuid.UUID, action rbac.Action) error {
	args := m.Called(ctx, itemLink, userLink, action)
	return args.Error(0)
}

func (m *MockRbacService) CheckPermissionOnComment(ctx context.Context, itemLink uuid.UUID, userLink uuid.UUID, action rbac.Action) error {
	args := m.Called(ctx, itemLink, userLink, action)
	return args.Error(0)
}

// --- TESTS ---

func TestGetCard(t *testing.T) {
	targetCardLink := uuid.New()
	targetUserLink := uuid.New()
	targetExecuterName := "John Doe"
	targetDataDeadLine := time.Now()

	repResponse := repositoryDto.InfoCard{
		Title:        "Title",
		Description:  "Desc",
		NameExecutor: &targetExecuterName,
		DataDeadLine: &targetDataDeadLine,
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockCardRep.CardRepository, r *MockRbacService)
		expectedError error
		expectedRes   dto.InfoCard
	}{
		{
			nameTest: "Success get card",
			mockBehavior: func(m *mockCardRep.CardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(nil)
				m.On("GetCard", mock.Anything, targetCardLink).Return(repResponse, nil)
			},
			expectedError: nil,
			expectedRes: dto.InfoCard{
				Title:        "Title",
				Description:  "Desc",
				NameExecutor: &targetExecuterName,
				DataDeadLine: &targetDataDeadLine,
			},
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *mockCardRep.CardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
			expectedRes:   dto.InfoCard{},
		},
		{
			nameTest: "Error from repository",
			mockBehavior: func(m *mockCardRep.CardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(nil)
				m.On("GetCard", mock.Anything, targetCardLink).Return(repositoryDto.InfoCard{}, errors.New("db error"))
			},
			expectedError: errors.New("rep.GetCard: db error"),
			expectedRes:   dto.InfoCard{},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := mockCardRep.NewCardRepository(t)
			mockPerm := new(MockRbacService)
			test.mockBehavior(mockRep, mockPerm)

			service := NewService(mockRep, mockPerm)
			res, err := service.GetCard(context.Background(), targetCardLink, targetUserLink)

			if test.expectedError != nil {
				assert.ErrorContains(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedRes, res)
			}
		})
	}
}

func TestDeleteCard(t *testing.T) {
	targetCardLink := uuid.New()
	targetUserLink := uuid.New()

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockCardRep.CardRepository, r *MockRbacService)
		expectedError error
	}{
		{
			nameTest: "Success delete card",
			mockBehavior: func(m *mockCardRep.CardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(nil)
				m.On("DeleteCard", mock.Anything, targetCardLink).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *mockCardRep.CardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := mockCardRep.NewCardRepository(t)
			mockPerm := new(MockRbacService)
			test.mockBehavior(mockRep, mockPerm)

			service := NewService(mockRep, mockPerm)
			err := service.DeleteCard(context.Background(), targetCardLink, targetUserLink)

			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateCardDetails(t *testing.T) {
	targetCardLink := uuid.New()
	targetUserLink := uuid.New()
	targetExecuterLink := uuid.New()
	targetDataDeadLine := time.Now()

	updateDto := dto.UpdatingCardDetails{
		LinkCard:     targetCardLink,
		Title:        "Upd Title",
		Description:  "Upd Desc",
		LinkExecutor: &targetExecuterLink,
		DataDeadLine: &targetDataDeadLine,
	}

	repUpdateDto := repositoryDto.UpdatingCardDetails{
		LinkCard:     targetCardLink,
		Title:        "Upd Title",
		Description:  "Upd Desc",
		LinkExecutor: &targetExecuterLink,
		DataDeadLine: &targetDataDeadLine,
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockCardRep.CardRepository, r *MockRbacService)
		expectedError error
	}{
		{
			nameTest: "Success update card",
			mockBehavior: func(m *mockCardRep.CardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(nil)
				m.On("UpdateCardDetails", mock.Anything, repUpdateDto).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *mockCardRep.CardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := mockCardRep.NewCardRepository(t)
			mockPerm := new(MockRbacService)
			test.mockBehavior(mockRep, mockPerm)

			service := NewService(mockRep, mockPerm)
			err := service.UpdateCardDetails(context.Background(), updateDto, targetUserLink)

			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReorderCard(t *testing.T) {
	targetCardLink := uuid.New()
	targetSectionLink := uuid.New()
	targetUserLink := uuid.New()

	placeDto := dto.PlaceCard{
		LinkCard:    targetCardLink,
		LinkSection: targetSectionLink,
		Position:    2,
	}

	repPlaceDto := repositoryDto.PlaceCard{
		LinkCard:    targetCardLink,
		LinkSection: targetSectionLink,
		Position:    2,
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockCardRep.CardRepository, r *MockRbacService)
		expectedError error
	}{
		{
			nameTest: "Success reorder card",
			mockBehavior: func(m *mockCardRep.CardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSection", mock.Anything, targetSectionLink, targetUserLink, mock.Anything).Return(nil)
				m.On("ReorderCard", mock.Anything, repPlaceDto).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *mockCardRep.CardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSection", mock.Anything, targetSectionLink, targetUserLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := mockCardRep.NewCardRepository(t)
			mockPerm := new(MockRbacService)
			test.mockBehavior(mockRep, mockPerm)

			service := NewService(mockRep, mockPerm)
			err := service.ReorderCard(context.Background(), placeDto, targetUserLink)

			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateCard(t *testing.T) {
	targetAuthorLink := uuid.New()
	targetSectionLink := uuid.New()
	targetExecuterLink := uuid.New()
	targetDataDeadLine := time.Now()

	newCardDto := dto.NewCard{
		LinkAuthor:   targetAuthorLink,
		LinkSection:  targetSectionLink,
		Title:        "Title",
		Description:  "Desc",
		LinkExecutor: &targetExecuterLink,
		DataDeadLine: &targetDataDeadLine,
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockCardRep.CardRepository, r *MockRbacService)
		expectedError error
	}{
		{
			nameTest: "Success create card",
			mockBehavior: func(m *mockCardRep.CardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSection", mock.Anything, targetSectionLink, targetAuthorLink, mock.Anything).Return(nil)
				m.On("CreateCard", mock.Anything, mock.AnythingOfType("dto.NewCard")).Return(5, nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *mockCardRep.CardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSection", mock.Anything, targetSectionLink, targetAuthorLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := mockCardRep.NewCardRepository(t)
			mockPerm := new(MockRbacService)
			test.mockBehavior(mockRep, mockPerm)

			service := NewService(mockRep, mockPerm)
			res, err := service.CreateCard(context.Background(), newCardDto)

			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 5, res.Position)
				assert.Equal(t, targetSectionLink, res.LinkSection)
				assert.NotEqual(t, uuid.Nil, res.LinkCard)
			}
		})
	}
}

func TestGetComments(t *testing.T) {
	targetCardLink := uuid.New()
	targetUserLink := uuid.New()
	commentLink := uuid.New()
	authorLink := uuid.New()
	parentLink := uuid.New()

	repComments := []repositoryDto.CommentInfo{
		{Link: commentLink, ParentLink: &parentLink, AuthorLink: authorLink, Text: "hello"},
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockCardRep.CardRepository, r *MockRbacService)
		expectedError error
		expectedLen   int
	}{
		{
			nameTest: "Success get comments",
			mockBehavior: func(m *mockCardRep.CardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(nil)
				m.On("GetComments", mock.Anything, targetCardLink).Return(repComments, nil)
			},
			expectedError: nil,
			expectedLen:   1,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *mockCardRep.CardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
			expectedLen:   0,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := mockCardRep.NewCardRepository(t)
			mockPerm := new(MockRbacService)
			test.mockBehavior(mockRep, mockPerm)

			service := NewService(mockRep, mockPerm)
			res, err := service.GetComments(context.Background(), targetCardLink, targetUserLink)

			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, test.expectedLen)
			}
		})
	}
}

func TestCreateComment(t *testing.T) {
	targetCardLink := uuid.New()
	targetAuthorLink := uuid.New()
	targetCommentLink := uuid.New()

	createDto := dto.CreateCommentInfo{
		CardLink:   targetCardLink,
		AuthorLink: targetAuthorLink,
		Text:       "test comment",
	}

	repResult := repositoryDto.CommentInfo{
		Link:       targetCommentLink,
		AuthorLink: targetAuthorLink,
		Text:       "test comment",
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockCardRep.CardRepository, r *MockRbacService)
		expectedError error
	}{
		{
			nameTest: "Success create comment",
			mockBehavior: func(m *mockCardRep.CardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetAuthorLink, mock.Anything).Return(nil)
				m.On("CreateComment", mock.Anything, mock.Anything).Return(repResult, nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *mockCardRep.CardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetAuthorLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := mockCardRep.NewCardRepository(t)
			mockPerm := new(MockRbacService)
			test.mockBehavior(mockRep, mockPerm)

			service := NewService(mockRep, mockPerm)
			res, err := service.CreateComment(context.Background(), createDto)

			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, targetCommentLink, res.Link)
				assert.Equal(t, "test comment", res.Text)
			}
		})
	}
}

func TestDeleteComment(t *testing.T) {
	targetCommentLink := uuid.New()
	targetUserLink := uuid.New()

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockCardRep.CardRepository, r *MockRbacService)
		expectedError error
	}{
		{
			nameTest: "Success delete comment",
			mockBehavior: func(m *mockCardRep.CardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnComment", mock.Anything, targetCommentLink, targetUserLink, mock.Anything).Return(nil)
				m.On("IsCommentAuthor", mock.Anything, targetCommentLink, targetUserLink).Return(true, nil)
				m.On("DeleteComment", mock.Anything, targetCommentLink).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error not comment author",
			mockBehavior: func(m *mockCardRep.CardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnComment", mock.Anything, targetCommentLink, targetUserLink, mock.Anything).Return(nil)
				m.On("IsCommentAuthor", mock.Anything, targetCommentLink, targetUserLink).Return(false, nil)
			},
			expectedError: common.ErrPermissionDenied,
		},
		{
			nameTest: "Error comment not found",
			mockBehavior: func(m *mockCardRep.CardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnComment", mock.Anything, targetCommentLink, targetUserLink, mock.Anything).Return(nil)
				m.On("IsCommentAuthor", mock.Anything, targetCommentLink, targetUserLink).Return(false, common.ErrCommentNotFound)
			},
			expectedError: common.ErrCommentNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := mockCardRep.NewCardRepository(t)
			mockPerm := new(MockRbacService)
			test.mockBehavior(mockRep, mockPerm)

			service := NewService(mockRep, mockPerm)
			err := service.DeleteComment(context.Background(), targetCommentLink, targetUserLink)

			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateComment(t *testing.T) {
	targetCommentLink := uuid.New()
	targetUserLink := uuid.New()

	updateDto := dto.UpdateCommentInfo{
		CommentLink: targetCommentLink,
		UserLink:    targetUserLink,
		Text:        "updated text",
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockCardRep.CardRepository, r *MockRbacService)
		expectedError error
	}{
		{
			nameTest: "Success update comment",
			mockBehavior: func(m *mockCardRep.CardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnComment", mock.Anything, targetCommentLink, targetUserLink, mock.Anything).Return(nil)
				m.On("IsCommentAuthor", mock.Anything, targetCommentLink, targetUserLink).Return(true, nil)
				m.On("UpdateComment", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error not comment author",
			mockBehavior: func(m *mockCardRep.CardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnComment", mock.Anything, targetCommentLink, targetUserLink, mock.Anything).Return(nil)
				m.On("IsCommentAuthor", mock.Anything, targetCommentLink, targetUserLink).Return(false, nil)
			},
			expectedError: common.ErrPermissionDenied,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *mockCardRep.CardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnComment", mock.Anything, targetCommentLink, targetUserLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := mockCardRep.NewCardRepository(t)
			mockPerm := new(MockRbacService)
			test.mockBehavior(mockRep, mockPerm)

			service := NewService(mockRep, mockPerm)
			err := service.UpdateComment(context.Background(), updateDto)

			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
