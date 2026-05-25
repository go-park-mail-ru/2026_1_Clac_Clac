package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	fixedCardLinkH       = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	fixedUserLinkH       = uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
	fixedCommentLinkH    = uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	fixedSubtaskLinkH    = uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	fixedAttachmentLinkH = uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")
)

var defaultCardCfg = CardConfig{MaxLenTitle: 128, MaxLenDescription: 500}

type mockCardUsecase struct {
	mock.Mock
}

func (m *mockCardUsecase) GetCard(ctx context.Context, infoCard domain.GetCardRequest) (domain.CardFullInfo, error) {
	args := m.Called(ctx, infoCard)
	return args.Get(0).(domain.CardFullInfo), args.Error(1)
}

func (m *mockCardUsecase) DeleteCard(ctx context.Context, infoCard domain.DeleteCardRequest) error {
	args := m.Called(ctx, infoCard)
	return args.Error(0)
}

func (m *mockCardUsecase) UpdateCard(ctx context.Context, infoCard domain.UpdateCardRequest) error {
	args := m.Called(ctx, infoCard)
	return args.Error(0)
}

func (m *mockCardUsecase) ReorderCards(ctx context.Context, infoCard domain.ReorderCardsRequest) error {
	args := m.Called(ctx, infoCard)
	return args.Error(0)
}

func (m *mockCardUsecase) CreateCard(ctx context.Context, infoCard domain.CreateCardRequest) (domain.CreateCardResponse, error) {
	args := m.Called(ctx, infoCard)
	return args.Get(0).(domain.CreateCardResponse), args.Error(1)
}

func (m *mockCardUsecase) GetComments(ctx context.Context, infoComments domain.GetCommentsRequest) (domain.GetCommentsResponse, error) {
	args := m.Called(ctx, infoComments)
	return args.Get(0).(domain.GetCommentsResponse), args.Error(1)
}

func (m *mockCardUsecase) CreateComment(ctx context.Context, infoComment domain.CreateCommentRequest) (domain.CreateCommentResponse, error) {
	args := m.Called(ctx, infoComment)
	return args.Get(0).(domain.CreateCommentResponse), args.Error(1)
}

func (m *mockCardUsecase) DeleteComment(ctx context.Context, infoComment domain.DeleteCommentRequest) error {
	args := m.Called(ctx, infoComment)
	return args.Error(0)
}

func (m *mockCardUsecase) UpdateComment(ctx context.Context, infoComment domain.UpdateCommentRequest) error {
	args := m.Called(ctx, infoComment)
	return args.Error(0)
}

func (m *mockCardUsecase) CreateSubtask(ctx context.Context, infoSubtask domain.CreateSubtaskRequest) (domain.SubtaskInfo, error) {
	args := m.Called(ctx, infoSubtask)
	return args.Get(0).(domain.SubtaskInfo), args.Error(1)
}

func (m *mockCardUsecase) UpdateSubtask(ctx context.Context, infoSubtask domain.UpdateSubtaskRequest) error {
	args := m.Called(ctx, infoSubtask)
	return args.Error(0)
}

func (m *mockCardUsecase) DeleteSubtask(ctx context.Context, infoSubtask domain.DeleteSubtaskRequest) error {
	args := m.Called(ctx, infoSubtask)
	return args.Error(0)
}

func (m *mockCardUsecase) CreateAttachment(ctx context.Context, infoAttachment domain.CreateAttachmentRequest) (domain.AttachmentInfo, error) {
	args := m.Called(ctx, infoAttachment)
	return args.Get(0).(domain.AttachmentInfo), args.Error(1)
}

func (m *mockCardUsecase) DeleteAttachment(ctx context.Context, infoAttachment domain.DeleteAttachmentRequest) error {
	args := m.Called(ctx, infoAttachment)
	return args.Error(0)
}

func (m *mockCardUsecase) UpdateStatusTask(ctx context.Context, info domain.NewStatusTask) error {
	args := m.Called(ctx, info)
	return args.Error(0)
}

func (m *mockCardUsecase) UpdateTimeLine(ctx context.Context, info domain.NewTimeLine) error {
	args := m.Called(ctx, info)
	return args.Error(0)
}

func (m *mockCardUsecase) UpdateCardPoints(ctx context.Context, req domain.UpdateCardPointsRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func newTestCardHandler(uc CardUsecase) *Card {
	return NewCard(uc, defaultCardCfg)
}

func withCardUserCtx(req *http.Request) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, fixedUserLinkH)
	return req.WithContext(ctx)
}

