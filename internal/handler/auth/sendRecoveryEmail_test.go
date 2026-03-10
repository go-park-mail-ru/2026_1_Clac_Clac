package auth

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
	mockAuthSrv "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/handler/auth/mock_auth_srv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendRecoveryEmail(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Success send email",
			Request: api.PasswordRecoveryRequest{
				Email: "test@mail.ru",
			},
			ExpectedResponse:   newResponse(api.StatusOK),
			ExpectedStatusCode: http.StatusOK,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("SendRecoveryCode", ctx, "test@mail.ru").Return(nil)
			},
		},
		{
			Name: "Service error",
			Request: api.PasswordRecoveryRequest{
				Email: "notfound@mail.ru",
			},
			ExpectedResponse:   newErrorResponse(http.StatusInternalServerError, cannotSendEmail),
			ExpectedStatusCode: http.StatusInternalServerError,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("SendRecoveryCode", ctx, "notfound@mail.ru").Return(errors.New("some internal error"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockSrv := mockAuthSrv.NewAuthService(t)
			if test.MockBehavior != nil {
				test.MockBehavior(mockSrv)
			}

			handler := NewAuthHandler(mockSrv)

			requestJson, err := json.Marshal(test.Request)
			require.NoError(t, err, "request marshal should not return error")

			requestReader := bytes.NewReader(requestJson)
			request := httptest.NewRequest(http.MethodPost, "/forgot-password", requestReader)
			response := httptest.NewRecorder()

			handler.SendRecoveryEmail(response, request)

			responseJson, err := json.Marshal(test.ExpectedResponse)
			require.NoError(t, err, "response marshal should not return error")

			assert.Equal(t, test.ExpectedStatusCode, response.Code, "incorrect status code")
			assert.Equal(t, string(responseJson), response.Body.String(), "incorrect body")
		})
	}
}

func TestSendRecoveryEmailWithRawJSON(t *testing.T) {
	t.Run("Incorrect JSON", func(t *testing.T) {
		incorrectJson := `{"email":"test@mail.ru",,,}`
		requestBody := strings.NewReader(incorrectJson)

		expectedResponse := newErrorResponse(http.StatusBadRequest, invalidDataMessage)
		expectedBody, err := json.Marshal(expectedResponse)
		require.NoError(t, err)

		mockSrv := mockAuthSrv.NewAuthService(t)
		handler := NewAuthHandler(mockSrv)

		req := httptest.NewRequest(http.MethodPost, "/forgot-password", requestBody)
		res := httptest.NewRecorder()

		handler.SendRecoveryEmail(res, req)

		assert.Equal(t, expectedResponse.Code, res.Code, "incorrect status code")
		assert.Equal(t, string(expectedBody), res.Body.String(), "incorrect body")
	})
}
