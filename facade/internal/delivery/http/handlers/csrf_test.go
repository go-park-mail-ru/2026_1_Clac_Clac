package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockCSRFUC "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers/mock_csrf_use_case" // Поправьте путь, если мок лежит в другой папке
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSetCSRFCookieHandler(t *testing.T) {
	fixedTime := time.Now().Add(1 * time.Hour)
	sessionID := "valid_session_id"
	expectedToken := "generated_csrf_token"

	tests := []struct {
		name               string
		setupRequest       func() *http.Request
		mockBehavior       func(m *mockCSRFUC.CSRFUsecase)
		expectedStatusCode int
		expectCookie       bool
	}{
		{
			name: "Success",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/api/csrf", nil)
				req.AddCookie(&http.Cookie{Name: middleware.SessiondIdKey, Value: sessionID})
				return req
			},
			mockBehavior: func(m *mockCSRFUC.CSRFUsecase) {
				m.On("GetExpireTime", mock.Anything).Return(fixedTime)
				m.On("Generate", mock.Anything, sessionID, fixedTime.Unix()).Return(expectedToken, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectCookie:       true,
		},
		{
			name: "Unauthorized_NoCookie",
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/csrf", nil)
			},
			mockBehavior:       nil,
			expectedStatusCode: http.StatusUnauthorized,
			expectCookie:       false,
		},
		{
			name: "InternalServerError_GenerateFails",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/api/csrf", nil)
				req.AddCookie(&http.Cookie{Name: middleware.SessiondIdKey, Value: sessionID})
				return req
			},
			mockBehavior: func(m *mockCSRFUC.CSRFUsecase) {
				m.On("GetExpireTime", mock.Anything).Return(fixedTime)
				m.On("Generate", mock.Anything, sessionID, fixedTime.Unix()).Return("", errors.New("internal error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectCookie:       false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockCSRFUC.NewCSRFUsecase(t)
			if tc.mockBehavior != nil {
				tc.mockBehavior(m)
			}

			handler := NewCSRF(m)
			req := tc.setupRequest()
			rr := httptest.NewRecorder()

			handler.SetCSRFCookieHandler(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)

			cookies := rr.Result().Cookies()
			if tc.expectCookie {
				require.Len(t, cookies, 1)
				assert.Equal(t, csrfCookieKey, cookies[0].Name)
				assert.Equal(t, expectedToken, cookies[0].Value)
			} else {
				assert.Len(t, cookies, 0)
			}
		})
	}
}
