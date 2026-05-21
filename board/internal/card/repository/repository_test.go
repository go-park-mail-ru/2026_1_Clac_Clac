package repository

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/models"
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
	targetExecutorLink := uuid.New()
	targetAttLink := uuid.New()
	targetSubLink := uuid.New()
	targetAttLink1 := uuid.New()
	targetAttLink2 := uuid.New()

	expectedInfo := dto.InfoCard{
		Title:        "Test Task",
		Description:  "Description",
		DataDeadLine: &targetDeadLine,
		ExecutorLink: &targetExecutorLink,
		Subtasks:     []models.SubtaskInfo{},
		Attachments:  []models.AttachmentInfo{},
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
				rows := pgxmock.NewRows([]string{"title", "description", "due_date", "executer_link", "position", "subtasks", "attachments"}).
					AddRow(expectedInfo.Title, expectedInfo.Description, expectedInfo.DataDeadLine, expectedInfo.ExecutorLink, expectedInfo.Position, []byte("[]"), []byte("[]"))

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
		{
			nameTest: "Success get card with attachments",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				attJSON, _ := json.Marshal([]rawAttachment{
					{AttachmentLink: targetAttLink.String(), Name: "photo.png", Path: "cards-attachments/uuid.png", Position: 1},
				})
				rows := pgxmock.NewRows([]string{"title", "description", "due_date", "executer_link", "position", "subtasks", "attachments"}).
					AddRow("Task", "Desc", &targetDeadLine, &targetExecutorLink, 1, []byte("[]"), attJSON)

				m.ExpectQuery(`(?s)SELECT.*t.title.*FROM task_actual.*WHERE t.task_link = \$1`).
					WithArgs(targetLink).
					WillReturnRows(rows)
			},
			expectedError: nil,
			expectedData: dto.InfoCard{
				Title:        "Task",
				Description:  "Desc",
				DataDeadLine: &targetDeadLine,
				ExecutorLink: &targetExecutorLink,
				Subtasks:     []models.SubtaskInfo{},
				Position:     1,
				Attachments: []models.AttachmentInfo{
					{AttachmentLink: targetAttLink, Name: "photo.png", Path: "cards-attachments/uuid.png", Position: 1},
				},
			},
		},
		{
			nameTest: "Success get card with subtasks and attachments",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				subJSON, _ := json.Marshal([]rawSubtask{
					{SubtaskLink: targetSubLink.String(), Description: "do stuff", IsDone: false, Position: 1},
				})
				attJSON, _ := json.Marshal([]rawAttachment{
					{AttachmentLink: targetAttLink1.String(), Name: "doc.pdf", Path: "cards-attachments/doc.pdf", Position: 1},
					{AttachmentLink: targetAttLink2.String(), Name: "img.png", Path: "cards-attachments/img.png", Position: 2},
				})
				rows := pgxmock.NewRows([]string{"title", "description", "due_date", "executer_link", "position", "subtasks", "attachments"}).
					AddRow("Task", "Desc", &targetDeadLine, &targetExecutorLink, 2, subJSON, attJSON)

				m.ExpectQuery(`(?s)SELECT.*t.title.*FROM task_actual.*WHERE t.task_link = \$1`).
					WithArgs(targetLink).
					WillReturnRows(rows)
			},
			expectedError: nil,
			expectedData: dto.InfoCard{
				Title:        "Task",
				Description:  "Desc",
				DataDeadLine: &targetDeadLine,
				ExecutorLink: &targetExecutorLink,
				Subtasks: []models.SubtaskInfo{
					{SubtaskLink: targetSubLink, Description: "do stuff", IsDone: false, Position: 1},
				},
				Position: 2,
				Attachments: []models.AttachmentInfo{
					{AttachmentLink: targetAttLink1, Name: "doc.pdf", Path: "cards-attachments/doc.pdf", Position: 1},
					{AttachmentLink: targetAttLink2, Name: "img.png", Path: "cards-attachments/img.png", Position: 2},
				},
			},
		},
		{
			nameTest: "Error invalid attachments json",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"title", "description", "due_date", "executer_link", "position", "subtasks", "attachments"}).
					AddRow("Task", "Desc", &targetDeadLine, &targetExecutorLink, 1, []byte("[]"), []byte("{invalid"))

				m.ExpectQuery(`(?s)SELECT.*t.title.*FROM task_actual.*WHERE t.task_link = \$1`).
					WithArgs(targetLink).
					WillReturnRows(rows)
			},
			expectedError: errors.New(msgInvalidUnmarshalAttachments),
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

			repo := NewRepository(mockDB, nil)
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

			repo := NewRepository(mockDB, nil)
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
		LinkExecutor: &targetExecuter,
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
			nameTest: "Error begin tx fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin().WillReturnError(errors.New("begin error"))
			},
			expectedError: errors.New("pool.Begin: begin error"),
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
			nameTest: "Error commit tx fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`(?s)UPDATE task_version.*RETURNING section_link, position;`).
					WithArgs(targetLink).
					WillReturnRows(pgxmock.NewRows([]string{"section_link", "position"}).AddRow(targetSection, 5))
				m.ExpectExec(`(?s)INSERT INTO task_version.*`).
					WithArgs(targetLink, targetSection, &targetExecuter, "Updated Title", "Updated Desc", 5, &targetDeadLine).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectCommit().WillReturnError(errors.New("commit error"))
			},
			expectedError: errors.New("tx.Commit: commit error"),
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

			repo := NewRepository(mockDB, nil)
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
			// card moves from oldSection(pos=1) to newSection(target pos=2):
			// 1. close old version
			// 2. shift cards in old section down (pos > 1 → pos-1)
			// 3. shift cards in new section up   (pos >= 2 → pos+1)
			// 4. insert at target position=2
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
				m.ExpectExec(`(?s)UPDATE task_version.*position = position - 1.*position > \$2`).
					WithArgs(oldSectionLink, 1).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
				m.ExpectExec(`(?s)UPDATE task_version.*position = position \+ 1.*position >= \$2`).
					WithArgs(newSectionLink, 2).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
				m.ExpectExec(`(?s)INSERT INTO task_version.*`).
					WithArgs(targetCardLink, newSectionLink, 2, "Title", "Desc", &targetExecuter, &targetDeadLine).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			// card moves within same section from pos=1 to pos=4 (moving down):
			// cards between old(1) and new(4) shift down by 1
			nameTest: "Success reorder same section",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`(?s)UPDATE task_version.*RETURNING.*`).
					WithArgs(targetCardLink).
					WillReturnRows(pgxmock.NewRows([]string{"section_link", "position", "title", "description", "executer_link", "due_date"}).
						AddRow(oldSectionLink, 1, "Title", "Desc", &targetExecuter, &targetDeadLine))
				m.ExpectExec(`(?s)UPDATE task_version.*position = position - 1.*position > \$2 AND position <= \$3`).
					WithArgs(oldSectionLink, 1, 4).
					WillReturnResult(pgxmock.NewResult("UPDATE", 3))
				m.ExpectExec(`(?s)INSERT INTO task_version.*`).
					WithArgs(targetCardLink, oldSectionLink, 4, "Title", "Desc", &targetExecuter, &targetDeadLine).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			nameTest: "Error begin tx fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin().WillReturnError(errors.New("begin error"))
			},
			expectedError: errors.New("pool.Begin: begin error"),
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
			nameTest: "Error CheckTaskLimit fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`(?s)UPDATE task_version.*RETURNING.*`).
					WithArgs(targetCardLink).
					WillReturnRows(pgxmock.NewRows([]string{"section_link", "position", "title", "description", "executer_link", "due_date"}).
						AddRow(oldSectionLink, 1, "Title", "Desc", &targetExecuter, &targetDeadLine))
				m.ExpectQuery(`(?s)WITH locked_section AS.*`).
					WithArgs(newSectionLink).
					WillReturnError(errors.New("limit error"))
				m.ExpectRollback()
			},
			expectedError: errors.New("CheckTaskLimit: tx.QueryRow: limit error"),
		},
		{
			nameTest: "Error check mandatory query fail",
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
					WillReturnError(errors.New("mandatory query failed"))
				m.ExpectRollback()
			},
			expectedError: errors.New("tx.QueryRow: mandatory query failed"),
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
			// different section: queryDownPos (shift old section) fails
			nameTest: "Error shift old section positions fail",
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
				m.ExpectExec(`(?s)UPDATE task_version.*position = position - 1.*position > \$2`).
					WithArgs(oldSectionLink, 1).
					WillReturnError(errors.New("shift down failed"))
				m.ExpectRollback()
			},
			expectedError: errors.New("tx.Exec: shift down failed"),
		},
		{
			// same section, card moves down (pos 1→4): shift query fails
			nameTest: "Error shift same section positions fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`(?s)UPDATE task_version.*RETURNING.*`).
					WithArgs(targetCardLink).
					WillReturnRows(pgxmock.NewRows([]string{"section_link", "position", "title", "description", "executer_link", "due_date"}).
						AddRow(oldSectionLink, 1, "Title", "Desc", &targetExecuter, &targetDeadLine))
				m.ExpectExec(`(?s)UPDATE task_version.*position = position - 1.*position > \$2 AND position <= \$3`).
					WithArgs(oldSectionLink, 1, 4).
					WillReturnError(errors.New("shift same section failed"))
				m.ExpectRollback()
			},
			expectedError: errors.New("tx.Exec: shift same section failed"),
		},
		{
			// same section, card moves down (pos 1→4): INSERT fails with check violation
			nameTest: "Error check violation data",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`(?s)UPDATE task_version.*RETURNING.*`).
					WithArgs(targetCardLink).
					WillReturnRows(pgxmock.NewRows([]string{"section_link", "position", "title", "description", "executer_link", "due_date"}).
						AddRow(oldSectionLink, 1, "Title", "Desc", &targetExecuter, &targetDeadLine))
				m.ExpectExec(`(?s)UPDATE task_version.*position = position - 1.*position > \$2 AND position <= \$3`).
					WithArgs(oldSectionLink, 1, 4).
					WillReturnResult(pgxmock.NewResult("UPDATE", 3))
				m.ExpectExec(`(?s)INSERT INTO task_version.*`).
					WithArgs(targetCardLink, oldSectionLink, 4, "Title", "Desc", &targetExecuter, &targetDeadLine).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.CheckViolation})
				m.ExpectRollback()
			},
			expectedError: common.ErrInvalidCardData,
		},
		{
			// same section, card moves down (pos 1→4): INSERT fails with not-null violation
			nameTest: "Error missing required field",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`(?s)UPDATE task_version.*RETURNING.*`).
					WithArgs(targetCardLink).
					WillReturnRows(pgxmock.NewRows([]string{"section_link", "position", "title", "description", "executer_link", "due_date"}).
						AddRow(oldSectionLink, 1, "Title", "Desc", &targetExecuter, &targetDeadLine))
				m.ExpectExec(`(?s)UPDATE task_version.*position = position - 1.*position > \$2 AND position <= \$3`).
					WithArgs(oldSectionLink, 1, 4).
					WillReturnResult(pgxmock.NewResult("UPDATE", 3))
				m.ExpectExec(`(?s)INSERT INTO task_version.*`).
					WithArgs(targetCardLink, oldSectionLink, 4, "Title", "Desc", &targetExecuter, &targetDeadLine).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.NotNullViolation})
				m.ExpectRollback()
			},
			expectedError: common.ErrMissingRequiredField,
		},
		{
			// same section, card moves down (pos 1→4): commit fails
			nameTest: "Error commit tx fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`(?s)UPDATE task_version.*RETURNING.*`).
					WithArgs(targetCardLink).
					WillReturnRows(pgxmock.NewRows([]string{"section_link", "position", "title", "description", "executer_link", "due_date"}).
						AddRow(oldSectionLink, 1, "Title", "Desc", &targetExecuter, &targetDeadLine))
				m.ExpectExec(`(?s)UPDATE task_version.*position = position - 1.*position > \$2 AND position <= \$3`).
					WithArgs(oldSectionLink, 1, 4).
					WillReturnResult(pgxmock.NewResult("UPDATE", 3))
				m.ExpectExec(`(?s)INSERT INTO task_version.*`).
					WithArgs(targetCardLink, oldSectionLink, 4, "Title", "Desc", &targetExecuter, &targetDeadLine).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectCommit().WillReturnError(errors.New("commit error"))
			},
			expectedError: errors.New("tx.Commit: commit error"),
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

			repo := NewRepository(mockDB, nil)

			var updateDto dto.PlaceCard
			if test.nameTest == "Success reorder same section" ||
				test.nameTest == "Error shift same section positions fail" ||
				test.nameTest == "Error commit tx fail" ||
				test.nameTest == "Error check violation data" ||
				test.nameTest == "Error missing required field" {
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
		LinkExecutor: &targetExecuterLink,
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
			nameTest: "Error begin tx fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin().WillReturnError(errors.New("begin error"))
			},
			expectedError: errors.New("pool.Begin: begin error"),
			expectedData:  -1,
		},
		{
			nameTest: "Error CheckTaskLimit fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`(?s)WITH locked_section AS.*`).
					WithArgs(targetSectionLink).
					WillReturnError(errors.New("limit error"))
				m.ExpectRollback()
			},
			expectedError: errors.New("CheckTaskLimit: tx.QueryRow: limit error"),
			expectedData:  -1,
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
			nameTest: "Error query max position fail",
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
					WillReturnError(errors.New("query position failed"))
				m.ExpectRollback()
			},
			expectedError: errors.New("tx.QueryRow: query position failed"),
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
			nameTest: "Error commit tx fail",
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
				m.ExpectCommit().WillReturnError(errors.New("commit error"))
			},
			expectedError: errors.New("tx.Commit: commit error"),
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

			repo := NewRepository(mockDB, nil)
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
				rows := pgxmock.NewRows([]string{"comment_link", "author_link", "parent_link", "text", "created_at"}).
					AddRow(commentLink1, authorLink, &parentLink, "first comment", time.Now()).
					AddRow(commentLink2, authorLink, nil, "second comment", time.Now())

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
				rows := pgxmock.NewRows([]string{"comment_link", "author_link", "parent_link", "text", "created_at"})

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

			repo := NewRepository(mockDB, nil)
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

	commentLink := uuid.New()
	createInfo := dto.CreateCommentInfo{
		CommentLink: commentLink,
		CardLink:    targetCardLink,
		ParentLink:  &targetParentLink,
		AuthorLink:  targetAuthorLink,
		Text:        "hello world",
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

			repo := NewRepository(mockDB, nil)
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
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expected      bool
		expectedError error
	}{
		{
			nameTest: "Returns true when author",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?s)SELECT EXISTS.*FROM comment_task.*WHERE link = \$1 AND author_link = \$2`).
					WithArgs(targetCommentLink, targetUserLink).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
			},
			expected:      true,
			expectedError: nil,
		},
		{
			nameTest: "Returns false when not author",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?s)SELECT EXISTS.*FROM comment_task.*WHERE link = \$1 AND author_link = \$2`).
					WithArgs(targetCommentLink, targetUserLink).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(false))
			},
			expected:      false,
			expectedError: nil,
		},
		{
			nameTest: "Error comment not found (pgx.ErrNoRows)",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?s)SELECT EXISTS.*FROM comment_task.*WHERE link = \$1 AND author_link = \$2`).
					WithArgs(targetCommentLink, targetUserLink).
					WillReturnError(pgx.ErrNoRows)
			},
			expected:      false,
			expectedError: common.ErrCommentNotFound,
		},
		{
			nameTest: "Returns false on query error",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?s)SELECT EXISTS.*FROM comment_task.*WHERE link = \$1 AND author_link = \$2`).
					WithArgs(targetCommentLink, targetUserLink).
					WillReturnError(errors.New("db disconnect"))
			},
			expected:      false,
			expectedError: errors.New("pool.QueryRow: db disconnect"),
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

			repo := NewRepository(mockDB, nil)
			result, err := repo.IsCommentAuthor(ctx, targetCommentLink, targetUserLink)

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

			repo := NewRepository(mockDB, nil)
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

			repo := NewRepository(mockDB, nil)
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

