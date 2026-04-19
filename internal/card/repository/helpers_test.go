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
	testMaxTask := 5

	queryCheckLimits := `
		WITH locked_section AS (
			SELECT section_link
			FROM section
			WHERE section_link = $1 AND deleted_at IS NULL
			FOR UPDATE
		),
		count_tasks AS (
			SELECT COUNT(t.section_link) AS count
			FROM task_version t
			JOIN locked_section ls ON t.section_link = ls.section_link
			WHERE t.valid_to IS NULL
		)

		SELECT c.count, v.max_tasks
		FROM locked_section ls
		JOIN section_version v
			ON v.section_link = ls.section_link AND v.valid_to IS NULL
		CROSS JOIN count_tasks c
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
					WillReturnRows(pgxmock.NewRows([]string{"count", "max_tasks"}).AddRow(3, &testMaxTask))
			},
		},
		{
			nameTest:      "Limit reached",
			linkSection:   targetSectionID,
			expectedError: common.ErrorRichLimitTasks,
			mockSetup: func(mock pgxmock.PgxPoolIface, sectionID uuid.UUID) {
				mock.ExpectQuery(regexp.QuoteMeta(queryCheckLimits)).
					WithArgs(sectionID).
					WillReturnRows(pgxmock.NewRows([]string{"count", "max_tasks"}).AddRow(10, &testMaxTask))
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
