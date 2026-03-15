package board

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	mockBoardService "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/handler/tests/mock_board_service"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/models"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type GetBoardsTestCase struct {
	Name               string
	CtxValue           any
	MockBehavior       func(m *mockBoardService.BoardService)
	ExpectedStatusCode int
	ExpectedResponse   any
}

func newOkResponse[T any](status string, data T) api.OkResponse[T] {
	return api.OkResponse[T]{
		Response: api.Response{
			Status: status,
		},
		Data: data,
	}
}

func newErrorResponse(code int, message string) api.ErrorResponse {
	return api.ErrorResponse{
		Response: api.Response{
			Status: api.StatusError,
		},
		Code:    code,
		Message: message,
	}
}

func TestGetUserBoards(t *testing.T) {
	targetUserID := common.FixedUserUuiD
	boardID := common.FixedBoardUuiD

	tests := []GetBoardsTestCase{
		{
			Name:     "Success get boards",
			CtxValue: targetUserID,
			MockBehavior: func(m *mockBoardService.BoardService) {
				m.On("GetBoards", mock.Anything, targetUserID).Return(
					[]models.Board{
						{ID: boardID},
					}, nil,
				)
			},
			ExpectedStatusCode: http.StatusOK,
			ExpectedResponse: newOkResponse(api.StatusOK, []models.Board{
				{ID: boardID},
			}),
		},
		{
			Name:               "Error user not authorized",
			CtxValue:           nil,
			MockBehavior:       nil,
			ExpectedStatusCode: http.StatusUnauthorized,
			ExpectedResponse:   newErrorResponse(http.StatusUnauthorized, unauthorizedMessage),
		},
		{
			Name:               "Error context value is not UUID",
			CtxValue:           "some-string-id",
			MockBehavior:       nil,
			ExpectedStatusCode: http.StatusUnauthorized,
			ExpectedResponse:   newErrorResponse(http.StatusUnauthorized, unauthorizedMessage),
		},
		{
			Name:     "User not found",
			CtxValue: targetUserID,
			MockBehavior: func(m *mockBoardService.BoardService) {
				m.On("GetBoards", mock.Anything, targetUserID).Return(
					[]models.Board{},
					fmt.Errorf("rep.GetBoards: %w", common.ErrorNonexistentUser),
				)
			},
			ExpectedStatusCode: http.StatusUnauthorized,
			ExpectedResponse:   newErrorResponse(http.StatusUnauthorized, unauthorizedMessage),
		},
		{
			Name:     "Success empty boards list",
			CtxValue: targetUserID,
			MockBehavior: func(m *mockBoardService.BoardService) {
				m.On("GetBoards", mock.Anything, targetUserID).Return([]models.Board{}, nil)
			},
			ExpectedStatusCode: http.StatusOK,
			ExpectedResponse:   newOkResponse(api.StatusOK, []models.Board{}),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockBoardService := mockBoardService.NewBoardService(t)
			if test.MockBehavior != nil {
				test.MockBehavior(mockBoardService)
			}

			handler := NewBoardHandler(mockBoardService)

			request := httptest.NewRequest(http.MethodGet, "/", nil)

			if test.CtxValue != nil {
				ctx := context.WithValue(request.Context(), middleware.UserIDKey{}, test.CtxValue)
				request = request.WithContext(ctx)
			}

			response := httptest.NewRecorder()
			handler.GetUserBoards(response, request)

			responseJson, err := json.Marshal(test.ExpectedResponse)
			require.NoError(t, err, "response marshal should not return error")

			assert.Equal(t, test.ExpectedStatusCode, response.Code, "incorrect status code")
			assert.Equal(t, string(responseJson), response.Body.String(), "incorrect response body")
		})
	}
}
