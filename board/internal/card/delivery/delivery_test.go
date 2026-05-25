package delivery

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/common"
	mockCardSrv "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/delivery/mock_card_srv"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/models"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/service/dto"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/boardRbac"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/card/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// testCardService extends the generated mock with methods missing from it.
type testCardService struct {
	*mockCardSrv.CardService
}

func (m *testCardService) CreateSubtask(ctx context.Context, createInfo serviceDto.CreateSubtaskInfo, userLink uuid.UUID) (models.SubtaskInfo, error) {
	args := m.Called(ctx, createInfo, userLink)
	return args.Get(0).(models.SubtaskInfo), args.Error(1)
}

func (m *testCardService) DeleteSubtask(ctx context.Context, deleteInfo serviceDto.DeleteSubtask, userLink uuid.UUID) error {
	args := m.Called(ctx, deleteInfo, userLink)
	return args.Error(0)
}

func (m *testCardService) UpdateSubtask(ctx context.Context, updateInfo serviceDto.UpdateSubtask, userLink uuid.UUID) error {
	args := m.Called(ctx, updateInfo, userLink)
	return args.Error(0)
}

func (m *testCardService) CreateAttachment(ctx context.Context, createInfo serviceDto.CreateAttachment) (serviceDto.AttachmentInfo, error) {
	args := m.Called(ctx, createInfo)
	return args.Get(0).(serviceDto.AttachmentInfo), args.Error(1)
}

func (m *testCardService) DeleteAttachment(ctx context.Context, deleteInfo serviceDto.DeleteAttachment) error {
	args := m.Called(ctx, deleteInfo)
	return args.Error(0)
}

func (m *testCardService) UpdateStatusTask(ctx context.Context, updateInfo serviceDto.UpdateStatusTask) error {
	args := m.Called(ctx, updateInfo)
	return args.Error(0)
}

func (m *testCardService) UpdateTimeLine(ctx context.Context, updateInfo serviceDto.UpdateTimeLine) error {
	args := m.Called(ctx, updateInfo)
	return args.Error(0)
}

func (m *testCardService) UpdateCardPoints(ctx context.Context, cardLink, userLink uuid.UUID, points *int) error {
	args := m.Called(ctx, cardLink, userLink, points)
	return args.Error(0)
}

func newTestCardService(t *testing.T) *testCardService {
	return &testCardService{mockCardSrv.NewCardService(t)}
}

func grpcCode(err error) codes.Code {
	return status.Code(err)
}

func grpcMsg(err error) string {
	return status.Convert(err).Message()
}

