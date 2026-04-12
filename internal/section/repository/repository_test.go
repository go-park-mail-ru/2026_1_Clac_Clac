package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/section/repository/dto"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryGetSectionInfo(t *testing.T) {
	ctx := context.Background()
	targetLink := common.FixedSectionUuiD
	maxTasks := 50

	expectedInfo := dto.FullSectionInfo{
		SectionName: "To Do",
		Position:    1,
		IsMandatory: true,
		Color:       "white",
		MaxTasks:    &maxTasks,
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedError error
		expectedData  dto.FullSectionInfo
	}{
		{
			nameTest: "Success get section info",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"section_name", "position", "is_mandatory", "color", "max_tasks"}).
					AddRow(expectedInfo.SectionName, expectedInfo.Position, expectedInfo.IsMandatory, expectedInfo.Color, expectedInfo.MaxTasks)

				m.ExpectQuery(`(?s)SELECT.*section_name.*FROM section_actual.*WHERE section_link = \$1`).
					WithArgs(targetLink).
					WillReturnRows(rows)
			},
			expectedError: nil,
			expectedData:  expectedInfo,
		},
		{
			nameTest: "Error not found pgx.ErrNoRows",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?s)SELECT.*section_name.*FROM section_actual.*WHERE section_link = \$1`).
					WithArgs(targetLink).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedError: common.ErrorNotExistingSection,
			expectedData:  dto.FullSectionInfo{},
		},
		{
			nameTest: "Error query fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?s)SELECT.*section_name.*FROM section_actual.*WHERE section_link = \$1`).
					WithArgs(targetLink).
					WillReturnError(errors.New("db disconnect"))
			},
			expectedError: errors.New("QueryRow: db disconnect"),
			expectedData:  dto.FullSectionInfo{},
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

			repo := NewRepository(Deps{Pool: mockDB})
			result, err := repo.GetSectionInfo(ctx, targetLink)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					if errors.Is(test.expectedError, common.ErrorNotExistingSection) {
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

func TestRepositoryCreateSection(t *testing.T) {
	ctx := context.Background()
	boardLink := common.FixedSectionUuiD
	sectionLink := common.FixedSectionUuiD
	maxTasks := 10

	creatingSection := dto.CreatingSection{
		SectionLink: sectionLink,
		BoardLink:   boardLink,
		SectionName: "In Progress",
		IsMandatory: false,
		Color:       "blue",
		MaxTasks:    &maxTasks,
	}

	expectedResult := dto.FullSectionInfo{
		SectionLink: sectionLink,
		SectionName: "In Progress",
		Position:    2,
		IsMandatory: false,
		Color:       "blue",
		MaxTasks:    &maxTasks,
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedError error
		expectedData  dto.FullSectionInfo
	}{
		{
			nameTest: "Success create section",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectExec(`(?s)INSERT INTO section \(section_link, board_link\).*`).
					WithArgs(creatingSection.SectionLink, creatingSection.BoardLink).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectQuery(`(?s)SELECT COALESCE\(MAX.*`).
					WithArgs(creatingSection.BoardLink).
					WillReturnRows(pgxmock.NewRows([]string{"position"}).AddRow(2))

				m.ExpectExec(`(?s)INSERT INTO section_version.*`).
					WithArgs(
						creatingSection.SectionLink,
						creatingSection.SectionName,
						2,
						creatingSection.IsMandatory,
						creatingSection.Color,
						creatingSection.MaxTasks,
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectCommit()
			},
			expectedError: nil,
			expectedData:  expectedResult,
		},
		{
			nameTest: "Error section already exist",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec(`(?s)INSERT INTO section \(section_link, board_link\).*`).
					WithArgs(creatingSection.SectionLink, creatingSection.BoardLink).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})
				m.ExpectRollback()
			},
			expectedError: common.ErrorSectionAlreadyExist,
			expectedData:  dto.FullSectionInfo{},
		},
		{
			nameTest: "Error missing required field in section",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec(`(?s)INSERT INTO section \(section_link, board_link\).*`).
					WithArgs(creatingSection.SectionLink, creatingSection.BoardLink).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.NotNullViolation})
				m.ExpectRollback()
			},
			expectedError: common.ErrorMissingRequiredField,
			expectedData:  dto.FullSectionInfo{},
		},
		{
			nameTest: "Error check violation in version",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec(`(?s)INSERT INTO section \(section_link, board_link\).*`).
					WithArgs(creatingSection.SectionLink, creatingSection.BoardLink).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectQuery(`(?s)SELECT COALESCE\(MAX.*`).
					WithArgs(creatingSection.BoardLink).
					WillReturnRows(pgxmock.NewRows([]string{"position"}).AddRow(2))

				m.ExpectExec(`(?s)INSERT INTO section_version.*`).
					WithArgs(
						creatingSection.SectionLink,
						creatingSection.SectionName,
						2,
						creatingSection.IsMandatory,
						creatingSection.Color,
						creatingSection.MaxTasks,
					).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.CheckViolation})

				m.ExpectRollback()
			},
			expectedError: common.ErrorInvalidSectionData,
			expectedData:  dto.FullSectionInfo{},
		},
		{
			nameTest: "Error execution fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec(`(?s)INSERT INTO section \(section_link, board_link\).*`).
					WithArgs(creatingSection.SectionLink, creatingSection.BoardLink).
					WillReturnError(errors.New("insert conflict"))
				m.ExpectRollback()
			},
			expectedError: errors.New("tx.Exec section: insert conflict"),
			expectedData:  dto.FullSectionInfo{},
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

			repo := NewRepository(Deps{Pool: mockDB})
			result, err := repo.CreateSection(ctx, creatingSection)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					if errors.Is(test.expectedError, common.ErrorSectionAlreadyExist) ||
						errors.Is(test.expectedError, common.ErrorMissingRequiredField) ||
						errors.Is(test.expectedError, common.ErrorInvalidSectionData) ||
						errors.Is(test.expectedError, common.ErrorInvalidReferenceSectionData) {
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

func TestRepositoryDeleteSection(t *testing.T) {
	ctx := context.Background()
	targetLink := common.FixedSectionUuiD
	boardLink := uuid.New()
	backlogLink := uuid.New()

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			nameTest: "Success delete section",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)SELECT s.board_link, v.position FROM section s.*`).
					WithArgs(targetLink).
					WillReturnRows(pgxmock.NewRows([]string{"board_link", "position"}).AddRow(boardLink, 2))

				m.ExpectQuery(`(?s)SELECT s.section_link FROM section s.*`).
					WithArgs(boardLink).
					WillReturnRows(pgxmock.NewRows([]string{"section_link"}).AddRow(backlogLink))

				m.ExpectExec(`(?s)UPDATE section SET deleted_at = NOW\(\).*`).
					WithArgs(targetLink).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectExec(`(?s)WITH closed_tasks AS \(.*`).
					WithArgs(targetLink, backlogLink).
					WillReturnResult(pgxmock.NewResult("INSERT", 3))

				m.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			nameTest: "Error section not found",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`(?s)SELECT s.board_link, v.position FROM section s.*`).
					WithArgs(targetLink).
					WillReturnError(pgx.ErrNoRows)
				m.ExpectRollback()
			},
			expectedError: common.ErrorNotExistingSection,
		},
		{
			nameTest: "Error try delete backlog",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`(?s)SELECT s.board_link, v.position FROM section s.*`).
					WithArgs(targetLink).
					WillReturnRows(pgxmock.NewRows([]string{"board_link", "position"}).AddRow(boardLink, 1))
				m.ExpectRollback()
			},
			expectedError: common.ErrorDeleteBacklog,
		},
		{
			nameTest: "Error move tasks invalid reference data",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)SELECT s.board_link, v.position FROM section s.*`).
					WithArgs(targetLink).
					WillReturnRows(pgxmock.NewRows([]string{"board_link", "position"}).AddRow(boardLink, 2))

				m.ExpectQuery(`(?s)SELECT s.section_link FROM section s.*`).
					WithArgs(boardLink).
					WillReturnRows(pgxmock.NewRows([]string{"section_link"}).AddRow(backlogLink))

				m.ExpectExec(`(?s)UPDATE section SET deleted_at = NOW\(\).*`).
					WithArgs(targetLink).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectExec(`(?s)WITH closed_tasks AS \(.*`).
					WithArgs(targetLink, backlogLink).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})

				m.ExpectRollback()
			},
			expectedError: common.ErrorInvalidReferenceSectionData,
		},
		{
			nameTest: "Error move tasks invalid card data",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)SELECT s.board_link, v.position FROM section s.*`).
					WithArgs(targetLink).
					WillReturnRows(pgxmock.NewRows([]string{"board_link", "position"}).AddRow(boardLink, 2))

				m.ExpectQuery(`(?s)SELECT s.section_link FROM section s.*`).
					WithArgs(boardLink).
					WillReturnRows(pgxmock.NewRows([]string{"section_link"}).AddRow(backlogLink))

				m.ExpectExec(`(?s)UPDATE section SET deleted_at = NOW\(\).*`).
					WithArgs(targetLink).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectExec(`(?s)WITH closed_tasks AS \(.*`).
					WithArgs(targetLink, backlogLink).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.CheckViolation})

				m.ExpectRollback()
			},
			expectedError: common.ErrorInvalidCardData,
		},
		{
			nameTest: "Error generic DB error on delete",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()

				m.ExpectQuery(`(?s)SELECT s.board_link, v.position FROM section s.*`).
					WithArgs(targetLink).
					WillReturnRows(pgxmock.NewRows([]string{"board_link", "position"}).AddRow(boardLink, 2))

				m.ExpectQuery(`(?s)SELECT s.section_link FROM section s.*`).
					WithArgs(boardLink).
					WillReturnRows(pgxmock.NewRows([]string{"section_link"}).AddRow(backlogLink))

				m.ExpectExec(`(?s)UPDATE section SET deleted_at = NOW\(\).*`).
					WithArgs(targetLink).
					WillReturnError(errors.New("db error on delete"))

				m.ExpectRollback()
			},
			expectedError: fmt.Errorf("tx.Exec delete section: %w", errors.New("db error on delete")),
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

			repo := NewRepository(Deps{Pool: mockDB})
			err = repo.DeleteSection(ctx, targetLink)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					if errors.Is(test.expectedError, common.ErrorNotExistingSection) ||
						errors.Is(test.expectedError, common.ErrorDeleteBacklog) ||
						errors.Is(test.expectedError, common.ErrorInvalidReferenceSectionData) ||
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

func TestRepositoryReorderSection(t *testing.T) {
	ctx := context.Background()
	boardLink := common.FixedSectionUuiD
	section1 := uuid.New()
	section2 := uuid.New()
	linksSection := []uuid.UUID{section1, section2}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			nameTest: "Success reorder sections",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec(`(?s)WITH new_positions AS \(.*`).
					WithArgs(linksSection, boardLink).
					WillReturnResult(pgxmock.NewResult("INSERT", 2))
				m.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			nameTest: "Error not all links updated",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec(`(?s)WITH new_positions AS \(.*`).
					WithArgs(linksSection, boardLink).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectRollback()
			},
			expectedError: common.ErrorNotFindAllLinks,
		},
		{
			nameTest: "Error check violation data",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec(`(?s)WITH new_positions AS \(.*`).
					WithArgs(linksSection, boardLink).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.CheckViolation})
				m.ExpectRollback()
			},
			expectedError: common.ErrorInvalidSectionData,
		},
		{
			nameTest: "Error missing required field",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec(`(?s)WITH new_positions AS \(.*`).
					WithArgs(linksSection, boardLink).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.NotNullViolation})
				m.ExpectRollback()
			},
			expectedError: common.ErrorMissingRequiredField,
		},
		{
			nameTest: "Error foreign key violation",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec(`(?s)WITH new_positions AS \(.*`).
					WithArgs(linksSection, boardLink).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})
				m.ExpectRollback()
			},
			expectedError: common.ErrorInvalidReferenceSectionData,
		},
		{
			nameTest: "Error query fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec(`(?s)WITH new_positions AS \(.*`).
					WithArgs(linksSection, boardLink).
					WillReturnError(errors.New("db error"))
				m.ExpectRollback()
			},
			expectedError: errors.New("tx.Exec reorder: db error"),
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

			repo := NewRepository(Deps{Pool: mockDB})
			err = repo.ReorderSection(ctx, boardLink, linksSection)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					if errors.Is(test.expectedError, common.ErrorNotFindAllLinks) ||
						errors.Is(test.expectedError, common.ErrorInvalidSectionData) ||
						errors.Is(test.expectedError, common.ErrorMissingRequiredField) ||
						errors.Is(test.expectedError, common.ErrorInvalidReferenceSectionData) {
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

func TestRepositoryUpdateSection(t *testing.T) {
	ctx := context.Background()
	targetLink := common.FixedSectionUuiD
	maxTasks := 50

	updateData := dto.FullSectionInfo{
		SectionLink: targetLink,
		SectionName: "Updated Name",
		Position:    3,
		IsMandatory: true,
		Color:       "red",
		MaxTasks:    &maxTasks,
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			nameTest: "Success update section",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`(?s)UPDATE section_version.*SET valid_to = NOW\(\).*`).
					WithArgs(updateData.SectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"position"}).AddRow(3))

				m.ExpectExec(`(?s)INSERT INTO section_version \(section_link, section_name.*`).
					WithArgs(
						updateData.SectionLink,
						updateData.SectionName,
						3,
						updateData.IsMandatory,
						updateData.Color,
						updateData.MaxTasks,
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			nameTest: "Error zero rows affected",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`(?s)UPDATE section_version.*SET valid_to = NOW\(\).*`).
					WithArgs(updateData.SectionLink).
					WillReturnError(pgx.ErrNoRows)
				m.ExpectRollback()
			},
			expectedError: common.ErrorNotExistingSection,
		},
		{
			nameTest: "Error missing required field",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`(?s)UPDATE section_version.*SET valid_to = NOW\(\).*`).
					WithArgs(updateData.SectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"position"}).AddRow(3))

				m.ExpectExec(`(?s)INSERT INTO section_version \(section_link, section_name.*`).
					WithArgs(
						updateData.SectionLink,
						updateData.SectionName,
						3,
						updateData.IsMandatory,
						updateData.Color,
						updateData.MaxTasks,
					).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.NotNullViolation})

				m.ExpectRollback()
			},
			expectedError: common.ErrorMissingRequiredField,
		},
		{
			nameTest: "Error foreign key violation",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`(?s)UPDATE section_version.*SET valid_to = NOW\(\).*`).
					WithArgs(updateData.SectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"position"}).AddRow(3))

				m.ExpectExec(`(?s)INSERT INTO section_version \(section_link, section_name.*`).
					WithArgs(
						updateData.SectionLink,
						updateData.SectionName,
						3,
						updateData.IsMandatory,
						updateData.Color,
						updateData.MaxTasks,
					).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})

				m.ExpectRollback()
			},
			expectedError: common.ErrorInvalidReferenceSectionData,
		},
		{
			nameTest: "Error invalid section data",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`(?s)UPDATE section_version.*SET valid_to = NOW\(\).*`).
					WithArgs(updateData.SectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"position"}).AddRow(3))

				m.ExpectExec(`(?s)INSERT INTO section_version \(section_link, section_name.*`).
					WithArgs(
						updateData.SectionLink,
						updateData.SectionName,
						3,
						updateData.IsMandatory,
						updateData.Color,
						updateData.MaxTasks,
					).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.CheckViolation})

				m.ExpectRollback()
			},
			expectedError: common.ErrorInvalidSectionData,
		},
		{
			nameTest: "Error generic DB error on exec",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`(?s)UPDATE section_version.*SET valid_to = NOW\(\).*`).
					WithArgs(updateData.SectionLink).
					WillReturnRows(pgxmock.NewRows([]string{"position"}).AddRow(3))

				m.ExpectExec(`(?s)INSERT INTO section_version \(section_link, section_name.*`).
					WithArgs(
						updateData.SectionLink,
						updateData.SectionName,
						3,
						updateData.IsMandatory,
						updateData.Color,
						updateData.MaxTasks,
					).
					WillReturnError(errors.New("db error"))

				m.ExpectRollback()
			},
			expectedError: fmt.Errorf("tx.Exec insert update: %w", errors.New("db error")),
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

			repo := NewRepository(Deps{Pool: mockDB})
			err = repo.UpdateSection(ctx, updateData)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					if errors.Is(test.expectedError, common.ErrorNotExistingSection) ||
						errors.Is(test.expectedError, common.ErrorInvalidSectionData) ||
						errors.Is(test.expectedError, common.ErrorMissingRequiredField) ||
						errors.Is(test.expectedError, common.ErrorInvalidReferenceSectionData) {
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

func TestRepositoryGetAllSections(t *testing.T) {
	ctx := context.Background()
	boardLink := common.FixedSectionUuiD
	sectionLink := common.FixedSectionUuiD
	maxTasks := 50

	expectedSections := []dto.FullSectionInfo{
		{
			SectionLink: sectionLink,
			SectionName: "To Do",
			Position:    1,
			IsMandatory: true,
			Color:       "white",
			MaxTasks:    &maxTasks,
		},
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedError error
		expectedData  []dto.FullSectionInfo
	}{
		{
			nameTest: "Success get all sections",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"section_link", "section_name", "position", "is_mandatory", "color", "max_tasks"}).
					AddRow(
						expectedSections[0].SectionLink,
						expectedSections[0].SectionName,
						expectedSections[0].Position,
						expectedSections[0].IsMandatory,
						expectedSections[0].Color,
						expectedSections[0].MaxTasks,
					)

				m.ExpectQuery(`(?s)SELECT.*FROM section_actual.*WHERE board_link = \$1.*ORDER BY position ASC`).
					WithArgs(boardLink).
					WillReturnRows(rows)
			},
			expectedError: nil,
			expectedData:  expectedSections,
		},
		{
			nameTest: "Error DB query fails",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?s)SELECT.*FROM section_actual.*WHERE board_link = \$1.*ORDER BY position ASC`).
					WithArgs(boardLink).
					WillReturnError(errors.New("db error"))
			},
			expectedError: fmt.Errorf("pool.Query: %w", errors.New("db error")),
			expectedData:  nil,
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

			repo := NewRepository(Deps{Pool: mockDB})
			result, err := repo.GetAllSections(ctx, boardLink)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					assert.EqualError(t, err, test.expectedError.Error())
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

func TestRepositoryGetCards(t *testing.T) {
	ctx := context.Background()
	targetSectionLink := uuid.New()
	targetCardLink1 := uuid.New()
	targetCardLink2 := uuid.New()
	targetExecuter := "John Doe"
	targetDeadLine := time.Now()

	expectedCards := []dto.Card{
		{
			CardLink:     targetCardLink1,
			ExecuterName: &targetExecuter,
			Title:        "Task 1",
			DeadLine:     &targetDeadLine,
		},
		{
			CardLink:     targetCardLink2,
			ExecuterName: nil,
			Title:        "Task 2",
			DeadLine:     nil,
		},
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedError error
		expectedData  []dto.Card
	}{
		{
			nameTest: "Success get cards",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"task_link", "name_executer", "title", "due_date"}).
					AddRow(expectedCards[0].CardLink, expectedCards[0].ExecuterName, expectedCards[0].Title, expectedCards[0].DeadLine).
					AddRow(expectedCards[1].CardLink, expectedCards[1].ExecuterName, expectedCards[1].Title, expectedCards[1].DeadLine)

				m.ExpectQuery(`(?s)SELECT.*FROM task_actual.*WHERE t\.section_link = \$1.*`).
					WithArgs(targetSectionLink).
					WillReturnRows(rows)
			},
			expectedError: nil,
			expectedData:  expectedCards,
		},
		{
			nameTest: "Success get empty cards",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"task_link", "name_executer", "title", "due_date"})

				m.ExpectQuery(`(?s)SELECT.*FROM task_actual.*WHERE t\.section_link = \$1.*`).
					WithArgs(targetSectionLink).
					WillReturnRows(rows)
			},
			expectedError: nil,
			expectedData:  []dto.Card{},
		},
		{
			nameTest: "Error pool query",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?s)SELECT.*FROM task_actual.*WHERE t\.section_link = \$1.*`).
					WithArgs(targetSectionLink).
					WillReturnError(errors.New("db disconnect"))
			},
			expectedError: errors.New("pool.Query: db disconnect"),
			expectedData:  []dto.Card{},
		},
		{
			nameTest: "Error rows scan type mismatch",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"task_link", "name_executer", "title", "due_date"}).
					AddRow("invalid-uuid", "Name", "Title", time.Now())

				m.ExpectQuery(`(?s)SELECT.*FROM task_actual.*WHERE t\.section_link = \$1.*`).
					WithArgs(targetSectionLink).
					WillReturnRows(rows)
			},
			expectedError: errors.New("rows.Scan:"),
			expectedData:  []dto.Card{},
		},
		{
			nameTest: "Error rows iteration",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"task_link", "name_executer", "title", "due_date"}).
					AddRow(expectedCards[0].CardLink, expectedCards[0].ExecuterName, expectedCards[0].Title, expectedCards[0].DeadLine).
					RowError(0, errors.New("iteration error"))

				m.ExpectQuery(`(?s)SELECT.*FROM task_actual.*WHERE t\.section_link = \$1.*`).
					WithArgs(targetSectionLink).
					WillReturnRows(rows)
			},
			expectedError: errors.New("rows.Scan: iteration error"),
			expectedData:  []dto.Card{},
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

			repo := NewRepository(Deps{Pool: mockDB})
			result, err := repo.GetCards(ctx, targetSectionLink)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					if test.nameTest == "Error rows scan type mismatch" {
						assert.Contains(t, err.Error(), test.expectedError.Error())
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
