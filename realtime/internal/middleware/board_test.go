package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

var testUserLink = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
var testBoardLink = uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

type stubBoardChecker struct {
	canViewFn func(ctx context.Context, userLink, boardLink uuid.UUID) error
}

func (s *stubBoardChecker) CanView(ctx context.Context, userLink, boardLink uuid.UUID) error {
	return s.canViewFn(ctx, userLink, boardLink)
}

func newBoardMiddleware(canViewFn func(context.Context, uuid.UUID, uuid.UUID) error) mux.MiddlewareFunc {
	stub := &stubBoardChecker{canViewFn: canViewFn}
	return middleware.BoardAccessMiddleware(stub)
}

func TestBoardAccessMiddleware(t *testing.T) {
	okCanView := func(_ context.Context, _ uuid.UUID, _ uuid.UUID) error { return nil }
	failCanView := func(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
		return errors.New("access denied")
	}

	tests := []struct {
		name            string
		userInCtx       bool
		boardLinkVar    string
		canViewFn       func(context.Context, uuid.UUID, uuid.UUID) error
		expectedStatus  int
		expectBoardLink bool
	}{
		{
			name:           "NoUserInContext",
			userInCtx:      false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "MissingBoardLinkVar",
			userInCtx:      true,
			boardLinkVar:   "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "InvalidBoardLinkUUID",
			userInCtx:      true,
			boardLinkVar:   "not-a-uuid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:            "AccessDenied",
			userInCtx:       true,
			boardLinkVar:    testBoardLink.String(),
			canViewFn:       failCanView,
			expectedStatus:  http.StatusForbidden,
		},
		{
			name:            "Success",
			userInCtx:       true,
			boardLinkVar:    testBoardLink.String(),
			canViewFn:       okCanView,
			expectedStatus:  http.StatusOK,
			expectBoardLink: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mw := newBoardMiddleware(tc.canViewFn)

			url := "/api/events/" + tc.boardLinkVar
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req = mux.SetURLVars(req, map[string]string{
				"board_link": tc.boardLinkVar,
			})

			if tc.userInCtx {
				ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, testUserLink)
				req = req.WithContext(ctx)
			}

			res := httptest.NewRecorder()

			var gotBoardLink uuid.UUID
			mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				val := r.Context().Value(middleware.BoardContextLink{})
				if link, ok := val.(uuid.UUID); ok {
					gotBoardLink = link
				}
				w.WriteHeader(http.StatusOK)
			})).ServeHTTP(res, req)

			assert.Equal(t, tc.expectedStatus, res.Code)

			if tc.expectBoardLink {
				assert.Equal(t, testBoardLink, gotBoardLink)
			}
		})
	}
}
