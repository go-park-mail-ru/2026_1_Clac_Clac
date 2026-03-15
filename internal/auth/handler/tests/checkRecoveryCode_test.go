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

func TestCheckRecoveryCode(t *testing.T) {
	tests := []TestCase{
		{
			Name: "Success check code",
			Request: api.RecoveryCodeRequest{
				Code: "123456",
			},
			ExpectedResponse:   newResponse(api.StatusOK),
			ExpectedStatusCode: http.StatusOK,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("CheckRecoveryCode", ctx, "123456").Return(nil)
			},
		},
		{
			Name: "Wrong or expired code",
			Request: api.RecoveryCodeRequest{
				Code: "000000",
			},
			ExpectedResponse:   newErrorResponse(http.StatusInternalServerError, handler.SomethingWentWrong),
			ExpectedStatusCode: http.StatusInternalServerError,
			MockBehavior: func(m *mockAuthSrv.AuthService) {
				ctx := context.Background()
				m.On("CheckRecoveryCode", ctx, "000000").Return(errors.New("invalid code"))
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
			request := httptest.NewRequest(http.MethodPost, "/verify-code", requestReader)
			response := httptest.NewRecorder()

			handler.CheckRecoveryCode(response, request)

			responseJson, err := json.Marshal(test.ExpectedResponse)
			require.NoError(t, err, "response marshal should not return error")

			assert.Equal(t, test.ExpectedStatusCode, response.Code, "incorrect status code")
			assert.Equal(t, string(responseJson), response.Body.String(), "incorrect body")
		})
	}
}

func TestCheckRecoveryCodeWithRawJSON(t *testing.T) {
	t.Run("Incorrect JSON", func(t *testing.T) {
		incorrectJson := `{"code":123456`
		requestBody := strings.NewReader(incorrectJson)

		expectedResponse := newErrorResponse(http.StatusBadRequest, handler.InvalidDataMessage)
		expectedBody, err := json.Marshal(expectedResponse)
		require.NoError(t, err)

		mockSrv := mockAuthSrv.NewAuthService(t)
		handler := handler.NewAuthHandler(mockSrv)

		req := httptest.NewRequest(http.MethodPost, "/verify-code", requestBody)
		res := httptest.NewRecorder()

		handler.CheckRecoveryCode(res, req)

		assert.Equal(t, expectedResponse.Code, res.Code)
		assert.Equal(t, string(expectedBody), res.Body.String())
	})
}
