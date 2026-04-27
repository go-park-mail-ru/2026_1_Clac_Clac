package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/common"
	dto "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/repository/dto"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryGetCard(t *testing.T) {
	ctx := context.Background()
	targetLink := uuid.New()
	targetDeadLine := time.Now()
	targetExecuter := "John Doe"

	expectedInfo := dto.InfoCard{
		Title:        "Test Task",
		Description:  "Description",
		DataDeadLine: &targetDeadLine,
		NameExecuter: &targetExecuter,
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedError error
		expectedData  dto.InfoCard
	}{
		{
			nameTest: "Success get card",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"title", "description", "due_date", "display_name"}).
					AddRow(expectedInfo.Title, expectedInfo.Description, expectedInfo.DataDeadLine, expectedInfo.NameExecuter)

				m.ExpectQuery(`(?s)SELECT.*t.title.*FROM task_actual.*WHERE t.task_link = \$1`).
					WithArgs(targetLink).
					WillReturnRows(rows)
			},
			expectedError: nil,
			expectedData:  expectedInfo,
		},
		{
			nameTest: "Error not found pgx.ErrNoRows",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?s)SELECT.*t.title.*FROM task_actual.*WHERE t.task_link = \$1`).
					WithArgs(targetLink).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedError: common.ErrCardNotFound,
			expectedData:  dto.InfoCard{},
		},
		{
			nameTest: "Error query fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?s)SELECT.*t.title.*FROM task_actual.*WHERE t.task_link = \$1`).
					WithArgs(targetLink).
					WillReturnError(errors.New("db disconnect"))
			},
			expectedError: errors.New("rep.QueryRow: db disconnect"),
			expectedData:  dto.InfoCard{},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockDB, err := pgxmock.NewPool()
			if !assert.NoError(t, err) {
				return
			}
			defer mockDB.Close()

			if test.mockBehavior != nil {
				test.mockBehavior(mockDB)
			}

			repo := NewRepository(mockDB)
			result, err := repo.GetCard(ctx, targetLink)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					if errors.Is(test.expectedError, common.ErrCardNotFound) {
						assert.ErrorIs(t, err, common.ErrCardNotFound)
					} else {
						assert.EqualError(t, err, test.expectedError.Error())
					}
				}
			} else {
				if assert.NoError(t, err) {
					assert.Equal(t, test.expectedData, result)
				}
			}

			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestRepositoryDeleteCard(t *testing.T) {
	ctx := context.Background()
	targetLink := uuid.New()

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			nameTest: "Success delete card",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)DELETE FROM task WHERE task_link = \$1;`).
					WithArgs(targetLink).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			expectedError: nil,
		},
		{
			nameTest: "Error not found",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)DELETE FROM task WHERE task_link = \$1;`).
					WithArgs(targetLink).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			expectedError: common.ErrCardNotFound,
		},
		{
			nameTest: "Error exec fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)DELETE FROM task WHERE task_link = \$1;`).
					WithArgs(targetLink).
					WillReturnError(errors.New("db disconnect"))
			},
			expectedError: errors.New("pool.Exec: db disconnect"),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockDB, err := pgxmock.NewPool()
			if !assert.NoError(t, err) {
				return
			}
			defer mockDB.Close()

			if test.mockBehavior != nil {
				test.mockBehavior(mockDB)
			}

			repo := NewRepository(mockDB)
			err = repo.DeleteCard(ctx, targetLink)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					if errors.Is(test.expectedError, common.ErrCardNotFound) {
						assert.ErrorIs(t, err, common.ErrCardNotFound)
					} else {
						assert.EqualError(t, err, test.expectedError.Error())
					}
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestRepositoryUpdateCardDetails(t *testing.T) {
	ctx := context.Background()
	targetLink := uuid.New()
	targetSection := uuid.New()
	targetExecuter := uuid.New()
	targetDeadLine := time.Now()

	updatingCard := dto.UpdatingCardDetails{
		LinkCard:     targetLink,
		Title:        "Updated Title",
		Description:  "Updated Desc",
		LinkExecuter: &targetExecuter,
		DataDeadLine: &targetDeadLine,
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			nameTest: "Success update card",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)UPDATE task_version.*RETURNING section_link, position;`).
					WithArgs(targetLink).
					WillReturnRows(pgxmock.NewRows([]string{"section_link", "position"}).AddRow(targetSection, 5))

				m.ExpectExec(`(?s)INSERT INTO task_version.*`).
					WithArgs(targetLink, targetSection, &targetExecuter, "Updated Title", "Updated Desc", 5, &targetDeadLine).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			nameTest: "Error not found",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)UPDATE task_version.*RETURNING section_link, position;`).
					WithArgs(targetLink).
					WillReturnError(pgx.ErrNoRows)

				m.ExpectRollback()
			},
			expectedError: common.ErrCardNotFound,
		},
		{
			nameTest: "Error invalid reference data",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)UPDATE task_version.*RETURNING section_link, position;`).
					WithArgs(targetLink).
					WillReturnRows(pgxmock.NewRows([]string{"section_link", "position"}).AddRow(targetSection, 5))

				m.ExpectExec(`(?s)INSERT INTO task_version.*`).
					WithArgs(targetLink, targetSection, &targetExecuter, "Updated Title", "Updated Desc", 5, &targetDeadLine).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})

				m.ExpectRollback()
			},
			expectedError: common.ErrInvalidReferenceCardData,
		},
		{
			nameTest: "Error check violation data",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)UPDATE task_version.*RETURNING section_link, position;`).
					WithArgs(targetLink).
					WillReturnRows(pgxmock.NewRows([]string{"section_link", "position"}).AddRow(targetSection, 5))

				m.ExpectExec(`(?s)INSERT INTO task_version.*`).
					WithArgs(targetLink, targetSection, &targetExecuter, "Updated Title", "Updated Desc", 5, &targetDeadLine).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.CheckViolation})

				m.ExpectRollback()
			},
			expectedError: common.ErrInvalidCardData,
		},
		{
			nameTest: "Error missing required field",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)UPDATE task_version.*RETURNING section_link, position;`).
					WithArgs(targetLink).
					WillReturnRows(pgxmock.NewRows([]string{"section_link", "position"}).AddRow(targetSection, 5))

				m.ExpectExec(`(?s)INSERT INTO task_version.*`).
					WithArgs(targetLink, targetSection, &targetExecuter, "Updated Title", "Updated Desc", 5, &targetDeadLine).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.NotNullViolation})

				m.ExpectRollback()
			},
			expectedError: common.ErrMissingRequiredField,
		},
		{
			nameTest: "Error insert fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)UPDATE task_version.*RETURNING section_link, position;`).
					WithArgs(targetLink).
					WillReturnRows(pgxmock.NewRows([]string{"section_link", "position"}).AddRow(targetSection, 5))

				m.ExpectExec(`(?s)INSERT INTO task_version.*`).
					WithArgs(targetLink, targetSection, &targetExecuter, "Updated Title", "Updated Desc", 5, &targetDeadLine).
					WillReturnError(errors.New("db insert err"))

				m.ExpectRollback()
			},
			expectedError: errors.New("tx.Exec: db insert err"),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockDB, err := pgxmock.NewPool()
			if !assert.NoError(t, err) {
				return
			}
			defer mockDB.Close()

			if test.mockBehavior != nil {
				test.mockBehavior(mockDB)
			}

			repo := NewRepository(mockDB)
			err = repo.UpdateCardDetails(ctx, updatingCard)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					if errors.Is(test.expectedError, common.ErrCardNotFound) ||
						errors.Is(test.expectedError, common.ErrInvalidReferenceCardData) ||
						errors.Is(test.expectedError, common.ErrInvalidCardData) ||
						errors.Is(test.expectedError, common.ErrMissingRequiredField) {
						assert.ErrorIs(t, err, test.expectedError)
					} else {
						assert.EqualError(t, err, test.expectedError.Error())
					}
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestRepositoryReorderCard(t *testing.T) {
	ctx := context.Background()
	targetCardLink := uuid.New()
	oldSectionLink := uuid.New()
	newSectionLink := uuid.New()
	targetExecuter := uuid.New()
	targetDeadLine := time.Now()

	updatingPlaceCardNewSection := dto.PlaceCard{
		LinkCard:    targetCardLink,
		LinkSection: newSectionLink,
		Position:    2,
	}

	updatingPlaceCardSameSection := dto.PlaceCard{
		LinkCard:    targetCardLink,
		LinkSection: oldSectionLink,
		Position:    4,
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			nameTest: "Success reorder different section",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)UPDATE task_version.*RETURNING.*`).
					WithArgs(targetCardLink).
					WillReturnRows(pgxmock.NewRows([]string{"section_link", "position", "title", "description", "executer_link", "due_date"}).
						AddRow(oldSectionLink, 1, "Title", "Desc", &targetExecuter, &targetDeadLine))

				m.ExpectQuery(`(?s)WITH locked_section AS.*`).
					WithArgs(newSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"count", "max_tasks"}).AddRow(0, nil))

				m.ExpectQuery(`(?s)WITH positions AS.*SELECT EXISTS.*`).
					WithArgs(oldSectionLink, newSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(false))

				m.ExpectQuery(`(?s)SELECT COALESCE.*`).
					WithArgs(newSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"position"}).AddRow(2))

				m.ExpectExec(`(?s)INSERT INTO task_version.*`).
					WithArgs(targetCardLink, newSectionLink, 2, "Title", "Desc", &targetExecuter, &targetDeadLine).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			nameTest: "Success reorder same section",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)UPDATE task_version.*RETURNING.*`).
					WithArgs(targetCardLink).
					WillReturnRows(pgxmock.NewRows([]string{"section_link", "position", "title", "description", "executer_link", "due_date"}).
						AddRow(oldSectionLink, 1, "Title", "Desc", &targetExecuter, &targetDeadLine))

				m.ExpectExec(`(?s)INSERT INTO task_version.*`).
					WithArgs(targetCardLink, oldSectionLink, 4, "Title", "Desc", &targetExecuter, &targetDeadLine).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			nameTest: "Error skip mandatory",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)UPDATE task_version.*RETURNING.*`).
					WithArgs(targetCardLink).
					WillReturnRows(pgxmock.NewRows([]string{"section_link", "position", "title", "description", "executer_link", "due_date"}).
						AddRow(oldSectionLink, 1, "Title", "Desc", &targetExecuter, &targetDeadLine))

				m.ExpectQuery(`(?s)WITH locked_section AS.*`).
					WithArgs(newSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"count", "max_tasks"}).AddRow(0, nil))

				m.ExpectQuery(`(?s)WITH positions AS.*SELECT EXISTS.*`).
					WithArgs(oldSectionLink, newSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

				m.ExpectRollback()
			},
			expectedError: common.ErrCannotSkipMandatorySection,
		},
		{
			nameTest: "Error not found",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)UPDATE task_version.*RETURNING.*`).
					WithArgs(targetCardLink).
					WillReturnError(pgx.ErrNoRows)

				m.ExpectRollback()
			},
			expectedError: common.ErrCardNotFound,
		},
		{
			nameTest: "Error check violation data",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)UPDATE task_version.*RETURNING.*`).
					WithArgs(targetCardLink).
					WillReturnRows(pgxmock.NewRows([]string{"section_link", "position", "title", "description", "executer_link", "due_date"}).
						AddRow(oldSectionLink, 1, "Title", "Desc", &targetExecuter, &targetDeadLine))

				m.ExpectExec(`(?s)INSERT INTO task_version.*`).
					WithArgs(targetCardLink, oldSectionLink, 4, "Title", "Desc", &targetExecuter, &targetDeadLine).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.CheckViolation})

				m.ExpectRollback()
			},
			expectedError: common.ErrInvalidCardData,
		},
		{
			nameTest: "Error missing required field",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)UPDATE task_version.*RETURNING.*`).
					WithArgs(targetCardLink).
					WillReturnRows(pgxmock.NewRows([]string{"section_link", "position", "title", "description", "executer_link", "due_date"}).
						AddRow(oldSectionLink, 1, "Title", "Desc", &targetExecuter, &targetDeadLine))

				m.ExpectExec(`(?s)INSERT INTO task_version.*`).
					WithArgs(targetCardLink, oldSectionLink, 4, "Title", "Desc", &targetExecuter, &targetDeadLine).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.NotNullViolation})

				m.ExpectRollback()
			},
			expectedError: common.ErrMissingRequiredField,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockDB, err := pgxmock.NewPool()
			if !assert.NoError(t, err) {
				return
			}
			defer mockDB.Close()

			if test.mockBehavior != nil {
				test.mockBehavior(mockDB)
			}

			repo := NewRepository(mockDB)

			var updateDto dto.PlaceCard
			if test.nameTest == "Success reorder same section" || test.nameTest == "Error check violation data" || test.nameTest == "Error missing required field" {
				updateDto = updatingPlaceCardSameSection
			} else {
				updateDto = updatingPlaceCardNewSection
			}

			err = repo.ReorderCard(ctx, updateDto)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					if errors.Is(test.expectedError, common.ErrCardNotFound) ||
						errors.Is(test.expectedError, common.ErrCannotSkipMandatorySection) ||
						errors.Is(test.expectedError, common.ErrInvalidCardData) ||
						errors.Is(test.expectedError, common.ErrMissingRequiredField) ||
						errors.Is(test.expectedError, common.ErrInvalidReferenceCardData) {
						assert.ErrorIs(t, err, test.expectedError)
					} else {
						assert.EqualError(t, err, test.expectedError.Error())
					}
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestRepositoryCreateCard(t *testing.T) {
	ctx := context.Background()
	targetCardLink := uuid.New()
	targetAuthorLink := uuid.New()
	targetSectionLink := uuid.New()
	targetExecuterLink := uuid.New()
	targetDeadLine := time.Now()

	newCard := dto.NewCard{
		LinkCard:     targetCardLink,
		LinkAuthor:   targetAuthorLink,
		LinkSection:  targetSectionLink,
		Title:        "New Task",
		Description:  "Desc",
		LinkExecuter: &targetExecuterLink,
		DataDeadLine: &targetDeadLine,
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedError error
		expectedData  int
	}{
		{
			nameTest: "Success create card",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)WITH locked_section AS.*`).
					WithArgs(targetSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"count", "max_tasks"}).AddRow(0, nil))

				m.ExpectExec(`(?s)INSERT INTO task.*`).
					WithArgs(targetCardLink, targetAuthorLink).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectExec(`(?s)SELECT 1.*FROM task_version.*FOR NO KEY UPDATE.*`).
					WithArgs(targetSectionLink).
					WillReturnResult(pgxmock.NewResult("SELECT", 0))

				m.ExpectQuery(`(?s)SELECT COALESCE.*`).
					WithArgs(targetSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"position"}).AddRow(1))

				m.ExpectExec(`(?s)INSERT INTO task_version.*`).
					WithArgs(targetCardLink, targetSectionLink, &targetExecuterLink, "New Task", "Desc", 1, &targetDeadLine).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectCommit()
			},
			expectedError: nil,
			expectedData:  1,
		},
		{
			nameTest: "Error card already exist",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)WITH locked_section AS.*`).
					WithArgs(targetSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"count", "max_tasks"}).AddRow(0, nil))

				m.ExpectExec(`(?s)INSERT INTO task.*`).
					WithArgs(targetCardLink, targetAuthorLink).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

				m.ExpectRollback()
			},
			expectedError: common.ErrCardAlreadyExists,
			expectedData:  -1,
		},
		{
			nameTest: "Error task missing required field",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)WITH locked_section AS.*`).
					WithArgs(targetSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"count", "max_tasks"}).AddRow(0, nil))

				m.ExpectExec(`(?s)INSERT INTO task.*`).
					WithArgs(targetCardLink, targetAuthorLink).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.NotNullViolation})

				m.ExpectRollback()
			},
			expectedError: common.ErrMissingRequiredField,
			expectedData:  -1,
		},
		{
			nameTest: "Error not existing section",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)WITH locked_section AS.*`).
					WithArgs(targetSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"count", "max_tasks"}).AddRow(0, nil))

				m.ExpectExec(`(?s)INSERT INTO task.*`).
					WithArgs(targetCardLink, targetAuthorLink).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectExec(`(?s)SELECT 1.*FROM task_version.*FOR NO KEY UPDATE.*`).
					WithArgs(targetSectionLink).
					WillReturnResult(pgxmock.NewResult("SELECT", 0))

				m.ExpectQuery(`(?s)SELECT COALESCE.*`).
					WithArgs(targetSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"position"}).AddRow(1))

				m.ExpectExec(`(?s)INSERT INTO task_version.*`).
					WithArgs(targetCardLink, targetSectionLink, &targetExecuterLink, "New Task", "Desc", 1, &targetDeadLine).
					WillReturnError(&pgconn.PgError{
						Code:           pgerrcode.ForeignKeyViolation,
						ConstraintName: "fk_version_section",
					})

				m.ExpectRollback()
			},
			expectedError: common.ErrSectionNotFound,
			expectedData:  -1,
		},
		{
			nameTest: "Error version check violation",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)WITH locked_section AS.*`).
					WithArgs(targetSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"count", "max_tasks"}).AddRow(0, nil))

				m.ExpectExec(`(?s)INSERT INTO task.*`).
					WithArgs(targetCardLink, targetAuthorLink).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectExec(`(?s)SELECT 1.*FROM task_version.*FOR NO KEY UPDATE.*`).
					WithArgs(targetSectionLink).
					WillReturnResult(pgxmock.NewResult("SELECT", 0))

				m.ExpectQuery(`(?s)SELECT COALESCE.*`).
					WithArgs(targetSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"position"}).AddRow(1))

				m.ExpectExec(`(?s)INSERT INTO task_version.*`).
					WithArgs(targetCardLink, targetSectionLink, &targetExecuterLink, "New Task", "Desc", 1, &targetDeadLine).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.CheckViolation})

				m.ExpectRollback()
			},
			expectedError: common.ErrInvalidCardData,
			expectedData:  -1,
		},
		{
			nameTest: "Error version missing required field",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)WITH locked_section AS.*`).
					WithArgs(targetSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"count", "max_tasks"}).AddRow(0, nil))

				m.ExpectExec(`(?s)INSERT INTO task.*`).
					WithArgs(targetCardLink, targetAuthorLink).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectExec(`(?s)SELECT 1.*FROM task_version.*FOR NO KEY UPDATE.*`).
					WithArgs(targetSectionLink).
					WillReturnResult(pgxmock.NewResult("SELECT", 0))

				m.ExpectQuery(`(?s)SELECT COALESCE.*`).
					WithArgs(targetSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"position"}).AddRow(1))

				m.ExpectExec(`(?s)INSERT INTO task_version.*`).
					WithArgs(targetCardLink, targetSectionLink, &targetExecuterLink, "New Task", "Desc", 1, &targetDeadLine).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.NotNullViolation})

				m.ExpectRollback()
			},
			expectedError: common.ErrMissingRequiredField,
			expectedData:  -1,
		},
		{
			nameTest: "Error exec fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)WITH locked_section AS.*`).
					WithArgs(targetSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"count", "max_tasks"}).AddRow(0, nil))

				m.ExpectExec(`(?s)INSERT INTO task.*`).
					WithArgs(targetCardLink, targetAuthorLink).
					WillReturnError(errors.New("db disconnect"))

				m.ExpectRollback()
			},
			expectedError: errors.New("tx.Exec insert task: db disconnect"),
			expectedData:  -1,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockDB, err := pgxmock.NewPool()
			if !assert.NoError(t, err) {
				return
			}
			defer mockDB.Close()

			if test.mockBehavior != nil {
				test.mockBehavior(mockDB)
			}

			repo := NewRepository(mockDB)
			result, err := repo.CreateCard(ctx, newCard)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					if errors.Is(test.expectedError, common.ErrSectionNotFound) ||
						errors.Is(test.expectedError, common.ErrCardAlreadyExists) ||
						errors.Is(test.expectedError, common.ErrMissingRequiredField) ||
						errors.Is(test.expectedError, common.ErrInvalidCardData) ||
						errors.Is(test.expectedError, common.ErrInvalidReferenceCardData) {
						assert.ErrorIs(t, err, test.expectedError)
					} else {
						assert.EqualError(t, err, test.expectedError.Error())
					}
				}
			} else {
				if assert.NoError(t, err) {
					assert.Equal(t, test.expectedData, result)
				}
			}

			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestRepositoryGetComments(t *testing.T) {
	ctx := context.Background()
	targetCardLink := uuid.New()
	commentLink1 := uuid.New()
	commentLink2 := uuid.New()
	authorLink := uuid.New()
	parentLink := uuid.New()

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedError error
		expectedLen   int
	}{
		{
			nameTest: "Success get comments with parent",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"comment_link", "author_link", "parent_link", "text"}).
					AddRow(commentLink1, authorLink, &parentLink, "first comment").
					AddRow(commentLink2, authorLink, nil, "second comment")

				m.ExpectQuery(`(?s)SELECT.*c.link.*FROM comment_task.*WHERE t.task_link = \$1`).
					WithArgs(targetCardLink).
					WillReturnRows(rows)
			},
			expectedError: nil,
			expectedLen:   2,
		},
		{
			nameTest: "Success empty comments",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"comment_link", "author_link", "parent_link", "text"})

				m.ExpectQuery(`(?s)SELECT.*c.link.*FROM comment_task.*WHERE t.task_link = \$1`).
					WithArgs(targetCardLink).
					WillReturnRows(rows)
			},
			expectedError: nil,
			expectedLen:   0,
		},
		{
			nameTest: "Error query fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?s)SELECT.*c.link.*FROM comment_task.*WHERE t.task_link = \$1`).
					WithArgs(targetCardLink).
					WillReturnError(errors.New("db disconnect"))
			},
			expectedError: errors.New("pool.Query: db disconnect"),
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
			result, err := repo.GetComments(ctx, targetCardLink)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					assert.EqualError(t, err, test.expectedError.Error())
				}
			} else {
				if assert.NoError(t, err) {
					assert.Len(t, result, test.expectedLen)
				}
			}

			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestRepositoryCreateComment(t *testing.T) {
	ctx := context.Background()
	targetCardLink := uuid.New()
	targetAuthorLink := uuid.New()
	targetParentLink := uuid.New()

	createInfo := dto.CreateCommentInfo{
		CardLink:   targetCardLink,
		ParentLink: &targetParentLink,
		AuthorLink: targetAuthorLink,
		Text:       "hello world",
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			nameTest: "Success create comment",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)INSERT INTO comment_task.*`).
					WithArgs(pgxmock.AnyArg(), targetCardLink, &targetParentLink, targetAuthorLink, "hello world").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: nil,
		},
		{
			nameTest: "Error missing required field",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)INSERT INTO comment_task.*`).
					WithArgs(pgxmock.AnyArg(), targetCardLink, &targetParentLink, targetAuthorLink, "hello world").
					WillReturnError(&pgconn.PgError{Code: pgerrcode.NotNullViolation})
			},
			expectedError: common.ErrMissingRequiredField,
		},
		{
			nameTest: "Error invalid reference data",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)INSERT INTO comment_task.*`).
					WithArgs(pgxmock.AnyArg(), targetCardLink, &targetParentLink, targetAuthorLink, "hello world").
					WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})
			},
			expectedError: common.ErrInvalidReferenceCardData,
		},
		{
			nameTest: "Error exec fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)INSERT INTO comment_task.*`).
					WithArgs(pgxmock.AnyArg(), targetCardLink, &targetParentLink, targetAuthorLink, "hello world").
					WillReturnError(errors.New("db disconnect"))
			},
			expectedError: errors.New("pool.Exec: db disconnect"),
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
			result, err := repo.CreateComment(ctx, createInfo)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					if errors.Is(test.expectedError, common.ErrMissingRequiredField) ||
						errors.Is(test.expectedError, common.ErrInvalidReferenceCardData) {
						assert.ErrorIs(t, err, test.expectedError)
					} else {
						assert.EqualError(t, err, test.expectedError.Error())
					}
				}
			} else {
				if assert.NoError(t, err) {
					assert.NotEqual(t, uuid.Nil, result.Link)
					assert.Equal(t, targetAuthorLink, result.AuthorLink)
					assert.Equal(t, "hello world", result.Text)
				}
			}

			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestRepositoryIsCommentAuthor(t *testing.T) {
	ctx := context.Background()
	targetCommentLink := uuid.New()
	targetUserLink := uuid.New()

	tests := []struct {
		nameTest     string
		mockBehavior func(m pgxmock.PgxPoolIface)
		expected     bool
	}{
		{
			nameTest: "Returns true when author",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?s)SELECT EXISTS.*FROM comment_task.*WHERE link = \$1 AND author_link = \$2`).
					WithArgs(targetCommentLink, targetUserLink).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
			},
			expected: true,
		},
		{
			nameTest: "Returns false when not author",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?s)SELECT EXISTS.*FROM comment_task.*WHERE link = \$1 AND author_link = \$2`).
					WithArgs(targetCommentLink, targetUserLink).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(false))
			},
			expected: false,
		},
		{
			nameTest: "Returns false on query error",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?s)SELECT EXISTS.*FROM comment_task.*WHERE link = \$1 AND author_link = \$2`).
					WithArgs(targetCommentLink, targetUserLink).
					WillReturnError(errors.New("db disconnect"))
			},
			expected: false,
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
			result := repo.IsCommentAuthor(ctx, targetCommentLink, targetUserLink)

			assert.Equal(t, test.expected, result)
			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestRepositoryDeleteComment(t *testing.T) {
	ctx := context.Background()
	targetCommentLink := uuid.New()

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			nameTest: "Success delete comment",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)DELETE FROM comment_task WHERE link = \$1`).
					WithArgs(targetCommentLink).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			expectedError: nil,
		},
		{
			nameTest: "Error comment not found",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)DELETE FROM comment_task WHERE link = \$1`).
					WithArgs(targetCommentLink).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			expectedError: common.ErrCommentNotFound,
		},
		{
			nameTest: "Error exec fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)DELETE FROM comment_task WHERE link = \$1`).
					WithArgs(targetCommentLink).
					WillReturnError(errors.New("db disconnect"))
			},
			expectedError: errors.New("pool.Exec: db disconnect"),
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
			err = repo.DeleteComment(ctx, targetCommentLink)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					if errors.Is(test.expectedError, common.ErrCommentNotFound) {
						assert.ErrorIs(t, err, common.ErrCommentNotFound)
					} else {
						assert.EqualError(t, err, test.expectedError.Error())
					}
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestRepositoryUpdateComment(t *testing.T) {
	ctx := context.Background()
	targetCommentLink := uuid.New()

	updateInfo := dto.UpdateCommentInfo{
		CommentLink: targetCommentLink,
		Text:        "new text",
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			nameTest: "Success update comment",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)UPDATE comment_task.*SET text = \$1.*WHERE link = \$2`).
					WithArgs("new text", targetCommentLink).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: nil,
		},
		{
			nameTest: "Error comment not found",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)UPDATE comment_task.*SET text = \$1.*WHERE link = \$2`).
					WithArgs("new text", targetCommentLink).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError: common.ErrCommentNotFound,
		},
		{
			nameTest: "Error exec fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)UPDATE comment_task.*SET text = \$1.*WHERE link = \$2`).
					WithArgs("new text", targetCommentLink).
					WillReturnError(errors.New("db disconnect"))
			},
			expectedError: errors.New("pool.Exec: db disconnect"),
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
			err = repo.UpdateComment(ctx, updateInfo)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					if errors.Is(test.expectedError, common.ErrCommentNotFound) {
						assert.ErrorIs(t, err, common.ErrCommentNotFound)
					} else {
						assert.EqualError(t, err, test.expectedError.Error())
					}
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}
