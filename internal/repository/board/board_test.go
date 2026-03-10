package board

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	dbConnection "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/db_connection"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

			boards, _ := baords.GetBoards(ctx, test.targetID)

			assert.Equal(t, test.expectedBoards, boards)
		})
	}
}

func TestGetBoardsError(t *testing.T) {
	userID1 := uuid.New()
	board1 := models.Board{ID: uuid.New()}

	tests := []struct {
		nameTest       string
		targetID       uuid.UUID
		users          []models.User
		expectedBoards []models.Board
		expectedError  error
	}{
		{
			nameTest: "User not found",
			targetID: uuid.New(),
			users: []models.User{
				{ID: userID1, Boards: []models.Board{board1}},
			},
			expectedBoards: nil,
			expectedError:  common.ErrorNonexistentUser,
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

			boards, err := baords.GetBoards(ctx, test.targetID)

			assert.Equal(t, test.expectedBoards, boards)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestCreateEmptyBoard(t *testing.T) {
	userID1 := uuid.New()
	userID2 := uuid.New()
	board1 := models.Board{ID: uuid.New()}

	tests := []struct {
		nameTest            string
		targetID            uuid.UUID
		users               []models.User
		expectedBoardsCount int
	}{
		{
			nameTest: "success create board for new user",
			targetID: userID1,
			users: []models.User{
				{ID: userID1, Boards: []models.Board{}},
			},
			expectedBoardsCount: 1,
		},
		{
			nameTest: "success create board for old user",
			targetID: userID2,
			users: []models.User{
				{ID: userID2, Boards: []models.Board{board1}},
			},
			expectedBoardsCount: 2,
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

			err := baords.CreateEmptyBoard(ctx, test.targetID)
			require.NoError(t, err)

			updatedUser := conectionDb.UsersDB[test.targetID]
			assert.Equal(t, test.expectedBoardsCount, len(updatedUser.Boards))

			if len(updatedUser.Boards) > 0 {
				lastBoard := updatedUser.Boards[len(updatedUser.Boards)-1]
				assert.NotEqual(t, uuid.Nil, lastBoard.ID)
			}
		})
	}
}

func TestCreateEmptyBoardError(t *testing.T) {
	userID1 := uuid.New()

	tests := []struct {
		nameTest      string
		targetID      uuid.UUID
		users         []models.User
		expectedError error
	}{
		{
			nameTest: "user not found",
			targetID: uuid.New(),
			users: []models.User{
				{ID: userID1, Boards: []models.Board{}},
			},
			expectedError: common.ErrorNonexistentUser,
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

			err := baords.CreateEmptyBoard(ctx, test.targetID)

			assert.Equal(t, test.expectedError, err)
		})
	}
}
