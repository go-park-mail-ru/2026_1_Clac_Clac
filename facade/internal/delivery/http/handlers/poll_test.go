package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/middleware"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/board/v1"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	fixedBoardLinkP = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	fixedUserLinkP  = uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
	errPollTest     = errors.New("poll error")
)

type mockPollUsecase struct {
	mock.Mock
}

func (m *mockPollUsecase) CreatePoll(ctx context.Context, boardLink, adminLink uuid.UUID, cards []uuid.UUID, invitees []uuid.UUID) error {
	args := m.Called(ctx, boardLink, adminLink, cards, invitees)
	return args.Error(0)
}

func (m *mockPollUsecase) DeletePoll(ctx context.Context, boardLink, userLink uuid.UUID) error {
	args := m.Called(ctx, boardLink, userLink)
	return args.Error(0)
}

func (m *mockPollUsecase) NextPollCard(ctx context.Context, boardLink, userLink uuid.UUID) error {
	args := m.Called(ctx, boardLink, userLink)
	return args.Error(0)
}

func (m *mockPollUsecase) VotePoll(ctx context.Context, boardLink, userLink uuid.UUID, points int) error {
	args := m.Called(ctx, boardLink, userLink, points)
	return args.Error(0)
}

func (m *mockPollUsecase) GetActivePoll(ctx context.Context, boardLink, userLink uuid.UUID) (*pb.GetActivePollResponse, error) {
	args := m.Called(ctx, boardLink, userLink)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.GetActivePollResponse), args.Error(1)
}

func newTestPollHandler(uc PollUsecase) *PollHandler {
	return NewPollHandler(uc, PollConfig{
		MinVotePoints: 1,
		MaxVotePoints: 21,
	})
}

func newPollRequest(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func setPollUserContext(r *http.Request) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserContextLink{}, fixedUserLinkP)
	return r.WithContext(ctx)
}

