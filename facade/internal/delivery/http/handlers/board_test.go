package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/dto"
	mockBoardUC "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers/mock_board_use_case"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	fixedBoardLink = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	defaultBoardCfg = BoardConfig{
		MultipartBackgroundFileKey: "background",
		MaxBackgroundSize:          5 << 20,
	}
)

func newTestBoardHandler(srv BoardUsecase) *Board {
	return NewBoard(srv, defaultBoardCfg)
}

func boardRequest(t *testing.T, method, url string, body any, withCtx bool) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		buf.Write(b)
	}
	req := httptest.NewRequest(method, url, &buf)
	if withCtx {
		ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, fixedLink)
		req = req.WithContext(ctx)
	}
	return req
}

func TestHandlerGetBoards(t *testing.T) {
	boards := []domain.BoardInfo{{Link: fixedBoardLink, Name: "Board 1"}}

	tests := []struct {
		name               string
		setContext         bool
		mockBehavior       func(m *mockBoardUC.BoardUsecase)
		expectedStatusCode int
		expectedContains   string
	}{
		{
			name:       "Success",
			setContext: true,
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("GetBoards", mock.Anything, fixedLink).Return(boards, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedContains:   "Board 1",
		},
		{
			name:               "Unauthorized",
			setContext:         false,
			mockBehavior:       func(m *mockBoardUC.BoardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:       "InternalError",
			setContext: true,
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("GetBoards", mock.Anything, fixedLink).Return([]domain.BoardInfo(nil), errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedContains:   ErrCannotGetBoards.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardUC.NewBoardUsecase(t)
			tc.mockBehavior(m)

			req := boardRequest(t, http.MethodGet, "/boards", nil, tc.setContext)
			rr := httptest.NewRecorder()

			newTestBoardHandler(m).GetBoards(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
			if tc.expectedContains != "" {
				assert.Contains(t, rr.Body.String(), tc.expectedContains)
			}
		})
	}
}

func TestHandlerGetBoard(t *testing.T) {
	board := domain.BoardInfo{Link: fixedBoardLink, Name: "Board 1"}

	tests := []struct {
		name               string
		setContext         bool
		linkParam          string
		mockBehavior       func(m *mockBoardUC.BoardUsecase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			setContext: true,
			linkParam:  fixedBoardLink.String(),
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("GetBoard", mock.Anything, domain.GetBoardRequest{
					UserLink:  fixedLink,
					BoardLink: fixedBoardLink,
				}).Return(board, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			setContext:         false,
			linkParam:          fixedBoardLink.String(),
			mockBehavior:       func(m *mockBoardUC.BoardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidUUID",
			setContext:         true,
			linkParam:          "not-a-uuid",
			mockBehavior:       func(m *mockBoardUC.BoardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "NotFound",
			setContext: true,
			linkParam:  fixedBoardLink.String(),
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("GetBoard", mock.Anything, mock.Anything).Return(domain.BoardInfo{}, common.ErrorBoardNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:       "InternalError",
			setContext: true,
			linkParam:  fixedBoardLink.String(),
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("GetBoard", mock.Anything, mock.Anything).Return(domain.BoardInfo{}, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardUC.NewBoardUsecase(t)
			tc.mockBehavior(m)

			req := boardRequest(t, http.MethodGet, "/boards/"+tc.linkParam, nil, tc.setContext)
			req = mux.SetURLVars(req, map[string]string{boardLinkKey: tc.linkParam})
			rr := httptest.NewRecorder()

			newTestBoardHandler(m).GetBoard(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestHandlerCreateBoard(t *testing.T) {
	createReq := dto.CreateBoardRequest{Name: "New Board", Description: "desc"}
	createdBoard := domain.BoardInfo{Link: fixedBoardLink, Name: "New Board"}

	tests := []struct {
		name               string
		setContext         bool
		request            any
		mockBehavior       func(m *mockBoardUC.BoardUsecase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			setContext: true,
			request:    createReq,
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("CreateBoard", mock.Anything, domain.CreateBoardRequest{
					UserLink:    fixedLink,
					Name:        "New Board",
					Description: "desc",
				}).Return(createdBoard, nil)
			},
			expectedStatusCode: http.StatusCreated,
		},
		{
			name:               "Unauthorized",
			setContext:         false,
			request:            createReq,
			mockBehavior:       func(m *mockBoardUC.BoardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidJSON",
			setContext:         true,
			request:            "{bad}",
			mockBehavior:       func(m *mockBoardUC.BoardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "EmptyName",
			setContext:         true,
			request:            dto.CreateBoardRequest{Name: ""},
			mockBehavior:       func(m *mockBoardUC.BoardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "InternalError",
			setContext: true,
			request:    createReq,
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("CreateBoard", mock.Anything, mock.Anything).Return(domain.BoardInfo{}, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardUC.NewBoardUsecase(t)
			tc.mockBehavior(m)

			var bodyBytes []byte
			if s, ok := tc.request.(string); ok {
				bodyBytes = []byte(s)
			} else {
				var err error
				bodyBytes, err = json.Marshal(tc.request)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/boards", bytes.NewReader(bodyBytes))
			if tc.setContext {
				ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, fixedLink)
				req = req.WithContext(ctx)
			}
			rr := httptest.NewRecorder()

			newTestBoardHandler(m).CreateBoard(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestHandlerDeleteBoard(t *testing.T) {
	tests := []struct {
		name               string
		setContext         bool
		linkParam          string
		mockBehavior       func(m *mockBoardUC.BoardUsecase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			setContext: true,
			linkParam:  fixedBoardLink.String(),
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("DeleteBoard", mock.Anything, domain.GetBoardRequest{
					UserLink:  fixedLink,
					BoardLink: fixedBoardLink,
				}).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			setContext:         false,
			linkParam:          fixedBoardLink.String(),
			mockBehavior:       func(m *mockBoardUC.BoardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidUUID",
			setContext:         true,
			linkParam:          "bad-uuid",
			mockBehavior:       func(m *mockBoardUC.BoardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "NotFound",
			setContext: true,
			linkParam:  fixedBoardLink.String(),
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("DeleteBoard", mock.Anything, mock.Anything).Return(common.ErrorBoardNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:       "InternalError",
			setContext: true,
			linkParam:  fixedBoardLink.String(),
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("DeleteBoard", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardUC.NewBoardUsecase(t)
			tc.mockBehavior(m)

			req := boardRequest(t, http.MethodDelete, "/boards/"+tc.linkParam, nil, tc.setContext)
			req = mux.SetURLVars(req, map[string]string{boardLinkKey: tc.linkParam})
			rr := httptest.NewRecorder()

			newTestBoardHandler(m).DeleteBoard(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestHandlerUpdateBoard(t *testing.T) {
	updateReq := dto.UpdateBoardRequest{Name: "Updated", Description: "new desc"}

	tests := []struct {
		name               string
		setContext         bool
		linkParam          string
		request            any
		mockBehavior       func(m *mockBoardUC.BoardUsecase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			setContext: true,
			linkParam:  fixedBoardLink.String(),
			request:    updateReq,
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("UpdateBoard", mock.Anything, domain.UpdateBoardRequest{
					UserLink:    fixedLink,
					BoardLink:   fixedBoardLink,
					Name:        "Updated",
					Description: "new desc",
				}).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			setContext:         false,
			linkParam:          fixedBoardLink.String(),
			request:            updateReq,
			mockBehavior:       func(m *mockBoardUC.BoardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidUUID",
			setContext:         true,
			linkParam:          "bad-uuid",
			request:            updateReq,
			mockBehavior:       func(m *mockBoardUC.BoardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "InvalidJSON",
			setContext:         true,
			linkParam:          fixedBoardLink.String(),
			request:            "{bad}",
			mockBehavior:       func(m *mockBoardUC.BoardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "NotFound",
			setContext: true,
			linkParam:  fixedBoardLink.String(),
			request:    updateReq,
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("UpdateBoard", mock.Anything, mock.Anything).Return(common.ErrorBoardNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:       "InternalError",
			setContext: true,
			linkParam:  fixedBoardLink.String(),
			request:    updateReq,
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("UpdateBoard", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardUC.NewBoardUsecase(t)
			tc.mockBehavior(m)

			var bodyBytes []byte
			if s, ok := tc.request.(string); ok {
				bodyBytes = []byte(s)
			} else {
				var err error
				bodyBytes, err = json.Marshal(tc.request)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, "/boards/"+tc.linkParam, bytes.NewReader(bodyBytes))
			if tc.setContext {
				ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, fixedLink)
				req = req.WithContext(ctx)
			}
			req = mux.SetURLVars(req, map[string]string{boardLinkKey: tc.linkParam})
			rr := httptest.NewRecorder()

			newTestBoardHandler(m).UpdateBoard(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func buildBackgroundRequest(t *testing.T, withContext bool, fileKey string, fileData []byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile(fileKey, "bg.png")
	require.NoError(t, err)
	_, err = part.Write(fileData)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	req := httptest.NewRequest(http.MethodPut, "/boards/"+fixedBoardLink.String()+"/background", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if withContext {
		ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, fixedLink)
		req = req.WithContext(ctx)
	}
	req = mux.SetURLVars(req, map[string]string{boardLinkKey: fixedBoardLink.String()})
	return req
}

func TestHandlerUploadBackground(t *testing.T) {
	tests := []struct {
		name               string
		setContext         bool
		linkParam          string
		fileKey            string
		mockBehavior       func(m *mockBoardUC.BoardUsecase)
		expectedStatusCode int
	}{
		{
			name:      "Success",
			setContext: true,
			linkParam: fixedBoardLink.String(),
			fileKey:   "background",
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("UploadBackground", mock.Anything, mock.MatchedBy(func(info domain.UploadBackgroundRequest) bool {
					return info.UserLink == fixedLink && info.BoardLink == fixedBoardLink
				}), mock.Anything).Return(domain.UploadBackgroundResponse{BackgroundKey: "key/bg.png"}, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			setContext:         false,
			linkParam:          fixedBoardLink.String(),
			fileKey:            "background",
			mockBehavior:       func(m *mockBoardUC.BoardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:      "WrongFileKey",
			setContext: true,
			linkParam: fixedBoardLink.String(),
			fileKey:   "wrong_key",
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:      "NotFound",
			setContext: true,
			linkParam: fixedBoardLink.String(),
			fileKey:   "background",
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("UploadBackground", mock.Anything, mock.Anything, mock.Anything).Return(domain.UploadBackgroundResponse{}, common.ErrorBoardNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:      "InternalError",
			setContext: true,
			linkParam: fixedBoardLink.String(),
			fileKey:   "background",
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("UploadBackground", mock.Anything, mock.Anything, mock.Anything).Return(domain.UploadBackgroundResponse{}, errors.New("storage error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardUC.NewBoardUsecase(t)
			tc.mockBehavior(m)

			req := buildBackgroundRequest(t, tc.setContext, tc.fileKey, []byte("fake image data"))
			rr := httptest.NewRecorder()

			newTestBoardHandler(m).UploadBackground(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestHandlerGetMembers(t *testing.T) {
	members := domain.GetMembersResponse{Members: []domain.MemberInfo{{Link: fixedLink, Role: "viewer"}}}

	tests := []struct {
		name               string
		setContext         bool
		linkParam          string
		mockBehavior       func(m *mockBoardUC.BoardUsecase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			setContext: true,
			linkParam:  fixedBoardLink.String(),
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("GetMembers", mock.Anything, domain.GetMembersRequest{
					UserLink:  fixedLink,
					BoardLink: fixedBoardLink,
				}).Return(members, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			setContext:         false,
			linkParam:          fixedBoardLink.String(),
			mockBehavior:       func(m *mockBoardUC.BoardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidUUID",
			setContext:         true,
			linkParam:          "bad-uuid",
			mockBehavior:       func(m *mockBoardUC.BoardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "NotFound",
			setContext: true,
			linkParam:  fixedBoardLink.String(),
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("GetMembers", mock.Anything, mock.Anything).Return(domain.GetMembersResponse{}, common.ErrorBoardNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:       "InternalError",
			setContext: true,
			linkParam:  fixedBoardLink.String(),
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("GetMembers", mock.Anything, mock.Anything).Return(domain.GetMembersResponse{}, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardUC.NewBoardUsecase(t)
			tc.mockBehavior(m)

			req := boardRequest(t, http.MethodGet, "/boards/"+tc.linkParam+"/users", nil, tc.setContext)
			req = mux.SetURLVars(req, map[string]string{boardLinkKey: tc.linkParam})
			rr := httptest.NewRecorder()

			newTestBoardHandler(m).GetMembers(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestHandlerCreateInvite(t *testing.T) {
	fixedInviteLink := uuid.New()

	tests := []struct {
		name               string
		setContext         bool
		linkParam          string
		body               any
		mockBehavior       func(m *mockBoardUC.BoardUsecase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			setContext: true,
			linkParam:  fixedBoardLink.String(),
			body: dto.CreateInviteRequest{
				DefaultRole:   "editor",
				ExpireSeconds: 86400,
			},
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("CreateInvite", mock.Anything, mock.Anything).Return(domain.CreateInviteResponse{
					InviteLink:  fixedInviteLink.String(),
					BoardLink:   fixedBoardLink.String(),
					DefaultRole: "editor",
					Status:      "active",
				}, nil)
			},
			expectedStatusCode: http.StatusCreated,
		},
		{
			name:       "Unauthorized",
			setContext: false,
			linkParam:  fixedBoardLink.String(),
			body: dto.CreateInviteRequest{
				DefaultRole:   "editor",
			},
			mockBehavior:       func(m *mockBoardUC.BoardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:       "EmptyRole",
			setContext: true,
			linkParam:  fixedBoardLink.String(),
			body: dto.CreateInviteRequest{
				DefaultRole: "",
			},
			mockBehavior:       func(m *mockBoardUC.BoardUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "BoardNotFound",
			setContext: true,
			linkParam:  fixedBoardLink.String(),
			body: dto.CreateInviteRequest{
				DefaultRole:   "editor",
			},
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("CreateInvite", mock.Anything, mock.Anything).Return(domain.CreateInviteResponse{}, common.ErrorBoardNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:       "PermissionDenied",
			setContext: true,
			linkParam:  fixedBoardLink.String(),
			body: dto.CreateInviteRequest{
				DefaultRole:   "editor",
			},
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("CreateInvite", mock.Anything, mock.Anything).Return(domain.CreateInviteResponse{}, common.ErrorBoardPermissionDenied)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:       "InternalError",
			setContext: true,
			linkParam:  fixedBoardLink.String(),
			body: dto.CreateInviteRequest{
				DefaultRole:   "editor",
			},
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("CreateInvite", mock.Anything, mock.Anything).Return(domain.CreateInviteResponse{}, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardUC.NewBoardUsecase(t)
			tc.mockBehavior(m)

			req := boardRequest(t, http.MethodPost, "/boards/"+tc.linkParam+"/invite", tc.body, tc.setContext)
			req = mux.SetURLVars(req, map[string]string{boardLinkKey: tc.linkParam})
			rr := httptest.NewRecorder()

			newTestBoardHandler(m).CreateInvite(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestHandlerAcceptInvite(t *testing.T) {
	fixedInviteLink := uuid.New()

	tests := []struct {
		name               string
		setContext         bool
		linkParam          string
		mockBehavior       func(m *mockBoardUC.BoardUsecase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			setContext: true,
			linkParam:  fixedInviteLink.String(),
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("AcceptInvite", mock.Anything, mock.Anything).Return(fixedBoardLink.String(), "editor", nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			setContext:         false,
			linkParam:          fixedInviteLink.String(),
			mockBehavior:       func(m *mockBoardUC.BoardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:       "InviteNotFound",
			setContext: true,
			linkParam:  fixedInviteLink.String(),
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("AcceptInvite", mock.Anything, mock.Anything).Return("", "", common.ErrorInviteNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:       "InviteClosedOrExpired",
			setContext: true,
			linkParam:  fixedInviteLink.String(),
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("AcceptInvite", mock.Anything, mock.Anything).Return("", "", common.ErrorInviteClosed)
			},
			expectedStatusCode: http.StatusPreconditionFailed,
		},
		{
			name:       "InviteNotForUser",
			setContext: true,
			linkParam:  fixedInviteLink.String(),
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("AcceptInvite", mock.Anything, mock.Anything).Return("", "", common.ErrorInviteNotForUser)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:       "UserAlreadyMember",
			setContext: true,
			linkParam:  fixedInviteLink.String(),
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("AcceptInvite", mock.Anything, mock.Anything).Return("", "", common.ErrorUserAlreadyMember)
			},
			expectedStatusCode: http.StatusConflict,
		},
		{
			name:       "InternalError",
			setContext: true,
			linkParam:  fixedInviteLink.String(),
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("AcceptInvite", mock.Anything, mock.Anything).Return("", "", errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardUC.NewBoardUsecase(t)
			tc.mockBehavior(m)

			req := boardRequest(t, http.MethodPost, "/invite/"+tc.linkParam, nil, tc.setContext)
			req = mux.SetURLVars(req, map[string]string{inviteLinkKey: tc.linkParam})
			rr := httptest.NewRecorder()

			newTestBoardHandler(m).AcceptInvite(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestHandlerCloseInvite(t *testing.T) {
	fixedInviteLink := uuid.New()

	tests := []struct {
		name               string
		setContext         bool
		linkParam          string
		mockBehavior       func(m *mockBoardUC.BoardUsecase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			setContext: true,
			linkParam:  fixedInviteLink.String(),
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("CloseInvite", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			setContext:         false,
			linkParam:          fixedInviteLink.String(),
			mockBehavior:       func(m *mockBoardUC.BoardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:       "InviteNotFound",
			setContext: true,
			linkParam:  fixedInviteLink.String(),
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("CloseInvite", mock.Anything, mock.Anything).Return(common.ErrorInviteNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:       "PermissionDenied",
			setContext: true,
			linkParam:  fixedInviteLink.String(),
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("CloseInvite", mock.Anything, mock.Anything).Return(common.ErrorBoardPermissionDenied)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:       "InternalError",
			setContext: true,
			linkParam:  fixedInviteLink.String(),
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("CloseInvite", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardUC.NewBoardUsecase(t)
			tc.mockBehavior(m)

			req := boardRequest(t, http.MethodPost, "/invite/"+tc.linkParam+"/close", nil, tc.setContext)
			req = mux.SetURLVars(req, map[string]string{inviteLinkKey: tc.linkParam})
			rr := httptest.NewRecorder()

			newTestBoardHandler(m).CloseInvite(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestHandlerGetActiveInvites(t *testing.T) {
	tests := []struct {
		name               string
		setContext         bool
		linkParam          string
		mockBehavior       func(m *mockBoardUC.BoardUsecase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			setContext: true,
			linkParam:  fixedBoardLink.String(),
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("GetActiveInvites", mock.Anything, fixedLink, fixedBoardLink).Return([]domain.InviteInfo{}, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			setContext:         false,
			linkParam:          fixedBoardLink.String(),
			mockBehavior:       func(m *mockBoardUC.BoardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:       "PermissionDenied",
			setContext: true,
			linkParam:  fixedBoardLink.String(),
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("GetActiveInvites", mock.Anything, fixedLink, fixedBoardLink).Return(nil, common.ErrorBoardPermissionDenied)
			},
			expectedStatusCode: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardUC.NewBoardUsecase(t)
			tc.mockBehavior(m)

			req := boardRequest(t, http.MethodGet, "/boards/"+tc.linkParam+"/invites", nil, tc.setContext)
			req = mux.SetURLVars(req, map[string]string{boardLinkKey: tc.linkParam})
			rr := httptest.NewRecorder()

			newTestBoardHandler(m).GetActiveInvites(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestHandlerUpdateMemberRole(t *testing.T) {
	fixedUserParam := uuid.New()

	tests := []struct {
		name               string
		setContext         bool
		boardLink          string
		userLink           string
		body               any
		mockBehavior       func(m *mockBoardUC.BoardUsecase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			setContext: true,
			boardLink:  fixedBoardLink.String(),
			userLink:   fixedUserParam.String(),
			body:       dto.UpdateMemberRoleRequest{NewRole: "editor"},
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("UpdateMemberRole", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			setContext:         false,
			boardLink:          fixedBoardLink.String(),
			userLink:           fixedUserParam.String(),
			mockBehavior:       func(m *mockBoardUC.BoardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:       "NotFound",
			setContext: true,
			boardLink:  fixedBoardLink.String(),
			userLink:   fixedUserParam.String(),
			body:       dto.UpdateMemberRoleRequest{NewRole: "editor"},
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("UpdateMemberRole", mock.Anything, mock.Anything).Return(common.ErrorBoardNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:       "PermissionDenied",
			setContext: true,
			boardLink:  fixedBoardLink.String(),
			userLink:   fixedUserParam.String(),
			body:       dto.UpdateMemberRoleRequest{NewRole: "editor"},
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("UpdateMemberRole", mock.Anything, mock.Anything).Return(common.ErrorBoardPermissionDenied)
			},
			expectedStatusCode: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardUC.NewBoardUsecase(t)
			tc.mockBehavior(m)

			req := boardRequest(t, http.MethodPut, "/boards/"+tc.boardLink+"/members/"+tc.userLink+"/role", tc.body, tc.setContext)
			req = mux.SetURLVars(req, map[string]string{boardLinkKey: tc.boardLink, "user_link": tc.userLink})
			rr := httptest.NewRecorder()

			newTestBoardHandler(m).UpdateMemberRole(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestHandlerRemoveMemberFromBoard(t *testing.T) {
	fixedUserParam := uuid.New()

	tests := []struct {
		name               string
		setContext         bool
		boardLink          string
		userLink           string
		mockBehavior       func(m *mockBoardUC.BoardUsecase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			setContext: true,
			boardLink:  fixedBoardLink.String(),
			userLink:   fixedUserParam.String(),
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("RemoveMemberFromBoard", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			setContext:         false,
			boardLink:          fixedBoardLink.String(),
			userLink:           fixedUserParam.String(),
			mockBehavior:       func(m *mockBoardUC.BoardUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:       "PermissionDenied",
			setContext: true,
			boardLink:  fixedBoardLink.String(),
			userLink:   fixedUserParam.String(),
			mockBehavior: func(m *mockBoardUC.BoardUsecase) {
				m.On("RemoveMemberFromBoard", mock.Anything, mock.Anything).Return(common.ErrorBoardPermissionDenied)
			},
			expectedStatusCode: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardUC.NewBoardUsecase(t)
			tc.mockBehavior(m)

			req := boardRequest(t, http.MethodDelete, "/boards/"+tc.boardLink+"/members/"+tc.userLink, nil, tc.setContext)
			req = mux.SetURLVars(req, map[string]string{boardLinkKey: tc.boardLink, "user_link": tc.userLink})
			rr := httptest.NewRecorder()

			newTestBoardHandler(m).RemoveMemberFromBoard(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}
