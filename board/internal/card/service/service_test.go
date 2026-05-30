package service

import (
	"context"
	"errors"
	"testing"
	"time"

	boardCommon "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/common"
	boardService "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/service"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/models"
	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/service/dto"
	mockCardRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/service/mock_card_rep"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/boardRbac"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/brokerEvents"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/pubsub"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func (m *MockRbacService) CheckPermissionOnSubtask(ctx context.Context, itemLink uuid.UUID, userLink uuid.UUID, action rbac.Action) error {
	args := m.Called(ctx, itemLink, userLink, action)
	return args.Error(0)
}

func (m *MockRbacService) CheckPermissionOnAttachment(ctx context.Context, itemLink uuid.UUID, userLink uuid.UUID, action rbac.Action) error {
	args := m.Called(ctx, itemLink, userLink, action)
	return args.Error(0)
}

func (m *MockRbacService) InvalidateUserBoardRole(ctx context.Context, userLink, boardLink uuid.UUID) error {
	args := m.Called(ctx, userLink, boardLink)
	return args.Error(0)
}

type mockPublisher struct {
	mock.Mock
}

func (m *mockPublisher) Publish(ctx context.Context, channel pubsub.Channel, event pubsub.Event[brokerEvents.BoardUpdateEvent]) (pubsub.ID, error) {
	args := m.Called(ctx, channel, event)
	var r0 pubsub.ID
	if args.Get(0) != nil {
		r0 = args.Get(0).(pubsub.ID)
	}
	return r0, args.Error(1)
}

type testCardRepository struct {
	*mockCardRep.CardRepository
}

func (m *testCardRepository) CreateSubtask(ctx context.Context, createInfo repositoryDto.CreateSubtaskInfo) (models.SubtaskInfo, error) {
	args := m.Called(ctx, createInfo)
	return args.Get(0).(models.SubtaskInfo), args.Error(1)
}

func (m *testCardRepository) DeleteSubtask(ctx context.Context, deleteInfo repositoryDto.DeleteSubtask) error {
	args := m.Called(ctx, deleteInfo)
	return args.Error(0)
}

func (m *testCardRepository) UpdateSubtask(ctx context.Context, updateInfo repositoryDto.UpdateSubtask) error {
	args := m.Called(ctx, updateInfo)
	return args.Error(0)
}

func (m *testCardRepository) UploadAttachment(ctx context.Context, uploadInfo repositoryDto.UploadAttachment) (string, error) {
	args := m.Called(ctx, uploadInfo)
	return args.String(0), args.Error(1)
}

func (m *testCardRepository) CreateAttachment(ctx context.Context, createInfo repositoryDto.CreateAttachment) (models.AttachmentInfo, error) {
	args := m.Called(ctx, createInfo)
	return args.Get(0).(models.AttachmentInfo), args.Error(1)
}

func (m *testCardRepository) DeleteAttachmentFromDB(ctx context.Context, attachmentLink uuid.UUID) (string, error) {
	args := m.Called(ctx, attachmentLink)
	return args.String(0), args.Error(1)
}

func (m *testCardRepository) DeleteAttachmentFromS3(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *testCardRepository) UpdateStatusTask(ctx context.Context, updateInfo repositoryDto.UpdateStatusTask) error {
	args := m.Called(ctx, updateInfo)
	return args.Error(0)
}

func (m *testCardRepository) UpdateTimeLine(ctx context.Context, updateInfo repositoryDto.UpdateTimeLine) error {
	args := m.Called(ctx, updateInfo)
	return args.Error(0)
}

func (m *testCardRepository) UpdateCardPoints(ctx context.Context, dto repositoryDto.UpdateCardPoints) error {
	args := m.Called(ctx, dto)
	return args.Error(0)
}

func newTestCardRepository(t *testing.T) *testCardRepository {
	return &testCardRepository{mockCardRep.NewCardRepository(t)}
}

func TestGetCard(t *testing.T) {
	targetCardLink := uuid.New()
	targetUserLink := uuid.New()
	targetExecutorLink := uuid.New()
	targetDataDeadLine := time.Now()
	targetAttachmentLink := uuid.New()
	targetAttachmentLink2 := uuid.New()

	repResponse := repositoryDto.InfoCard{
		Title:        "Title",
		Description:  "Desc",
		ExecutorLink: &targetExecutorLink,
		DataDeadLine: &targetDataDeadLine,
	}

	tests := []struct {
		nameTest      string
		cfg           Config
		mockBehavior  func(m *testCardRepository, r *MockRbacService)
		expectedError error
		expectedRes   dto.InfoCard
	}{
		{
			nameTest: "Success get card",
			cfg:      Config{},
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(nil)
				m.On("GetCard", mock.Anything, targetCardLink).Return(repResponse, nil)
			},
			expectedError: nil,
			expectedRes: dto.InfoCard{
				Title:        "Title",
				Description:  "Desc",
				ExecutorLink: &targetExecutorLink,
				DataDeadLine: &targetDataDeadLine,
			},
		},
		{
			nameTest: "Error permission denied",
			cfg:      Config{},
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
			expectedRes:   dto.InfoCard{},
		},
		{
			nameTest: "Error from repository",
			cfg:      Config{},
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(nil)
				m.On("GetCard", mock.Anything, targetCardLink).Return(repositoryDto.InfoCard{}, errors.New("db error"))
			},
			expectedError: errors.New("rep.GetCard: db error"),
			expectedRes:   dto.InfoCard{},
		},
		{
			nameTest: "Success get card with attachments with correct full URL",
			cfg:      Config{BaseURLAttachment: "https://bucket.endpoint"},
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(nil)
				m.On("GetCard", mock.Anything, targetCardLink).Return(repositoryDto.InfoCard{
					Title:        "Title",
					Description:  "Desc",
					ExecutorLink: &targetExecutorLink,
					DataDeadLine: &targetDataDeadLine,
					Attachments: []models.AttachmentInfo{
						{AttachmentLink: targetAttachmentLink, Path: "cards-attachments/photo.png", Name: "photo.png", Position: 1},
					},
				}, nil)
			},
			expectedError: nil,
			expectedRes: dto.InfoCard{
				Title:        "Title",
				Description:  "Desc",
				ExecutorLink: &targetExecutorLink,
				DataDeadLine: &targetDataDeadLine,
				Attachments: []models.AttachmentInfo{
					{AttachmentLink: targetAttachmentLink, Path: "https://bucket.endpoint/cards-attachments/photo.png", Name: "photo.png", Position: 1},
				},
			},
		},
		{
			nameTest: "Success get card with multiple attachments all with full URL",
			cfg:      Config{BaseURLAttachment: "https://bucket.endpoint"},
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(nil)
				m.On("GetCard", mock.Anything, targetCardLink).Return(repositoryDto.InfoCard{
					Title:        "Title",
					Description:  "Desc",
					ExecutorLink: &targetExecutorLink,
					DataDeadLine: &targetDataDeadLine,
					Attachments: []models.AttachmentInfo{
						{AttachmentLink: targetAttachmentLink, Path: "cards-attachments/doc.pdf", Name: "doc.pdf", Position: 1},
						{AttachmentLink: targetAttachmentLink2, Path: "cards-attachments/img.png", Name: "img.png", Position: 2},
					},
				}, nil)
			},
			expectedError: nil,
			expectedRes: dto.InfoCard{
				Title:        "Title",
				Description:  "Desc",
				ExecutorLink: &targetExecutorLink,
				DataDeadLine: &targetDataDeadLine,
				Attachments: []models.AttachmentInfo{
					{AttachmentLink: targetAttachmentLink, Path: "https://bucket.endpoint/cards-attachments/doc.pdf", Name: "doc.pdf", Position: 1},
					{AttachmentLink: targetAttachmentLink2, Path: "https://bucket.endpoint/cards-attachments/img.png", Name: "img.png", Position: 2},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := newTestCardRepository(t)
			mockPerm := new(MockRbacService)
			test.mockBehavior(mockRep, mockPerm)

			service := NewService(mockRep, mockPerm, nil, nil, test.cfg)
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
		mockBehavior  func(m *testCardRepository, r *MockRbacService)
		expectedError error
	}{
		{
			nameTest: "Success delete card",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(nil)
				m.On("DeleteCard", mock.Anything, targetCardLink).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := newTestCardRepository(t)
			mockPerm := new(MockRbacService)
			test.mockBehavior(mockRep, mockPerm)

			service := NewService(mockRep, mockPerm, nil, nil, Config{})
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
		mockBehavior  func(m *testCardRepository, r *MockRbacService)
		expectedError error
	}{
		{
			nameTest: "Success update card",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(nil)
				m.On("UpdateCardDetails", mock.Anything, repUpdateDto).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := newTestCardRepository(t)
			mockPerm := new(MockRbacService)
			test.mockBehavior(mockRep, mockPerm)

			service := NewService(mockRep, mockPerm, nil, nil, Config{})
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
		mockBehavior  func(m *testCardRepository, r *MockRbacService)
		expectedError error
	}{
		{
			nameTest: "Success reorder card",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSection", mock.Anything, targetSectionLink, targetUserLink, mock.Anything).Return(nil)
				m.On("ReorderCard", mock.Anything, repPlaceDto).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSection", mock.Anything, targetSectionLink, targetUserLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := newTestCardRepository(t)
			mockPerm := new(MockRbacService)
			test.mockBehavior(mockRep, mockPerm)

			service := NewService(mockRep, mockPerm, nil, nil, Config{})
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
		mockBehavior  func(m *testCardRepository, r *MockRbacService)
		expectedError error
	}{
		{
			nameTest: "Success create card",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSection", mock.Anything, targetSectionLink, targetAuthorLink, mock.Anything).Return(nil)
				m.On("CreateCard", mock.Anything, mock.AnythingOfType("dto.NewCard")).Return(5, nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSection", mock.Anything, targetSectionLink, targetAuthorLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := newTestCardRepository(t)
			mockPerm := new(MockRbacService)
			test.mockBehavior(mockRep, mockPerm)

			service := NewService(mockRep, mockPerm, nil, nil, Config{})
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
		mockBehavior  func(m *testCardRepository, r *MockRbacService)
		expectedError error
		expectedLen   int
	}{
		{
			nameTest: "Success get comments",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(nil)
				m.On("GetComments", mock.Anything, targetCardLink).Return(repComments, nil)
			},
			expectedError: nil,
			expectedLen:   1,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
			expectedLen:   0,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := newTestCardRepository(t)
			mockPerm := new(MockRbacService)
			test.mockBehavior(mockRep, mockPerm)

			service := NewService(mockRep, mockPerm, nil, nil, Config{})
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
		mockBehavior  func(m *testCardRepository, r *MockRbacService, p *mockPublisher)
		expectedError error
	}{
		{
			nameTest: "Success create comment",
			mockBehavior: func(m *testCardRepository, r *MockRbacService, p *mockPublisher) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetAuthorLink, mock.Anything).Return(nil)
				m.On("CreateComment", mock.Anything, mock.Anything).Return(repResult, nil)
				m.On("GetBoardLinkByCard", mock.Anything, targetCardLink).Return(uuid.New(), nil)
				p.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(pubsub.ID(""), nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *testCardRepository, r *MockRbacService, p *mockPublisher) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetAuthorLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := newTestCardRepository(t)
			mockPerm := new(MockRbacService)
			pub := new(mockPublisher)
			test.mockBehavior(mockRep, mockPerm, pub)

			service := NewService(mockRep, mockPerm, nil, pub, Config{})
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
		mockBehavior  func(m *testCardRepository, r *MockRbacService, p *mockPublisher)
		expectedError error
	}{
		{
			nameTest: "Success delete comment",
			mockBehavior: func(m *testCardRepository, r *MockRbacService, p *mockPublisher) {
				r.On("CheckPermissionOnComment", mock.Anything, targetCommentLink, targetUserLink, mock.Anything).Return(nil)
				m.On("IsCommentAuthor", mock.Anything, targetCommentLink, targetUserLink).Return(true, nil)
				m.On("DeleteComment", mock.Anything, targetCommentLink).Return(nil)
				m.On("GetBoardLinkByComment", mock.Anything, targetCommentLink).Return(uuid.New(), nil)
				p.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(pubsub.ID(""), nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error not comment author",
			mockBehavior: func(m *testCardRepository, r *MockRbacService, p *mockPublisher) {
				r.On("CheckPermissionOnComment", mock.Anything, targetCommentLink, targetUserLink, mock.Anything).Return(nil)
				m.On("IsCommentAuthor", mock.Anything, targetCommentLink, targetUserLink).Return(false, nil)
			},
			expectedError: common.ErrPermissionDenied,
		},
		{
			nameTest: "Error comment not found",
			mockBehavior: func(m *testCardRepository, r *MockRbacService, p *mockPublisher) {
				r.On("CheckPermissionOnComment", mock.Anything, targetCommentLink, targetUserLink, mock.Anything).Return(nil)
				m.On("IsCommentAuthor", mock.Anything, targetCommentLink, targetUserLink).Return(false, common.ErrCommentNotFound)
			},
			expectedError: common.ErrCommentNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := newTestCardRepository(t)
			mockPerm := new(MockRbacService)
			pub := new(mockPublisher)
			test.mockBehavior(mockRep, mockPerm, pub)

			service := NewService(mockRep, mockPerm, nil, pub, Config{})
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
		mockBehavior  func(m *testCardRepository, r *MockRbacService, p *mockPublisher)
		expectedError error
	}{
		{
			nameTest: "Success update comment",
			mockBehavior: func(m *testCardRepository, r *MockRbacService, p *mockPublisher) {
				r.On("CheckPermissionOnComment", mock.Anything, targetCommentLink, targetUserLink, mock.Anything).Return(nil)
				m.On("IsCommentAuthor", mock.Anything, targetCommentLink, targetUserLink).Return(true, nil)
				m.On("UpdateComment", mock.Anything, mock.Anything).Return(nil)
				m.On("GetBoardLinkByComment", mock.Anything, targetCommentLink).Return(uuid.New(), nil)
				p.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(pubsub.ID(""), nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error not comment author",
			mockBehavior: func(m *testCardRepository, r *MockRbacService, p *mockPublisher) {
				r.On("CheckPermissionOnComment", mock.Anything, targetCommentLink, targetUserLink, mock.Anything).Return(nil)
				m.On("IsCommentAuthor", mock.Anything, targetCommentLink, targetUserLink).Return(false, nil)
			},
			expectedError: common.ErrPermissionDenied,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *testCardRepository, r *MockRbacService, p *mockPublisher) {
				r.On("CheckPermissionOnComment", mock.Anything, targetCommentLink, targetUserLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
		{
			nameTest: "Error generic rbac",
			mockBehavior: func(m *testCardRepository, r *MockRbacService, p *mockPublisher) {
				r.On("CheckPermissionOnComment", mock.Anything, targetCommentLink, targetUserLink, mock.Anything).Return(errors.New("rbac fail"))
			},
			expectedError: errors.New("rbac fail"),
		},
		{
			nameTest: "Error IsCommentAuthor generic",
			mockBehavior: func(m *testCardRepository, r *MockRbacService, p *mockPublisher) {
				r.On("CheckPermissionOnComment", mock.Anything, targetCommentLink, targetUserLink, mock.Anything).Return(nil)
				m.On("IsCommentAuthor", mock.Anything, targetCommentLink, targetUserLink).Return(false, errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
		{
			nameTest: "Error update comment repo",
			mockBehavior: func(m *testCardRepository, r *MockRbacService, p *mockPublisher) {
				r.On("CheckPermissionOnComment", mock.Anything, targetCommentLink, targetUserLink, mock.Anything).Return(nil)
				m.On("IsCommentAuthor", mock.Anything, targetCommentLink, targetUserLink).Return(true, nil)
				m.On("UpdateComment", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := newTestCardRepository(t)
			mockPerm := new(MockRbacService)
			pub := new(mockPublisher)
			test.mockBehavior(mockRep, mockPerm, pub)

			service := NewService(mockRep, mockPerm, nil, pub, Config{})
			err := service.UpdateComment(context.Background(), updateDto)

			if test.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateSubtask(t *testing.T) {
	targetCardLink := uuid.New()
	targetUserLink := uuid.New()

	createDto := dto.CreateSubtaskInfo{
		TaskLink:    targetCardLink,
		Description: "Do something",
	}

	repResult := models.SubtaskInfo{
		SubtaskLink: uuid.New(),
		Description: "Do something",
		IsDone:      false,
		Position:    1,
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *testCardRepository, r *MockRbacService)
		expectedError error
	}{
		{
			nameTest: "Success create subtask",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(nil)
				m.On("CreateSubtask", mock.Anything, mock.Anything).Return(repResult, nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
		{
			nameTest: "Error generic rbac",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(errors.New("rbac fail"))
			},
			expectedError: errors.New("rbac fail"),
		},
		{
			nameTest: "Error from repository",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(nil)
				m.On("CreateSubtask", mock.Anything, mock.Anything).Return(models.SubtaskInfo{}, errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := newTestCardRepository(t)
			mockPerm := new(MockRbacService)
			test.mockBehavior(mockRep, mockPerm)

			service := NewService(mockRep, mockPerm, nil, nil, Config{})
			res, err := service.CreateSubtask(context.Background(), createDto, targetUserLink)

			if test.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, res.SubtaskLink)
			}
		})
	}
}

func TestDeleteSubtask(t *testing.T) {
	targetSubtaskLink := uuid.New()
	targetUserLink := uuid.New()

	deleteDto := dto.DeleteSubtask{SubTaskLink: targetSubtaskLink}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *testCardRepository, r *MockRbacService)
		expectedError error
	}{
		{
			nameTest: "Success delete subtask",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSubtask", mock.Anything, targetSubtaskLink, targetUserLink, mock.Anything).Return(nil)
				m.On("DeleteSubtask", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSubtask", mock.Anything, targetSubtaskLink, targetUserLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
		{
			nameTest: "Error generic rbac",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSubtask", mock.Anything, targetSubtaskLink, targetUserLink, mock.Anything).Return(errors.New("rbac fail"))
			},
			expectedError: errors.New("rbac fail"),
		},
		{
			nameTest: "Error from repository",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSubtask", mock.Anything, targetSubtaskLink, targetUserLink, mock.Anything).Return(nil)
				m.On("DeleteSubtask", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := newTestCardRepository(t)
			mockPerm := new(MockRbacService)
			test.mockBehavior(mockRep, mockPerm)

			service := NewService(mockRep, mockPerm, nil, nil, Config{})
			err := service.DeleteSubtask(context.Background(), deleteDto, targetUserLink)

			if test.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateSubtask(t *testing.T) {
	targetSubtaskLink := uuid.New()
	targetUserLink := uuid.New()

	updateDto := dto.UpdateSubtask{
		SubTaskLink: targetSubtaskLink,
		Description: "updated desc",
		IsDone:      true,
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *testCardRepository, r *MockRbacService)
		expectedError error
	}{
		{
			nameTest: "Success update subtask",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSubtask", mock.Anything, targetSubtaskLink, targetUserLink, mock.Anything).Return(nil)
				m.On("UpdateSubtask", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSubtask", mock.Anything, targetSubtaskLink, targetUserLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
		{
			nameTest: "Error generic rbac",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSubtask", mock.Anything, targetSubtaskLink, targetUserLink, mock.Anything).Return(errors.New("rbac fail"))
			},
			expectedError: errors.New("rbac fail"),
		},
		{
			nameTest: "Error from repository",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSubtask", mock.Anything, targetSubtaskLink, targetUserLink, mock.Anything).Return(nil)
				m.On("UpdateSubtask", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := newTestCardRepository(t)
			mockPerm := new(MockRbacService)
			test.mockBehavior(mockRep, mockPerm)

			service := NewService(mockRep, mockPerm, nil, nil, Config{})
			err := service.UpdateSubtask(context.Background(), updateDto, targetUserLink)

			if test.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateAttachment(t *testing.T) {
	targetTaskLink := uuid.New()
	targetUserLink := uuid.New()
	attachmentLink := uuid.New()
	s3Key := "some-uuid.png"

	createDto := dto.CreateAttachment{
		TaskLink:    targetTaskLink,
		UserLink:    targetUserLink,
		ContentType: "image/png",
		Extension:   ".png",
		DisplayName: "photo.png",
	}

	repResult := models.AttachmentInfo{
		AttachmentLink: attachmentLink,
		Path:           s3Key,
		Name:           "photo.png",
		Position:       1,
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *testCardRepository, r *MockRbacService)
		expectedError error
	}{
		{
			nameTest: "Success create attachment",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetTaskLink, targetUserLink, mock.Anything).Return(nil)
				m.On("UploadAttachment", mock.Anything, mock.Anything).Return(s3Key, nil)
				m.On("CreateAttachment", mock.Anything, mock.Anything).Return(repResult, nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetTaskLink, targetUserLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
		{
			nameTest: "Error generic rbac",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetTaskLink, targetUserLink, mock.Anything).Return(errors.New("rbac fail"))
			},
			expectedError: errors.New("rbac fail"),
		},
		{
			nameTest: "Error upload to S3",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetTaskLink, targetUserLink, mock.Anything).Return(nil)
				m.On("UploadAttachment", mock.Anything, mock.Anything).Return("", errors.New("s3 error"))
			},
			expectedError: errors.New("s3 error"),
		},
		{
			nameTest: "Error create in DB",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnCard", mock.Anything, targetTaskLink, targetUserLink, mock.Anything).Return(nil)
				m.On("UploadAttachment", mock.Anything, mock.Anything).Return(s3Key, nil)
				m.On("CreateAttachment", mock.Anything, mock.Anything).Return(models.AttachmentInfo{}, errors.New("db error"))
				m.On("DeleteAttachmentFromS3", mock.Anything, s3Key).Return(nil)
			},
			expectedError: errors.New("db error"),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := newTestCardRepository(t)
			mockPerm := new(MockRbacService)
			test.mockBehavior(mockRep, mockPerm)

			service := NewService(mockRep, mockPerm, nil, nil, Config{})
			res, err := service.CreateAttachment(context.Background(), createDto)

			if test.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, res.AttachmentLink)
				assert.Equal(t, 1, res.Position)
			}
		})
	}
}

func TestDeleteAttachment(t *testing.T) {
	targetAttachmentLink := uuid.New()
	targetUserLink := uuid.New()
	s3Key := "some-uuid.png"

	deleteDto := dto.DeleteAttachment{
		AttachmentLink: targetAttachmentLink,
		UserLink:       targetUserLink,
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *testCardRepository, r *MockRbacService)
		expectedError error
	}{
		{
			nameTest: "Success delete attachment",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnAttachment", mock.Anything, targetAttachmentLink, targetUserLink, mock.Anything).Return(nil)
				m.On("DeleteAttachmentFromDB", mock.Anything, targetAttachmentLink).Return(s3Key, nil)
				m.On("DeleteAttachmentFromS3", mock.Anything, s3Key).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnAttachment", mock.Anything, targetAttachmentLink, targetUserLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
		{
			nameTest: "Error generic rbac",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnAttachment", mock.Anything, targetAttachmentLink, targetUserLink, mock.Anything).Return(errors.New("rbac fail"))
			},
			expectedError: errors.New("rbac fail"),
		},
		{
			nameTest: "Error attachment not found in DB",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnAttachment", mock.Anything, targetAttachmentLink, targetUserLink, mock.Anything).Return(nil)
				m.On("DeleteAttachmentFromDB", mock.Anything, targetAttachmentLink).Return("", common.ErrAttachmentNotFound)
			},
			expectedError: common.ErrAttachmentNotFound,
		},
		{
			nameTest: "Error delete from S3",
			mockBehavior: func(m *testCardRepository, r *MockRbacService) {
				r.On("CheckPermissionOnAttachment", mock.Anything, targetAttachmentLink, targetUserLink, mock.Anything).Return(nil)
				m.On("DeleteAttachmentFromDB", mock.Anything, targetAttachmentLink).Return(s3Key, nil)
				m.On("DeleteAttachmentFromS3", mock.Anything, s3Key).Return(errors.New("s3 error"))
			},
			expectedError: errors.New("s3 error"),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := newTestCardRepository(t)
			mockPerm := new(MockRbacService)
			test.mockBehavior(mockRep, mockPerm)

			service := NewService(mockRep, mockPerm, nil, nil, Config{})
			err := service.DeleteAttachment(context.Background(), deleteDto)

			if test.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateCardPoints(t *testing.T) {
	targetCardLink := uuid.New()
	targetUserLink := uuid.New()
	targetBoardLink := uuid.New()
	pollAdminLink := uuid.New()
	nonAdminLink := uuid.New()
	points := 5

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *testCardRepository, perm *MockRbacService) *boardService.PollStore
		userLink      uuid.UUID
		expectedError error
	}{
		{
			nameTest: "Success_NoActivePoll",
			mockBehavior: func(m *testCardRepository, perm *MockRbacService) *boardService.PollStore {
				perm.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(nil)
				m.On("GetBoardLinkByCard", mock.Anything, targetCardLink).Return(targetBoardLink, nil)
				m.On("UpdateCardPoints", mock.Anything, repositoryDto.UpdateCardPoints{CardLink: targetCardLink, Points: &points}).Return(nil)
				return boardService.NewPollStore()
			},
			expectedError: nil,
		},
		{
			nameTest: "Success_PollAdmin",
			mockBehavior: func(m *testCardRepository, perm *MockRbacService) *boardService.PollStore {
				ps := boardService.NewPollStore()
				_ = ps.Create(targetBoardLink, pollAdminLink, []boardService.CardInfo{}, nil)
				perm.On("CheckPermissionOnCard", mock.Anything, targetCardLink, pollAdminLink, mock.Anything).Return(nil)
				m.On("GetBoardLinkByCard", mock.Anything, targetCardLink).Return(targetBoardLink, nil)
				m.On("UpdateCardPoints", mock.Anything, repositoryDto.UpdateCardPoints{CardLink: targetCardLink, Points: &points}).Return(nil)
				return ps
			},
			userLink:      pollAdminLink,
			expectedError: nil,
		},
		{
			nameTest: "Error_RbacDenied",
			mockBehavior: func(m *testCardRepository, perm *MockRbacService) *boardService.PollStore {
				perm.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(rbac.ErrActionDenied)
				return boardService.NewPollStore()
			},
			expectedError: rbac.ErrActionDenied,
		},
		{
			nameTest: "Error_NotPollAdmin",
			mockBehavior: func(m *testCardRepository, perm *MockRbacService) *boardService.PollStore {
				ps := boardService.NewPollStore()
				_ = ps.Create(targetBoardLink, pollAdminLink, []boardService.CardInfo{}, nil)
				perm.On("CheckPermissionOnCard", mock.Anything, targetCardLink, nonAdminLink, mock.Anything).Return(nil)
				m.On("GetBoardLinkByCard", mock.Anything, targetCardLink).Return(targetBoardLink, nil)
				return ps
			},
			userLink:      nonAdminLink,
			expectedError: boardCommon.ErrNotPollAdmin,
		},
		{
			nameTest: "Error_CardNotFound",
			mockBehavior: func(m *testCardRepository, perm *MockRbacService) *boardService.PollStore {
				perm.On("CheckPermissionOnCard", mock.Anything, targetCardLink, targetUserLink, mock.Anything).Return(nil)
				m.On("GetBoardLinkByCard", mock.Anything, targetCardLink).Return(uuid.Nil, common.ErrCardNotFound)
				return boardService.NewPollStore()
			},
			expectedError: common.ErrCardNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := newTestCardRepository(t)
			mockPerm := new(MockRbacService)
			pollStore := test.mockBehavior(mockRep, mockPerm)

			userLink := test.userLink
			if userLink == uuid.Nil {
				userLink = targetUserLink
			}

			service := NewService(mockRep, mockPerm, pollStore, nil, Config{})
			err := service.UpdateCardPoints(context.Background(), targetCardLink, userLink, &points)

			if test.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, test.expectedError))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
