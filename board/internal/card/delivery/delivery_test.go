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
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/service/dto"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/card/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
	execName := "John Doe"

	serviceCardInfo := serviceDto.InfoCard{
		Title:        "TestTitle",
		Description:  "Test Desc",
		NameExecutor: &execName,
		DataDeadLine: &targetDeadline,
	}

	tests := []struct {
		nameTest     string
		req          *pb.GetCardRequest
		mockBehavior func(m *mockCardSrv.CardService)
		expectedCode codes.Code
		checkResp    func(t *testing.T, resp *pb.GetCardResponse)
	}{
		{
			nameTest: "Success get card",
			req:      &pb.GetCardRequest{CardLink: targetCardLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("GetCard", mock.Anything, targetCardLink, mock.Anything).Return(serviceCardInfo, nil)
			},
			expectedCode: codes.OK,
			checkResp: func(t *testing.T, resp *pb.GetCardResponse) {
				assert.NotNil(t, resp)
				assert.Equal(t, targetCardLink.String(), resp.CardInfo.Link)
				assert.Equal(t, "TestTitle", resp.CardInfo.Title)
				assert.Equal(t, "Test Desc", resp.CardInfo.Description)
				assert.Equal(t, execName, resp.CardInfo.GetExecutorName())
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
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("GetCard", mock.Anything, targetCardLink, mock.Anything).Return(serviceDto.InfoCard{}, common.ErrCardNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Error internal server",
			req:      &pb.GetCardRequest{CardLink: targetCardLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("GetCard", mock.Anything, targetCardLink, mock.Anything).Return(serviceDto.InfoCard{}, errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := mockCardSrv.NewCardService(t)
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
		mockBehavior func(m *mockCardSrv.CardService)
		expectedCode codes.Code
	}{
		{
			nameTest: "Success delete card",
			req:      &pb.DeleteCardRequest{CardLink: targetCardLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *mockCardSrv.CardService) {
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
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("DeleteCard", mock.Anything, targetCardLink, mock.Anything).Return(common.ErrCardNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Error internal server",
			req:      &pb.DeleteCardRequest{CardLink: targetCardLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("DeleteCard", mock.Anything, targetCardLink, mock.Anything).Return(errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := mockCardSrv.NewCardService(t)
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
		mockBehavior func(m *mockCardSrv.CardService)
		expectedCode codes.Code
	}{
		{
			nameTest: "Success update card",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
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
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("UpdateCardDetails", mock.Anything, mock.Anything, mock.Anything).Return(common.ErrCardNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Error invalid reference data",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("UpdateCardDetails", mock.Anything, mock.Anything, mock.Anything).Return(common.ErrInvalidReferenceCardData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error check violation data",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("UpdateCardDetails", mock.Anything, mock.Anything, mock.Anything).Return(common.ErrInvalidCardData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error missing required field",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("UpdateCardDetails", mock.Anything, mock.Anything, mock.Anything).Return(common.ErrMissingRequiredField)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error internal server",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("UpdateCardDetails", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := mockCardSrv.NewCardService(t)
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
		mockBehavior func(m *mockCardSrv.CardService)
		expectedCode codes.Code
	}{
		{
			nameTest: "Success reorder card",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
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
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("ReorderCard", mock.Anything, mock.Anything, mock.Anything).Return(common.ErrCardNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Error skip mandatory section",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("ReorderCard", mock.Anything, mock.Anything, mock.Anything).Return(common.ErrCannotSkipMandatorySection)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error invalid reference data",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("ReorderCard", mock.Anything, mock.Anything, mock.Anything).Return(common.ErrInvalidReferenceCardData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error check violation data",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("ReorderCard", mock.Anything, mock.Anything, mock.Anything).Return(common.ErrInvalidCardData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error missing required field",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("ReorderCard", mock.Anything, mock.Anything, mock.Anything).Return(common.ErrMissingRequiredField)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error internal server",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("ReorderCard", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := mockCardSrv.NewCardService(t)
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
		Description: "Task desc",
		SectionLink: targetSectionLink.String(),
	}

	tests := []struct {
		nameTest     string
		req          *pb.CreateCardRequest
		mockBehavior func(m *mockCardSrv.CardService)
		expectedCode codes.Code
		checkResp    func(t *testing.T, resp *pb.CreateCardResponse)
	}{
		{
			nameTest: "Success create card",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
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
				Description: "Task desc",
				SectionLink: targetSectionLink.String(),
			},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error section not found",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("CreateCard", mock.Anything, mock.Anything).Return(serviceDto.PlaceCard{}, common.ErrSectionNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Error card already exists",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("CreateCard", mock.Anything, mock.Anything).Return(serviceDto.PlaceCard{}, common.ErrCardAlreadyExists)
			},
			expectedCode: codes.AlreadyExists,
		},
		{
			nameTest: "Error invalid reference data",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("CreateCard", mock.Anything, mock.Anything).Return(serviceDto.PlaceCard{}, common.ErrInvalidReferenceCardData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error check violation data",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("CreateCard", mock.Anything, mock.Anything).Return(serviceDto.PlaceCard{}, common.ErrInvalidCardData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error missing required field",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("CreateCard", mock.Anything, mock.Anything).Return(serviceDto.PlaceCard{}, common.ErrMissingRequiredField)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error internal server",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("CreateCard", mock.Anything, mock.Anything).Return(serviceDto.PlaceCard{}, errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := mockCardSrv.NewCardService(t)
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
		mockBehavior func(m *mockCardSrv.CardService)
		expectedCode codes.Code
		checkResp    func(t *testing.T, resp *pb.GetCommentsResponse)
	}{
		{
			nameTest: "Success get comments",
			req:      &pb.GetCommentsRequest{CardLink: targetCardLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *mockCardSrv.CardService) {
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
			mockBehavior: func(m *mockCardSrv.CardService) {
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
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("GetComments", mock.Anything, targetCardLink, mock.Anything).Return([]serviceDto.CommentInfo{}, common.ErrCardNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Error internal server",
			req:      &pb.GetCommentsRequest{CardLink: targetCardLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("GetComments", mock.Anything, targetCardLink, mock.Anything).Return([]serviceDto.CommentInfo{}, errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := mockCardSrv.NewCardService(t)
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
		mockBehavior func(m *mockCardSrv.CardService)
		expectedCode codes.Code
		checkResp    func(t *testing.T, resp *pb.CreateCommentResponse)
	}{
		{
			nameTest: "Success create comment",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
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
			mockBehavior: func(m *mockCardSrv.CardService) {
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
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("CreateComment", mock.Anything, mock.Anything).Return(serviceDto.CommentInfo{}, common.ErrMissingRequiredField)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error invalid reference data",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("CreateComment", mock.Anything, mock.Anything).Return(serviceDto.CommentInfo{}, common.ErrInvalidReferenceCardData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error internal server",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("CreateComment", mock.Anything, mock.Anything).Return(serviceDto.CommentInfo{}, errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := mockCardSrv.NewCardService(t)
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
		mockBehavior func(m *mockCardSrv.CardService)
		expectedCode codes.Code
	}{
		{
			nameTest: "Success delete comment",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
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
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("DeleteComment", mock.Anything, targetCommentLink, targetUserLink).Return(common.ErrCommentNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Error permission denied",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("DeleteComment", mock.Anything, targetCommentLink, targetUserLink).Return(common.ErrPermissionDenied)
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			nameTest: "Error internal server",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("DeleteComment", mock.Anything, targetCommentLink, targetUserLink).Return(errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := mockCardSrv.NewCardService(t)
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
		mockBehavior func(m *mockCardSrv.CardService)
		expectedCode codes.Code
	}{
		{
			nameTest: "Success update comment",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
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
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("UpdateComment", mock.Anything, mock.Anything).Return(common.ErrCommentNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Error permission denied",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("UpdateComment", mock.Anything, mock.Anything).Return(common.ErrPermissionDenied)
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			nameTest: "Error internal server",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("UpdateComment", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := mockCardSrv.NewCardService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(mockService, Config{})
			_, err := handler.UpdateComment(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
		})
	}
}
