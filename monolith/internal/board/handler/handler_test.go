package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/board/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/board/handler"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/board/handler/dto"
	mocks "github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/board/handler/mock_board_srv"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/board/service/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/middleware"
)

var testConf = handler.Config{
	MaxBackgroundSize:          10 << 20,
	MultipartBackgroundFileKey: "background",
}

func reqWithUser(req *http.Request, userLink uuid.UUID) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, userLink)
	return req.WithContext(ctx)
}

func TestGetBoards(t *testing.T) {
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
			name: "Success get boards",
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
			name: "Error unauthorized",
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/boards", nil)
			},
			setupMock:      func(m *mocks.BoardService) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Error internal server",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/boards", nil)
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetBoards", mock.Anything, userLink).Return(nil, errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			req := test.setupRequest()
			w := httptest.NewRecorder()

			h.GetBoards(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestCreateBoard(t *testing.T) {
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
			name: "Success create board",
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(createReq)
				req := httptest.NewRequest(http.MethodPost, "/boards", bytes.NewBuffer(body))
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("CreateBoard", mock.Anything, mock.Anything, userLink).
					Return(boardInfo, nil).Once()
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Error unauthorized",
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(createReq)
				return httptest.NewRequest(http.MethodPost, "/boards", bytes.NewBuffer(body))
			},
			setupMock:      func(m *mocks.BoardService) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Error invalid json",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/boards", bytes.NewBuffer([]byte("{invalid}")))
				return reqWithUser(req, userLink)
			},
			setupMock:      func(m *mocks.BoardService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error missing required field",
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(createReq)
				req := httptest.NewRequest(http.MethodPost, "/boards", bytes.NewBuffer(body))
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("CreateBoard", mock.Anything, mock.Anything, userLink).
					Return(serviceDto.BoardInfo{}, common.ErrorNotNullValue).Once()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error invalid board data",
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(createReq)
				req := httptest.NewRequest(http.MethodPost, "/boards", bytes.NewBuffer(body))
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("CreateBoard", mock.Anything, mock.Anything, userLink).
					Return(serviceDto.BoardInfo{}, common.ErrorInvalidBoardData).Once()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error invalid board reference",
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(createReq)
				req := httptest.NewRequest(http.MethodPost, "/boards", bytes.NewBuffer(body))
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("CreateBoard", mock.Anything, mock.Anything, userLink).
					Return(serviceDto.BoardInfo{}, common.ErrorInvalidBoardReference).Once()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error user already member",
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(createReq)
				req := httptest.NewRequest(http.MethodPost, "/boards", bytes.NewBuffer(body))
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("CreateBoard", mock.Anything, mock.Anything, userLink).
					Return(serviceDto.BoardInfo{}, common.ErrorUserAlreadyMember).Once()
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "Error internal server",
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(createReq)
				req := httptest.NewRequest(http.MethodPost, "/boards", bytes.NewBuffer(body))
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("CreateBoard", mock.Anything, mock.Anything, userLink).
					Return(serviceDto.BoardInfo{}, errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			req := test.setupRequest()
			w := httptest.NewRecorder()

			h.CreateBoard(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestDeleteBoard(t *testing.T) {
	userLink := uuid.New()
	boardLink := uuid.New()

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		setupMock      func(m *mocks.BoardService)
		expectedStatus int
	}{
		{
			name: "Success delete board",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/boards/{link}", nil)
				req = mux.SetURLVars(req, map[string]string{"link": boardLink.String()})
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("DeleteBoard", mock.Anything, boardLink, userLink).Return(nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Error unauthorized",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/boards/{link}", nil)
				return mux.SetURLVars(req, map[string]string{"link": boardLink.String()})
			},
			setupMock:      func(m *mocks.BoardService) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Error missing board link var",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/boards/", nil)
				return reqWithUser(req, userLink)
			},
			setupMock:      func(m *mocks.BoardService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error invalid board link format",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/boards/invalid-uuid", nil)
				req = mux.SetURLVars(req, map[string]string{"link": "invalid-uuid"})
				return reqWithUser(req, userLink)
			},
			setupMock:      func(m *mocks.BoardService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error forbidden",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/boards/{link}", nil)
				req = mux.SetURLVars(req, map[string]string{"link": boardLink.String()})
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("DeleteBoard", mock.Anything, boardLink, userLink).Return(common.ErrActionDenied).Once()
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "Error not found",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/boards/{link}", nil)
				req = mux.SetURLVars(req, map[string]string{"link": boardLink.String()})
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("DeleteBoard", mock.Anything, boardLink, userLink).Return(common.ErrBoardNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Error internal server",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/boards/{link}", nil)
				req = mux.SetURLVars(req, map[string]string{"link": boardLink.String()})
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("DeleteBoard", mock.Anything, boardLink, userLink).Return(errors.New("db crash")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			req := test.setupRequest()
			w := httptest.NewRecorder()

			h.DeleteBoard(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestUpdateBoard(t *testing.T) {
	userLink := uuid.New()
	boardLink := uuid.New()
	updateReq := dto.UpdateBoardRequest{Name: "Updated Name"}

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		setupMock      func(m *mocks.BoardService)
		expectedStatus int
	}{
		{
			name: "Success update board",
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(updateReq)
				req := httptest.NewRequest(http.MethodPut, "/boards/{link}", bytes.NewBuffer(body))
				req = mux.SetURLVars(req, map[string]string{"link": boardLink.String()})
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateBoard", mock.Anything, mock.Anything, userLink).Return(nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Error invalid json",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPut, "/boards/{link}", bytes.NewBuffer([]byte(`{bad_json`)))
				req = mux.SetURLVars(req, map[string]string{"link": boardLink.String()})
				return reqWithUser(req, userLink)
			},
			setupMock:      func(m *mocks.BoardService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error forbidden",
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(updateReq)
				req := httptest.NewRequest(http.MethodPut, "/boards/{link}", bytes.NewBuffer(body))
				req = mux.SetURLVars(req, map[string]string{"link": boardLink.String()})
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateBoard", mock.Anything, mock.Anything, userLink).Return(common.ErrActionDenied).Once()
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "Error not found",
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(updateReq)
				req := httptest.NewRequest(http.MethodPut, "/boards/{link}", bytes.NewBuffer(body))
				req = mux.SetURLVars(req, map[string]string{"link": boardLink.String()})
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateBoard", mock.Anything, mock.Anything, userLink).Return(common.ErrBoardNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Error missing required field",
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(updateReq)
				req := httptest.NewRequest(http.MethodPut, "/boards/{link}", bytes.NewBuffer(body))
				req = mux.SetURLVars(req, map[string]string{"link": boardLink.String()})
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateBoard", mock.Anything, mock.Anything, userLink).
					Return(common.ErrorNotNullValue).Once()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error invalid board data",
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(updateReq)
				req := httptest.NewRequest(http.MethodPut, "/boards/{link}", bytes.NewBuffer(body))
				req = mux.SetURLVars(req, map[string]string{"link": boardLink.String()})
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateBoard", mock.Anything, mock.Anything, userLink).
					Return(common.ErrorInvalidBoardData).Once()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error internal server",
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(updateReq)
				req := httptest.NewRequest(http.MethodPut, "/boards/{link}", bytes.NewBuffer(body))
				req = mux.SetURLVars(req, map[string]string{"link": boardLink.String()})
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateBoard", mock.Anything, mock.Anything, userLink).
					Return(errors.New("some unexpected error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			req := test.setupRequest()
			w := httptest.NewRecorder()

			h.UpdateBoard(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestGetBoard(t *testing.T) {
	userLink := uuid.New()
	boardLink := uuid.New()
	boardInfo := serviceDto.BoardInfo{Link: boardLink, Name: "Target Board", CreatedAt: time.Now()}

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		setupMock      func(m *mocks.BoardService)
		expectedStatus int
	}{
		{
			name: "Success get board",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/boards/{link}", nil)
				req = mux.SetURLVars(req, map[string]string{"link": boardLink.String()})
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetBoard", mock.Anything, boardLink, userLink).Return(boardInfo, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Error unauthorized",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/boards/{link}", nil)
				req = mux.SetURLVars(req, map[string]string{"link": boardLink.String()})
				return req
			},
			setupMock:      func(m *mocks.BoardService) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Error missing board link var",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/boards/", nil)
				return reqWithUser(req, userLink)
			},
			setupMock:      func(m *mocks.BoardService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error invalid board link format",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/boards/invalid-uuid", nil)
				req = mux.SetURLVars(req, map[string]string{"link": "invalid-uuid"})
				return reqWithUser(req, userLink)
			},
			setupMock:      func(m *mocks.BoardService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error forbidden",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/boards/{link}", nil)
				req = mux.SetURLVars(req, map[string]string{"link": boardLink.String()})
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetBoard", mock.Anything, boardLink, userLink).Return(serviceDto.BoardInfo{}, common.ErrActionDenied).Once()
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "Error not found",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/boards/{link}", nil)
				req = mux.SetURLVars(req, map[string]string{"link": boardLink.String()})
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetBoard", mock.Anything, boardLink, userLink).Return(serviceDto.BoardInfo{}, common.ErrBoardNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Error internal server",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/boards/{link}", nil)
				req = mux.SetURLVars(req, map[string]string{"link": boardLink.String()})
				return reqWithUser(req, userLink)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetBoard", mock.Anything, boardLink, userLink).Return(serviceDto.BoardInfo{}, errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			req := test.setupRequest()
			w := httptest.NewRecorder()

			h.GetBoard(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestUploadBackground(t *testing.T) {
	userLink := uuid.New()
	boardLink := uuid.New()
	backgroundURL := "https://s3.example.com/bg.png"

	createMultipartRequest := func(fileKey, fileName string, fileContent []byte, setupVars bool, withUser bool) *http.Request {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		if fileKey != "" {
			part, _ := writer.CreateFormFile(fileKey, fileName)
			part.Write(fileContent)
		}
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/boards/{link}/background", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		if setupVars {
			req = mux.SetURLVars(req, map[string]string{"link": boardLink.String()})
		}

		if withUser {
			return reqWithUser(req, userLink)
		}
		return req
	}

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		setupMock      func(m *mocks.BoardService)
		expectedStatus int
	}{
		{
			name: "Success upload background",
			setupRequest: func() *http.Request {
				pngContent := append([]byte("\x89PNG\r\n\x1a\n"), []byte("dummy image content")...)
				return createMultipartRequest("background", "bg.png", pngContent, true, true)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateBackground", mock.Anything, mock.Anything, "image/png", ".png", boardLink, userLink).
					Return(backgroundURL, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Error unauthorized",
			setupRequest: func() *http.Request {
				return createMultipartRequest("background", "bg.png", []byte("\x89PNG\r\n\x1a\n"), true, false)
			},
			setupMock:      func(m *mocks.BoardService) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Error invalid board link",
			setupRequest: func() *http.Request {
				req := createMultipartRequest("background", "bg.png", []byte("\x89PNG\r\n\x1a\n"), false, true)
				req = mux.SetURLVars(req, map[string]string{"link": "not-a-uuid"})
				return req
			},
			setupMock:      func(m *mocks.BoardService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error missing background key",
			setupRequest: func() *http.Request {
				return createMultipartRequest("wrong_key", "bg.png", []byte("\x89PNG\r\n\x1a\n"), true, true)
			},
			setupMock:      func(m *mocks.BoardService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error invalid content type",
			setupRequest: func() *http.Request {
				textContent := []byte("just a regular text string")
				return createMultipartRequest("background", "text.txt", textContent, true, true)
			},
			setupMock:      func(m *mocks.BoardService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error board not found",
			setupRequest: func() *http.Request {
				pngContent := append([]byte("\x89PNG\r\n\x1a\n"), []byte("content")...)
				return createMultipartRequest("background", "bg.png", pngContent, true, true)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateBackground", mock.Anything, mock.Anything, "image/png", ".png", boardLink, userLink).
					Return("", common.ErrBoardNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Error internal service",
			setupRequest: func() *http.Request {
				pngContent := append([]byte("\x89PNG\r\n\x1a\n"), []byte("content")...)
				return createMultipartRequest("background", "bg.png", pngContent, true, true)
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("UpdateBackground", mock.Anything, mock.Anything, "image/png", ".png", boardLink, userLink).
					Return("", errors.New("s3 upload failed")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			req := test.setupRequest()
			w := httptest.NewRecorder()

			h.UploadBackground(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestGetUsersOfBoard(t *testing.T) {
	boardLink := uuid.New()
	usersLinks := []uuid.UUID{uuid.New(), uuid.New()}

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		setupMock      func(m *mocks.BoardService)
		expectedStatus int
	}{
		{
			name: "Success get users",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/boards/{link}/users", nil)
				return mux.SetURLVars(req, map[string]string{"link": boardLink.String()})
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetUsersOfBoard", mock.Anything, boardLink).Return(usersLinks, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Error missing board link var",
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/boards//users", nil)
			},
			setupMock:      func(m *mocks.BoardService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error invalid board link format",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/boards/invalid-uuid/users", nil)
				return mux.SetURLVars(req, map[string]string{"link": "invalid-uuid"})
			},
			setupMock:      func(m *mocks.BoardService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error board not found",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/boards/{link}/users", nil)
				return mux.SetURLVars(req, map[string]string{"link": boardLink.String()})
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetUsersOfBoard", mock.Anything, boardLink).Return(nil, common.ErrBoardNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Error internal server",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/boards/{link}/users", nil)
				return mux.SetURLVars(req, map[string]string{"link": boardLink.String()})
			},
			setupMock: func(m *mocks.BoardService) {
				m.On("GetUsersOfBoard", mock.Anything, boardLink).Return(nil, errors.New("db error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSrv := new(mocks.BoardService)
			test.setupMock(mockSrv)

			h := handler.NewHandler(mockSrv, testConf)
			req := test.setupRequest()
			w := httptest.NewRecorder()

			h.GetUsersOfBoard(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)
			mockSrv.AssertExpectations(t)
		})
	}
}
