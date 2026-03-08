package board

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	mockBoardService "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/handler/board/mock_board_service"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetUserBoards(t *testing.T) {
	targetUserID := common.FixedUserUuiD
	boardID := common.FixedBoardUuiD

	tests := []struct {
		nameTest           string
		ctxValue           interface{}
		mockBehavior       func(m *mockBoardService.BoardService)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			nameTest: "Success get boards",
			ctxValue: targetUserID,
			mockBehavior: func(m *mockBoardService.BoardService) {
				m.On("GetBoards", mock.Anything, targetUserID).Return(
					[]models.Board{
						{ID: boardID},
					}, nil,
				)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   "[{\"id\":\"22222222-2222-2222-2222-222222222222\"}]\n",
		},
		{
			nameTest:           "Error user not authorized",
			ctxValue:           nil,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   "{\"error\":\"user was not authorized\"}\n",
		},
		{
			nameTest:           "Error context value is not UUID",
			ctxValue:           "some-string-id",
			mockBehavior:       nil,
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   "{\"error\":\"user was not authorized\"}\n",
		},
		{
			nameTest: "User not found",
			ctxValue: targetUserID,
			mockBehavior: func(m *mockBoardService.BoardService) {
				m.On("GetBoards", mock.Anything, targetUserID).Return([]models.Board{}, fmt.Errorf("rep.GetBoards: %w", repository.ErrorNonexistentUser))
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   "{\"error\":\"user not found: rep.GetBoards: user with this ID not exist\"}\n",
		},
		{
			nameTest: "Success empty boards list",
			ctxValue: targetUserID,
			mockBehavior: func(m *mockBoardService.BoardService) {
				m.On("GetBoards", mock.Anything, targetUserID).Return([]models.Board{}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   "[]\n",
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockBoardService := mockBoardService.NewBoardService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockBoardService)
			}

			handler := NewBoardHandler(mockBoardService)

			request := httptest.NewRequest(http.MethodGet, "/boards", nil)

			if test.ctxValue != nil {
				ctx := context.WithValue(request.Context(), middleware.UserIDKey, test.ctxValue)
				request = request.WithContext(ctx)
			}

			response := httptest.NewRecorder()
			handler.GetUserBoards(response, request)

			assert.Equal(t, test.expectedStatusCode, response.Code, "incorrect status code")
			assert.Equal(t, test.expectedResponse, response.Body.String(), "incorrect response body")
		})
	}
}
