package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckLimitTasks(t *testing.T) {
	targetSectionID := uuid.New()

	queryCheckLimits := `
		WITH count_tasks AS (
			SELECT COUNT(section_link) AS count
			FROM task_version
			WHERE section_link = $1 AND valid_to IS NULL
		)
		SELECT c.count, s.max_tasks
		FROM section s
		CROSS JOIN count_tasks c
		WHERE section_link = $1
		FOR UPDATE OF s
		`

	tests := []struct {
		nameTest      string
		linkSection   uuid.UUID
		expectedError error
		mockSetup     func(mock pgxmock.PgxPoolIface, sectionID uuid.UUID)
	}{
		{
			nameTest:      "No limit max tasks is NULL",
			linkSection:   targetSectionID,
			expectedError: nil,
			mockSetup: func(mock pgxmock.PgxPoolIface, sectionID uuid.UUID) {
				mock.ExpectQuery(regexp.QuoteMeta(queryCheckLimits)).
					WithArgs(sectionID).
					WillReturnRows(pgxmock.NewRows([]string{"count", "max_tasks"}).AddRow(5, nil))
			},
		},
		{
			nameTest:      "Limit not reached",
			linkSection:   targetSectionID,
			expectedError: nil,
			mockSetup: func(mock pgxmock.PgxPoolIface, sectionID uuid.UUID) {
				mock.ExpectQuery(regexp.QuoteMeta(queryCheckLimits)).
					WithArgs(sectionID).
					WillReturnRows(pgxmock.NewRows([]string{"count", "max_tasks"}).AddRow(3, 10))
			},
		},
		{
			nameTest:      "Generic DB Error",
			linkSection:   targetSectionID,
			expectedError: errors.New("db error"),
			mockSetup: func(mock pgxmock.PgxPoolIface, sectionID uuid.UUID) {
				mock.ExpectQuery(regexp.QuoteMeta(queryCheckLimits)).
					WithArgs(sectionID).
					WillReturnError(errors.New("db error"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			mockPool.ExpectBegin()

			if test.mockSetup != nil {
				test.mockSetup(mockPool, test.linkSection)
			}

			mockPool.ExpectRollback()

			ctx := context.Background()
			tx, err := mockPool.Begin(ctx)
			require.NoError(t, err)

			err = checkLimitTasks(ctx, tx, test.linkSection)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					if errors.Is(test.expectedError, common.ErrorRichLimitTasks) {
						assert.ErrorIs(t, err, test.expectedError)
					}
				}
			} else {
				assert.NoError(t, err)
			}

			_ = tx.Rollback(ctx)

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}