func TestRepositoryCreateSubtask(t *testing.T) {
	ctx := context.Background()
	taskLink := uuid.New()
	subtaskLink := uuid.New()

	createInfo := dto.CreateSubtaskInfo{
		SubtaskLink: subtaskLink,
		TaskLink:    taskLink,
		Description: "do something",
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			nameTest: "Success create subtask",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`(?s)SELECT COALESCE.*`).
					WithArgs(taskLink).
					WillReturnRows(pgxmock.NewRows([]string{"position"}).AddRow(1))
				m.ExpectQuery(`(?s)INSERT INTO subtask.*`).
					WithArgs(taskLink, subtaskLink, "do something", 1).
					WillReturnRows(pgxmock.NewRows([]string{"subtask_link", "is_done", "position"}).AddRow(subtaskLink, false, 1))
				m.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			nameTest: "Error begin tx fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin().WillReturnError(errors.New("begin error"))
			},
			expectedError: errors.New("pool.Begin: begin error"),
		},
		{
			nameTest: "Error invalid reference data",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`(?s)SELECT COALESCE.*`).
					WithArgs(taskLink).
					WillReturnRows(pgxmock.NewRows([]string{"position"}).AddRow(1))
				m.ExpectQuery(`(?s)INSERT INTO subtask.*`).
					WithArgs(taskLink, subtaskLink, "do something", 1).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})
				m.ExpectRollback()
			},
			expectedError: common.ErrInvalidReferenceCardData,
		},
		{
			nameTest: "Error query fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`(?s)SELECT COALESCE.*`).
					WithArgs(taskLink).
					WillReturnError(errors.New("db disconnect"))
				m.ExpectRollback()
			},
			expectedError: errors.New("tx.QueryRow: db disconnect"),
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

			repo := NewRepository(mockDB, nil)
			result, err := repo.CreateSubtask(ctx, createInfo)

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
					assert.Equal(t, subtaskLink, result.SubtaskLink)
					assert.Equal(t, "do something", result.Description)
					assert.False(t, result.IsDone)
					assert.Equal(t, 1, result.Position)
				}
			}

			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestRepositoryDeleteSubtask(t *testing.T) {
	ctx := context.Background()
	subtaskLink := uuid.New()

	deleteInfo := dto.DeleteSubtask{SubTaskLink: subtaskLink}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			nameTest: "Success delete subtask",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)DELETE FROM subtask.*WHERE subtask_link = \$1`).
					WithArgs(subtaskLink).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			expectedError: nil,
		},
		{
			nameTest: "Error subtask not found",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)DELETE FROM subtask.*WHERE subtask_link = \$1`).
					WithArgs(subtaskLink).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			expectedError: common.ErrSubtaskNotFound,
		},
		{
			nameTest: "Error exec fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)DELETE FROM subtask.*WHERE subtask_link = \$1`).
					WithArgs(subtaskLink).
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

			repo := NewRepository(mockDB, nil)
			err = repo.DeleteSubtask(ctx, deleteInfo)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					if errors.Is(test.expectedError, common.ErrSubtaskNotFound) {
						assert.ErrorIs(t, err, common.ErrSubtaskNotFound)
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

func TestRepositoryUpdateSubtask(t *testing.T) {
	ctx := context.Background()
	subtaskLink := uuid.New()

	updateInfo := dto.UpdateSubtask{
		SubTaskLink: subtaskLink,
		Description: "updated desc",
		IsDone:      true,
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			nameTest: "Success update subtask",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)UPDATE subtask.*WHERE subtask_link = \$3`).
					WithArgs("updated desc", true, subtaskLink).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: nil,
		},
		{
			nameTest: "Error subtask not found",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)UPDATE subtask.*WHERE subtask_link = \$3`).
					WithArgs("updated desc", true, subtaskLink).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError: common.ErrSubtaskNotFound,
		},
		{
			nameTest: "Error exec fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)UPDATE subtask.*WHERE subtask_link = \$3`).
					WithArgs("updated desc", true, subtaskLink).
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

			repo := NewRepository(mockDB, nil)
			err = repo.UpdateSubtask(ctx, updateInfo)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					if errors.Is(test.expectedError, common.ErrSubtaskNotFound) {
						assert.ErrorIs(t, err, common.ErrSubtaskNotFound)
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

// --- S3 mock ---

type mockS3Bucket struct {
	putFunc    func(ctx context.Context, key string) (string, error)
	deleteFunc func(ctx context.Context, key string) error
}

func (m *mockS3Bucket) Put(ctx context.Context, data io.Reader, key string, contentType string) (string, error) {
	if m.putFunc != nil {
		return m.putFunc(ctx, key)
	}
	return key, nil
}

func (m *mockS3Bucket) Delete(ctx context.Context, key string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, key)
	}
	return nil
}

func TestRepositoryUploadAttachment(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		nameTest      string
		s3Behavior    func() *mockS3Bucket
		expectedKey   string
		expectedError error
	}{
		{
			nameTest: "Success upload",
			s3Behavior: func() *mockS3Bucket {
				return &mockS3Bucket{
					putFunc: func(ctx context.Context, key string) (string, error) {
						return "uploads/file.png", nil
					},
				}
			},
			expectedKey:   "uploads/file.png",
			expectedError: nil,
		},
		{
			nameTest: "Error S3 put fails",
			s3Behavior: func() *mockS3Bucket {
				return &mockS3Bucket{
					putFunc: func(ctx context.Context, key string) (string, error) {
						return "", errors.New("s3 unavailable")
					},
				}
			},
			expectedKey:   "",
			expectedError: errors.New("s3 unavailable"),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockDB, err := pgxmock.NewPool()
			if !assert.NoError(t, err) {
				return
			}
			defer mockDB.Close()

			s3 := test.s3Behavior()
			repo := NewRepository(mockDB, s3)

			key, err := repo.UploadAttachment(ctx, dto.UploadAttachment{
				FilePath:    "some-uuid.png",
				ContentType: "image/png",
			})

			if test.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedKey, key)
			}
		})
	}
}

