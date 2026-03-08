package board

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	dbConnection "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/db_connection"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetBoards(t *testing.T) {
	userID1 := uuid.New()
	userID2 := uuid.New()

	board1 := models.Board{ID: uuid.New()}
	board2 := models.Board{ID: uuid.New()}
	board3 := models.Board{ID: uuid.New()}

	tests := []struct {
		nameTest       string
		targetID       uuid.UUID
		users          []models.User
		expectedBoards []models.Board
	}{
		{
			nameTest: "Success get user boards",
			targetID: userID1,
			users: []models.User{
				{ID: userID1, Boards: []models.Board{board1, board2}},
				{ID: userID2, Boards: []models.Board{board3}},
			},
			expectedBoards: []models.Board{board1, board2},
		},
		{
			nameTest: "User has no boards",
			targetID: userID1,
			users: []models.User{
				{ID: userID1, Boards: []models.Board{}},
			},
			expectedBoards: []models.Board{},
		},
		{
			nameTest: "User not found",
			targetID: uuid.New(),
			users: []models.User{
				{ID: userID2, Boards: []models.Board{board3}},
			},
			expectedBoards: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			conectionDb := dbConnection.NewMapDatabse()

			for _, user := range test.users {
				conectionDb.UsersDB[user.ID] = user
			}

			baords := NewBoardRepository(conectionDb)
			ctx := context.Background()

			boards := baords.GetBoards(ctx, test.targetID)

			assert.Equal(t, test.expectedBoards, boards)
		})
	}
}