func TestGetCard(t *testing.T) {
	targetCardLink := uuid.New()
	targetUserLink := uuid.New()
	targetDeadline := time.Now().Add(24 * time.Hour)
	execLink := uuid.New()

	serviceCardInfo := serviceDto.InfoCard{
		Title:        "TestTitle",
		Description:  "Test Desc",
		ExecutorLink: &execLink,
		DataDeadLine: &targetDeadline,
	}

	tests := []struct {
		nameTest     string
		req          *pb.GetCardRequest
		mockBehavior func(m *testCardService)
		expectedCode codes.Code
		checkResp    func(t *testing.T, resp *pb.GetCardResponse)
	}{
		{
			nameTest: "Success get card",
			req:      &pb.GetCardRequest{CardLink: targetCardLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *testCardService) {
				m.On("GetCard", mock.Anything, targetCardLink, mock.Anything).Return(serviceCardInfo, nil)
			},
			expectedCode: codes.OK,
			checkResp: func(t *testing.T, resp *pb.GetCardResponse) {
				assert.NotNil(t, resp)
				assert.Equal(t, targetCardLink.String(), resp.CardInfo.Link)
				assert.Equal(t, "TestTitle", resp.CardInfo.Title)
				assert.Equal(t, "Test Desc", resp.CardInfo.Description)
				assert.Equal(t, execLink.String(), resp.CardInfo.GetExecutorLink())
			},
		},
		{
			nameTest:     "Error invalid uuid format",
			req:          &pb.GetCardRequest{CardLink: "invalid-uuid"},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error card not found",
			req:      &pb.GetCardRequest{CardLink: targetCardLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *testCardService) {
				m.On("GetCard", mock.Anything, targetCardLink, mock.Anything).Return(serviceDto.InfoCard{}, common.ErrCardNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Error internal server",
			req:      &pb.GetCardRequest{CardLink: targetCardLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *testCardService) {
				m.On("GetCard", mock.Anything, targetCardLink, mock.Anything).Return(serviceDto.InfoCard{}, errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
		{
			nameTest: "Success get card with attachments",
			req:      &pb.GetCardRequest{CardLink: targetCardLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *testCardService) {
				attLink := uuid.New()
				m.On("GetCard", mock.Anything, targetCardLink, mock.Anything).Return(serviceDto.InfoCard{
					Title:        "TestTitle",
					Description:  "Test Desc",
					ExecutorLink: &execLink,
					DataDeadLine: &targetDeadline,
					Attachments: []models.AttachmentInfo{
						{AttachmentLink: attLink, Path: "https://s3.example.com/file.png", Name: "photo.png", Position: 1},
					},
				}, nil)
			},
			expectedCode: codes.OK,
			checkResp: func(t *testing.T, resp *pb.GetCardResponse) {
				assert.NotNil(t, resp)
				assert.Len(t, resp.CardInfo.Attachments, 1)
				assert.Equal(t, "https://s3.example.com/file.png", resp.CardInfo.Attachments[0].Path)
				assert.Equal(t, "photo.png", resp.CardInfo.Attachments[0].Name)
				assert.Equal(t, int64(1), resp.CardInfo.Attachments[0].Position)
			},
		},
		{
			nameTest: "Success get card with empty attachments",
			req:      &pb.GetCardRequest{CardLink: targetCardLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *testCardService) {
				m.On("GetCard", mock.Anything, targetCardLink, mock.Anything).Return(serviceDto.InfoCard{
					Title:        "TestTitle",
					Description:  "Test Desc",
					ExecutorLink: &execLink,
					DataDeadLine: &targetDeadline,
				}, nil)
			},
			expectedCode: codes.OK,
			checkResp: func(t *testing.T, resp *pb.GetCardResponse) {
				assert.NotNil(t, resp)
				assert.Empty(t, resp.CardInfo.Attachments)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := newTestCardService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(mockService, Config{})
			resp, err := handler.GetCard(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			if test.checkResp != nil {
				test.checkResp(t, resp)
			}
		})
	}
}

func TestDeleteCard(t *testing.T) {
	targetCardLink := uuid.New()
	targetUserLink := uuid.New()

	tests := []struct {
		nameTest     string
		req          *pb.DeleteCardRequest
		mockBehavior func(m *testCardService)
		expectedCode codes.Code
	}{
		{
			nameTest: "Success delete card",
			req:      &pb.DeleteCardRequest{CardLink: targetCardLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *testCardService) {
				m.On("DeleteCard", mock.Anything, targetCardLink, mock.Anything).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			nameTest:     "Error invalid uuid format",
			req:          &pb.DeleteCardRequest{CardLink: "invalid"},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error card not found",
			req:      &pb.DeleteCardRequest{CardLink: targetCardLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *testCardService) {
				m.On("DeleteCard", mock.Anything, targetCardLink, mock.Anything).Return(common.ErrCardNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Error internal server",
			req:      &pb.DeleteCardRequest{CardLink: targetCardLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *testCardService) {
				m.On("DeleteCard", mock.Anything, targetCardLink, mock.Anything).Return(errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := newTestCardService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(mockService, Config{})
			_, err := handler.DeleteCard(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
		})
	}
}

func TestUpdateCard(t *testing.T) {
	targetCardLink := uuid.New()
	targetUserLink := uuid.New()
	targetExecutorLink := uuid.New()
	executorLinkStr := targetExecutorLink.String()
	deadline := time.Now().Add(24 * time.Hour)

	validReq := &pb.UpdateCardRequest{
		CardLink:     targetCardLink.String(),
		UserLink:     targetUserLink.String(),
		Title:        "UpdTitle",
		Description:  "Updated Desc",
		ExecutorLink: &executorLinkStr,
		Deadline:     timestamppb.New(deadline),
	}

	tests := []struct {
		nameTest     string
		req          *pb.UpdateCardRequest
		mockBehavior func(m *testCardService)
		expectedCode codes.Code
	}{
		{
			nameTest: "Success update card",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("UpdateCardDetails", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			nameTest:     "Error invalid card uuid",
			req:          &pb.UpdateCardRequest{CardLink: "invalid"},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error max len title",
			req: &pb.UpdateCardRequest{
				CardLink:    targetCardLink.String(),
				Title:       "very long title exceeding the limit",
				Description: "desc",
			},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error card not found",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("UpdateCardDetails", mock.Anything, mock.Anything, mock.Anything).Return(common.ErrCardNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Error invalid reference data",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("UpdateCardDetails", mock.Anything, mock.Anything, mock.Anything).Return(common.ErrInvalidReferenceCardData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error check violation data",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("UpdateCardDetails", mock.Anything, mock.Anything, mock.Anything).Return(common.ErrInvalidCardData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error missing required field",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("UpdateCardDetails", mock.Anything, mock.Anything, mock.Anything).Return(common.ErrMissingRequiredField)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error internal server",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("UpdateCardDetails", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := newTestCardService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(mockService, Config{
				MaxLenTitle:       10,
				MaxLenDescription: 1000,
			})
			_, err := handler.UpdateCard(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
		})
	}
}

func TestReorderCards(t *testing.T) {
	targetCardLink := uuid.New()
	targetSectionLink := uuid.New()
	targetUserLink := uuid.New()

	validReq := &pb.ReorderCardsRequest{
		CardLink:    targetCardLink.String(),
		SectionLink: targetSectionLink.String(),
		UserLink:    targetUserLink.String(),
		Position:    3,
	}

	tests := []struct {
		nameTest     string
		req          *pb.ReorderCardsRequest
		mockBehavior func(m *testCardService)
		expectedCode codes.Code
	}{
		{
			nameTest: "Success reorder card",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("ReorderCard", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			nameTest:     "Error invalid card uuid",
			req:          &pb.ReorderCardsRequest{CardLink: "invalid", SectionLink: targetSectionLink.String()},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest:     "Error invalid section uuid",
			req:          &pb.ReorderCardsRequest{CardLink: targetCardLink.String(), SectionLink: "invalid"},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error card not found",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("ReorderCard", mock.Anything, mock.Anything, mock.Anything).Return(common.ErrCardNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Error skip mandatory section",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("ReorderCard", mock.Anything, mock.Anything, mock.Anything).Return(common.ErrCannotSkipMandatorySection)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error invalid reference data",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("ReorderCard", mock.Anything, mock.Anything, mock.Anything).Return(common.ErrInvalidReferenceCardData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error check violation data",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("ReorderCard", mock.Anything, mock.Anything, mock.Anything).Return(common.ErrInvalidCardData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error missing required field",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("ReorderCard", mock.Anything, mock.Anything, mock.Anything).Return(common.ErrMissingRequiredField)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error internal server",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("ReorderCard", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := newTestCardService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(mockService, Config{})
			_, err := handler.ReorderCards(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
		})
	}
}

func TestCreateCard(t *testing.T) {
	targetSectionLink := uuid.New()
	targetAuthorLink := uuid.New()
	targetCardLink := uuid.New()

	serviceResult := serviceDto.PlaceCard{
		LinkCard:    targetCardLink,
		LinkSection: targetSectionLink,
		Position:    1,
	}

	validReq := &pb.CreateCardRequest{
		UserLink:    targetAuthorLink.String(),
		Title:       "New Task",
		SectionLink: targetSectionLink.String(),
	}

	tests := []struct {
		nameTest     string
		req          *pb.CreateCardRequest
		mockBehavior func(m *testCardService)
		expectedCode codes.Code
		checkResp    func(t *testing.T, resp *pb.CreateCardResponse)
	}{
		{
			nameTest: "Success create card",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("CreateCard", mock.Anything, mock.Anything).Return(serviceResult, nil)
			},
			expectedCode: codes.OK,
			checkResp: func(t *testing.T, resp *pb.CreateCardResponse) {
				assert.NotNil(t, resp)
				assert.Equal(t, targetCardLink.String(), resp.CardLink)
				assert.Equal(t, targetSectionLink.String(), resp.SectionLink)
				assert.Equal(t, int64(1), resp.Position)
			},
		},
		{
			nameTest:     "Error invalid author uuid",
			req:          &pb.CreateCardRequest{UserLink: "invalid", SectionLink: targetSectionLink.String()},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest:     "Error invalid section uuid",
			req:          &pb.CreateCardRequest{UserLink: targetAuthorLink.String(), SectionLink: "invalid"},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error max len title",
			req: &pb.CreateCardRequest{
				UserLink:    targetAuthorLink.String(),
				Title:       "very long title exceeding the limit",
				SectionLink: targetSectionLink.String(),
			},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error section not found",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("CreateCard", mock.Anything, mock.Anything).Return(serviceDto.PlaceCard{}, common.ErrSectionNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Error card already exists",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("CreateCard", mock.Anything, mock.Anything).Return(serviceDto.PlaceCard{}, common.ErrCardAlreadyExists)
			},
			expectedCode: codes.AlreadyExists,
		},
		{
			nameTest: "Error invalid reference data",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("CreateCard", mock.Anything, mock.Anything).Return(serviceDto.PlaceCard{}, common.ErrInvalidReferenceCardData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error check violation data",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("CreateCard", mock.Anything, mock.Anything).Return(serviceDto.PlaceCard{}, common.ErrInvalidCardData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error missing required field",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("CreateCard", mock.Anything, mock.Anything).Return(serviceDto.PlaceCard{}, common.ErrMissingRequiredField)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error internal server",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("CreateCard", mock.Anything, mock.Anything).Return(serviceDto.PlaceCard{}, errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := newTestCardService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(mockService, Config{
				MaxLenTitle:       10,
				MaxLenDescription: 1000,
			})
			resp, err := handler.CreateCard(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			if test.checkResp != nil {
				test.checkResp(t, resp)
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

	tests := []struct {
		nameTest     string
		req          *pb.GetCommentsRequest
		mockBehavior func(m *testCardService)
		expectedCode codes.Code
		checkResp    func(t *testing.T, resp *pb.GetCommentsResponse)
	}{
		{
			nameTest: "Success get comments",
			req:      &pb.GetCommentsRequest{CardLink: targetCardLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *testCardService) {
				m.On("GetComments", mock.Anything, targetCardLink, mock.Anything).Return([]serviceDto.CommentInfo{
					{Link: commentLink, ParentLink: &parentLink, AuthorLink: authorLink, Text: "hello"},
				}, nil)
			},
			expectedCode: codes.OK,
			checkResp: func(t *testing.T, resp *pb.GetCommentsResponse) {
				assert.NotNil(t, resp)
				assert.Len(t, resp.CommentsInfo, 1)
				assert.Equal(t, commentLink.String(), resp.CommentsInfo[0].CommentLink)
				assert.Equal(t, "hello", resp.CommentsInfo[0].Text)
			},
		},
		{
			nameTest: "Success empty comments",
			req:      &pb.GetCommentsRequest{CardLink: targetCardLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *testCardService) {
				m.On("GetComments", mock.Anything, targetCardLink, mock.Anything).Return([]serviceDto.CommentInfo{}, nil)
			},
			expectedCode: codes.OK,
			checkResp: func(t *testing.T, resp *pb.GetCommentsResponse) {
				assert.NotNil(t, resp)
				assert.Empty(t, resp.CommentsInfo)
			},
		},
		{
			nameTest:     "Error invalid card uuid",
			req:          &pb.GetCommentsRequest{CardLink: "invalid"},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error card not found",
			req:      &pb.GetCommentsRequest{CardLink: targetCardLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *testCardService) {
				m.On("GetComments", mock.Anything, targetCardLink, mock.Anything).Return([]serviceDto.CommentInfo{}, common.ErrCardNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Error internal server",
			req:      &pb.GetCommentsRequest{CardLink: targetCardLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *testCardService) {
				m.On("GetComments", mock.Anything, targetCardLink, mock.Anything).Return([]serviceDto.CommentInfo{}, errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := newTestCardService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(mockService, Config{})
			resp, err := handler.GetComments(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			if test.checkResp != nil {
				test.checkResp(t, resp)
			}
		})
	}
}

func TestCreateComment(t *testing.T) {
	targetCardLink := uuid.New()
	targetAuthorLink := uuid.New()
	targetParentLink := uuid.New()
	parentLinkStr := targetParentLink.String()
	newCommentLink := uuid.New()

	validReq := &pb.CreateCommentRequest{
		CardLink: targetCardLink.String(),
		UserLink: targetAuthorLink.String(),
		Text:     "test comment",
	}

	tests := []struct {
		nameTest     string
		req          *pb.CreateCommentRequest
		mockBehavior func(m *testCardService)
		expectedCode codes.Code
		checkResp    func(t *testing.T, resp *pb.CreateCommentResponse)
	}{
		{
			nameTest: "Success create comment",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("CreateComment", mock.Anything, mock.Anything).Return(serviceDto.CommentInfo{
					Link:       newCommentLink,
					AuthorLink: targetAuthorLink,
					Text:       "test comment",
				}, nil)
			},
			expectedCode: codes.OK,
			checkResp: func(t *testing.T, resp *pb.CreateCommentResponse) {
				assert.NotNil(t, resp)
				assert.Equal(t, newCommentLink.String(), resp.CommentLink)
			},
		},
		{
			nameTest:     "Error invalid card uuid",
			req:          &pb.CreateCommentRequest{CardLink: "invalid", UserLink: targetAuthorLink.String()},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest:     "Error invalid author uuid",
			req:          &pb.CreateCommentRequest{CardLink: targetCardLink.String(), UserLink: "invalid"},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error invalid parent uuid",
			req: &pb.CreateCommentRequest{
				CardLink:   targetCardLink.String(),
				UserLink:   targetAuthorLink.String(),
				ParentLink: func() *string { s := "invalid"; return &s }(),
			},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Success create comment with parent",
			req: &pb.CreateCommentRequest{
				CardLink:   targetCardLink.String(),
				UserLink:   targetAuthorLink.String(),
				ParentLink: &parentLinkStr,
				Text:       "reply",
			},
			mockBehavior: func(m *testCardService) {
				m.On("CreateComment", mock.Anything, mock.Anything).Return(serviceDto.CommentInfo{
					Link:       newCommentLink,
					ParentLink: &targetParentLink,
					AuthorLink: targetAuthorLink,
					Text:       "reply",
				}, nil)
			},
			expectedCode: codes.OK,
		},
		{
			nameTest: "Error missing required field",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("CreateComment", mock.Anything, mock.Anything).Return(serviceDto.CommentInfo{}, common.ErrMissingRequiredField)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error invalid reference data",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("CreateComment", mock.Anything, mock.Anything).Return(serviceDto.CommentInfo{}, common.ErrInvalidReferenceCardData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error internal server",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("CreateComment", mock.Anything, mock.Anything).Return(serviceDto.CommentInfo{}, errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := newTestCardService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(mockService, Config{})
			resp, err := handler.CreateComment(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			if test.checkResp != nil {
				test.checkResp(t, resp)
			}
		})
	}
}

func TestDeleteComment(t *testing.T) {
	targetCommentLink := uuid.New()
	targetUserLink := uuid.New()

	validReq := &pb.DeleteCommentRequest{
		CommentLink: targetCommentLink.String(),
		UserLink:    targetUserLink.String(),
	}

	tests := []struct {
		nameTest     string
		req          *pb.DeleteCommentRequest
		mockBehavior func(m *testCardService)
		expectedCode codes.Code
	}{
		{
			nameTest: "Success delete comment",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("DeleteComment", mock.Anything, targetCommentLink, targetUserLink).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			nameTest:     "Error invalid comment uuid",
			req:          &pb.DeleteCommentRequest{CommentLink: "invalid", UserLink: targetUserLink.String()},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest:     "Error invalid user uuid",
			req:          &pb.DeleteCommentRequest{CommentLink: targetCommentLink.String(), UserLink: "invalid"},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error comment not found",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("DeleteComment", mock.Anything, targetCommentLink, targetUserLink).Return(common.ErrCommentNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Error permission denied",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("DeleteComment", mock.Anything, targetCommentLink, targetUserLink).Return(common.ErrPermissionDenied)
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			nameTest: "Error internal server",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("DeleteComment", mock.Anything, targetCommentLink, targetUserLink).Return(errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := newTestCardService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(mockService, Config{})
			_, err := handler.DeleteComment(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
		})
	}
}

func TestUpdateComment(t *testing.T) {
	targetCommentLink := uuid.New()
	targetUserLink := uuid.New()

	validReq := &pb.UpdateCommentRequest{
		CommentLink: targetCommentLink.String(),
		UserLink:    targetUserLink.String(),
		Text:        "updated text",
	}

	tests := []struct {
		nameTest     string
		req          *pb.UpdateCommentRequest
		mockBehavior func(m *testCardService)
		expectedCode codes.Code
	}{
		{
			nameTest: "Success update comment",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("UpdateComment", mock.Anything, mock.Anything).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			nameTest:     "Error invalid comment uuid",
			req:          &pb.UpdateCommentRequest{CommentLink: "invalid", UserLink: targetUserLink.String()},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest:     "Error invalid user uuid",
			req:          &pb.UpdateCommentRequest{CommentLink: targetCommentLink.String(), UserLink: "invalid"},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error comment not found",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("UpdateComment", mock.Anything, mock.Anything).Return(common.ErrCommentNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Error permission denied",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("UpdateComment", mock.Anything, mock.Anything).Return(common.ErrPermissionDenied)
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			nameTest: "Error internal server",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("UpdateComment", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := newTestCardService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(mockService, Config{})
			_, err := handler.UpdateComment(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
		})
	}
}

func TestCreateSubtask(t *testing.T) {
	targetCardLink := uuid.New()
	targetUserLink := uuid.New()
	subtaskLink := uuid.New()

	validReq := &pb.CreateSubtaskRequest{
		CardLink:    targetCardLink.String(),
		UserLink:    targetUserLink.String(),
		Description: "test subtask",
	}

	tests := []struct {
		nameTest     string
		req          *pb.CreateSubtaskRequest
		mockBehavior func(m *testCardService)
		expectedCode codes.Code
	}{
		{
			nameTest: "Success create subtask",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("CreateSubtask", mock.Anything, mock.Anything, mock.Anything).Return(
					models.SubtaskInfo{SubtaskLink: subtaskLink, Description: "test subtask", Position: 1},
					nil,
				)
			},
			expectedCode: codes.OK,
		},
		{
			nameTest:     "Error invalid card uuid",
			req:          &pb.CreateSubtaskRequest{CardLink: "invalid", UserLink: targetUserLink.String()},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest:     "Error invalid user uuid",
			req:          &pb.CreateSubtaskRequest{CardLink: targetCardLink.String(), UserLink: "invalid"},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error missing required field",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("CreateSubtask", mock.Anything, mock.Anything, mock.Anything).Return(
					models.SubtaskInfo{}, common.ErrMissingRequiredField,
				)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error internal server",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("CreateSubtask", mock.Anything, mock.Anything, mock.Anything).Return(
					models.SubtaskInfo{}, errors.New("db error"),
				)
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := newTestCardService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(mockService, Config{})
			_, err := handler.CreateSubtask(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
		})
	}
}

func TestDeleteSubtask(t *testing.T) {
	targetSubtaskLink := uuid.New()
	targetUserLink := uuid.New()

	validReq := &pb.DeleteSubtaskRequest{
		SubtaskLink: targetSubtaskLink.String(),
		UserLink:    targetUserLink.String(),
	}

	tests := []struct {
		nameTest     string
		req          *pb.DeleteSubtaskRequest
		mockBehavior func(m *testCardService)
		expectedCode codes.Code
	}{
		{
			nameTest: "Success delete subtask",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("DeleteSubtask", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			nameTest:     "Error invalid subtask uuid",
			req:          &pb.DeleteSubtaskRequest{SubtaskLink: "invalid", UserLink: targetUserLink.String()},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest:     "Error invalid user uuid",
			req:          &pb.DeleteSubtaskRequest{SubtaskLink: targetSubtaskLink.String(), UserLink: "invalid"},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error subtask not found",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("DeleteSubtask", mock.Anything, mock.Anything, mock.Anything).Return(common.ErrSubtaskNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Error internal server",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("DeleteSubtask", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := newTestCardService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(mockService, Config{})
			_, err := handler.DeleteSubtask(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
		})
	}
}

func TestUpdateSubtask(t *testing.T) {
	targetSubtaskLink := uuid.New()
	targetUserLink := uuid.New()

	validReq := &pb.UpdateSubtaskRequest{
		SubtaskLink: targetSubtaskLink.String(),
		UserLink:    targetUserLink.String(),
		Description: "updated desc",
		IsDone:      true,
	}

	tests := []struct {
		nameTest     string
		req          *pb.UpdateSubtaskRequest
		mockBehavior func(m *testCardService)
		expectedCode codes.Code
	}{
		{
			nameTest: "Success update subtask",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("UpdateSubtask", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			nameTest:     "Error invalid subtask uuid",
			req:          &pb.UpdateSubtaskRequest{SubtaskLink: "invalid", UserLink: targetUserLink.String()},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest:     "Error invalid user uuid",
			req:          &pb.UpdateSubtaskRequest{SubtaskLink: targetSubtaskLink.String(), UserLink: "invalid"},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error subtask not found",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("UpdateSubtask", mock.Anything, mock.Anything, mock.Anything).Return(common.ErrSubtaskNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Error internal server",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("UpdateSubtask", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := newTestCardService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(mockService, Config{})
			_, err := handler.UpdateSubtask(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
		})
	}
}

func TestCreateAttachment(t *testing.T) {
	targetTaskLink := uuid.New()
	targetUserLink := uuid.New()
	attachmentLink := uuid.New()

	validReq := &pb.CreateAttachmentRequest{
		TaskLink: targetTaskLink.String(),
		UserLink: targetUserLink.String(),
		Data:     []byte("fake-image-data"),
		Name:     "photo.png",
	}

	serviceResult := serviceDto.AttachmentInfo{
		AttachmentLink: attachmentLink,
		Path:           "https://s3.example.com/file.png",
		Position:       1,
		DisplayName:    "photo.png",
	}

	tests := []struct {
		nameTest     string
		req          *pb.CreateAttachmentRequest
		mockBehavior func(m *testCardService)
		expectedCode codes.Code
		checkResp    func(t *testing.T, resp *pb.CreateAttachmentResponse)
	}{
		{
			nameTest: "Success create attachment",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("CreateAttachment", mock.Anything, mock.Anything).Return(serviceResult, nil)
			},
			expectedCode: codes.OK,
			checkResp: func(t *testing.T, resp *pb.CreateAttachmentResponse) {
				assert.NotNil(t, resp)
				assert.Equal(t, attachmentLink.String(), resp.AttachmentLink)
				assert.Equal(t, "https://s3.example.com/file.png", resp.Path)
				assert.Equal(t, int64(1), resp.Position)
				assert.Equal(t, "photo.png", resp.Name)
			},
		},
		{
			nameTest:     "Error invalid task uuid",
			req:          &pb.CreateAttachmentRequest{TaskLink: "invalid", UserLink: targetUserLink.String()},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest:     "Error invalid user uuid",
			req:          &pb.CreateAttachmentRequest{TaskLink: targetTaskLink.String(), UserLink: "invalid"},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error permission denied",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("CreateAttachment", mock.Anything, mock.Anything).Return(serviceDto.AttachmentInfo{}, rbac.ErrActionDenied)
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			nameTest: "Error missing required field",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("CreateAttachment", mock.Anything, mock.Anything).Return(serviceDto.AttachmentInfo{}, common.ErrMissingRequiredField)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error attachment limit reached",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("CreateAttachment", mock.Anything, mock.Anything).Return(serviceDto.AttachmentInfo{}, common.ErrAttachmentLimitReached)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error internal server",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("CreateAttachment", mock.Anything, mock.Anything).Return(serviceDto.AttachmentInfo{}, errors.New("s3 error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := newTestCardService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(mockService, Config{})
			resp, err := handler.CreateAttachment(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			if test.checkResp != nil {
				test.checkResp(t, resp)
			}
		})
	}
}

func TestDeleteAttachment(t *testing.T) {
	targetAttachmentLink := uuid.New()
	targetUserLink := uuid.New()

	validReq := &pb.DeleteAttachmentRequest{
		AttachmentLink: targetAttachmentLink.String(),
		UserLink:       targetUserLink.String(),
	}

	tests := []struct {
		nameTest     string
		req          *pb.DeleteAttachmentRequest
		mockBehavior func(m *testCardService)
		expectedCode codes.Code
	}{
		{
			nameTest: "Success delete attachment",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("DeleteAttachment", mock.Anything, mock.Anything).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			nameTest:     "Error invalid attachment uuid",
			req:          &pb.DeleteAttachmentRequest{AttachmentLink: "invalid", UserLink: targetUserLink.String()},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest:     "Error invalid user uuid",
			req:          &pb.DeleteAttachmentRequest{AttachmentLink: targetAttachmentLink.String(), UserLink: "invalid"},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error permission denied",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("DeleteAttachment", mock.Anything, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			nameTest: "Error attachment not found",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("DeleteAttachment", mock.Anything, mock.Anything).Return(common.ErrAttachmentNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Error internal server",
			req:      validReq,
			mockBehavior: func(m *testCardService) {
				m.On("DeleteAttachment", mock.Anything, mock.Anything).Return(errors.New("s3 error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := newTestCardService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(mockService, Config{})
			_, err := handler.DeleteAttachment(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
		})
	}
}
