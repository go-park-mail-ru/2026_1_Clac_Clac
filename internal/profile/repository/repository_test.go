package repository

import (
	"context"
	"regexp"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/repository/dto"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProfile(t *testing.T) {
	targetID := common.FixedUserUuiD

	expectedDTO := dto.UserInfoEntity{
		Link:        targetID,
		DisplayName: "Bobr",
		Email:       "bobr@mail.ru",
		Avatar:      "avatar.jpg",
	}

	tests := []struct {
		nameTest     string
		targetID     uuid.UUID
		mockSetup    func(mock pgxmock.PgxPoolIface, targetID uuid.UUID)
		expectedUser dto.UserInfoEntity
	}{
		{
			nameTest: "Success get user profile",
			targetID: targetID,
			mockSetup: func(mock pgxmock.PgxPoolIface, targetID uuid.UUID) {
				getProfileQuery := `SELECT link, display_name, email, avatar
				FROM "user"
				WHERE link = $1
				`

				rows := pgxmock.NewRows([]string{"link", "display_name", "email", "avatar"}).
					AddRow(expectedDTO.Link, expectedDTO.DisplayName, expectedDTO.Email, expectedDTO.Avatar)

				mock.ExpectQuery(regexp.QuoteMeta(getProfileQuery)).
					WithArgs(targetID).
					WillReturnRows(rows)
			},
			expectedUser: expectedDTO,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			test.mockSetup(mock, test.targetID)

			repoProfile := NewRepository(mock)
			ctx := context.Background()

			user, err := repoProfile.GetProfile(ctx, test.targetID)

			assert.NoError(t, err)
			assert.Equal(t, test.expectedUser, user)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err, "not wait error")
		})
	}
}

func TestGetProfileError(t *testing.T) {
	targetID := uuid.New()

	tests := []struct {
		nameTest      string
		targetID      uuid.UUID
		mockSetup     func(mock pgxmock.PgxPoolIface, targetID uuid.UUID)
		expectedUser  dto.UserInfoEntity
		expectedError error
	}{
		{
			nameTest: "User not found",
			targetID: targetID,
			mockSetup: func(mock pgxmock.PgxPoolIface, targetID uuid.UUID) {
				getProfileQuery := `SELECT link, display_name, email, avatar
				FROM "user"
				WHERE link = $1
				`
				mock.ExpectQuery(regexp.QuoteMeta(getProfileQuery)).
					WithArgs(targetID).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedUser:  dto.UserInfoEntity{},
			expectedError: common.ErrorNonexistentUser,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			test.mockSetup(mock, test.targetID)

			repoProfile := NewRepository(mock)
			ctx := context.Background()

			user, err := repoProfile.GetProfile(ctx, test.targetID)

			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedUser, user)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err, "not wait error")
		})
	}
}
