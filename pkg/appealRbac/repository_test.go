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

func TestRepository_GetUserRoleByLink(t *testing.T) {
	ctx := context.Background()
	userLink := uuid.New()

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedRole  Role
		expectedError error
	}{
		{
			nameTest: "Success get Support role",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("(?is)SELECT role FROM support.*").
					WithArgs(userLink).
					WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(Roles.Support))
			},
			expectedRole:  Roles.Support,
			expectedError: nil,
		},
		{
			nameTest: "User not found - returns User role",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("(?is)SELECT role FROM support.*").
					WithArgs(userLink).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedRole:  Roles.User,
			expectedError: nil,
		},
		{
			nameTest: "DB Error",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery("(?is)SELECT role FROM support.*").
					WithArgs(userLink).
					WillReturnError(errors.New("db error"))
			},
			expectedRole:  Roles.User,
			expectedError: errors.New("support get user role: db error"),
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
			role, err := repo.GetUserRoleByLink(ctx, userLink)

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