func TestCardHandler_GetCard(t *testing.T) {
	cardInfo := domain.CardFullInfo{
		CardLink:    fixedCardLinkH,
		Title:       "Test Card",
		Description: "Desc",
		Subtasks:    []domain.SubtaskInfo{},
		Attachments: []domain.AttachmentInfo{},
	}

	tests := []struct {
		name               string
		linkParam          string
		setContext         bool
		mockBehavior       func(m *mockCardUsecase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			linkParam:  fixedCardLinkH.String(),
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("GetCard", mock.Anything, domain.GetCardRequest{
					UserLink: fixedUserLinkH,
					CardLink: fixedCardLinkH,
				}).Return(cardInfo, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			linkParam:          fixedCardLinkH.String(),
			setContext:         false,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidPathParam",
			linkParam:          "not-a-uuid",
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "NotFound",
			linkParam:  fixedCardLinkH.String(),
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("GetCard", mock.Anything, mock.Anything).Return(domain.CardFullInfo{}, common.ErrorCardNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:       "PermissionDenied",
			linkParam:  fixedCardLinkH.String(),
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("GetCard", mock.Anything, mock.Anything).Return(domain.CardFullInfo{}, common.ErrorPermissionDenied)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:       "InternalError",
			linkParam:  fixedCardLinkH.String(),
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("GetCard", mock.Anything, mock.Anything).Return(domain.CardFullInfo{}, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardUsecase)
			tc.mockBehavior(m)

			req := httptest.NewRequest(http.MethodGet, "/cards/"+tc.linkParam, nil)
			req = mux.SetURLVars(req, map[string]string{"link": tc.linkParam})
			if tc.setContext {
				req = withCardUserCtx(req)
			}
			rr := httptest.NewRecorder()

			newTestCardHandler(m).GetCard(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestCardHandler_DeleteCard(t *testing.T) {
	tests := []struct {
		name               string
		linkParam          string
		setContext         bool
		mockBehavior       func(m *mockCardUsecase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			linkParam:  fixedCardLinkH.String(),
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("DeleteCard", mock.Anything, domain.DeleteCardRequest{
					UserLink: fixedUserLinkH,
					CardLink: fixedCardLinkH,
				}).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			linkParam:          fixedCardLinkH.String(),
			setContext:         false,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidPathParam",
			linkParam:          "bad-uuid",
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "NotFound",
			linkParam:  fixedCardLinkH.String(),
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("DeleteCard", mock.Anything, mock.Anything).Return(common.ErrorCardNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:       "PermissionDenied",
			linkParam:  fixedCardLinkH.String(),
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("DeleteCard", mock.Anything, mock.Anything).Return(common.ErrorPermissionDenied)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:       "InternalError",
			linkParam:  fixedCardLinkH.String(),
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("DeleteCard", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardUsecase)
			tc.mockBehavior(m)

			req := httptest.NewRequest(http.MethodDelete, "/cards/"+tc.linkParam, nil)
			req = mux.SetURLVars(req, map[string]string{"link": tc.linkParam})
			if tc.setContext {
				req = withCardUserCtx(req)
			}
			rr := httptest.NewRecorder()

			newTestCardHandler(m).DeleteCard(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestCardHandler_UpdateCard(t *testing.T) {
	validReq := dto.UpdateCardRequest{Title: "New Title", Description: "New Desc"}

	tests := []struct {
		name               string
		linkParam          string
		request            any
		setContext         bool
		mockBehavior       func(m *mockCardUsecase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			linkParam:  fixedCardLinkH.String(),
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("UpdateCard", mock.Anything, domain.UpdateCardRequest{
					UserLink:    fixedUserLinkH,
					CardLink:    fixedCardLinkH,
					Title:       "New Title",
					Description: "New Desc",
				}).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			linkParam:          fixedCardLinkH.String(),
			request:            validReq,
			setContext:         false,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidPathParam",
			linkParam:          "bad-uuid",
			request:            validReq,
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "InvalidJSON",
			linkParam:          fixedCardLinkH.String(),
			request:            "{bad}",
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "TitleTooLong",
			linkParam:          fixedCardLinkH.String(),
			request:            dto.UpdateCardRequest{Title: strings.Repeat("a", 200)},
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "DescriptionTooLong",
			linkParam:          fixedCardLinkH.String(),
			request:            dto.UpdateCardRequest{Title: "ok", Description: strings.Repeat("d", 600)},
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "NotFound",
			linkParam:  fixedCardLinkH.String(),
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("UpdateCard", mock.Anything, mock.Anything).Return(common.ErrorCardNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:       "PermissionDenied",
			linkParam:  fixedCardLinkH.String(),
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("UpdateCard", mock.Anything, mock.Anything).Return(common.ErrorPermissionDenied)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:       "InternalError",
			linkParam:  fixedCardLinkH.String(),
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("UpdateCard", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardUsecase)
			tc.mockBehavior(m)

			var bodyReader *bytes.Reader
			if strBody, ok := tc.request.(string); ok {
				bodyReader = bytes.NewReader([]byte(strBody))
			} else {
				b, _ := json.Marshal(tc.request)
				bodyReader = bytes.NewReader(b)
			}

			req := httptest.NewRequest(http.MethodPut, "/cards/"+tc.linkParam, bodyReader)
			req = mux.SetURLVars(req, map[string]string{"link": tc.linkParam})
			if tc.setContext {
				req = withCardUserCtx(req)
			}
			rr := httptest.NewRecorder()

			newTestCardHandler(m).UpdateCard(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestCardHandler_ReorderCards(t *testing.T) {
	sectionLinkStr := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb").String()
	validReq := dto.ReorderCardsRequest{SectionLink: sectionLinkStr, Position: 2}

	tests := []struct {
		name               string
		linkParam          string
		request            any
		setContext         bool
		mockBehavior       func(m *mockCardUsecase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			linkParam:  fixedCardLinkH.String(),
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("ReorderCards", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			linkParam:          fixedCardLinkH.String(),
			request:            validReq,
			setContext:         false,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidPathParam",
			linkParam:          "bad-uuid",
			request:            validReq,
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "InvalidJSON",
			linkParam:          fixedCardLinkH.String(),
			request:            "{bad}",
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "InvalidSectionLink",
			linkParam:          fixedCardLinkH.String(),
			request:            dto.ReorderCardsRequest{SectionLink: "not-a-uuid", Position: 1},
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "NotFound",
			linkParam:  fixedCardLinkH.String(),
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("ReorderCards", mock.Anything, mock.Anything).Return(common.ErrorCardNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:       "PermissionDenied",
			linkParam:  fixedCardLinkH.String(),
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("ReorderCards", mock.Anything, mock.Anything).Return(common.ErrorPermissionDenied)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:       "InternalError",
			linkParam:  fixedCardLinkH.String(),
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("ReorderCards", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardUsecase)
			tc.mockBehavior(m)

			var bodyReader *bytes.Reader
			if strBody, ok := tc.request.(string); ok {
				bodyReader = bytes.NewReader([]byte(strBody))
			} else {
				b, _ := json.Marshal(tc.request)
				bodyReader = bytes.NewReader(b)
			}

			req := httptest.NewRequest(http.MethodPatch, "/cards/"+tc.linkParam+"/reorder", bodyReader)
			req = mux.SetURLVars(req, map[string]string{"link": tc.linkParam})
			if tc.setContext {
				req = withCardUserCtx(req)
			}
			rr := httptest.NewRecorder()

			newTestCardHandler(m).ReorderCards(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestCardHandler_CreateCard(t *testing.T) {
	sectionLinkStr := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb").String()
	validReq := dto.CreateCardRequest{
		SectionLink: sectionLinkStr,
		Title:       "New Card",
		Description: "Desc",
	}

	tests := []struct {
		name               string
		request            any
		setContext         bool
		mockBehavior       func(m *mockCardUsecase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("CreateCard", mock.Anything, mock.Anything).Return(domain.CreateCardResponse{
					CardLink:    fixedCardLinkH,
					SectionLink: uuid.MustParse(sectionLinkStr),
					Position:    1,
				}, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			request:            validReq,
			setContext:         false,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidJSON",
			request:            "{bad}",
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "TitleTooLong",
			request:            dto.CreateCardRequest{SectionLink: sectionLinkStr, Title: strings.Repeat("a", 200)},
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "InvalidSectionLink",
			request:            dto.CreateCardRequest{SectionLink: "bad-uuid", Title: "Title"},
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "SectionNotFound",
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("CreateCard", mock.Anything, mock.Anything).Return(domain.CreateCardResponse{}, common.ErrorSectionNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:       "PermissionDenied",
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("CreateCard", mock.Anything, mock.Anything).Return(domain.CreateCardResponse{}, common.ErrorPermissionDenied)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:       "TaskLimitReached",
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("CreateCard", mock.Anything, mock.Anything).Return(domain.CreateCardResponse{}, common.ErrorTaskLimitReached)
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "InternalError",
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("CreateCard", mock.Anything, mock.Anything).Return(domain.CreateCardResponse{}, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardUsecase)
			tc.mockBehavior(m)

			var bodyReader *bytes.Reader
			if strBody, ok := tc.request.(string); ok {
				bodyReader = bytes.NewReader([]byte(strBody))
			} else {
				b, _ := json.Marshal(tc.request)
				bodyReader = bytes.NewReader(b)
			}

			req := httptest.NewRequest(http.MethodPost, "/cards", bodyReader)
			if tc.setContext {
				req = withCardUserCtx(req)
			}
			rr := httptest.NewRecorder()

			newTestCardHandler(m).CreateCard(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestCardHandler_GetComments(t *testing.T) {
	tests := []struct {
		name               string
		linkParam          string
		setContext         bool
		mockBehavior       func(m *mockCardUsecase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			linkParam:  fixedCardLinkH.String(),
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("GetComments", mock.Anything, domain.GetCommentsRequest{
					UserLink: fixedUserLinkH,
					CardLink: fixedCardLinkH,
				}).Return(domain.GetCommentsResponse{CommentsInfo: []domain.CommentInfo{}}, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			linkParam:          fixedCardLinkH.String(),
			setContext:         false,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidPathParam",
			linkParam:          "bad-uuid",
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "NotFound",
			linkParam:  fixedCardLinkH.String(),
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("GetComments", mock.Anything, mock.Anything).Return(domain.GetCommentsResponse{}, common.ErrorCardNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:       "PermissionDenied",
			linkParam:  fixedCardLinkH.String(),
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("GetComments", mock.Anything, mock.Anything).Return(domain.GetCommentsResponse{}, common.ErrorPermissionDenied)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:       "InternalError",
			linkParam:  fixedCardLinkH.String(),
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("GetComments", mock.Anything, mock.Anything).Return(domain.GetCommentsResponse{}, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardUsecase)
			tc.mockBehavior(m)

			req := httptest.NewRequest(http.MethodGet, "/cards/"+tc.linkParam+"/comments", nil)
			req = mux.SetURLVars(req, map[string]string{"link": tc.linkParam})
			if tc.setContext {
				req = withCardUserCtx(req)
			}
			rr := httptest.NewRecorder()

			newTestCardHandler(m).GetComments(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestCardHandler_CreateComment(t *testing.T) {
	validReq := dto.CreateCommentRequest{Text: "hello"}

	tests := []struct {
		name               string
		linkParam          string
		request            any
		setContext         bool
		mockBehavior       func(m *mockCardUsecase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			linkParam:  fixedCardLinkH.String(),
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("CreateComment", mock.Anything, domain.CreateCommentRequest{
					UserLink: fixedUserLinkH,
					CardLink: fixedCardLinkH,
					Text:     "hello",
				}).Return(domain.CreateCommentResponse{CommentLink: fixedCommentLinkH}, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			linkParam:          fixedCardLinkH.String(),
			request:            validReq,
			setContext:         false,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidPathParam",
			linkParam:          "bad-uuid",
			request:            validReq,
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "InvalidJSON",
			linkParam:          fixedCardLinkH.String(),
			request:            "{bad}",
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "NotFound",
			linkParam:  fixedCardLinkH.String(),
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("CreateComment", mock.Anything, mock.Anything).Return(domain.CreateCommentResponse{}, common.ErrorCardNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:       "PermissionDenied",
			linkParam:  fixedCardLinkH.String(),
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("CreateComment", mock.Anything, mock.Anything).Return(domain.CreateCommentResponse{}, common.ErrorPermissionDenied)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:       "InternalError",
			linkParam:  fixedCardLinkH.String(),
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("CreateComment", mock.Anything, mock.Anything).Return(domain.CreateCommentResponse{}, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardUsecase)
			tc.mockBehavior(m)

			var bodyReader *bytes.Reader
			if strBody, ok := tc.request.(string); ok {
				bodyReader = bytes.NewReader([]byte(strBody))
			} else {
				b, _ := json.Marshal(tc.request)
				bodyReader = bytes.NewReader(b)
			}

			req := httptest.NewRequest(http.MethodPost, "/cards/"+tc.linkParam+"/comments", bodyReader)
			req = mux.SetURLVars(req, map[string]string{"link": tc.linkParam})
			if tc.setContext {
				req = withCardUserCtx(req)
			}
			rr := httptest.NewRecorder()

			newTestCardHandler(m).CreateComment(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestCardHandler_DeleteComment(t *testing.T) {
	tests := []struct {
		name               string
		commentParam       string
		setContext         bool
		mockBehavior       func(m *mockCardUsecase)
		expectedStatusCode int
	}{
		{
			name:         "Success",
			commentParam: fixedCommentLinkH.String(),
			setContext:   true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("DeleteComment", mock.Anything, domain.DeleteCommentRequest{
					UserLink:    fixedUserLinkH,
					CommentLink: fixedCommentLinkH,
				}).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			commentParam:       fixedCommentLinkH.String(),
			setContext:         false,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidPathParam",
			commentParam:       "bad-uuid",
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:         "NotFound",
			commentParam: fixedCommentLinkH.String(),
			setContext:   true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("DeleteComment", mock.Anything, mock.Anything).Return(common.ErrorCommentNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:         "PermissionDenied",
			commentParam: fixedCommentLinkH.String(),
			setContext:   true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("DeleteComment", mock.Anything, mock.Anything).Return(common.ErrorPermissionDenied)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:         "InternalError",
			commentParam: fixedCommentLinkH.String(),
			setContext:   true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("DeleteComment", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardUsecase)
			tc.mockBehavior(m)

			req := httptest.NewRequest(http.MethodDelete, "/comments/"+tc.commentParam, nil)
			req = mux.SetURLVars(req, map[string]string{"comment_link": tc.commentParam})
			if tc.setContext {
				req = withCardUserCtx(req)
			}
			rr := httptest.NewRecorder()

			newTestCardHandler(m).DeleteComment(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestCardHandler_UpdateComment(t *testing.T) {
	validReq := dto.UpdateCommentRequest{Text: "updated text"}

	tests := []struct {
		name               string
		commentParam       string
		request            any
		setContext         bool
		mockBehavior       func(m *mockCardUsecase)
		expectedStatusCode int
	}{
		{
			name:         "Success",
			commentParam: fixedCommentLinkH.String(),
			request:      validReq,
			setContext:   true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("UpdateComment", mock.Anything, domain.UpdateCommentRequest{
					UserLink:    fixedUserLinkH,
					CommentLink: fixedCommentLinkH,
					Text:        "updated text",
				}).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			commentParam:       fixedCommentLinkH.String(),
			request:            validReq,
			setContext:         false,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidPathParam",
			commentParam:       "bad-uuid",
			request:            validReq,
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "InvalidJSON",
			commentParam:       fixedCommentLinkH.String(),
			request:            "{bad}",
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:         "NotFound",
			commentParam: fixedCommentLinkH.String(),
			request:      validReq,
			setContext:   true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("UpdateComment", mock.Anything, mock.Anything).Return(common.ErrorCommentNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:         "PermissionDenied",
			commentParam: fixedCommentLinkH.String(),
			request:      validReq,
			setContext:   true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("UpdateComment", mock.Anything, mock.Anything).Return(common.ErrorPermissionDenied)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:         "InternalError",
			commentParam: fixedCommentLinkH.String(),
			request:      validReq,
			setContext:   true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("UpdateComment", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardUsecase)
			tc.mockBehavior(m)

			var bodyReader *bytes.Reader
			if strBody, ok := tc.request.(string); ok {
				bodyReader = bytes.NewReader([]byte(strBody))
			} else {
				b, _ := json.Marshal(tc.request)
				bodyReader = bytes.NewReader(b)
			}

			req := httptest.NewRequest(http.MethodPut, "/comments/"+tc.commentParam, bodyReader)
			req = mux.SetURLVars(req, map[string]string{"comment_link": tc.commentParam})
			if tc.setContext {
				req = withCardUserCtx(req)
			}
			rr := httptest.NewRecorder()

			newTestCardHandler(m).UpdateComment(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestCardHandler_CreateSubtask(t *testing.T) {
	validReq := dto.CreateSubtaskRequest{Description: "do something"}

	tests := []struct {
		name               string
		linkParam          string
		request            any
		setContext         bool
		mockBehavior       func(m *mockCardUsecase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			linkParam:  fixedCardLinkH.String(),
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("CreateSubtask", mock.Anything, domain.CreateSubtaskRequest{
					UserLink:    fixedUserLinkH,
					CardLink:    fixedCardLinkH,
					Description: "do something",
				}).Return(domain.SubtaskInfo{
					SubtaskLink: fixedSubtaskLinkH,
					Description: "do something",
					IsDone:      false,
					Position:    1,
				}, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			linkParam:          fixedCardLinkH.String(),
			request:            validReq,
			setContext:         false,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidPathParam",
			linkParam:          "bad-uuid",
			request:            validReq,
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "InvalidJSON",
			linkParam:          fixedCardLinkH.String(),
			request:            "{bad}",
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "NotFound",
			linkParam:  fixedCardLinkH.String(),
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("CreateSubtask", mock.Anything, mock.Anything).Return(domain.SubtaskInfo{}, common.ErrorCardNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:       "PermissionDenied",
			linkParam:  fixedCardLinkH.String(),
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("CreateSubtask", mock.Anything, mock.Anything).Return(domain.SubtaskInfo{}, common.ErrorPermissionDenied)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:       "InternalError",
			linkParam:  fixedCardLinkH.String(),
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("CreateSubtask", mock.Anything, mock.Anything).Return(domain.SubtaskInfo{}, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardUsecase)
			tc.mockBehavior(m)

			var bodyReader *bytes.Reader
			if strBody, ok := tc.request.(string); ok {
				bodyReader = bytes.NewReader([]byte(strBody))
			} else {
				b, _ := json.Marshal(tc.request)
				bodyReader = bytes.NewReader(b)
			}

			req := httptest.NewRequest(http.MethodPost, "/cards/"+tc.linkParam+"/subtasks", bodyReader)
			req = mux.SetURLVars(req, map[string]string{"link": tc.linkParam})
			if tc.setContext {
				req = withCardUserCtx(req)
			}
			rr := httptest.NewRecorder()

			newTestCardHandler(m).CreateSubtask(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestCardHandler_UpdateSubtask(t *testing.T) {
	validReq := dto.UpdateSubtaskRequest{IsDone: true, Description: "updated"}

	tests := []struct {
		name               string
		subtaskParam       string
		request            any
		setContext         bool
		mockBehavior       func(m *mockCardUsecase)
		expectedStatusCode int
	}{
		{
			name:         "Success",
			subtaskParam: fixedSubtaskLinkH.String(),
			request:      validReq,
			setContext:   true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("UpdateSubtask", mock.Anything, domain.UpdateSubtaskRequest{
					UserLink:    fixedUserLinkH,
					SubtaskLink: fixedSubtaskLinkH,
					IsDone:      true,
					Description: "updated",
				}).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			subtaskParam:       fixedSubtaskLinkH.String(),
			request:            validReq,
			setContext:         false,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidPathParam",
			subtaskParam:       "bad-uuid",
			request:            validReq,
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "InvalidJSON",
			subtaskParam:       fixedSubtaskLinkH.String(),
			request:            "{bad}",
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:         "NotFound",
			subtaskParam: fixedSubtaskLinkH.String(),
			request:      validReq,
			setContext:   true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("UpdateSubtask", mock.Anything, mock.Anything).Return(common.ErrorSubtaskNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:         "PermissionDenied",
			subtaskParam: fixedSubtaskLinkH.String(),
			request:      validReq,
			setContext:   true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("UpdateSubtask", mock.Anything, mock.Anything).Return(common.ErrorPermissionDenied)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:         "InternalError",
			subtaskParam: fixedSubtaskLinkH.String(),
			request:      validReq,
			setContext:   true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("UpdateSubtask", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardUsecase)
			tc.mockBehavior(m)

			var bodyReader *bytes.Reader
			if strBody, ok := tc.request.(string); ok {
				bodyReader = bytes.NewReader([]byte(strBody))
			} else {
				b, _ := json.Marshal(tc.request)
				bodyReader = bytes.NewReader(b)
			}

			req := httptest.NewRequest(http.MethodPut, "/subtasks/"+tc.subtaskParam, bodyReader)
			req = mux.SetURLVars(req, map[string]string{"subtask_link": tc.subtaskParam})
			if tc.setContext {
				req = withCardUserCtx(req)
			}
			rr := httptest.NewRecorder()

			newTestCardHandler(m).UpdateSubtask(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestCardHandler_DeleteSubtask(t *testing.T) {
	tests := []struct {
		name               string
		subtaskParam       string
		setContext         bool
		mockBehavior       func(m *mockCardUsecase)
		expectedStatusCode int
	}{
		{
			name:         "Success",
			subtaskParam: fixedSubtaskLinkH.String(),
			setContext:   true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("DeleteSubtask", mock.Anything, domain.DeleteSubtaskRequest{
					UserLink:    fixedUserLinkH,
					SubtaskLink: fixedSubtaskLinkH,
				}).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			subtaskParam:       fixedSubtaskLinkH.String(),
			setContext:         false,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidPathParam",
			subtaskParam:       "bad-uuid",
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:         "NotFound",
			subtaskParam: fixedSubtaskLinkH.String(),
			setContext:   true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("DeleteSubtask", mock.Anything, mock.Anything).Return(common.ErrorSubtaskNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:         "PermissionDenied",
			subtaskParam: fixedSubtaskLinkH.String(),
			setContext:   true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("DeleteSubtask", mock.Anything, mock.Anything).Return(common.ErrorPermissionDenied)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:         "InternalError",
			subtaskParam: fixedSubtaskLinkH.String(),
			setContext:   true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("DeleteSubtask", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardUsecase)
			tc.mockBehavior(m)

			req := httptest.NewRequest(http.MethodDelete, "/subtasks/"+tc.subtaskParam, nil)
			req = mux.SetURLVars(req, map[string]string{"subtask_link": tc.subtaskParam})
			if tc.setContext {
				req = withCardUserCtx(req)
			}
			rr := httptest.NewRecorder()

			newTestCardHandler(m).DeleteSubtask(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func buildMultipartBody(fieldKey, filename string, content []byte) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	part, _ := w.CreateFormFile(fieldKey, filename)
	_, _ = part.Write(content)
	_ = w.Close()
	return body, w.FormDataContentType()
}

func TestCardHandler_CreateAttachment(t *testing.T) {
	attachmentResult := domain.AttachmentInfo{
		AttachmentLink: fixedAttachmentLinkH,
		DisplayName:    "photo.png",
		Path:           "https://s3.example.com/photo.png",
		Position:       1,
	}

	cfg := CardConfig{
		MaxLenTitle:                128,
		MaxLenDescription:          500,
		MultipartAttachmentFileKey: "attachment",
		MaxAttachmentSize:          10 * 1024 * 1024,
	}

	tests := []struct {
		name               string
		cardParam          string
		setContext         bool
		mockBehavior       func(m *mockCardUsecase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			cardParam:  fixedCardLinkH.String(),
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("CreateAttachment", mock.Anything, mock.Anything).Return(attachmentResult, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "No auth context",
			cardParam:          fixedCardLinkH.String(),
			setContext:         false,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:       "Permission denied",
			cardParam:  fixedCardLinkH.String(),
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("CreateAttachment", mock.Anything, mock.Anything).Return(domain.AttachmentInfo{}, common.ErrorPermissionDenied)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:       "Attachment limit reached",
			cardParam:  fixedCardLinkH.String(),
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("CreateAttachment", mock.Anything, mock.Anything).Return(domain.AttachmentInfo{}, common.ErrorAttachmentLimitReached)
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "Internal error",
			cardParam:  fixedCardLinkH.String(),
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("CreateAttachment", mock.Anything, mock.Anything).Return(domain.AttachmentInfo{}, errors.New("s3 error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardUsecase)
			if tc.mockBehavior != nil {
				tc.mockBehavior(m)
			}

			body, contentType := buildMultipartBody("attachment", "photo.png", []byte("fake image data"))
			req := httptest.NewRequest(http.MethodPost, "/cards/"+tc.cardParam+"/attachments", body)
			req.Header.Set("Content-Type", contentType)
			req = mux.SetURLVars(req, map[string]string{"link": tc.cardParam})
			if tc.setContext {
				req = withCardUserCtx(req)
			}
			rr := httptest.NewRecorder()

			NewCard(m, cfg).CreateAttachment(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestCardHandler_DeleteAttachment(t *testing.T) {
	tests := []struct {
		name               string
		attachmentParam    string
		setContext         bool
		mockBehavior       func(m *mockCardUsecase)
		expectedStatusCode int
	}{
		{
			name:            "Success",
			attachmentParam: fixedAttachmentLinkH.String(),
			setContext:      true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("DeleteAttachment", mock.Anything, domain.DeleteAttachmentRequest{
					UserLink:       fixedUserLinkH,
					AttachmentLink: fixedAttachmentLinkH,
				}).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "No auth context",
			attachmentParam:    fixedAttachmentLinkH.String(),
			setContext:         false,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "Invalid attachment uuid",
			attachmentParam:    "not-a-uuid",
			setContext:         true,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:            "Attachment not found",
			attachmentParam: fixedAttachmentLinkH.String(),
			setContext:      true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("DeleteAttachment", mock.Anything, mock.Anything).Return(common.ErrorAttachmentNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:            "Permission denied",
			attachmentParam: fixedAttachmentLinkH.String(),
			setContext:      true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("DeleteAttachment", mock.Anything, mock.Anything).Return(common.ErrorPermissionDenied)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:            "Internal error",
			attachmentParam: fixedAttachmentLinkH.String(),
			setContext:      true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("DeleteAttachment", mock.Anything, mock.Anything).Return(errors.New("s3 error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardUsecase)
			if tc.mockBehavior != nil {
				tc.mockBehavior(m)
			}

			req := httptest.NewRequest(http.MethodDelete, "/attachments/"+tc.attachmentParam, nil)
			req = mux.SetURLVars(req, map[string]string{"attachment_link": tc.attachmentParam})
			if tc.setContext {
				req = withCardUserCtx(req)
			}
			rr := httptest.NewRecorder()

			newTestCardHandler(m).DeleteAttachment(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestCardHandler_UpdateStatusTask(t *testing.T) {
	validReq := dto.NewStatusTask{Done: true}

	tests := []struct {
		name               string
		linkParam          string
		request            any
		setContext         bool
		mockBehavior       func(m *mockCardUsecase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			linkParam:  fixedCardLinkH.String(),
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("UpdateStatusTask", mock.Anything, domain.NewStatusTask{
					UserLink: fixedUserLinkH,
					CardLink: fixedCardLinkH,
					Status:   true,
				}).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			linkParam:          fixedCardLinkH.String(),
			request:            validReq,
			setContext:         false,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidPathParam",
			linkParam:          "bad-uuid",
			request:            validReq,
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "InvalidJSON",
			linkParam:          fixedCardLinkH.String(),
			request:            "{bad}",
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "NotFound",
			linkParam:  fixedCardLinkH.String(),
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("UpdateStatusTask", mock.Anything, mock.Anything).Return(common.ErrorCardNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:       "PermissionDenied",
			linkParam:  fixedCardLinkH.String(),
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("UpdateStatusTask", mock.Anything, mock.Anything).Return(common.ErrorPermissionDenied)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:       "InternalError",
			linkParam:  fixedCardLinkH.String(),
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("UpdateStatusTask", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardUsecase)
			tc.mockBehavior(m)

			var bodyReader *bytes.Reader
			if strBody, ok := tc.request.(string); ok {
				bodyReader = bytes.NewReader([]byte(strBody))
			} else {
				b, _ := json.Marshal(tc.request)
				bodyReader = bytes.NewReader(b)
			}

			req := httptest.NewRequest(http.MethodPatch, "/cards/"+tc.linkParam+"/status", bodyReader)
			req = mux.SetURLVars(req, map[string]string{"link": tc.linkParam})
			if tc.setContext {
				req = withCardUserCtx(req)
			}
			rr := httptest.NewRecorder()

			newTestCardHandler(m).UpdateStatusTask(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestCardHandler_UpdateTimeLine(t *testing.T) {
	validReq := dto.NewTimeLine{
		Start:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		DeadLine: time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
	}

	invalidTimeLineReq := dto.NewTimeLine{
		Start:    time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
		DeadLine: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name               string
		linkParam          string
		request            any
		setContext         bool
		mockBehavior       func(m *mockCardUsecase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			linkParam:  fixedCardLinkH.String(),
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("UpdateTimeLine", mock.Anything, domain.NewTimeLine{
					UserLink: fixedUserLinkH,
					CardLink: fixedCardLinkH,
					DeadLine: validReq.DeadLine,
					Start:    validReq.Start,
				}).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			linkParam:          fixedCardLinkH.String(),
			request:            validReq,
			setContext:         false,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidPathParam",
			linkParam:          "bad-uuid",
			request:            validReq,
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "InvalidJSON",
			linkParam:          fixedCardLinkH.String(),
			request:            "{bad}",
			setContext:         true,
			mockBehavior:       func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "StartAfterDeadline",
			linkParam:  fixedCardLinkH.String(),
			request:    invalidTimeLineReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "NotFound",
			linkParam:  fixedCardLinkH.String(),
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("UpdateTimeLine", mock.Anything, mock.Anything).Return(common.ErrorCardNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:       "PermissionDenied",
			linkParam:  fixedCardLinkH.String(),
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("UpdateTimeLine", mock.Anything, mock.Anything).Return(common.ErrorPermissionDenied)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:       "InternalError",
			linkParam:  fixedCardLinkH.String(),
			request:    validReq,
			setContext: true,
			mockBehavior: func(m *mockCardUsecase) {
				m.On("UpdateTimeLine", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardUsecase)
			tc.mockBehavior(m)

			var bodyReader *bytes.Reader
			if strBody, ok := tc.request.(string); ok {
				bodyReader = bytes.NewReader([]byte(strBody))
			} else {
				b, _ := json.Marshal(tc.request)
				bodyReader = bytes.NewReader(b)
			}

			req := httptest.NewRequest(http.MethodPatch, "/cards/"+tc.linkParam+"/timeline", bodyReader)
			req = mux.SetURLVars(req, map[string]string{"link": tc.linkParam})
			if tc.setContext {
				req = withCardUserCtx(req)
			}
			rr := httptest.NewRecorder()

			newTestCardHandler(m).UpdateTimeLine(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}
