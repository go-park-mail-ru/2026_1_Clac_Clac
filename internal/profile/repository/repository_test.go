package repository

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/repository/dto"
	mockS3 "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/repository/mock_s3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetProfile(t *testing.T) {
	targetLink := common.FixedUserUuiD
	avatar := "avatar.jpg"

	expectedDTO := dto.UserInfoEntity{
		Link:            targetLink,
		DisplayName:     "Bobr",
		DescriptionUser: "desc",
		Email:           "bobr@mail.ru",
		AvatarKey:       avatar,
	}

	tests := []struct {
		nameTest      string
		targetID      uuid.UUID
		mockSetup     func(mock pgxmock.PgxPoolIface)
		expectedUser  dto.UserInfoEntity
		expectedError error
	}{
		{
			nameTest: "Success get user profile",
			targetID: targetLink,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				query := `SELECT link, display_name, description_user, email, avatar_key
				FROM "user"
				WHERE link = $1
				`
				rows := pgxmock.NewRows([]string{"link", "display_name", "description_user", "email", "avatar_key"}).
					AddRow(expectedDTO.Link, expectedDTO.DisplayName, expectedDTO.DescriptionUser, expectedDTO.Email, &avatar)

				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(targetLink).WillReturnRows(rows)
			},
			expectedUser:  expectedDTO,
			expectedError: nil,
		},
		{
			nameTest: "Success get user profile with nil avatar",
			targetID: targetLink,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				query := `SELECT link, display_name, description_user, email, avatar_key
				FROM "user"
				WHERE link = $1
				`

				rows := pgxmock.NewRows([]string{"link", "display_name", "description_user", "email", "avatar_key"}).
					AddRow(expectedDTO.Link, expectedDTO.DisplayName, expectedDTO.DescriptionUser, expectedDTO.Email, nil)

				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(targetLink).WillReturnRows(rows)
			},
			expectedUser: dto.UserInfoEntity{
				Link:            targetLink,
				DisplayName:     "Bobr",
				DescriptionUser: "desc",
				Email:           "bobr@mail.ru",
				AvatarKey:       "",
			},
			expectedError: nil,
		},
		{
			nameTest: "Error user not found",
			targetID: targetLink,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				query := `SELECT link, display_name, description_user, email, avatar_key
				FROM "user"
				WHERE link = $1
				`
				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(targetLink).WillReturnError(pgx.ErrNoRows)
			},
			expectedUser:  dto.UserInfoEntity{},
			expectedError: common.ErrorNonexistentUser,
		},
		{
			nameTest: "Error from db",
			targetID: targetLink,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				query := `SELECT link, display_name, description_user, email, avatar_key
				FROM "user"
				WHERE link = $1
				`
				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(targetLink).WillReturnError(errors.New("db error"))
			},
			expectedUser:  dto.UserInfoEntity{},
			expectedError: fmt.Errorf("pool.QueryRow: %w", errors.New("db error")),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			test.mockSetup(mockPool)

			repoProfile := NewRepository(Deps{Pool: mockPool})
			user, err := repoProfile.GetProfile(context.Background(), test.targetID)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expectedUser, user)

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUpdateProfile(t *testing.T) {
	targetLink := uuid.New()
	updatedInfo := dto.UpdatedInfo{
		Link:            targetLink,
		NameUser:        "Bobr",
		DescriptionUser: "desc",
	}

	tests := []struct {
		nameTest      string
		info          dto.UpdatedInfo
		mockSetup     func(mock pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			nameTest: "Success update profile",
			info:     updatedInfo,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				query := `
				UPDATE "user"
				SET
					display_name = $1,
					description_user = $2,
					updated_at = NOW()
				WHERE link = $3 AND (
					display_name IS DISTINCT FROM $1 OR
					description_user IS DISTINCT FROM $2
				)`

				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(updatedInfo.NameUser, updatedInfo.DescriptionUser, updatedInfo.Link).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: nil,
		},
		{
			nameTest: "Error from DB",
			info:     updatedInfo,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				query := `
				UPDATE "user"
				SET
					display_name = $1,
					description_user = $2,
					updated_at = NOW()
				WHERE link = $3 AND (
					display_name IS DISTINCT FROM $1 OR
					description_user IS DISTINCT FROM $2
				)`

				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(updatedInfo.NameUser, updatedInfo.DescriptionUser, updatedInfo.Link).
					WillReturnError(errors.New("db error"))
			},
			expectedError: fmt.Errorf("pool.Exec: %w", errors.New("db error")),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			test.mockSetup(mockPool)

			repoProfile := NewRepository(Deps{Pool: mockPool})
			err = repoProfile.UpdateProfile(context.Background(), test.info)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestGetAvatarKey(t *testing.T) {
	targetLink := uuid.New()
	avatar := "avatar.jpg"

	tests := []struct {
		nameTest      string
		targetID      uuid.UUID
		mockSetup     func(mock pgxmock.PgxPoolIface)
		expectedKey   string
		expectedError error
	}{
		{
			nameTest: "Success get avatar key",
			targetID: targetLink,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				query := `
				SELECT avatar_key
				FROM "user"
				WHERE link = $1
				`
				rows := pgxmock.NewRows([]string{"avatar_key"}).AddRow(&avatar)
				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(targetLink).WillReturnRows(rows)
			},
			expectedKey:   avatar,
			expectedError: nil,
		},
		{
			nameTest: "Success get nil avatar key",
			targetID: targetLink,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				query := `
				SELECT avatar_key
				FROM "user"
				WHERE link = $1
				`
				rows := pgxmock.NewRows([]string{"avatar_key"}).AddRow(nil)
				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(targetLink).WillReturnRows(rows)
			},
			expectedKey:   "",
			expectedError: nil,
		},
		{
			nameTest: "Error DB returns error",
			targetID: targetLink,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				query := `
				SELECT avatar_key
				FROM "user"
				WHERE link = $1
				`
				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(targetLink).WillReturnError(errors.New("db error"))
			},
			expectedKey:   "",
			expectedError: fmt.Errorf("pool.QueryRow: db error"),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			test.mockSetup(mockPool)

			repoProfile := NewRepository(Deps{Pool: mockPool})
			key, err := repoProfile.GetAvatarKey(context.Background(), test.targetID)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expectedKey, key)

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUploadAvatarS3(t *testing.T) {
	file := bytes.NewReader([]byte("test_avatar"))
	pathFile := "user123/avatar.jpg"
	contentType := "image/jpg"
	expectedKey := "test_key_s3"

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockS3.S3Bucket)
		expectedKey   string
		expectedError error
	}{
		{
			nameTest: "Success upload to S3",
			mockBehavior: func(m *mockS3.S3Bucket) {
				m.On("Put", mock.Anything, file, pathFile, contentType).Return(expectedKey, nil)
			},
			expectedKey:   expectedKey,
			expectedError: nil,
		},
		{
			nameTest: "Error from S3",
			mockBehavior: func(m *mockS3.S3Bucket) {
				m.On("Put", mock.Anything, file, pathFile, contentType).Return("", errors.New("s3 upload failed"))
			},
			expectedKey:   "",
			expectedError: fmt.Errorf("avatars.Put: %w", errors.New("s3 upload failed")),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockBucket := mockS3.NewS3Bucket(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockBucket)
			}

			repoProfile := NewRepository(Deps{Avatars: mockBucket})
			key, err := repoProfile.UploadAvatarS3(context.Background(), file, pathFile, contentType)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expectedKey, key)
		})
	}
}

