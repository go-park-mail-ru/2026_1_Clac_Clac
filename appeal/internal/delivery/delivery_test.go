package delivery_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/common"
	handler "github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/delivery"
	mocks "github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/delivery/mock_appeal_srv"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/service/dto"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/appealRbac"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/appeal/v1"
)

var testConf = handler.Config{
	AttachmentBaseURL: "https://storage.example.com",
}

func grpcCode(err error) codes.Code {
	return status.Code(err)
}

func TestHandler_CreateAppeal(t *testing.T) {
	userLink := uuid.New()

	tests := []struct {
		name         string
		req          *pb.CreateAppealRequest
		setupMock    func(m *mocks.AppealService)
		expectedCode codes.Code
	}{
		{
			name: "Success",
			req: &pb.CreateAppealRequest{
				UserLink:    userLink.String(),
				Email:       "test@test.com",
				Category:    pb.Category_CATEGORY_BUG,
				Description: "test description",
				DisplayName: "Test User",
			},
			setupMock: func(m *mocks.AppealService) {
				m.On("CreateAppeal", context.Background(), serviceDto.EntityAppeal{
					UserLink:    userLink,
					Mail:        "test@test.com",
					Category:    common.Categories.Bug,
					Description: "test description",
					DisplayName: "Test User",
				}).Return(uuid.New(), nil)
			},
			expectedCode: codes.OK,
		},
		{
			name:         "Invalid user link",
			req:          &pb.CreateAppealRequest{UserLink: "not-a-uuid", Email: "test@test.com", Category: pb.Category_CATEGORY_BUG},
			setupMock:    func(m *mocks.AppealService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "Invalid category",
			req:          &pb.CreateAppealRequest{UserLink: userLink.String(), Email: "test@test.com", Category: pb.Category_CATEGORY_UNSPECIFIED},
			setupMock:    func(m *mocks.AppealService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Invalid email format",
			req: &pb.CreateAppealRequest{
				UserLink:    userLink.String(),
				Email:       "not-an-email",
				Category:    pb.Category_CATEGORY_BUG,
				Description: "desc",
				DisplayName: "Name",
			},
			setupMock:    func(m *mocks.AppealService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Internal error",
			req: &pb.CreateAppealRequest{
				UserLink:    userLink.String(),
				Email:       "test@test.com",
				Category:    pb.Category_CATEGORY_BUG,
				Description: "desc",
				DisplayName: "Name",
			},
			setupMock: func(m *mocks.AppealService) {
				m.On("CreateAppeal", context.Background(), serviceDto.EntityAppeal{
					UserLink:    userLink,
					Mail:        "test@test.com",
					Category:    common.Categories.Bug,
					Description: "desc",
					DisplayName: "Name",
				}).Return(uuid.Nil, errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.AppealService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			_, err := h.CreateAppeal(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestHandler_GetAppeals(t *testing.T) {
	userLink := uuid.New()
	appealLink := uuid.New()

	appeals := serviceDto.Appeals{
		Role: rbac.Roles.User,
		Appeals: []serviceDto.Appeal{
			{
				AppealID:    1,
				AppealLink:  appealLink,
				Email:       "user@test.com",
				DisplayName: "Test User",
				Status:      common.Statuses.Open,
				Category:    common.Categories.Bug,
				Description: "desc",
				CreatedAt:   time.Now(),
			},
		},
	}

	tests := []struct {
		name         string
		req          *pb.GetAppealsRequest
		setupMock    func(m *mocks.AppealService)
		expectedCode codes.Code
	}{
		{
			name: "Success",
			req:  &pb.GetAppealsRequest{UserLink: userLink.String()},
			setupMock: func(m *mocks.AppealService) {
				m.On("GetAppeals", context.Background(), userLink).Return(appeals, nil)
			},
			expectedCode: codes.OK,
		},
		{
			name:         "Invalid user link",
			req:          &pb.GetAppealsRequest{UserLink: "not-a-uuid"},
			setupMock:    func(m *mocks.AppealService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Internal error",
			req:  &pb.GetAppealsRequest{UserLink: userLink.String()},
			setupMock: func(m *mocks.AppealService) {
				m.On("GetAppeals", context.Background(), userLink).Return(serviceDto.Appeals{}, errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.AppealService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			_, err := h.GetAppeals(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestHandler_DeleteAppeal(t *testing.T) {
	userLink := uuid.New()
	appealLink := uuid.New()

	tests := []struct {
		name         string
		req          *pb.DeleteAppealRequest
		setupMock    func(m *mocks.AppealService)
		expectedCode codes.Code
	}{
		{
			name: "Success",
			req:  &pb.DeleteAppealRequest{UserLink: userLink.String(), AppealLink: appealLink.String()},
			setupMock: func(m *mocks.AppealService) {
				m.On("DeleteAppeal", context.Background(), appealLink, userLink).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name:         "Invalid user link",
			req:          &pb.DeleteAppealRequest{UserLink: "bad", AppealLink: appealLink.String()},
			setupMock:    func(m *mocks.AppealService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "Invalid appeal link",
			req:          &pb.DeleteAppealRequest{UserLink: userLink.String(), AppealLink: "bad"},
			setupMock:    func(m *mocks.AppealService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Internal error",
			req:  &pb.DeleteAppealRequest{UserLink: userLink.String(), AppealLink: appealLink.String()},
			setupMock: func(m *mocks.AppealService) {
				m.On("DeleteAppeal", context.Background(), appealLink, userLink).Return(errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.AppealService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			_, err := h.DeleteAppeal(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestHandler_GetStats(t *testing.T) {
	userLink := uuid.New()

	tests := []struct {
		name         string
		req          *pb.GetStatsRequest
		setupMock    func(m *mocks.AppealService)
		expectedCode codes.Code
	}{
		{
			name: "Success",
			req:  &pb.GetStatsRequest{UserLink: userLink.String()},
			setupMock: func(m *mocks.AppealService) {
				m.On("GetStats", context.Background(), userLink).Return(serviceDto.AppealStats{Open: 3, InWork: 1, Close: 10}, nil)
			},
			expectedCode: codes.OK,
		},
		{
			name:         "Invalid user link",
			req:          &pb.GetStatsRequest{UserLink: "bad"},
			setupMock:    func(m *mocks.AppealService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Internal error",
			req:  &pb.GetStatsRequest{UserLink: userLink.String()},
			setupMock: func(m *mocks.AppealService) {
				m.On("GetStats", context.Background(), userLink).Return(serviceDto.AppealStats{}, errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.AppealService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			_, err := h.GetStats(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestHandler_ChangeAppealStatus(t *testing.T) {
	userLink := uuid.New()
	appealLink := uuid.New()

	tests := []struct {
		name         string
		req          *pb.ChangeAppealStatusRequest
		setupMock    func(m *mocks.AppealService)
		expectedCode codes.Code
	}{
		{
			name: "Success",
			req: &pb.ChangeAppealStatusRequest{
				UserLink:   userLink.String(),
				AppealLink: appealLink.String(),
				NewStatus:  pb.Status_STATUS_IN_WORK,
			},
			setupMock: func(m *mocks.AppealService) {
				m.On("ChangeAppealStatus", context.Background(), serviceDto.ChangeAppealStatusInfo{
					SupporterLink: userLink,
					AppealLink:    appealLink,
					Status:        common.Statuses.InWork,
				}).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name:         "Invalid user link",
			req:          &pb.ChangeAppealStatusRequest{UserLink: "bad", AppealLink: appealLink.String(), NewStatus: pb.Status_STATUS_IN_WORK},
			setupMock:    func(m *mocks.AppealService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "Invalid appeal link",
			req:          &pb.ChangeAppealStatusRequest{UserLink: userLink.String(), AppealLink: "bad", NewStatus: pb.Status_STATUS_IN_WORK},
			setupMock:    func(m *mocks.AppealService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "Invalid status",
			req:          &pb.ChangeAppealStatusRequest{UserLink: userLink.String(), AppealLink: appealLink.String(), NewStatus: pb.Status_STATUS_UNSPECIFIED},
			setupMock:    func(m *mocks.AppealService) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Internal error",
			req: &pb.ChangeAppealStatusRequest{
				UserLink:   userLink.String(),
				AppealLink: appealLink.String(),
				NewStatus:  pb.Status_STATUS_IN_WORK,
			},
			setupMock: func(m *mocks.AppealService) {
				m.On("ChangeAppealStatus", context.Background(), serviceDto.ChangeAppealStatusInfo{
					SupporterLink: userLink,
					AppealLink:    appealLink,
					Status:        common.Statuses.InWork,
				}).Return(errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.AppealService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			_, err := h.ChangeAppealStatus(context.Background(), test.req)

			assert.Equal(t, test.expectedCode, grpcCode(err))
			mockSrv.AssertExpectations(t)
		})
	}
}
