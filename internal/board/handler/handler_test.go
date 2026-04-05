package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/handler"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/handler/dto"
	mocks "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/handler/mock_board_srv"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/service/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
)

func reqWithUser(req *http.Request, userLink uuid.UUID) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, userLink)
	return req.WithContext(ctx)
}

func TestBoardHandler_GetBoards(t *testing.T) {
	userLink := uuid.New()
	boardsInfo := []serviceDto.BoardInfo{
		{Link: uuid.New(), Name: "Board 1", CreatedAt: time.Now()},
		{Link: uuid.New(), Name: "Board 2", CreatedAt: time.Now()},
	}

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		setupMock      func(m *mocks.BoardService)
		expectedStatus int
	}{
		{
			name: "success get boards",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/boards", nil)
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetBoards", mock.Anything, userLink).Return(boardsInfo, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "unauthorized - no context",
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/boards", nil)
			},
			setupMock:      func(m *mocks.BoardService) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "service error",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/boards", nil)
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetBoards", mock.Anything, userLink).Return(nil, errors.New("db error")).Once()
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv)
			req := test.setupRequest()
			w := httptest.NewRecorder()

			h.GetBoards(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestBoardHandler_CreateBoard(t *testing.T) {
	userLink := uuid.New()
	createReq := dto.CreateBoardRequest{Name: "New Board", Description: "Desc"}
	boardInfo := serviceDto.BoardInfo{Link: uuid.New(), Name: "New Board"}

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		setupMock      func(m *mocks.BoardService)
		expectedStatus int
	}{
		{
			name: "success create board",
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(createReq)
				req := httptest.NewRequest(http.MethodPost, "/boards", bytes.NewBuffer(body))
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("CreateBoard", mock.Anything, mock.AnythingOfType("dto.NewBoardInfo"), userLink).
					Return(boardInfo, nil).Once()
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "unauthorized",
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(createReq)
				return httptest.NewRequest(http.MethodPost, "/boards", bytes.NewBuffer(body))
			},
			setupMock:      func(m *mocks.BoardService) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid json",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/boards", bytes.NewBuffer([]byte("{invalid}")))
				return reqWithUser(req, userLink)
			},
			setupMock:      func(m *mocks.BoardService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(createReq)
				req := httptest.NewRequest(http.MethodPost, "/boards", bytes.NewBuffer(body))
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("CreateBoard", mock.Anything, mock.AnythingOfType("dto.NewBoardInfo"), userLink).
					Return(serviceDto.BoardInfo{}, errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv)
			req := test.setupRequest()
			w := httptest.NewRecorder()

			h.CreateBoard(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestBoardHandler_DeleteBoard(t *testing.T) {
	userLink := uuid.New()
	deleteReq := dto.DeleteBoardRequest{Link: uuid.New()}

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		setupMock      func(m *mocks.BoardService)
		expectedStatus int
	}{
		{
			name: "success delete board",
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(deleteReq)
				req := httptest.NewRequest(http.MethodDelete, "/boards", bytes.NewBuffer(body))
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("DeleteBoard", mock.Anything, deleteReq.Link, userLink).Return(nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "forbidden (action denied)",
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(deleteReq)
				req := httptest.NewRequest(http.MethodDelete, "/boards", bytes.NewBuffer(body))
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("DeleteBoard", mock.Anything, deleteReq.Link, userLink).Return(common.ErrActionDenied).Once()
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "not found",
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(deleteReq)
				req := httptest.NewRequest(http.MethodDelete, "/boards", bytes.NewBuffer(body))
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("DeleteBoard", mock.Anything, deleteReq.Link, userLink).Return(common.ErrBoardNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "internal server error",
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(deleteReq)
				req := httptest.NewRequest(http.MethodDelete, "/boards", bytes.NewBuffer(body))
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("DeleteBoard", mock.Anything, deleteReq.Link, userLink).Return(errors.New("db crash")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv)
			req := test.setupRequest()
			w := httptest.NewRecorder()

			h.DeleteBoard(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestBoardHandler_UpdateBoard(t *testing.T) {
	userLink := uuid.New()
	updateReq := dto.UpdateBoardRequest{Link: uuid.New(), Name: "Updated Name"}

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		setupMock      func(m *mocks.BoardService)
		expectedStatus int
	}{
		{
			name: "success update board",
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(updateReq)
				req := httptest.NewRequest(http.MethodPut, "/boards", bytes.NewBuffer(body))
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateBoard", mock.Anything, mock.AnythingOfType("dto.UpdateBoardInfo"), userLink).Return(nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid json",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPut, "/boards", bytes.NewBuffer([]byte(`{bad_json`)))
				return reqWithUser(req, userLink)
			},
			setupMock:      func(m *mocks.BoardService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "forbidden (action denied)",
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(updateReq)
				req := httptest.NewRequest(http.MethodPut, "/boards", bytes.NewBuffer(body))
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateBoard", mock.Anything, mock.AnythingOfType("dto.UpdateBoardInfo"), userLink).Return(common.ErrActionDenied).Once()
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "not found",
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(updateReq)
				req := httptest.NewRequest(http.MethodPut, "/boards", bytes.NewBuffer(body))
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateBoard", mock.Anything, mock.AnythingOfType("dto.UpdateBoardInfo"), userLink).Return(common.ErrBoardNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv)
			req := test.setupRequest()
			w := httptest.NewRecorder()

			h.UpdateBoard(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)
			mockSrv.AssertExpectations(t)
		})
	}
}
