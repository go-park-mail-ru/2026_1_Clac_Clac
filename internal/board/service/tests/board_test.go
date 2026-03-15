package board

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/models"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/service"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"

	mockBoardRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/service/tests/mock_board_rep"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

			boardService := service.NewBoardService(mockRepo)

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
				m.On("GetBoards", context.Background(), targetUserID).Return(nil, common.ErrorNonexistentUser)
			},
			expectedError: fmt.Errorf("rep.GetBoards: %w", common.ErrorNonexistentUser),
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

			boardService := service.NewBoardService(mockRepo)

			boards, err := boardService.GetBoards(ctx, test.userID)

			assert.Equal(t, test.expectedBoard, boards, "returned boards mismatch")
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestAddEmptyBoard(t *testing.T) {
	targetUserID := uuid.New()
	expectedRepoError := errors.New("database connection error")

	tests := []struct {
		nameTest      string
		userID        uuid.UUID
		mockBehavior  func(m *mockBoardRep.BoardRepository)
		expectedError error
	}{
		{
			nameTest: "Success create empty board",
			userID:   targetUserID,
			mockBehavior: func(m *mockBoardRep.BoardRepository) {
				m.On("AddEmptyBoard", context.Background(), mock.AnythingOfType("db.Board"), targetUserID).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Repository error",
			userID:   targetUserID,
			mockBehavior: func(m *mockBoardRep.BoardRepository) {
				m.On("AddEmptyBoard", context.Background(), mock.AnythingOfType("db.Board"), targetUserID).Return(expectedRepoError)
			},

			expectedError: fmt.Errorf("rep.AddEmptyBoard: %w", expectedRepoError),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockBoardRep.NewBoardRepository(t)
			ctx := context.Background()

			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			boardService := service.NewBoardService(mockRepo)

			err := boardService.CreateEmptyBoard(ctx, test.userID)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err, "expected no error")
			}
		})
	}
}