func TestUploadURLAvatar(t *testing.T) {
	targetLink := uuid.New()
	objectKey := "new_avatar.jpg"

	tests := []struct {
		nameTest      string
		mockSetup     func(mock pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			nameTest: "Success upload URL avatar",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				query := `
				UPDATE "user"
				SET avatar_key = $1,
				updated_at = NOW()
				WHERE link = $2
				`
				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(objectKey, targetLink).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: nil,
		},
		{
			nameTest: "Error user not found",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				query := `
				UPDATE "user"
				SET avatar_key = $1,
				updated_at = NOW()
				WHERE link = $2
				`
				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(objectKey, targetLink).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError: common.ErrorNonexistentUser,
		},
		{
			nameTest: "Error DB",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				query := `
				UPDATE "user"
				SET avatar_key = $1,
				updated_at = NOW()
				WHERE link = $2
				`
				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(objectKey, targetLink).
					WillReturnError(errors.New("db error"))
			},
			expectedError: fmt.Errorf("pool.Exec: %w", errors.New("db error")),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			test.mockSetup(mockPool)

			repoProfile := NewRepository(Deps{Pool: mockPool})
			err = repoProfile.UploadURLAvatar(context.Background(), targetLink, objectKey)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestDeleteAvatarS3(t *testing.T) {
	deleteKey := "delete_key_s3.jpg"

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockS3.S3Bucket)
		expectedError error
	}{
		{
			nameTest: "Success delete S3",
			mockBehavior: func(m *mockS3.S3Bucket) {
				m.On("Delete", mock.Anything, deleteKey).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error delete S3",
			mockBehavior: func(m *mockS3.S3Bucket) {
				m.On("Delete", mock.Anything, deleteKey).Return(errors.New("s3 delete failed"))
			},
			expectedError: fmt.Errorf("avatars.Delete: %w", errors.New("s3 delete failed")),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockBucket := mockS3.NewS3Bucket(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockBucket)
			}

			repoProfile := NewRepository(Deps{Avatars: mockBucket})
			err := repoProfile.DeleteAvatarS3(context.Background(), deleteKey)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteURLAvatar(t *testing.T) {
	targetLink := uuid.New()

	tests := []struct {
		nameTest      string
		mockSetup     func(mock pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			nameTest: "Success delete URL avatar",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				query := `
				UPDATE "user"
				SET avatar_key = $1,
				updated_at = NOW()
				WHERE link = $2
				`
				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(nil, targetLink).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: nil,
		},
		{
			nameTest: "Error user not found",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				query := `
				UPDATE "user"
				SET avatar_key = $1,
				updated_at = NOW()
				WHERE link = $2
				`
				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(nil, targetLink).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError: common.ErrorNonexistentUser,
		},
		{
			nameTest: "Error DB",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				query := `
				UPDATE "user"
				SET avatar_key = $1,
				updated_at = NOW()
				WHERE link = $2
				`
				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(nil, targetLink).
					WillReturnError(errors.New("db error"))
			},
			expectedError: fmt.Errorf("pool.Exec: %w", errors.New("db error")),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			test.mockSetup(mockPool)

			repoProfile := NewRepository(Deps{Pool: mockPool})
			err = repoProfile.DeleteURLAvatar(context.Background(), targetLink)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
