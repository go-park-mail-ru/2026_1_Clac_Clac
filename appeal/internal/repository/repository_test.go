package repository_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/repository"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/s3"
)

type mockS3Bucket struct {
	mock.Mock
}

func (m *mockS3Bucket) Put(ctx context.Context, data io.Reader, key string, contentType string) (string, error) {
	args := m.Called(ctx, data, key, contentType)
	return args.String(0), args.Error(1)
}

func (m *mockS3Bucket) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

var _ s3.S3Bucket = (*mockS3Bucket)(nil)

func setup(dbMock pgxmock.PgxPoolIface) *repository.Repository {
	return repository.NewRepository(dbMock, new(mockS3Bucket))
}

func TestRepository_CreateAppeal(t *testing.T) {
	ctx := context.Background()
	userLink := uuid.New()
	appealLink := uuid.New()

	info := dto.CreateAppealInfo{
		UserLink:    &userLink,
		Email:       "test@test.com",
		DisplayName: "Test User",
		Category:    common.Categories.Bug,
		Description: "test desc",
	}

	tests := []struct {
		name          string
		mockSetup     func(m pgxmock.PgxPoolIface)
		expectedLink  uuid.UUID
		expectedError error
	}{
		{
			name: "Success",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?s)INSERT INTO appeal.*`).
					WithArgs(info.UserLink, info.Email, info.DisplayName, info.Category, info.Description, info.AttachmentKey).
					WillReturnRows(pgxmock.NewRows([]string{"appeal_link"}).AddRow(appealLink))
			},
			expectedLink: appealLink,
		},
		{
			name: "DB error",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?s)INSERT INTO appeal.*`).
					WithArgs(info.UserLink, info.Email, info.DisplayName, info.Category, info.Description, info.AttachmentKey).
					WillReturnError(errors.New("db error"))
			},
			expectedError: errors.New("create appeal: db error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockDB, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockDB.Close()

			test.mockSetup(mockDB)
			repo := setup(mockDB)
			link, err := repo.CreateAppeal(ctx, info)

			if test.expectedError != nil {
				assert.ErrorContains(t, err, test.expectedError.Error())
				assert.Equal(t, uuid.UUID{}, link)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedLink, link)
			}
			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestRepository_GetUserAppeals(t *testing.T) {
	ctx := context.Background()
	userLink := uuid.New()
	appealLink := uuid.New()
	createdAt := time.Now()

	tests := []struct {
		name          string
		mockSetup     func(m pgxmock.PgxPoolIface)
		expectedCount int
		expectedError error
	}{
		{
			name: "Success - returns appeals",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"appeal_id", "appeal_link", "mail", "display_name",
					"status", "category", "description", "attachment_key", "created_at",
				}).AddRow(1, appealLink, "u@test.com", "User", common.Statuses.Open, common.Categories.Bug, "desc", "", createdAt)
				m.ExpectQuery(`(?s)SELECT.*FROM appeal.*WHERE user_link.*`).
					WithArgs(userLink).
					WillReturnRows(rows)
			},
			expectedCount: 1,
		},
		{
			name: "Success - empty result",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"appeal_id", "appeal_link", "mail", "display_name",
					"status", "category", "description", "attachment_key", "created_at",
				})
				m.ExpectQuery(`(?s)SELECT.*FROM appeal.*WHERE user_link.*`).
					WithArgs(userLink).
					WillReturnRows(rows)
			},
			expectedCount: 0,
		},
		{
			name: "DB error",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?s)SELECT.*FROM appeal.*WHERE user_link.*`).
					WithArgs(userLink).
					WillReturnError(errors.New("db error"))
			},
			expectedError: errors.New("query user appeals: db error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockDB, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockDB.Close()

			test.mockSetup(mockDB)
			repo := setup(mockDB)
			appeals, err := repo.GetUserAppeals(ctx, userLink)

			if test.expectedError != nil {
				assert.ErrorContains(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, appeals, test.expectedCount)
			}
			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestRepository_GetOpenAppeals(t *testing.T) {
	ctx := context.Background()
	supportLink := uuid.New()
	appealLink := uuid.New()
	createdAt := time.Now()

	tests := []struct {
		name          string
		mockSetup     func(m pgxmock.PgxPoolIface)
		expectedCount int
		expectedError error
	}{
		{
			name: "Success",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"appeal_id", "appeal_link", "mail", "display_name",
					"status", "category", "description", "attachment_key", "created_at",
				}).AddRow(2, appealLink, "s@test.com", "Support", common.Statuses.Open, common.Categories.Complaint, "issue", "", createdAt)
				m.ExpectQuery(`(?s)SELECT.*FROM appeal.*WHERE status.*`).
					WithArgs(common.Statuses.Open, supportLink).
					WillReturnRows(rows)
			},
			expectedCount: 1,
		},
		{
			name: "DB error",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?s)SELECT.*FROM appeal.*WHERE status.*`).
					WithArgs(common.Statuses.Open, supportLink).
					WillReturnError(errors.New("db error"))
			},
			expectedError: errors.New("query open appeals: db error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockDB, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockDB.Close()

			test.mockSetup(mockDB)
			repo := setup(mockDB)
			appeals, err := repo.GetOpenAppeals(ctx, supportLink)

			if test.expectedError != nil {
				assert.ErrorContains(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, appeals, test.expectedCount)
			}
			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestRepository_DeleteAppeal(t *testing.T) {
	ctx := context.Background()
	appealLink := uuid.New()

	tests := []struct {
		name          string
		mockSetup     func(m pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			name: "Success",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)DELETE FROM appeal.*`).
					WithArgs(appealLink).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
		},
		{
			name: "DB error",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)DELETE FROM appeal.*`).
					WithArgs(appealLink).
					WillReturnError(errors.New("db error"))
			},
			expectedError: errors.New("delete appeal: db error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockDB, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockDB.Close()

			test.mockSetup(mockDB)
			repo := setup(mockDB)
			err = repo.DeleteAppeal(ctx, appealLink)

			if test.expectedError != nil {
				assert.ErrorContains(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestRepository_ChangeAppealStatus(t *testing.T) {
	ctx := context.Background()
	supportLink := uuid.New()
	appealLink := uuid.New()

	info := dto.ChangeAppealStatusInfo{
		SupporterLink: supportLink,
		AppealLink:    appealLink,
		Status:        common.Statuses.InWork,
	}

	tests := []struct {
		name          string
		mockSetup     func(m pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			name: "Success",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)UPDATE appeal.*SET status.*`).
					WithArgs(info.Status, info.SupporterLink, info.AppealLink).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
		},
		{
			name: "DB error",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)UPDATE appeal.*SET status.*`).
					WithArgs(info.Status, info.SupporterLink, info.AppealLink).
					WillReturnError(errors.New("db error"))
			},
			expectedError: errors.New("change appeal status: db error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockDB, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockDB.Close()

			test.mockSetup(mockDB)
			repo := setup(mockDB)
			err = repo.ChangeAppealStatus(ctx, info)

			if test.expectedError != nil {
				assert.ErrorContains(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestRepository_GetStats(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		mockSetup     func(m pgxmock.PgxPoolIface)
		expectedStats dto.AppealStats
		expectedError error
	}{
		{
			name: "Success",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?s)SELECT.*COUNT.*FROM appeal.*`).
					WithArgs(common.Statuses.Open, common.Statuses.InWork, common.Statuses.Close).
					WillReturnRows(pgxmock.NewRows([]string{"open_count", "in_work_count", "closed_count"}).AddRow(5, 2, 10))
			},
			expectedStats: dto.AppealStats{Open: 5, InWork: 2, Close: 10},
		},
		{
			name: "DB error",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?s)SELECT.*COUNT.*FROM appeal.*`).
					WithArgs(common.Statuses.Open, common.Statuses.InWork, common.Statuses.Close).
					WillReturnError(errors.New("db error"))
			},
			expectedError: errors.New("get appeal stats: db error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockDB, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockDB.Close()

			test.mockSetup(mockDB)
			repo := setup(mockDB)
			stats, err := repo.GetStats(ctx)

			if test.expectedError != nil {
				assert.ErrorContains(t, err, test.expectedError.Error())
				assert.Equal(t, dto.AppealStats{}, stats)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedStats, stats)
			}
			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestRepository_UpdateAttachmentKey(t *testing.T) {
	ctx := context.Background()
	appealLink := uuid.New()
	key := "attachments/file.jpg"

	tests := []struct {
		name          string
		mockSetup     func(m pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			name: "Success",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)UPDATE appeal SET attachment_key.*`).
					WithArgs(key, appealLink).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
		},
		{
			name: "Appeal not found",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)UPDATE appeal SET attachment_key.*`).
					WithArgs(key, appealLink).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError: common.ErrorAppealNotFound,
		},
		{
			name: "DB error",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`(?s)UPDATE appeal SET attachment_key.*`).
					WithArgs(key, appealLink).
					WillReturnError(errors.New("db error"))
			},
			expectedError: errors.New("update appeal attachment_key: db error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockDB, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockDB.Close()

			test.mockSetup(mockDB)
			repo := setup(mockDB)
			err = repo.UpdateAttachmentKey(ctx, key, appealLink)

			if test.expectedError != nil {
				if errors.Is(test.expectedError, common.ErrorAppealNotFound) {
					assert.ErrorIs(t, err, common.ErrorAppealNotFound)
				} else {
					assert.ErrorContains(t, err, test.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}

func TestRepository_GetUserRole(t *testing.T) {
	ctx := context.Background()
	userLink := uuid.New()

	tests := []struct {
		name          string
		mockSetup     func(m pgxmock.PgxPoolIface)
		expectedRole  string
		expectedError error
	}{
		{
			name: "Success - Support role",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?is)SELECT role FROM support.*`).
					WithArgs(userLink).
					WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow("support"))
			},
			expectedRole: "support",
		},
		{
			name: "User not found - default User role",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?is)SELECT role FROM support.*`).
					WithArgs(userLink).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedRole: "user",
		},
		{
			name: "DB error",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`(?is)SELECT role FROM support.*`).
					WithArgs(userLink).
					WillReturnError(errors.New("db error"))
			},
			expectedRole:  "user",
			expectedError: errors.New("get user role on board: db error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockDB, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockDB.Close()

			test.mockSetup(mockDB)
			repo := setup(mockDB)
			role, err := repo.GetUserRole(ctx, userLink)

			if test.expectedError != nil {
				assert.ErrorContains(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expectedRole, string(role))
			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}
