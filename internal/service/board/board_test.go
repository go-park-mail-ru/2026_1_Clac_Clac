package board

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository"
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
		mockBehavior   func(m *mockBoardRep.BoardRepository)
		expectedBoards []models.Board
	}{
		{
			nameTest: "Success get boards",
			userID:   targetUserID,
			mockBehavior: func(m *mockBoardRep.BoardRepository) {
				m.On("GetBoards", context.Background(), targetUserID).Return(expectedBoards, nil)
			},
			expectedBoards: expectedBoards,
		},
		{
			nameTest: "User has no boards",
			userID:   targetUserID,
			mockBehavior: func(m *mockBoardRep.BoardRepository) {
				m.On("GetBoards", context.Background(), targetUserID).Return([]models.Board{}, nil)
			},
			expectedBoards: []models.Board{},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockBoardRep.NewBoardRepository(t)
			ctx := context.Background()

			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			boardService := NewBoardService(mockRepo)

			boards, _ := boardService.GetBoards(ctx, test.userID)

			assert.Equal(t, test.expectedBoards, boards, "returned boards mismatch")
		})
	}
}

func TestGetBoardsError(t *testing.T) {
	targetUserID := uuid.New()

	tests := []struct {
		nameTest      string
		userID        uuid.UUID
		mockBehavior  func(m *mockBoardRep.BoardRepository)
		expectedError error
		expectedBoard []models.Board
	}{
		{
			nameTest: "User not found",
			userID:   targetUserID,
			mockBehavior: func(m *mockBoardRep.BoardRepository) {
				m.On("GetBoards", context.Background(), targetUserID).Return(nil, repository.ErrorNonexistentUser)
			},
			expectedError: fmt.Errorf("rep.GetBoards: %w", repository.ErrorNonexistentUser),
			expectedBoard: []models.Board{},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockBoardRep.NewBoardRepository(t)
			ctx := context.Background()

			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			boardService := NewBoardService(mockRepo)

			boards, err := boardService.GetBoards(ctx, test.userID)

			assert.Equal(t, test.expectedBoard, boards, "returned boards mismatch")
			assert.Equal(t, test.expectedError, err)
		})
	}
}
