package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/user/repository/dto"
	mockS3Bucket "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/user/repository/mock_s3_bucket"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var fixedUUID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

func TestAddUser(t *testing.T) {
	tests := []struct {
		nameTest      string
		user          dto.UserInitialize
		mockSetup     func(m pgxmock.PgxPoolIface, user dto.UserInitialize)
		expectedError error
	}{
		{
			nameTest: "Success registration",
			user: dto.UserInitialize{
				Link:         fixedUUID,
				DisplayName:  "Artem",
				PasswordHash: "hash123",
				Email:        "artem@mail.ru",
			},
			mockSetup: func(m pgxmock.PgxPoolIface, user dto.UserInitialize) {
				query := `INSERT INTO "user"\s+\(link, display_name, password_hash, email\)\s+VALUES\s+\(\$1, \$2, \$3, \$4\)`
				m.ExpectExec(query).
					WithArgs(user.Link, user.DisplayName, user.PasswordHash, user.Email).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: nil,
		},
		{
			nameTest: "Error email already exists",
			user:     dto.UserInitialize{Email: "artem@mail.ru"},
			mockSetup: func(m pgxmock.PgxPoolIface, user dto.UserInitialize) {
				query := `INSERT INTO "user"\s+\(link, display_name, password_hash, email\)\s+VALUES\s+\(\$1, \$2, \$3, \$4\)`
				m.ExpectExec(query).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), user.Email).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})
			},
			expectedError: common.ErrorExistingUser,
		},
		{
			nameTest: "Error not null violation",
			user:     dto.UserInitialize{Email: "artem@mail.ru"},
			mockSetup: func(m pgxmock.PgxPoolIface, user dto.UserInitialize) {
				query := `INSERT INTO "user"\s+\(link, display_name, password_hash, email\)\s+VALUES\s+\(\$1, \$2, \$3, \$4\)`
				m.ExpectExec(query).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), user.Email).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.NotNullViolation})
			},
			expectedError: common.ErrorNotNullValue,
		},
		{
			nameTest: "Error generic DB failure",
			user:     dto.UserInitialize{Email: "artem@mail.ru"},
			mockSetup: func(m pgxmock.PgxPoolIface, user dto.UserInitialize) {
				query := `INSERT INTO "user"\s+\(link, display_name, password_hash, email\)\s+VALUES\s+\(\$1, \$2, \$3, \$4\)`
				m.ExpectExec(query).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), user.Email).
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

			test.mockSetup(mockPool, test.user)

			err = NewRepository(mockPool, nil).AddUser(context.Background(), test.user)

			if test.expectedError != nil {
				if errors.Is(test.expectedError, common.ErrorExistingUser) || errors.Is(test.expectedError, common.ErrorNotNullValue) {
					assert.ErrorIs(t, err, test.expectedError)
				} else {
					assert.EqualError(t, err, test.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestGetUser(t *testing.T) {
	tests := []struct {
		nameTest     string
		email        string
		mockSetup    func(m pgxmock.PgxPoolIface)
		expectedUser dto.UserEntity
		expectedError error
	}{
		{
			nameTest: "Success get user",
			email:    "artem@mail.ru",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `SELECT link, display_name, password_hash, email, avatar_key FROM "user" WHERE email = \$1`
				rows := pgxmock.NewRows([]string{"link", "display_name", "password_hash", "email", "avatar_key"}).
					AddRow(fixedUUID, "Artem", "hash", "artem@mail.ru", "avatar.jpg")
				m.ExpectQuery(query).WithArgs("artem@mail.ru").WillReturnRows(rows)
			},
			expectedUser: dto.UserEntity{
				Link:         fixedUUID,
				DisplayName:  "Artem",
				PasswordHash: "hash",
				Email:        "artem@mail.ru",
				Avatar:       "avatar.jpg",
			},
		},
		{
			nameTest: "Error user not found",
			email:    "unknown@mail.ru",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `SELECT link, display_name, password_hash, email, avatar_key\s+FROM "user"\s+WHERE email = \$1`
				m.ExpectQuery(query).WithArgs("unknown@mail.ru").WillReturnError(pgx.ErrNoRows)
			},
			expectedError: common.ErrorNonexistentEmail,
		},
		{
			nameTest: "Error generic DB failure",
			email:    "artem@mail.ru",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `SELECT link, display_name, password_hash, email, avatar_key FROM "user" WHERE email = \$1`
				m.ExpectQuery(query).WithArgs("artem@mail.ru").WillReturnError(errors.New("db error"))
			},
			expectedError: fmt.Errorf("pool.QueryRow: %w", errors.New("db error")),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			test.mockSetup(mockPool)

			user, err := NewRepository(mockPool, nil).GetUser(context.Background(), test.email)

			if test.expectedError != nil {
				if errors.Is(test.expectedError, common.ErrorNonexistentEmail) {
					assert.ErrorIs(t, err, test.expectedError)
				} else {
					assert.EqualError(t, err, test.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedUser, user)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestGetUserLink(t *testing.T) {
	tests := []struct {
		nameTest         string
		email            string
		mockSetup        func(m pgxmock.PgxPoolIface)
		expectedUserLink uuid.UUID
		expectedError    error
	}{
		{
			nameTest: "Success get user link",
			email:    "artem@mail.ru",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `SELECT link\s+FROM "user"\s+WHERE email = \$1`
				rows := pgxmock.NewRows([]string{"link"}).AddRow(fixedUUID)
				m.ExpectQuery(query).WithArgs("artem@mail.ru").WillReturnRows(rows)
			},
			expectedUserLink: fixedUUID,
		},
		{
			nameTest: "Error user not found",
			email:    "unknown@mail.ru",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `SELECT link\s+FROM "user"\s+WHERE email = \$1`
				m.ExpectQuery(query).WithArgs("unknown@mail.ru").WillReturnError(pgx.ErrNoRows)
			},
			expectedError: common.ErrorNonexistentEmail,
		},
		{
			nameTest: "Error generic DB failure",
			email:    "artem@mail.ru",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `SELECT link\s+FROM "user"\s+WHERE email = \$1`
				m.ExpectQuery(query).WithArgs("artem@mail.ru").WillReturnError(errors.New("db error"))
			},
			expectedError: fmt.Errorf("pool.QueryRow: %w", errors.New("db error")),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			test.mockSetup(mockPool)

			link, err := NewRepository(mockPool, nil).GetUserLink(context.Background(), test.email)

			if test.expectedError != nil {
				if errors.Is(test.expectedError, common.ErrorNonexistentEmail) {
					assert.ErrorIs(t, err, test.expectedError)
				} else {
					assert.EqualError(t, err, test.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedUserLink, link)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUpdatePassword(t *testing.T) {
	newHash := "newhash123"

	tests := []struct {
		nameTest      string
		userID        uuid.UUID
		hash          string
		mockSetup     func(m pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			nameTest: "Success update password",
			userID:   fixedUUID,
			hash:     newHash,
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `UPDATE "user" SET password_hash = $1, updated_at = NOW() WHERE link = $2`
				m.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(newHash, fixedUUID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
		},
		{
			nameTest: "Error user not found",
			userID:   fixedUUID,
			hash:     newHash,
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `UPDATE "user" SET password_hash = $1, updated_at = NOW() WHERE link = $2`
				m.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(newHash, fixedUUID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError: common.ErrorNonexistentUser,
		},
		{
			nameTest: "Error not null violation",
			userID:   fixedUUID,
			hash:     newHash,
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `UPDATE "user" SET password_hash = $1, updated_at = NOW() WHERE link = $2`
				m.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(newHash, fixedUUID).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.NotNullViolation})
			},
			expectedError: common.ErrorNotNullValue,
		},
		{
			nameTest: "Error generic DB failure",
			userID:   fixedUUID,
			hash:     newHash,
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `UPDATE "user" SET password_hash = $1, updated_at = NOW() WHERE link = $2`
				m.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(newHash, fixedUUID).
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

			err = NewRepository(mockPool, nil).UpdatePassword(context.Background(), test.userID, test.hash)

			if test.expectedError != nil {
				if errors.Is(test.expectedError, common.ErrorNonexistentUser) || errors.Is(test.expectedError, common.ErrorNotNullValue) {
					assert.ErrorIs(t, err, test.expectedError)
				} else {
					assert.EqualError(t, err, test.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestGetProfile(t *testing.T) {
	avatarKey := "avatars/key.jpg"

	tests := []struct {
		nameTest      string
		userLink      uuid.UUID
		mockSetup     func(m pgxmock.PgxPoolIface)
		expectedInfo  dto.UserInfoEntity
		expectedError error
	}{
		{
			nameTest: "Success get profile with avatar",
			userLink: fixedUUID,
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `SELECT link, display_name, description_user, email, avatar_key\s+FROM "user"\s+WHERE link = \$1`
				rows := pgxmock.NewRows([]string{"link", "display_name", "description_user", "email", "avatar_key"}).
					AddRow(fixedUUID, "Artem", "Developer", "artem@mail.ru", &avatarKey)
				m.ExpectQuery(query).WithArgs(fixedUUID).WillReturnRows(rows)
			},
			expectedInfo: dto.UserInfoEntity{
				Link:            fixedUUID,
				DisplayName:     "Artem",
				DescriptionUser: "Developer",
				Email:           "artem@mail.ru",
				AvatarKey:       avatarKey,
			},
		},
		{
			nameTest: "Success get profile without avatar",
			userLink: fixedUUID,
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `SELECT link, display_name, description_user, email, avatar_key\s+FROM "user"\s+WHERE link = \$1`
				rows := pgxmock.NewRows([]string{"link", "display_name", "description_user", "email", "avatar_key"}).
					AddRow(fixedUUID, "Artem", "Developer", "artem@mail.ru", (*string)(nil))
				m.ExpectQuery(query).WithArgs(fixedUUID).WillReturnRows(rows)
			},
			expectedInfo: dto.UserInfoEntity{
				Link:            fixedUUID,
				DisplayName:     "Artem",
				DescriptionUser: "Developer",
				Email:           "artem@mail.ru",
				AvatarKey:       "",
			},
		},
		{
			nameTest: "Error user not found",
			userLink: fixedUUID,
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `SELECT link, display_name, description_user, email, avatar_key\s+FROM "user"\s+WHERE link = \$1`
				m.ExpectQuery(query).WithArgs(fixedUUID).WillReturnError(pgx.ErrNoRows)
			},
			expectedError: common.ErrorNonexistentUser,
		},
		{
			nameTest: "Error generic DB failure",
			userLink: fixedUUID,
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `SELECT link, display_name, description_user, email, avatar_key\s+FROM "user"\s+WHERE link = \$1`
				m.ExpectQuery(query).WithArgs(fixedUUID).WillReturnError(errors.New("db error"))
			},
			expectedError: fmt.Errorf("pool.QueryRow: %w", errors.New("db error")),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			test.mockSetup(mockPool)

			info, err := NewRepository(mockPool, nil).GetProfile(context.Background(), test.userLink)

			if test.expectedError != nil {
				if errors.Is(test.expectedError, common.ErrorNonexistentUser) {
					assert.ErrorIs(t, err, test.expectedError)
				} else {
					assert.EqualError(t, err, test.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedInfo, info)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUpdateProfile(t *testing.T) {
	updated := dto.UpdatedInfo{
		Link:            fixedUUID,
		NameUser:        "New Name",
		DescriptionUser: "New bio",
	}

	tests := []struct {
		nameTest      string
		info          dto.UpdatedInfo
		mockSetup     func(m pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			nameTest: "Success update profile",
			info:     updated,
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `UPDATE "user"\s+SET display_name = \$1, description_user = \$2, updated_at = NOW\(\)\s+WHERE link = \$3`
				m.ExpectExec(query).
					WithArgs(updated.NameUser, updated.DescriptionUser, updated.Link).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
		},
		{
			nameTest: "Error not null violation",
			info:     updated,
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `UPDATE "user"\s+SET display_name = \$1, description_user = \$2, updated_at = NOW\(\)\s+WHERE link = \$3`
				m.ExpectExec(query).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.NotNullViolation})
			},
			expectedError: common.ErrorMissingRequiredField,
		},
		{
			nameTest: "Error check violation",
			info:     updated,
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `UPDATE "user"\s+SET display_name = \$1, description_user = \$2, updated_at = NOW\(\)\s+WHERE link = \$3`
				m.ExpectExec(query).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.CheckViolation})
			},
			expectedError: common.ErrorInvalidProfileData,
		},
		{
			nameTest: "Error generic DB failure",
			info:     updated,
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `UPDATE "user"\s+SET display_name = \$1, description_user = \$2, updated_at = NOW\(\)\s+WHERE link = \$3`
				m.ExpectExec(query).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
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

			err = NewRepository(mockPool, nil).UpdateProfile(context.Background(), test.info)

			if test.expectedError != nil {
				if errors.Is(test.expectedError, common.ErrorMissingRequiredField) || errors.Is(test.expectedError, common.ErrorInvalidProfileData) {
					assert.ErrorIs(t, err, test.expectedError)
				} else {
					assert.EqualError(t, err, test.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestGetAvatarKey(t *testing.T) {
	avatarKey := "avatars/key.jpg"

	tests := []struct {
		nameTest      string
		userLink      uuid.UUID
		mockSetup     func(m pgxmock.PgxPoolIface)
		expectedKey   string
		expectedError error
	}{
		{
			nameTest: "Success get avatar key",
			userLink: fixedUUID,
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `SELECT avatar_key\s+FROM "user"\s+WHERE link = \$1`
				rows := pgxmock.NewRows([]string{"avatar_key"}).AddRow(&avatarKey)
				m.ExpectQuery(query).WithArgs(fixedUUID).WillReturnRows(rows)
			},
			expectedKey: avatarKey,
		},
		{
			nameTest: "Success no avatar (null)",
			userLink: fixedUUID,
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `SELECT avatar_key\s+FROM "user"\s+WHERE link = \$1`
				rows := pgxmock.NewRows([]string{"avatar_key"}).AddRow((*string)(nil))
				m.ExpectQuery(query).WithArgs(fixedUUID).WillReturnRows(rows)
			},
			expectedKey: "",
		},
		{
			nameTest: "Error user not found",
			userLink: fixedUUID,
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `SELECT avatar_key\s+FROM "user"\s+WHERE link = \$1`
				m.ExpectQuery(query).WithArgs(fixedUUID).WillReturnError(pgx.ErrNoRows)
			},
			expectedError: common.ErrorNonexistentUser,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			test.mockSetup(mockPool)

			key, err := NewRepository(mockPool, nil).GetAvatarKey(context.Background(), test.userLink)

			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedKey, key)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUploadURLAvatar(t *testing.T) {
	tests := []struct {
		nameTest      string
		userLink      uuid.UUID
		objectKey     string
		mockSetup     func(m pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			nameTest:  "Success upload avatar URL",
			userLink:  fixedUUID,
			objectKey: "avatars/key.jpg",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `UPDATE "user"\s+SET avatar_key = \$1,\s+updated_at = NOW\(\)\s+WHERE link = \$2`
				m.ExpectExec(query).
					WithArgs("avatars/key.jpg", fixedUUID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
		},
		{
			nameTest:  "Error user not found",
			userLink:  fixedUUID,
			objectKey: "avatars/key.jpg",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `UPDATE "user"\s+SET avatar_key = \$1,\s+updated_at = NOW\(\)\s+WHERE link = \$2`
				m.ExpectExec(query).
					WithArgs("avatars/key.jpg", fixedUUID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError: common.ErrorNonexistentUser,
		},
		{
			nameTest:  "Error generic DB failure",
			userLink:  fixedUUID,
			objectKey: "avatars/key.jpg",
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `UPDATE "user"\s+SET avatar_key = \$1,\s+updated_at = NOW\(\)\s+WHERE link = \$2`
				m.ExpectExec(query).
					WithArgs("avatars/key.jpg", fixedUUID).
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

			err = NewRepository(mockPool, nil).UploadURLAvatar(context.Background(), test.userLink, test.objectKey)

			if test.expectedError != nil {
				if errors.Is(test.expectedError, common.ErrorNonexistentUser) {
					assert.ErrorIs(t, err, test.expectedError)
				} else {
					assert.EqualError(t, err, test.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestDeleteURLAvatar(t *testing.T) {
	tests := []struct {
		nameTest      string
		userLink      uuid.UUID
		mockSetup     func(m pgxmock.PgxPoolIface)
		expectedError error
	}{
		{
			nameTest: "Success delete avatar URL",
			userLink: fixedUUID,
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `UPDATE "user"\s+SET avatar_key = \$1,\s+updated_at = NOW\(\)\s+WHERE link = \$2`
				m.ExpectExec(query).
					WithArgs(nil, fixedUUID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
		},
		{
			nameTest: "Error user not found",
			userLink: fixedUUID,
			mockSetup: func(m pgxmock.PgxPoolIface) {
				query := `UPDATE "user"\s+SET avatar_key = \$1,\s+updated_at = NOW\(\)\s+WHERE link = \$2`
				m.ExpectExec(query).
					WithArgs(nil, fixedUUID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError: common.ErrorNonexistentUser,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			test.mockSetup(mockPool)

			err = NewRepository(mockPool, nil).DeleteURLAvatar(context.Background(), test.userLink)

			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUploadAvatarS3(t *testing.T) {
	t.Run("Success upload to S3", func(t *testing.T) {
		m := mockS3Bucket.NewS3Bucket(t)
		m.On("Put", mock.Anything, mock.Anything, "path/key.jpg", "image/jpeg").Return("path/key.jpg", nil)

		key, err := NewRepository(nil, m).UploadAvatarS3(context.Background(), strings.NewReader("data"), "path/key.jpg", "image/jpeg")

		assert.NoError(t, err)
		assert.Equal(t, "path/key.jpg", key)
	})

	t.Run("Error S3 put fails", func(t *testing.T) {
		m := mockS3Bucket.NewS3Bucket(t)
		m.On("Put", mock.Anything, mock.Anything, "path/key.jpg", "image/jpeg").Return("", errors.New("s3 error"))

		_, err := NewRepository(nil, m).UploadAvatarS3(context.Background(), strings.NewReader("data"), "path/key.jpg", "image/jpeg")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "avatars.Put")
	})
}

func TestDeleteAvatarS3(t *testing.T) {
	t.Run("Success delete from S3", func(t *testing.T) {
		m := mockS3Bucket.NewS3Bucket(t)
		m.On("Delete", mock.Anything, "path/key.jpg").Return(nil)

		err := NewRepository(nil, m).DeleteAvatarS3(context.Background(), "path/key.jpg")

		assert.NoError(t, err)
	})

	t.Run("Error S3 delete fails", func(t *testing.T) {
		m := mockS3Bucket.NewS3Bucket(t)
		m.On("Delete", mock.Anything, "path/key.jpg").Return(errors.New("s3 error"))

		err := NewRepository(nil, m).DeleteAvatarS3(context.Background(), "path/key.jpg")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "avatars.Delete")
	})
}
