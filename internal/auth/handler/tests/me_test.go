package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/handler"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMeHandler(t *testing.T) {
	handler := &handler.AuthHandler{}

	tests := []struct {
		Name           string
		SetupRequest   func(req *http.Request) *http.Request
		ExpectedStatus int
	}{
		{
			Name: "success",
			SetupRequest: func(req *http.Request) *http.Request {
				userID := uuid.New()
				ctx := context.WithValue(req.Context(), middleware.UserIDKey{}, userID)
				return req.WithContext(ctx)
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name: "unauthorized no context value",
			SetupRequest: func(req *http.Request) *http.Request {
				return req
			},
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name: "unauthorized wrong type",
			SetupRequest: func(req *http.Request) *http.Request {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey{}, "invalid-uuid-string")
				return req.WithContext(ctx)
			},
			ExpectedStatus: http.StatusUnauthorized,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
			req = test.SetupRequest(req)
			res := httptest.NewRecorder()

			handler.MeHandler(res, req)

			assert.Equal(t, test.ExpectedStatus, res.Code)
		})
	}
}