func TestPollHandler_CreatePoll(t *testing.T) {
	validBody := `{"card_links":["aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"],"invitees":["eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"]}`
	invalidBody := `{"card_links": "not an array"}`

	tests := []struct {
		name         string
		boardLinkVar string
		body         string
		mockBehavior func(m *mockPollUsecase)
		setupCtx     bool
		expectedCode int
	}{
		{
			name:         "Success",
			boardLinkVar: fixedBoardLinkP.String(),
			body:         validBody,
			mockBehavior: func(m *mockPollUsecase) {
				m.On("CreatePoll", mock.Anything, fixedBoardLinkP, fixedUserLinkP, mock.Anything, mock.Anything).Return(nil)
			},
			setupCtx:     true,
			expectedCode: http.StatusCreated,
		},
		{
			name:         "Error_InvalidBoardLink",
			boardLinkVar: "not-a-uuid",
			body:         validBody,
			mockBehavior: func(m *mockPollUsecase) {},
			setupCtx:     true,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Error_InvalidBody",
			boardLinkVar: fixedBoardLinkP.String(),
			body:         invalidBody,
			mockBehavior: func(m *mockPollUsecase) {},
			setupCtx:     true,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Error_Unauthorized",
			boardLinkVar: fixedBoardLinkP.String(),
			body:         validBody,
			mockBehavior: func(m *mockPollUsecase) {
				m.On("CreatePoll", mock.Anything, fixedBoardLinkP, fixedUserLinkP, mock.Anything, mock.Anything).Return(errPollTest)
			},
			setupCtx:     true,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockPollUsecase)
			tt.mockBehavior(m)

			handler := newTestPollHandler(m)
			path := "/boards/" + tt.boardLinkVar + "/polls"
			req := newPollRequest(http.MethodPost, path, tt.body)
			if tt.setupCtx {
				req = setPollUserContext(req)
			}
			req = mux.SetURLVars(req, map[string]string{pollBoardLinkKey: tt.boardLinkVar})

			w := httptest.NewRecorder()
			handler.CreatePoll(w, req)
			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

func TestPollHandler_Vote(t *testing.T) {
	validBody := `{"points":5}`
	invalidBody := `{"points":"bad"}`

	tests := []struct {
		name         string
		boardLinkVar string
		body         string
		mockBehavior func(m *mockPollUsecase)
		setupCtx     bool
		expectedCode int
	}{
		{
			name:         "Success",
			boardLinkVar: fixedBoardLinkP.String(),
			body:         validBody,
			mockBehavior: func(m *mockPollUsecase) {
				m.On("VotePoll", mock.Anything, fixedBoardLinkP, fixedUserLinkP, 5).Return(nil)
			},
			setupCtx:     true,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Error_InvalidBoardLink",
			boardLinkVar: "not-a-uuid",
			body:         validBody,
			mockBehavior: func(m *mockPollUsecase) {},
			setupCtx:     true,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Error_InvalidBody",
			boardLinkVar: fixedBoardLinkP.String(),
			body:         invalidBody,
			mockBehavior: func(m *mockPollUsecase) {},
			setupCtx:     true,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Error_NotInvited",
			boardLinkVar: fixedBoardLinkP.String(),
			body:         validBody,
			mockBehavior: func(m *mockPollUsecase) {
				m.On("VotePoll", mock.Anything, fixedBoardLinkP, fixedUserLinkP, 5).Return(errPollTest)
			},
			setupCtx:     true,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockPollUsecase)
			tt.mockBehavior(m)

			handler := newTestPollHandler(m)
			path := "/boards/" + tt.boardLinkVar + "/polls"
			req := newPollRequest(http.MethodPut, path, tt.body)
			if tt.setupCtx {
				req = setPollUserContext(req)
			}
			req = mux.SetURLVars(req, map[string]string{pollBoardLinkKey: tt.boardLinkVar})

			w := httptest.NewRecorder()
			handler.Vote(w, req)
			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

func TestPollHandler_DeletePoll(t *testing.T) {
	tests := []struct {
		name         string
		boardLinkVar string
		mockBehavior func(m *mockPollUsecase)
		setupCtx     bool
		expectedCode int
	}{
		{
			name:         "Success",
			boardLinkVar: fixedBoardLinkP.String(),
			mockBehavior: func(m *mockPollUsecase) {
				m.On("DeletePoll", mock.Anything, fixedBoardLinkP, fixedUserLinkP).Return(nil)
			},
			setupCtx:     true,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Error_InvalidBoardLink",
			boardLinkVar: "not-a-uuid",
			mockBehavior: func(m *mockPollUsecase) {
			},
			setupCtx:     true,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Error_NotPollAdmin",
			boardLinkVar: fixedBoardLinkP.String(),
			mockBehavior: func(m *mockPollUsecase) {
				m.On("DeletePoll", mock.Anything, fixedBoardLinkP, fixedUserLinkP).Return(errPollTest)
			},
			setupCtx:     true,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockPollUsecase)
			tt.mockBehavior(m)

			handler := newTestPollHandler(m)
			path := "/boards/" + tt.boardLinkVar + "/polls"
			req := newPollRequest(http.MethodDelete, path, "")
			if tt.setupCtx {
				req = setPollUserContext(req)
			}
			req = mux.SetURLVars(req, map[string]string{pollBoardLinkKey: tt.boardLinkVar})

			w := httptest.NewRecorder()
			handler.DeletePoll(w, req)
			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

func TestPollHandler_NextCard(t *testing.T) {
	tests := []struct {
		name         string
		boardLinkVar string
		mockBehavior func(m *mockPollUsecase)
		setupCtx     bool
		expectedCode int
	}{
		{
			name:         "Success",
			boardLinkVar: fixedBoardLinkP.String(),
			mockBehavior: func(m *mockPollUsecase) {
				m.On("NextPollCard", mock.Anything, fixedBoardLinkP, fixedUserLinkP).Return(nil)
			},
			setupCtx:     true,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Error_InvalidBoardLink",
			boardLinkVar: "not-a-uuid",
			mockBehavior: func(m *mockPollUsecase) {
			},
			setupCtx:     true,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Error_NotPollAdmin",
			boardLinkVar: fixedBoardLinkP.String(),
			mockBehavior: func(m *mockPollUsecase) {
				m.On("NextPollCard", mock.Anything, fixedBoardLinkP, fixedUserLinkP).Return(errPollTest)
			},
			setupCtx:     true,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockPollUsecase)
			tt.mockBehavior(m)

			handler := newTestPollHandler(m)
			path := "/boards/" + tt.boardLinkVar + "/polls/next"
			req := newPollRequest(http.MethodPost, path, "")
			if tt.setupCtx {
				req = setPollUserContext(req)
			}
			req = mux.SetURLVars(req, map[string]string{pollBoardLinkKey: tt.boardLinkVar})

			w := httptest.NewRecorder()
			handler.NextCard(w, req)
			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

func TestPollHandler_GetActivePoll(t *testing.T) {
	mockResponse := &pb.GetActivePollResponse{
		AdminLink:  fixedUserLinkP.String(),
		CurrentIdx: 0,
		Tasks: []*pb.PollTaskInfo{
			{
				CardLink:    fixedBoardLinkP.String(),
				Title:       "Task 1",
				Description: "Task description",
				Votes:       []*pb.VoteEntry{},
			},
		},
		Invitees: []string{fixedUserLinkP.String()},
	}

	tests := []struct {
		name         string
		boardLinkVar string
		mockBehavior func(m *mockPollUsecase)
		setupCtx     bool
		expectedCode int
	}{
		{
			name:         "Success",
			boardLinkVar: fixedBoardLinkP.String(),
			mockBehavior: func(m *mockPollUsecase) {
				m.On("GetActivePoll", mock.Anything, fixedBoardLinkP, fixedUserLinkP).Return(mockResponse, nil)
			},
			setupCtx:     true,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Error_InvalidBoardLink",
			boardLinkVar: "not-a-uuid",
			mockBehavior: func(m *mockPollUsecase) {},
			setupCtx:     true,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Error_Unauthorized",
			boardLinkVar: fixedBoardLinkP.String(),
			mockBehavior: func(m *mockPollUsecase) {
				m.On("GetActivePoll", mock.Anything, fixedBoardLinkP, fixedUserLinkP).Return(nil, errPollTest)
			},
			setupCtx:     true,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockPollUsecase)
			tt.mockBehavior(m)

			handler := newTestPollHandler(m)
			path := "/boards/" + tt.boardLinkVar + "/polls"
			req := newPollRequest(http.MethodGet, path, "")
			if tt.setupCtx {
				req = setPollUserContext(req)
			}
			req = mux.SetURLVars(req, map[string]string{pollBoardLinkKey: tt.boardLinkVar})

			w := httptest.NewRecorder()
			handler.GetActivePoll(w, req)
			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}
