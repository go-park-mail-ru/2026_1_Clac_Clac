package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	common "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	mockAuthSrv "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/handler/auth/mock_auth_srv"

	"github.com/stretchr/testify/assert"
)

func TestLogOutUser(t *testing.T) {
	tests := []struct {
		nameTest           string
		addCookie          bool
		cookieValue        string
		mockBehavior       func(m *mockAuthSrv.AuthService)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			nameTest:    "Success logout",
			addCookie:   true,
			cookieValue: common.FixedSessionID,
			mockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("LogOut", ctx, common.FixedSessionID).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   "{message: \"successfully logged out\"}",
		},
		{
			nameTest:           "No cookie provided",
			addCookie:          false,
			cookieValue:        "",
			mockBehavior:       nil,
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   "{\"error\":\"user not authorized\"}\n",
		},
		{
			nameTest:    "Service error",
			addCookie:   true,
			cookieValue: common.FixedSessionID,
			mockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("LogOut", ctx, common.FixedSessionID).Return(fmt.Errorf("database down"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   "{\"error\":\"failed to logout: database down\"}\n",
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockAuthService := mockAuthSrv.NewAuthService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockAuthService)
			}

			handler := NewAuthHandler(mockAuthService)

			request := httptest.NewRequest(http.MethodPost, "/logout", nil)
			if test.addCookie {
				request.AddCookie(&http.Cookie{
					Name:  "session_id",
					Value: test.cookieValue,
				})
			}
			response := httptest.NewRecorder()

			handler.LogOutUser(response, request)

			assert.Equal(t, test.expectedStatusCode, response.Code, "incorrect status code")
			assert.Equal(t, test.expectedResponse, response.Body.String(), "incorrect response")

			if test.expectedStatusCode == http.StatusOK {
				res := response.Result()
				cookies := res.Cookies()

				var sessionCookie *http.Cookie
				for _, c := range cookies {
					if c.Name == "session_id" {
						sessionCookie = c
						break
					}
				}

				assert.NotNil(t, sessionCookie, "cookie wasn't found in response")
				assert.Empty(t, sessionCookie.Value, "cookie value must be empty")
			}
		})
	}
}
