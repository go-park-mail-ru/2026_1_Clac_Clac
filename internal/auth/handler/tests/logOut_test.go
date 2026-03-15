package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/handler"
	mockAuthSrv "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/handler/tests/mock_auth_srv"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type LogOutTestCase struct {
	Name               string
	AddCookie          bool
	CookieValue        string
	ExpectedResponse   any
	ExpectedStatusCode int
	MockBehavior       func(m *mockAuthSrv.AuthService)
}

func TestLogOutUser(t *testing.T) {
	tests := []LogOutTestCase{
		{
			Name:               "Success logout",
			AddCookie:          true,
			CookieValue:        common.FixedSessionID,
			ExpectedResponse:   newResponse(api.StatusOK),
			ExpectedStatusCode: http.StatusOK,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("LogOut", ctx, common.FixedSessionID).Return(nil)
			},
		},
		{
			Name:               "Service error",
			AddCookie:          true,
			CookieValue:        common.FixedSessionID,
			ExpectedResponse:   newResponse(api.StatusOK),
			ExpectedStatusCode: http.StatusOK,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("LogOut", ctx, common.FixedSessionID).Return(fmt.Errorf("database down"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockAuthService := mockAuthSrv.NewAuthService(t)
			if test.MockBehavior != nil {
				test.MockBehavior(mockAuthService)
			}

			handler := handler.NewAuthHandler(mockAuthService)

			request := httptest.NewRequest(http.MethodPost, "/", nil)
			if test.AddCookie {
				request.AddCookie(&http.Cookie{
					Name:  "session_id",
					Value: test.CookieValue,
				})
			}
			response := httptest.NewRecorder()

			handler.LogOutUser(response, request)

			responseJson, err := json.Marshal(test.ExpectedResponse)
			require.NoError(t, err, "response marshal should not return error")

			assert.Equal(t, test.ExpectedStatusCode, response.Code, "incorrect status code")
			assert.Equal(t, string(responseJson), response.Body.String(), "incorrect body")

			if test.ExpectedStatusCode == http.StatusOK {
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
