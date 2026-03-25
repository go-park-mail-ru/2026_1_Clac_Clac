package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/models"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBoards(t *testing.T) {
	userID1 := uuid.New()

	board1 := models.Board{Link: uuid.New(), Created_at: time.Now()}
	board2 := models.Board{Link: uuid.New(), Created_at: time.Now()}

	tests := []struct {
		nameTest       string
		targetID       uuid.UUID
		mockSetup      func(mock pgxmock.PgxPoolIface, targetID uuid.UUID)
		expectedBoards []models.Board
	}{
		{
			nameTest: "Success get user boards",
			targetID: userID1,
			mockSetup: func(mock pgxmock.PgxPoolIface, targetID uuid.UUID) {
				checkQuery := `SELECT EXISTS\(SELECT 1 FROM "user" WHERE link = \$1\)`
				mock.ExpectQuery(checkQuery).
					WithArgs(targetID).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

				getBoardQuery := `SELECT b\.link, b\.created_at FROM board b JOIN member_board mb ON b\.link = mb\.board_link WHERE mb\.user_link = \$1`
				rows := pgxmock.NewRows([]string{"link", "created_at"}).
					AddRow(board1.Link, board1.Created_at).
					AddRow(board2.Link, board2.Created_at)

				mock.ExpectQuery(getBoardQuery).
					WithArgs(targetID).
					WillReturnRows(rows)
			},
			expectedBoards: []models.Board{board1, board2},
		},
		{
			nameTest: "User has no boards",
			targetID: userID1,
			mockSetup: func(mock pgxmock.PgxPoolIface, targetID uuid.UUID) {
				checkQuery := `SELECT EXISTS\(SELECT 1 FROM "user" WHERE link = \$1\)`
				mock.ExpectQuery(checkQuery).
					WithArgs(targetID).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

				getBoardQuery := `SELECT b\.link, b\.created_at FROM board b JOIN member_board mb ON b\.link = mb\.board_link WHERE mb\.user_link = \$1`
				rows := pgxmock.NewRows([]string{"link", "created_at"})

				mock.ExpectQuery(getBoardQuery).
					WithArgs(targetID).
					WillReturnRows(rows)
			},
			expectedBoards: []models.Board{},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			test.mockSetup(mock, test.targetID)

			repoBoards := NewRepository(mock)
			ctx := context.Background()

			boards, err := repoBoards.GetBoards(ctx, test.targetID)

			assert.NoError(t, err)
			assert.Equal(t, test.expectedBoards, boards)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err, "not wait error")
		})
	}
}

func TestGetBoardsError(t *testing.T) {
	tests := []struct {
		nameTest       string
		targetID       uuid.UUID
		mockSetup      func(mock pgxmock.PgxPoolIface, targetID uuid.UUID)
		expectedBoards []models.Board
		expectedError  error
	}{
		{
			nameTest: "User not found",
			targetID: uuid.New(),
			mockSetup: func(mock pgxmock.PgxPoolIface, targetID uuid.UUID) {
				checkQuery := `SELECT EXISTS\(SELECT 1 FROM "user" WHERE link = \$1\)`
				mock.ExpectQuery(checkQuery).
					WithArgs(targetID).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(false))
			},
			expectedBoards: []models.Board{},
			expectedError:  common.ErrorNonexistentUser,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mock.Close()

			test.mockSetup(mock, test.targetID)

			repoBoards := NewRepository(mock)
			ctx := context.Background()

			boards, err := repoBoards.GetBoards(ctx, test.targetID)

			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedBoards, boards)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err, "not wait error")
		})
	}
}

func TestCreateEmptyBoard(t *testing.T) {
	tests := []struct {
		nameTest      string
		targetID      uuid.UUID
		board         models.Board
		mockSetup     func(mock pgxmock.PgxPoolIface, targetID uuid.UUID, board models.Board)
		expectedError error
	}{
		{
			nameTest: "Success create empty board",
			targetID: uuid.New(),
			board:    models.Board{Link: uuid.New()},
			mockSetup: func(mock pgxmock.PgxPoolIface, targetID uuid.UUID, board models.Board) {

				addEmptyBoardQuery := `INSERT INTO board (link) VALUES ($1)`
				mock.ExpectExec(regexp.QuoteMeta(addEmptyBoardQuery)).
					WithArgs(board.Link).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				addMemberBoardQuery := `INSERT INTO member_board (board_link, user_link) VALUES ($1, $2)`
				mock.ExpectExec(regexp.QuoteMeta(addMemberBoardQuery)).
					WithArgs(board.Link, targetID).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			test.mockSetup(mock, test.targetID, test.board)

			repoBoards := NewRepository(mock)
			ctx := context.Background()

			err = repoBoards.AddEmptyBoard(ctx, test.board, test.targetID)

			if test.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, test.expectedError, err)
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err, "not wait error")
		})
	}
}
