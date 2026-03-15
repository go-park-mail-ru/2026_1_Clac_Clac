package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/handler"
	mockAuthSrv "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/handler/tests/mock_auth_srv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResetUserPassword(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Success reset password",
			Request: api.NewPasswordRequest{
				TokenID:          "valid-token-123",
				Password:         "new_secure_password",
				RepeatedPassword: "new_secure_password",
			},
			ExpectedResponse:   newResponse(api.StatusOK),
			ExpectedStatusCode: http.StatusOK,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("ResetPassword", ctx, "valid-token-123", "new_secure_password").Return(nil)
			},
		},
		{
			Name: "Validation failed (passwords do not match)",
			Request: api.NewPasswordRequest{
				TokenID:          "valid-token-123",
				Password:         "new_secure_password",
				RepeatedPassword: "different_password",
			},
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, handler.InvalidEmailOrPassword),
			ExpectedStatusCode: http.StatusBadRequest,
			MockBehavior:       nil,
		},
		{
			Name: "Service error (e.g. token expired)",
			Request: api.NewPasswordRequest{
				TokenID:          "expired-token",
				Password:         "new_secure_password",
				RepeatedPassword: "new_secure_password",
			},
			ExpectedResponse:   newErrorResponse(http.StatusInternalServerError, handler.CannotResetPassword),
			ExpectedStatusCode: http.StatusInternalServerError,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("ResetPassword", ctx, "expired-token", "new_secure_password").Return(errors.New("token expired"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockSrv := mockAuthSrv.NewAuthService(t)
			if test.MockBehavior != nil {
				test.MockBehavior(mockSrv)
			}

			handler := handler.NewAuthHandler(mockSrv)

			requestJson, err := json.Marshal(test.Request)
			require.NoError(t, err, "request marshal should not return error")

			requestReader := bytes.NewReader(requestJson)
			request := httptest.NewRequest(http.MethodPost, "/reset-password", requestReader)
			response := httptest.NewRecorder()

			handler.ResetUserPassword(response, request)

			responseJson, err := json.Marshal(test.ExpectedResponse)
			require.NoError(t, err, "response marshal should not return error")

			assert.Equal(t, test.ExpectedStatusCode, response.Code, "incorrect status code")
			assert.Equal(t, string(responseJson), response.Body.String(), "incorrect body")
		})
	}
}

func TestResetUserPasswordWithRawJSON(t *testing.T) {
	t.Run("Incorrect JSON", func(t *testing.T) {
		incorrectJson := `{"password":"123", "repeat"`
		requestBody := strings.NewReader(incorrectJson)

		expectedResponse := newErrorResponse(http.StatusBadRequest, handler.InvalidDataMessage)
		expectedBody, err := json.Marshal(expectedResponse)
		require.NoError(t, err)

		mockSrv := mockAuthSrv.NewAuthService(t)
		handler := handler.NewAuthHandler(mockSrv)

		req := httptest.NewRequest(http.MethodPost, "/reset-password", requestBody)
		res := httptest.NewRecorder()

		handler.ResetUserPassword(res, req)

		assert.Equal(t, expectedResponse.Code, res.Code)
		assert.Equal(t, string(expectedBody), res.Body.String())
	})
}
