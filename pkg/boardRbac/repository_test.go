package rbac

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
)

func TestRepository_GetUserRoleByBoardLink(t *testing.T) {
	ctx := context.Background()
	boardLink := uuid.New()
	userLink := uuid.New()

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedRole  Role
		expectedError error
	}{
		{
			nameTest: "Success get role",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("(?is)SELECT level_member FROM member_board.*").
					WithArgs(boardLink, userLink).
					WillReturnRows(pgxmock.NewRows([]string{"level_member"}).AddRow(Roles.Admin))
			},
			expectedRole:  Roles.Admin,
			expectedError: nil,
		},
		{
			nameTest: "User not found (pgx.ErrNoRows)",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("(?is)SELECT level_member FROM member_board.*").
					WithArgs(boardLink, userLink).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedRole:  Roles.None,
			expectedError: nil,
		},
		{
			nameTest: "DB Error",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("(?is)SELECT level_member FROM member_board.*").
					WithArgs(boardLink, userLink).
					WillReturnError(errors.New("db error"))
			},
			expectedRole:  Roles.None,
			expectedError: errors.New("get user role on board: db error"),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockDB, err := pgxmock.NewPool()
			if !assert.NoError(t, err) {
				return
			}
			defer mockDB.Close()

			test.mockBehavior(mockDB)

			repo := NewRepository(mockDB)
			role, err := repo.GetUserRoleByBoardLink(ctx, boardLink, userLink)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expectedRole, role)
			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestRepository_GetUserRoleByLink_Generic(t *testing.T) {
	ctx := context.Background()
	itemLink := uuid.New()
	userLink := uuid.New()
	expectedBoardLink := uuid.New()

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedRole  Role
		expectedBoard uuid.UUID
		expectedError error
	}{
		{
			nameTest: "Success get role",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("(?is)SELECT m.level_member, s.board_link.*").
					WithArgs(itemLink, userLink).
					WillReturnRows(pgxmock.NewRows([]string{"level_member", "board_link"}).AddRow(Roles.Editor, expectedBoardLink))
			},
			expectedRole:  Roles.Editor,
			expectedBoard: expectedBoardLink,
			expectedError: nil,
		},
		{
			nameTest: "Not found (pgx.ErrNoRows)",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("(?is)SELECT m.level_member, s.board_link.*").
					WithArgs(itemLink, userLink).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedRole:  Roles.None,
			expectedBoard: uuid.Nil,
			expectedError: nil,
		},
		{
			nameTest: "DB Error",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("(?is)SELECT m.level_member, s.board_link.*").
					WithArgs(itemLink, userLink).
					WillReturnError(errors.New("db error"))
			},
			expectedRole:  Roles.None,
			expectedBoard: uuid.Nil,
			expectedError: errors.New("db error"),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest+"_Section", func(t *testing.T) {
			mockDB, _ := pgxmock.NewPool()
			defer mockDB.Close()

			test.mockBehavior(mockDB)
			repo := NewRepository(mockDB)
			role, bLink, err := repo.GetUserRoleBySectionLink(ctx, itemLink, userLink)

			if test.expectedError != nil {
				assert.ErrorContains(t, err, test.expectedError.Error())
				assert.ErrorContains(t, err, "get user role on section")
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expectedRole, role)
			assert.Equal(t, test.expectedBoard, bLink)
		})

		t.Run(test.nameTest+"_Card", func(t *testing.T) {
			mockDB, _ := pgxmock.NewPool()
			defer mockDB.Close()

			test.mockBehavior(mockDB)
			repo := NewRepository(mockDB)
			role, bLink, err := repo.GetUserRoleByCardLink(ctx, itemLink, userLink)

			if test.expectedError != nil {
				assert.ErrorContains(t, err, test.expectedError.Error())
				assert.ErrorContains(t, err, "get user role on card")
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expectedRole, role)
			assert.Equal(t, test.expectedBoard, bLink)
		})

		t.Run(test.nameTest+"_Comment", func(t *testing.T) {
			mockDB, _ := pgxmock.NewPool()
			defer mockDB.Close()

			test.mockBehavior(mockDB)
			repo := NewRepository(mockDB)
			role, bLink, err := repo.GetUserRoleByCommentLink(ctx, itemLink, userLink)

			if test.expectedError != nil {
				assert.ErrorContains(t, err, test.expectedError.Error())
				assert.ErrorContains(t, err, "get user role on comment")
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expectedRole, role)
			assert.Equal(t, test.expectedBoard, bLink)
		})

		t.Run(test.nameTest+"_Subtask", func(t *testing.T) {
			mockDB, _ := pgxmock.NewPool()
			defer mockDB.Close()

			test.mockBehavior(mockDB)
			repo := NewRepository(mockDB)
			role, bLink, err := repo.GetUserRoleBySubtaskLink(ctx, itemLink, userLink)

			if test.expectedError != nil {
				assert.ErrorContains(t, err, test.expectedError.Error())
				assert.ErrorContains(t, err, "get user role on subtask")
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expectedRole, role)
			assert.Equal(t, test.expectedBoard, bLink)
		})
	}
}
