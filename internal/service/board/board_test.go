package board

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	mockBoardRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/board/mock_board_rep"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetBoards(t *testing.T) {
	targetUserID := uuid.New()

	expectedBoards := []models.Board{
		{ID: uuid.New()},
		{ID: uuid.New()},
	}

	tests := []struct {
		nameTest       string
		userID         uuid.UUID
		mockBehavior   func(m *mockBoardRep.BoardRepositpry)
		expectedBoards []models.Board
	}{
		{
			nameTest: "Success get boards",
			userID:   targetUserID,
			mockBehavior: func(m *mockBoardRep.BoardRepositpry) {
				m.On("GetBoards", context.Background(), targetUserID).Return(expectedBoards)
			},
			expectedBoards: expectedBoards,
		},
		{
			nameTest: "User has no boards",
			userID:   targetUserID,
			mockBehavior: func(m *mockBoardRep.BoardRepositpry) {
				m.On("GetBoards", context.Background(), targetUserID).Return([]models.Board{})
			},
			expectedBoards: []models.Board{},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockBoardRep.NewBoardRepositpry(t)
			ctx := context.Background()

			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			boardService := NewBoardService(mockRepo)

			boards := boardService.GetBoards(ctx, test.userID)

			assert.Equal(t, test.expectedBoards, boards, "returned boards mismatch")
		})
	}
}