func TestRepositoryCreateAttachment(t *testing.T) {
	ctx := context.Background()
	taskLink := uuid.New()
	attachmentLink := uuid.New()

	createInfo := dto.CreateAttachment{
		AttachmentLink: attachmentLink,
		TaskLink:       taskLink,
		Key:            "uploads/file.png",
		Name:           "photo.png",
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			nameTest: "Success create attachment",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`(?s)SELECT COALESCE.*FROM attachment WHERE task_link`).
					WithArgs(taskLink).
					WillReturnRows(pgxmock.NewRows([]string{"position"}).AddRow(1))
				m.ExpectExec(`(?s)INSERT INTO attachment.*`).
					WithArgs(attachmentLink, taskLink, "photo.png", "uploads/file.png", 1).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			nameTest: "Error begin tx fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin().WillReturnError(errors.New("begin error"))
			},
			expectedError: errors.New("CreateAttachment pool.Begin: begin error"),
		},
		{
			nameTest: "Error position query fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`(?s)SELECT COALESCE.*FROM attachment WHERE task_link`).
					WithArgs(taskLink).
					WillReturnError(errors.New("db disconnect"))
				m.ExpectRollback()
			},
			expectedError: errors.New("db disconnect"),
		},
		{
			nameTest: "Error insert fk violation",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`(?s)SELECT COALESCE.*FROM attachment WHERE task_link`).
					WithArgs(taskLink).
					WillReturnRows(pgxmock.NewRows([]string{"position"}).AddRow(1))
				m.ExpectExec(`(?s)INSERT INTO attachment.*`).
					WithArgs(attachmentLink, taskLink, "photo.png", "uploads/file.png", 1).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})
				m.ExpectRollback()
			},
			expectedError: common.ErrInvalidReferenceCardData,
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
			repo := NewRepository(mockDB, nil)
			result, err := repo.CreateAttachment(ctx, createInfo)

			if test.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, attachmentLink, result.AttachmentLink)
				assert.Equal(t, 1, result.Position)
			}

			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestRepositoryDeleteAttachmentFromDB(t *testing.T) {
	ctx := context.Background()
	attachmentLink := uuid.New()
	s3Key := "uploads/file.png"

	tests := []struct {
		nameTest      string
		mockBehavior  func(m pgxmock.PgxPoolIface)
		expectedKey   string
		expectedError error
	}{
		{
			nameTest: "Success delete",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?s)DELETE FROM attachment WHERE attachment_link = \$1 RETURNING attachment_path`).
					WithArgs(attachmentLink).
					WillReturnRows(pgxmock.NewRows([]string{"attachment_path"}).AddRow(s3Key))
			},
			expectedKey:   s3Key,
			expectedError: nil,
		},
		{
			nameTest: "Error not found",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?s)DELETE FROM attachment WHERE attachment_link = \$1 RETURNING attachment_path`).
					WithArgs(attachmentLink).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedKey:   "",
			expectedError: common.ErrAttachmentNotFound,
		},
		{
			nameTest: "Error query fail",
			mockBehavior: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?s)DELETE FROM attachment WHERE attachment_link = \$1 RETURNING attachment_path`).
					WithArgs(attachmentLink).
					WillReturnError(errors.New("db disconnect"))
			},
			expectedKey:   "",
			expectedError: errors.New("db disconnect"),
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
			repo := NewRepository(mockDB, nil)
			key, err := repo.DeleteAttachmentFromDB(ctx, attachmentLink)

			if test.expectedError != nil {
				assert.Error(t, err)
				if errors.Is(test.expectedError, common.ErrAttachmentNotFound) {
					assert.ErrorIs(t, err, common.ErrAttachmentNotFound)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedKey, key)
			}

			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestRepositoryDeleteAttachmentFromS3(t *testing.T) {
	ctx := context.Background()
	s3Key := "uploads/file.png"

	tests := []struct {
		nameTest      string
		s3Behavior    func() *mockS3Bucket
		expectedError error
	}{
		{
			nameTest: "Success delete from S3",
			s3Behavior: func() *mockS3Bucket {
				return &mockS3Bucket{
					deleteFunc: func(ctx context.Context, key string) error { return nil },
				}
			},
			expectedError: nil,
		},
		{
			nameTest: "Error S3 delete fails",
			s3Behavior: func() *mockS3Bucket {
				return &mockS3Bucket{
					deleteFunc: func(ctx context.Context, key string) error {
						return errors.New("s3 unavailable")
					},
				}
			},
			expectedError: errors.New("s3 unavailable"),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockDB, err := pgxmock.NewPool()
			if !assert.NoError(t, err) {
				return
			}
			defer mockDB.Close()

			s3 := test.s3Behavior()
			repo := NewRepository(mockDB, s3)
			err = repo.DeleteAttachmentFromS3(ctx, s3Key)

			if test.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
