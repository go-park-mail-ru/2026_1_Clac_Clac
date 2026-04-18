package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	dto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/card/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
)

func TestGetCard(t *testing.T) {
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
			expectedError: common.ErrorNotExistingCard,
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
					if errors.Is(test.expectedError, common.ErrorNotExistingCard) {
						assert.ErrorIs(t, err, common.ErrorNotExistingCard)
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

func TestDeleteCard(t *testing.T) {
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
			expectedError: common.ErrorNotExistingCard,
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
					if errors.Is(test.expectedError, common.ErrorNotExistingCard) {
						assert.ErrorIs(t, err, common.ErrorNotExistingCard)
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

func TestUpdateCardDetails(t *testing.T) {
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
			expectedError: common.ErrorNotExistingCard,
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
			expectedError: common.ErrorInvalidReferenceCardData,
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
			expectedError: common.ErrorInvalidCardData,
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
			expectedError: common.ErrorMissingRequiredField,
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
					if errors.Is(test.expectedError, common.ErrorNotExistingCard) ||
						errors.Is(test.expectedError, common.ErrorInvalidReferenceCardData) ||
						errors.Is(test.expectedError, common.ErrorInvalidCardData) ||
						errors.Is(test.expectedError, common.ErrorMissingRequiredField) {
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

func TestReorderCard(t *testing.T) {
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

				m.ExpectQuery(`(?s)WITH count_tasks AS.*SELECT c\.count, s\.max_tasks.*`).
					WithArgs(newSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"count", "max_tasks"}).AddRow(2, 10))

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

				m.ExpectQuery(`(?s)WITH count_tasks AS.*SELECT c\.count, s\.max_tasks.*`).
					WithArgs(newSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"count", "max_tasks"}).AddRow(2, 10))

				m.ExpectQuery(`(?s)WITH positions AS.*SELECT EXISTS.*`).
					WithArgs(oldSectionLink, newSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

				m.ExpectRollback()
			},
			expectedError: common.ErrorSkipMandatorySection,
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
			expectedError: common.ErrorNotExistingCard,
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
			expectedError: common.ErrorInvalidCardData,
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
			expectedError: common.ErrorMissingRequiredField,
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
					if errors.Is(test.expectedError, common.ErrorNotExistingCard) ||
						errors.Is(test.expectedError, common.ErrorSkipMandatorySection) ||
						errors.Is(test.expectedError, common.ErrorInvalidCardData) ||
						errors.Is(test.expectedError, common.ErrorMissingRequiredField) ||
						errors.Is(test.expectedError, common.ErrorInvalidReferenceCardData) {
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

func TestCreateCard(t *testing.T) {
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

				m.ExpectQuery(`(?s)WITH count_tasks AS.*SELECT c\.count, s\.max_tasks.*`).
					WithArgs(targetSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"count", "max_tasks"}).AddRow(1, 10))

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

				m.ExpectQuery(`(?s)WITH count_tasks AS.*SELECT c\.count, s\.max_tasks.*`).
					WithArgs(targetSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"count", "max_tasks"}).AddRow(1, 10))

				m.ExpectExec(`(?s)INSERT INTO task.*`).
					WithArgs(targetCardLink, targetAuthorLink).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

				m.ExpectRollback()
			},
			expectedError: common.ErrorCardAlreadyExist,
			expectedData:  -1,
		},
		{
			nameTest: "Error task missing required field",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)WITH count_tasks AS.*SELECT c\.count, s\.max_tasks.*`).
					WithArgs(targetSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"count", "max_tasks"}).AddRow(1, 10))

				m.ExpectExec(`(?s)INSERT INTO task.*`).
					WithArgs(targetCardLink, targetAuthorLink).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.NotNullViolation})

				m.ExpectRollback()
			},
			expectedError: common.ErrorMissingRequiredField,
			expectedData:  -1,
		},
		{
			nameTest: "Error not existing section",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)WITH count_tasks AS.*SELECT c\.count, s\.max_tasks.*`).
					WithArgs(targetSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"count", "max_tasks"}).AddRow(1, 10))

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
			expectedError: common.ErrorNotExistingSection,
			expectedData:  -1,
		},
		{
			nameTest: "Error version check violation",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)WITH count_tasks AS.*SELECT c\.count, s\.max_tasks.*`).
					WithArgs(targetSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"count", "max_tasks"}).AddRow(1, 10))

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
			expectedError: common.ErrorInvalidCardData,
			expectedData:  -1,
		},
		{
			nameTest: "Error version missing required field",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)WITH count_tasks AS.*SELECT c\.count, s\.max_tasks.*`).
					WithArgs(targetSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"count", "max_tasks"}).AddRow(1, 10))

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
			expectedError: common.ErrorMissingRequiredField,
			expectedData:  -1,
		},
		{
			nameTest: "Error exec fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)WITH count_tasks AS.*SELECT c\.count, s\.max_tasks.*`).
					WithArgs(targetSectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"count", "max_tasks"}).AddRow(1, 10))

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
					if errors.Is(test.expectedError, common.ErrorNotExistingSection) ||
						errors.Is(test.expectedError, common.ErrorCardAlreadyExist) ||
						errors.Is(test.expectedError, common.ErrorMissingRequiredField) ||
						errors.Is(test.expectedError, common.ErrorInvalidCardData) ||
						errors.Is(test.expectedError, common.ErrorInvalidReferenceCardData) {
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
