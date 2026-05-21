package api_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Константы намеренно продублированы, чтобы тесты не зависели от возможных опечаток в пакете api.
const (
	HeaderContentType   = "Content-Type"
	MIMEApplicationJSON = "application/json"
	MIMETextPlain       = "text/plain"
	StatusOK            = "ok"
	StatusError         = "error"
)

func TestSetContentType(t *testing.T) {
	tests := []struct {
		name        string
		first       string
		second      string
		expectedCT  string
	}{
		{
			name:       "SingleWrite",
			first:      MIMETextPlain,
			second:     "",
			expectedCT: MIMETextPlain,
		},
		{
			name:       "DoubleWrite_IdempotentFirstValue",
			first:      MIMETextPlain,
			second:     MIMEApplicationJSON,
			expectedCT: MIMETextPlain,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := httptest.NewRecorder()
			api.SetContentType(res, tc.first)
			if tc.second != "" {
				api.SetContentType(res, tc.second)
			}
			assert.Equal(t, tc.expectedCT, res.Header().Get(HeaderContentType))
		})
	}
}

func TestRespond(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		status       string
		expectedBody string
	}{
		{
			name:         "OK",
			statusCode:   http.StatusOK,
			status:       StatusOK,
			expectedBody: fmt.Sprintf(`{"status":"%s"}`, StatusOK),
		},
		{
			name:         "Error",
			statusCode:   http.StatusBadRequest,
			status:       StatusError,
			expectedBody: fmt.Sprintf(`{"status":"%s"}`, StatusError),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := httptest.NewRecorder()
			_, err := api.Respond(res, tc.statusCode, tc.status)

			require.NoError(t, err)
			assert.Equal(t, MIMEApplicationJSON, res.Header().Get(HeaderContentType))
			assert.Equal(t, tc.statusCode, res.Result().StatusCode)
			assert.Equal(t, tc.expectedBody, res.Body.String())
		})
	}
}

func TestRespondOk(t *testing.T) {
	type simpleUser struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	tests := []struct {
		name         string
		data         simpleUser
		expectedBody string
	}{
		{
			name:         "SimpleUser",
			data:         simpleUser{ID: 5, Name: "TempName"},
			expectedBody: fmt.Sprintf(`{"status":"%s","data":{"id":5,"name":"TempName"}}`, StatusOK),
		},
		{
			name:         "EmptyUser",
			data:         simpleUser{},
			expectedBody: fmt.Sprintf(`{"status":"%s","data":{"id":0,"name":""}}`, StatusOK),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := httptest.NewRecorder()
			_, err := api.RespondOk(res, tc.data)

			require.NoError(t, err)
			assert.Equal(t, MIMEApplicationJSON, res.Header().Get(HeaderContentType))
			assert.Equal(t, http.StatusOK, res.Result().StatusCode)
			assert.Equal(t, tc.expectedBody, res.Body.String())
		})
	}
}

func TestRespondError(t *testing.T) {
	tests := []struct {
		name         string
		code         int
		message      string
		expectedBody string
	}{
		{
			name:         "BadRequest",
			code:         http.StatusBadRequest,
			message:      "this is error message",
			expectedBody: fmt.Sprintf(`{"code":%d,"message":"this is error message","status":"%s"}`, http.StatusBadRequest, StatusError),
		},
		{
			name:         "NotFound",
			code:         http.StatusNotFound,
			message:      "not found",
			expectedBody: fmt.Sprintf(`{"code":%d,"message":"not found","status":"%s"}`, http.StatusNotFound, StatusError),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := httptest.NewRecorder()
			_, err := api.RespondError(res, tc.code, tc.message)

			require.NoError(t, err)
			assert.Equal(t, MIMEApplicationJSON, res.Header().Get(HeaderContentType))
			assert.Equal(t, tc.code, res.Result().StatusCode)
			assert.Equal(t, tc.expectedBody, res.Body.String())
		})
	}
}

func TestHandleError(t *testing.T) {
	tests := []struct {
		name           string
		inputErr       error
		expectedStatus int
		expectBody     bool
		expectErr      bool
	}{
		{
			name:           "NoError",
			inputErr:       nil,
			expectedStatus: 0,
			expectBody:     false,
			expectErr:      false,
		},
		{
			name:           "WithError",
			inputErr:       errors.New("oh no..."),
			expectedStatus: http.StatusInternalServerError,
			expectBody:     true,
			expectErr:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			const zeroStatus = 0
			res := httptest.NewRecorder()
			if tc.inputErr == nil {
				res.Code = zeroStatus
			}

			err := api.HandleError(res, tc.inputErr)

			if tc.expectErr {
				require.Error(t, err)
				assert.NotEmpty(t, res.Body.String())
				assert.Equal(t, tc.expectedStatus, res.Code)
			} else {
				require.NoError(t, err)
				assert.Empty(t, res.Body.String())
				assert.Equal(t, zeroStatus, res.Code)
			}
		})
	}
}
