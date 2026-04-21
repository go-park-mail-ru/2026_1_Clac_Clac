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
	targetDeadline := time.Now().Add(24 * time.Hour)
	execName := "John Doe"

	serviceCardInfo := serviceDto.InfoCard{
		Title:        "TestTitle",
		Description:  "Test Desc",
		NameExecuter: &execName,
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
			req:      &pb.GetCardRequest{CardLink: targetCardLink.String()},
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("GetCard", mock.Anything, targetCardLink).Return(serviceCardInfo, nil)
			},
			expectedCode: codes.OK,
			checkResp: func(t *testing.T, resp *pb.GetCardResponse) {
				assert.NotNil(t, resp)
				assert.Equal(t, targetCardLink.String(), resp.CardInfo.Link)
				assert.Equal(t, "TestTitle", resp.CardInfo.Title)
				assert.Equal(t, "Test Desc", resp.CardInfo.Description)
				assert.Equal(t, execName, resp.CardInfo.GetExecuterName())
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
			req:      &pb.GetCardRequest{CardLink: targetCardLink.String()},
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("GetCard", mock.Anything, targetCardLink).Return(serviceDto.InfoCard{}, common.ErrCardNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Error internal server",
			req:      &pb.GetCardRequest{CardLink: targetCardLink.String()},
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("GetCard", mock.Anything, targetCardLink).Return(serviceDto.InfoCard{}, errors.New("db error"))
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

	tests := []struct {
		nameTest     string
		req          *pb.DeleteCardRequest
		mockBehavior func(m *mockCardSrv.CardService)
		expectedCode codes.Code
	}{
		{
			nameTest: "Success delete card",
			req:      &pb.DeleteCardRequest{CardLink: targetCardLink.String()},
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("DeleteCard", mock.Anything, targetCardLink).Return(nil)
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
			req:      &pb.DeleteCardRequest{CardLink: targetCardLink.String()},
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("DeleteCard", mock.Anything, targetCardLink).Return(common.ErrCardNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Error internal server",
			req:      &pb.DeleteCardRequest{CardLink: targetCardLink.String()},
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("DeleteCard", mock.Anything, targetCardLink).Return(errors.New("db error"))
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
	targetExecutorLink := uuid.New()
	executorLinkStr := targetExecutorLink.String()
	deadline := time.Now().Add(24 * time.Hour)

	validReq := &pb.UpdateCardRequest{
		CardLink:     targetCardLink.String(),
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
				m.On("UpdateCardDetails", mock.Anything, mock.Anything).Return(nil)
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
				m.On("UpdateCardDetails", mock.Anything, mock.Anything).Return(common.ErrCardNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Error invalid reference data",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("UpdateCardDetails", mock.Anything, mock.Anything).Return(common.ErrInvalidReferenceCardData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error check violation data",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("UpdateCardDetails", mock.Anything, mock.Anything).Return(common.ErrInvalidCardData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error missing required field",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("UpdateCardDetails", mock.Anything, mock.Anything).Return(common.ErrMissingRequiredField)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error internal server",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("UpdateCardDetails", mock.Anything, mock.Anything).Return(errors.New("db error"))
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

	validReq := &pb.ReorderCardsRequest{
		CardLink:    targetCardLink.String(),
		SectionLink: targetSectionLink.String(),
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
				m.On("ReorderCard", mock.Anything, mock.Anything).Return(nil)
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
				m.On("ReorderCard", mock.Anything, mock.Anything).Return(common.ErrCardNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			nameTest: "Error skip mandatory section",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("ReorderCard", mock.Anything, mock.Anything).Return(common.ErrCannotSkipMandatorySection)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error invalid reference data",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("ReorderCard", mock.Anything, mock.Anything).Return(common.ErrInvalidReferenceCardData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error check violation data",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("ReorderCard", mock.Anything, mock.Anything).Return(common.ErrInvalidCardData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error missing required field",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("ReorderCard", mock.Anything, mock.Anything).Return(common.ErrMissingRequiredField)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error internal server",
			req:      validReq,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("ReorderCard", mock.Anything, mock.Anything).Return(errors.New("db error"))
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
		AuthorLink:  targetAuthorLink.String(),
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
			req:          &pb.CreateCardRequest{AuthorLink: "invalid", SectionLink: targetSectionLink.String()},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest:     "Error invalid section uuid",
			req:          &pb.CreateCardRequest{AuthorLink: targetAuthorLink.String(), SectionLink: "invalid"},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			nameTest: "Error max len title",
			req: &pb.CreateCardRequest{
				AuthorLink:  targetAuthorLink.String(),
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
