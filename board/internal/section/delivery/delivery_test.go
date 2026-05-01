package delivery_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/delivery"
	mockSectionService "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/delivery/mock_section_service"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/models"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/service/dto"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/boardRbac"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/section/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func grpcCode(err error) codes.Code {
	st, _ := status.FromError(err)
	return st.Code()
}

func TestGetSection(t *testing.T) {
	sectionLink := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	targetUserLink := uuid.New()
	maxTasks := 50
	maxTasks64 := int64(maxTasks)

	serviceSectionInfo := serviceDto.FullSectionInfo{
		SectionLink: sectionLink,
		SectionName: "To Do",
		Position:    1,
		IsMandatory: true,
		Color:       "white",
		MaxTasks:    &maxTasks,
	}

	tests := []struct {
		name         string
		req          *pb.GetSectionRequest
		mockBehavior func(m *mockSectionService.SectionService)
		expectedCode codes.Code
		checkResp    func(t *testing.T, resp *pb.GetSectionResponse)
	}{
		{
			name: "success",
			req:  &pb.GetSectionRequest{SectionLink: sectionLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("GetSection", mock.Anything, sectionLink, mock.Anything).Return(serviceSectionInfo, nil)
			},
			expectedCode: codes.OK,
			checkResp: func(t *testing.T, resp *pb.GetSectionResponse) {
				assert.Equal(t, sectionLink.String(), resp.SectionInfo.Link)
				assert.Equal(t, "To Do", resp.SectionInfo.Name)
				assert.Equal(t, int64(1), resp.SectionInfo.Position)
				assert.True(t, resp.SectionInfo.IsMandatory)
				assert.Equal(t, "white", resp.SectionInfo.Color)
				assert.Equal(t, maxTasks64, resp.SectionInfo.GetMaxTasks())
			},
		},
		{
			name:         "invalid uuid",
			req:          &pb.GetSectionRequest{SectionLink: "not-a-uuid"},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "section not found",
			req:  &pb.GetSectionRequest{SectionLink: sectionLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("GetSection", mock.Anything, sectionLink, mock.Anything).Return(serviceDto.FullSectionInfo{}, common.ErrSectionNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "internal error",
			req:  &pb.GetSectionRequest{SectionLink: sectionLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("GetSection", mock.Anything, sectionLink, mock.Anything).Return(serviceDto.FullSectionInfo{}, errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mockSectionService.NewSectionService(t)
			if tc.mockBehavior != nil {
				tc.mockBehavior(mockSvc)
			}
			h := delivery.NewHandler(mockSvc, delivery.Config{})
			resp, err := h.GetSection(context.Background(), tc.req)
			assert.Equal(t, tc.expectedCode, grpcCode(err))
			if tc.checkResp != nil && err == nil {
				tc.checkResp(t, resp)
			}
		})
	}
}

func TestGetSections(t *testing.T) {
	boardLink := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	sectionLink := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	targetUserLink := uuid.New()
	maxTasks := 50

	serviceSections := []serviceDto.FullSectionInfo{
		{
			SectionLink: sectionLink,
			SectionName: "To Do",
			Position:    1,
			IsMandatory: true,
			Color:       "white",
			MaxTasks:    &maxTasks,
		},
	}

	tests := []struct {
		name         string
		req          *pb.GetSectionsRequest
		mockBehavior func(m *mockSectionService.SectionService)
		expectedCode codes.Code
		checkResp    func(t *testing.T, resp *pb.GetSectionsResponse)
	}{
		{
			name: "success",
			req:  &pb.GetSectionsRequest{BoardLink: boardLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("GetSections", mock.Anything, boardLink, mock.Anything).Return(serviceSections, nil)
			},
			expectedCode: codes.OK,
			checkResp: func(t *testing.T, resp *pb.GetSectionsResponse) {
				assert.Len(t, resp.SectionsInfo, 1)
				assert.Equal(t, sectionLink.String(), resp.SectionsInfo[0].Link)
				assert.Equal(t, "To Do", resp.SectionsInfo[0].Name)
			},
		},
		{
			name:         "invalid board uuid",
			req:          &pb.GetSectionsRequest{BoardLink: "bad-uuid"},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "internal error",
			req:  &pb.GetSectionsRequest{BoardLink: boardLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("GetSections", mock.Anything, boardLink, mock.Anything).Return(nil, errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mockSectionService.NewSectionService(t)
			if tc.mockBehavior != nil {
				tc.mockBehavior(mockSvc)
			}
			h := delivery.NewHandler(mockSvc, delivery.Config{})
			resp, err := h.GetSections(context.Background(), tc.req)
			assert.Equal(t, tc.expectedCode, grpcCode(err))
			if tc.checkResp != nil && err == nil {
				tc.checkResp(t, resp)
			}
		})
	}
}

func TestGetCards(t *testing.T) {
	sectionLink := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	targetUserLink := uuid.New()
	cardLink := uuid.New()
	subtaskLink := uuid.New()
	executer := "John Doe"
	deadline := time.Now().Add(24 * time.Hour)

	serviceCards := []serviceDto.Card{
		{
			CardLink:     cardLink,
			ExecutorName: &executer,
			Title:        "Task 1",
			DeadLine:     &deadline,
			Subtasks: []models.SubtaskInfo{
				{
					SubtaskLink: subtaskLink,
					Description: "Subtask 1",
					IsDone:      true,
					Position:    1,
				},
			},
		},
	}

	tests := []struct {
		name         string
		req          *pb.GetCardsRequest
		mockBehavior func(m *mockSectionService.SectionService)
		expectedCode codes.Code
		checkResp    func(t *testing.T, resp *pb.GetCardsResponse)
	}{
		{
			name: "success",
			req:  &pb.GetCardsRequest{SectionLink: sectionLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				// Ожидаем передачу targetUserLink
				m.On("GetCards", mock.Anything, sectionLink, targetUserLink).Return(serviceCards, nil)
			},
			expectedCode: codes.OK,
			checkResp: func(t *testing.T, resp *pb.GetCardsResponse) {
				assert.Len(t, resp.CardsInfo, 1)
				card := resp.CardsInfo[0]

				// Проверяем основные поля
				assert.Equal(t, cardLink.String(), card.Link)
				assert.Equal(t, "Task 1", card.Title)
				assert.Equal(t, executer, *card.ExecutorName)
				assert.NotNil(t, card.Deadline)

				// Проверяем маппинг подзадач
				assert.Len(t, card.Subtasks, 1)
				assert.Equal(t, subtaskLink.String(), card.Subtasks[0].SubtaskLink)
				assert.Equal(t, "Subtask 1", card.Subtasks[0].Description)
				assert.True(t, card.Subtasks[0].IsDone)
				assert.Equal(t, int64(1), card.Subtasks[0].Position)
			},
		},
		{
			name:         "invalid section uuid",
			req:          &pb.GetCardsRequest{SectionLink: "bad-uuid", UserLink: targetUserLink.String()},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "invalid user uuid", // НОВЫЙ КЕЙС
			req:          &pb.GetCardsRequest{SectionLink: sectionLink.String(), UserLink: "bad-uuid"},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "permission denied (rbac)", // НОВЫЙ КЕЙС
			req:  &pb.GetCardsRequest{SectionLink: sectionLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("GetCards", mock.Anything, sectionLink, targetUserLink).Return(nil, rbac.ErrActionDenied)
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "section not found",
			req:  &pb.GetCardsRequest{SectionLink: sectionLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("GetCards", mock.Anything, sectionLink, targetUserLink).Return(nil, common.ErrSectionNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "internal error",
			req:  &pb.GetCardsRequest{SectionLink: sectionLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("GetCards", mock.Anything, sectionLink, targetUserLink).Return(nil, errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mockSectionService.NewSectionService(t)
			if tc.mockBehavior != nil {
				tc.mockBehavior(mockSvc)
			}

			// Инициализация хендлера (подставь свой способ, если он отличается)
			h := delivery.NewHandler(mockSvc, delivery.Config{})

			resp, err := h.GetCards(context.Background(), tc.req)

			// Предполагается, что grpcCode - это твоя вспомогательная функция,
			// которая делает status.Code(err)
			assert.Equal(t, tc.expectedCode, grpcCode(err))

			if tc.checkResp != nil && err == nil {
				tc.checkResp(t, resp)
			}
		})
	}
}
func TestCreateSection(t *testing.T) {
	boardLink := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	sectionLink := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	targetUserLink := uuid.New()
	validMaxTasks := 50
	validMaxTasks64 := int64(validMaxTasks)
	invalidMaxTasks64 := int64(101)

	serviceResult := serviceDto.EntitySection{
		SectionLink: sectionLink,
		SectionName: "In Progress",
		IsMandatory: false,
		Position:    2,
		Color:       "blue",
		MaxTasks:    &validMaxTasks,
	}

	makeReq := func(maxTasks *int64) *pb.CreateSectionRequest {
		return &pb.CreateSectionRequest{
			BoardLink:   boardLink.String(),
			UserLink:    targetUserLink.String(),
			Name:        "In Progress",
			IsMandatory: false,
			Color:       "blue",
			MaxTasks:    maxTasks,
		}
	}

	serviceInput := serviceDto.CreatingSection{
		BoardLink:   boardLink,
		SectionName: "In Progress",
		IsMandatory: false,
		Color:       "blue",
		MaxTasks:    &validMaxTasks,
	}

	tests := []struct {
		name         string
		req          *pb.CreateSectionRequest
		mockBehavior func(m *mockSectionService.SectionService)
		expectedCode codes.Code
	}{
		{
			name: "success",
			req:  makeReq(&validMaxTasks64),
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("CreateSection", mock.Anything, serviceInput, mock.Anything).Return(serviceResult, nil)
			},
			expectedCode: codes.OK,
		},
		{
			name:         "invalid board uuid",
			req:          &pb.CreateSectionRequest{BoardLink: "bad-uuid"},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "max tasks exceeded",
			req:          makeReq(&invalidMaxTasks64),
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "section already exists",
			req:  makeReq(&validMaxTasks64),
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("CreateSection", mock.Anything, serviceInput, mock.Anything).Return(serviceDto.EntitySection{}, common.ErrSectionAlreadyExists)
			},
			expectedCode: codes.AlreadyExists,
		},
		{
			name: "invalid reference data",
			req:  makeReq(&validMaxTasks64),
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("CreateSection", mock.Anything, serviceInput, mock.Anything).Return(serviceDto.EntitySection{}, common.ErrInvalidReferenceSectionData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "invalid section data",
			req:  makeReq(&validMaxTasks64),
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("CreateSection", mock.Anything, serviceInput, mock.Anything).Return(serviceDto.EntitySection{}, common.ErrInvalidSectionData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "missing required field",
			req:  makeReq(&validMaxTasks64),
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("CreateSection", mock.Anything, serviceInput, mock.Anything).Return(serviceDto.EntitySection{}, common.ErrMissingRequiredField)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "internal error",
			req:  makeReq(&validMaxTasks64),
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("CreateSection", mock.Anything, serviceInput, mock.Anything).Return(serviceDto.EntitySection{}, errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mockSectionService.NewSectionService(t)
			if tc.mockBehavior != nil {
				tc.mockBehavior(mockSvc)
			}
			h := delivery.NewHandler(mockSvc, delivery.Config{
				MaxQuantityTasks: 100,
				MinQuantityTasks: 0,
			})
			_, err := h.CreateSection(context.Background(), tc.req)
			assert.Equal(t, tc.expectedCode, grpcCode(err))
		})
	}
}

func TestDeleteSection(t *testing.T) {
	sectionLink := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	targetUserLink := uuid.New()

	tests := []struct {
		name         string
		req          *pb.DeleteSectionRequest
		mockBehavior func(m *mockSectionService.SectionService)
		expectedCode codes.Code
	}{
		{
			name: "success",
			req:  &pb.DeleteSectionRequest{SectionLink: sectionLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("DeleteSection", mock.Anything, sectionLink, mock.Anything).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name:         "invalid uuid",
			req:          &pb.DeleteSectionRequest{SectionLink: "bad-uuid"},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "section not found",
			req:  &pb.DeleteSectionRequest{SectionLink: sectionLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("DeleteSection", mock.Anything, sectionLink, mock.Anything).Return(common.ErrSectionNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "cannot delete backlog",
			req:  &pb.DeleteSectionRequest{SectionLink: sectionLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("DeleteSection", mock.Anything, sectionLink, mock.Anything).Return(common.ErrCannotDeleteBacklog)
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "invalid reference data",
			req:  &pb.DeleteSectionRequest{SectionLink: sectionLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("DeleteSection", mock.Anything, sectionLink, mock.Anything).Return(common.ErrInvalidReferenceSectionData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "invalid section data",
			req:  &pb.DeleteSectionRequest{SectionLink: sectionLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("DeleteSection", mock.Anything, sectionLink, mock.Anything).Return(common.ErrInvalidSectionData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "missing required field",
			req:  &pb.DeleteSectionRequest{SectionLink: sectionLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("DeleteSection", mock.Anything, sectionLink, mock.Anything).Return(common.ErrMissingRequiredField)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "internal error",
			req:  &pb.DeleteSectionRequest{SectionLink: sectionLink.String(), UserLink: targetUserLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("DeleteSection", mock.Anything, sectionLink, mock.Anything).Return(errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mockSectionService.NewSectionService(t)
			if tc.mockBehavior != nil {
				tc.mockBehavior(mockSvc)
			}
			h := delivery.NewHandler(mockSvc, delivery.Config{})
			_, err := h.DeleteSection(context.Background(), tc.req)
			assert.Equal(t, tc.expectedCode, grpcCode(err))
		})
	}
}

func TestUpdateSection(t *testing.T) {
	sectionLink := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	targetUserLink := uuid.New()
	maxTasks := 50
	maxTasks64 := int64(maxTasks)
	invalidMaxTasks64 := int64(101)

	serviceUpdateInfo := serviceDto.FullSectionInfo{
		SectionLink: sectionLink,
		SectionName: "Updated Name",
		Position:    0,
		IsMandatory: true,
		Color:       "red",
		MaxTasks:    &maxTasks,
	}

	makeReq := func(name, color string, maxT *int64) *pb.UpdateSectionRequest {
		return &pb.UpdateSectionRequest{
			SectionLink: sectionLink.String(),
			UserLink:    targetUserLink.String(),
			Name:        name,
			IsMandatory: true,
			Color:       color,
			MaxTasks:    maxT,
		}
	}

	tests := []struct {
		name         string
		req          *pb.UpdateSectionRequest
		mockBehavior func(m *mockSectionService.SectionService)
		expectedCode codes.Code
	}{
		{
			name: "success",
			req:  makeReq("Updated Name", "red", &maxTasks64),
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("UpdateSection", mock.Anything, serviceUpdateInfo, mock.Anything).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name:         "invalid section uuid",
			req:          &pb.UpdateSectionRequest{SectionLink: "bad-uuid"},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "name too long",
			req:          makeReq(string(make([]byte, 129)), "red", &maxTasks64),
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "invalid color",
			req:          makeReq("Updated Name", "invisible", &maxTasks64),
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "max tasks exceeded",
			req:          makeReq("Updated Name", "red", &invalidMaxTasks64),
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "section not found",
			req:  makeReq("Updated Name", "red", &maxTasks64),
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("UpdateSection", mock.Anything, serviceUpdateInfo, mock.Anything).Return(common.ErrSectionNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "cannot update backlog",
			req:  makeReq("Updated Name", "red", &maxTasks64),
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("UpdateSection", mock.Anything, serviceUpdateInfo, mock.Anything).Return(common.ErrCannotUpdateBacklog)
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "invalid reference data",
			req:  makeReq("Updated Name", "red", &maxTasks64),
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("UpdateSection", mock.Anything, serviceUpdateInfo, mock.Anything).Return(common.ErrInvalidReferenceSectionData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "invalid section data",
			req:  makeReq("Updated Name", "red", &maxTasks64),
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("UpdateSection", mock.Anything, serviceUpdateInfo, mock.Anything).Return(common.ErrInvalidSectionData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "missing required field",
			req:  makeReq("Updated Name", "red", &maxTasks64),
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("UpdateSection", mock.Anything, serviceUpdateInfo, mock.Anything).Return(common.ErrMissingRequiredField)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "internal error",
			req:  makeReq("Updated Name", "red", &maxTasks64),
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("UpdateSection", mock.Anything, serviceUpdateInfo, mock.Anything).Return(errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mockSectionService.NewSectionService(t)
			if tc.mockBehavior != nil {
				tc.mockBehavior(mockSvc)
			}
			h := delivery.NewHandler(mockSvc, delivery.Config{
				MaxLenNameSection: 128,
				MaxQuantityTasks:  100,
				MinQuantityTasks:  0,
			})
			_, err := h.UpdateSection(context.Background(), tc.req)
			assert.Equal(t, tc.expectedCode, grpcCode(err))
		})
	}
}

func TestReorderSection(t *testing.T) {
	boardLink := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	section1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	section2 := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	targetUserLink := uuid.New()
	linksList := []uuid.UUID{section1, section2}

	tests := []struct {
		name         string
		req          *pb.ReorderSectionRequest
		mockBehavior func(m *mockSectionService.SectionService)
		expectedCode codes.Code
	}{
		{
			name: "success",
			req: &pb.ReorderSectionRequest{
				BoardLink: boardLink.String(),
				UserLink:  targetUserLink.String(),
				LinksList: []string{section1.String(), section2.String()},
			},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("ReorderSection", mock.Anything, boardLink, linksList, mock.Anything).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name:         "invalid board uuid",
			req:          &pb.ReorderSectionRequest{BoardLink: "bad-uuid"},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "invalid section uuid in list",
			req: &pb.ReorderSectionRequest{
				BoardLink: boardLink.String(),
				UserLink:  targetUserLink.String(),
				LinksList: []string{"bad-uuid"},
			},
			mockBehavior: nil,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "not find all links",
			req: &pb.ReorderSectionRequest{
				BoardLink: boardLink.String(),
				UserLink:  targetUserLink.String(),
				LinksList: []string{section1.String(), section2.String()},
			},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("ReorderSection", mock.Anything, boardLink, linksList, mock.Anything).Return(common.ErrNotFindAllLinks)
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "invalid reference data",
			req: &pb.ReorderSectionRequest{
				BoardLink: boardLink.String(),
				UserLink:  targetUserLink.String(),
				LinksList: []string{section1.String(), section2.String()},
			},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("ReorderSection", mock.Anything, boardLink, linksList, mock.Anything).Return(common.ErrInvalidReferenceSectionData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "invalid section data",
			req: &pb.ReorderSectionRequest{
				BoardLink: boardLink.String(),
				UserLink:  targetUserLink.String(),
				LinksList: []string{section1.String(), section2.String()},
			},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("ReorderSection", mock.Anything, boardLink, linksList, mock.Anything).Return(common.ErrInvalidSectionData)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "missing required field",
			req: &pb.ReorderSectionRequest{
				BoardLink: boardLink.String(),
				UserLink:  targetUserLink.String(),
				LinksList: []string{section1.String(), section2.String()},
			},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("ReorderSection", mock.Anything, boardLink, linksList, mock.Anything).Return(common.ErrMissingRequiredField)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "internal error",
			req: &pb.ReorderSectionRequest{
				BoardLink: boardLink.String(),
				UserLink:  targetUserLink.String(),
				LinksList: []string{section1.String(), section2.String()},
			},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("ReorderSection", mock.Anything, boardLink, linksList, mock.Anything).Return(errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mockSectionService.NewSectionService(t)
			if tc.mockBehavior != nil {
				tc.mockBehavior(mockSvc)
			}
			h := delivery.NewHandler(mockSvc, delivery.Config{})
			_, err := h.ReorderSection(context.Background(), tc.req)
			assert.Equal(t, tc.expectedCode, grpcCode(err))
		})
	}
}
