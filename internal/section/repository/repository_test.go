package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/section/repository/dto"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

				m.ExpectQuery(`(?is)SELECT.*section_name.*FROM section_actual.*WHERE section_link = \$1`).
					WithArgs(targetLink).
					WillReturnRows(rows)
			},
			expectedError: nil,
			expectedData:  expectedInfo,
		},
		{
			nameTest: "Error not found pgx.ErrNoRows",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?is)SELECT.*section_name.*FROM section_actual.*WHERE section_link = \$1`).
					WithArgs(targetLink).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedError: common.ErrorNotExistingSection,
			expectedData:  dto.FullSectionInfo{},
		},
		{
			nameTest: "Error query fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?is)SELECT.*section_name.*FROM section_actual.*WHERE section_link = \$1`).
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

			repo := NewRepository(mockDB)
			result, err := repo.GetSectionInfo(ctx, targetLink)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					if errors.Is(test.expectedError, common.ErrorNotExistingSection) {
						assert.ErrorIs(t, err, common.ErrorNotExistingSection)
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
				rows := pgxmock.NewRows([]string{"section_name", "position", "is_mandatory", "color", "max_tasks"}).
					AddRow(expectedResult.SectionName, expectedResult.Position, expectedResult.IsMandatory, expectedResult.Color, expectedResult.MaxTasks)

				m.ExpectQuery(`(?is)INSERT INTO section_actual.*VALUES.*RETURNING.*`).
					WithArgs(
						creatingSection.SectionLink,
						creatingSection.BoardLink,
						creatingSection.SectionName,
						creatingSection.IsMandatory,
						creatingSection.Color,
						creatingSection.MaxTasks,
					).
					WillReturnRows(rows)
			},
			expectedError: nil,
			expectedData:  expectedResult,
		},
		{
			nameTest: "Error execution fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?is)INSERT INTO section_actual.*VALUES.*RETURNING.*`).
					WithArgs(
						creatingSection.SectionLink,
						creatingSection.BoardLink,
						creatingSection.SectionName,
						creatingSection.IsMandatory,
						creatingSection.Color,
						creatingSection.MaxTasks,
					).
					WillReturnError(errors.New("insert conflict"))
			},
			expectedError: errors.New("QueryRow: insert conflict"),
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

			repo := NewRepository(mockDB)
			result, err := repo.CreateSection(ctx, creatingSection)

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

func TestRepositoryDeleteSection(t *testing.T) {
	ctx := context.Background()
	targetLink := common.FixedSectionUuiD

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			nameTest: "Success delete section",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?is)WITH target_info AS.*UPDATE section.*SET deleted_at.*WHERE section_link = \$1`).
					WithArgs(targetLink).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: nil,
		},
		{
			nameTest: "Error zero rows affected",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?is)WITH target_info AS.*UPDATE section.*SET deleted_at.*WHERE section_link = \$1`).
					WithArgs(targetLink).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError: common.ErrorNotExistingSection,
		},
		{
			nameTest: "Error connection fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?is)WITH target_info AS.*UPDATE section.*SET deleted_at.*WHERE section_link = \$1`).
					WithArgs(targetLink).
					WillReturnError(errors.New("db down"))
			},
			expectedError: errors.New("pool.Exec: db down"),
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
			err = repo.DeleteSection(ctx, targetLink)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					if errors.Is(test.expectedError, common.ErrorNotExistingSection) {
						assert.ErrorIs(t, err, common.ErrorNotExistingSection)
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
	section1 := common.FixedSectionUuiD
	section2 := common.FixedSectionUuiD
	linksSection := []uuid.UUID{section1, section2}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			nameTest: "Success reorder sections",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?is)UPDATE section_actual sa.*SET position = data\.new_pos.*`).
					WithArgs(section1, 1, section2, 2, boardLink).
					WillReturnResult(pgxmock.NewResult("UPDATE", 2))
			},
			expectedError: nil,
		},
		{
			nameTest: "Error not all links updated",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?is)UPDATE section_actual sa.*SET position = data\.new_pos.*`).
					WithArgs(section1, 1, section2, 2, boardLink).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: common.ErrorNotFindAllLinks,
		},
		{
			nameTest: "Error query fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?is)UPDATE section_actual sa.*SET position = data\.new_pos.*`).
					WithArgs(section1, 1, section2, 2, boardLink).
					WillReturnError(errors.New("db error"))
			},
			expectedError: errors.New("pool.Exec: db error"),
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
			err = repo.ReorderSection(ctx, boardLink, linksSection)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					if errors.Is(test.expectedError, common.ErrorNotFindAllLinks) {
						assert.ErrorIs(t, err, common.ErrorNotFindAllLinks)
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
				m.ExpectExec(`(?is)UPDATE section_actual.*SET section_name = \$1.*WHERE section_link = \$5`).
					WithArgs(
						updateData.SectionName,
						updateData.IsMandatory,
						updateData.Color,
						updateData.MaxTasks,
						updateData.SectionLink,
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: nil,
		},
		{
			nameTest: "Error zero rows affected",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?is)UPDATE section_actual.*SET section_name = \$1.*WHERE section_link = \$5`).
					WithArgs(
						updateData.SectionName,
						updateData.IsMandatory,
						updateData.Color,
						updateData.MaxTasks,
						updateData.SectionLink,
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError: common.ErrorNotExistingSection,
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
			err = repo.UpdateSection(ctx, updateData)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					if errors.Is(test.expectedError, common.ErrorNotExistingSection) {
						assert.ErrorIs(t, err, common.ErrorNotExistingSection)
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

				m.ExpectQuery(`(?is)SELECT.*FROM section_actual.*WHERE board_link = \$1.*ORDER BY position ASC`).
					WithArgs(boardLink).
					WillReturnRows(rows)
			},
			expectedError: nil,
			expectedData:  expectedSections,
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
