package repository

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/s3"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/user/repository/dto"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBEngine interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Repository struct {
	pool    DBEngine
	avatars s3.S3Bucket
}

func NewRepository(pool DBEngine, avatars s3.S3Bucket) *Repository {
	return &Repository{
		pool:    pool,
		avatars: avatars,
	}
}

func (r *Repository) AddUser(ctx context.Context, user dto.UserInitialize) error {
	addUserQuery := `
		INSERT INTO "user" (link, display_name, password_hash, email)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.pool.Exec(ctx, addUserQuery,
		user.Link,
		user.DisplayName,
		user.PasswordHash,
		user.Email,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return common.ErrorExistingUser
			case pgerrcode.NotNullViolation:
				return common.ErrorNotNullValue
			}
		}

		return fmt.Errorf("pool.Exec: %w", err)
	}

	return nil
}

func (r *Repository) GetUser(ctx context.Context, email string) (dto.UserEntity, error) {
	getUserQuery := `
		SELECT link, display_name, password_hash, email, avatar_key
		FROM "user"
		WHERE email = $1
	`
	var user dto.UserEntity
	err := r.pool.QueryRow(ctx, getUserQuery, email).Scan(
		&user.Link,
		&user.DisplayName,
		&user.PasswordHash,
		&user.Email,
		&user.Avatar,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dto.UserEntity{}, common.ErrorNonexistentEmail
		}
		return dto.UserEntity{}, fmt.Errorf("pool.QueryRow: %w", err)
	}

	return user, nil
}

func (r *Repository) GetUserLink(ctx context.Context, email string) (uuid.UUID, error) {
	query := `
	SELECT link
	FROM "user"
	WHERE email = $1`

	var link uuid.UUID

	err := r.pool.QueryRow(ctx, query, email).Scan(&link)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, common.ErrorNonexistentEmail
		}
		return uuid.Nil, fmt.Errorf("pool.QueryRow: %w", err)
	}

	return link, nil
}

func (r *Repository) UpdatePassword(ctx context.Context, link uuid.UUID, newPasswordHash string) error {
	updatePasswordQuery := `
	UPDATE "user"
	SET password_hash = $1,
	updated_at = NOW()
	WHERE link = $2
	`

	countModifies, err := r.pool.Exec(ctx, updatePasswordQuery, newPasswordHash, link)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.NotNullViolation {
			return common.ErrorNotNullValue
		}

		return fmt.Errorf("pool.Exec: %w", err)
	}

	if countModifies.RowsAffected() == 0 {
		return common.ErrorNonexistentUser
	}

	return nil
}

func (r *Repository) GetProfile(ctx context.Context, userLink uuid.UUID) (dto.UserInfoEntity, error) {
	getProfileQuery := `SELECT link, display_name, description_user, email, avatar_key
	FROM "user"
	WHERE link = $1
	`

	var avatarKeyPtr *string

	var userInfo dto.UserInfoEntity
	err := r.pool.QueryRow(ctx, getProfileQuery, userLink).Scan(
		&userInfo.Link,
		&userInfo.DisplayName,
		&userInfo.DescriptionUser,
		&userInfo.Email,
		&avatarKeyPtr,
	)

	if avatarKeyPtr != nil {
		userInfo.AvatarKey = *avatarKeyPtr
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dto.UserInfoEntity{}, common.ErrorNonexistentUser
		}

		return dto.UserInfoEntity{}, fmt.Errorf("pool.QueryRow: %w", err)
	}

	return userInfo, nil
}

func (r *Repository) UpdateProfile(ctx context.Context, updatedInfo dto.UpdatedInfo) error {
	query := `
    UPDATE "user"
    SET display_name = $1, description_user = $2, updated_at = NOW()
    WHERE link = $3`

	_, err := r.pool.Exec(ctx, query, updatedInfo.NameUser, updatedInfo.DescriptionUser, updatedInfo.Link)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.NotNullViolation:
				return common.ErrorMissingRequiredField
			case pgerrcode.CheckViolation:
				return common.ErrorInvalidProfileData
			}
		}
		return fmt.Errorf("pool.Exec: %w", err)
	}

	return nil
}

func (r *Repository) GetAvatarKey(ctx context.Context, userLink uuid.UUID) (string, error) {
	query := `
	SELECT avatar_key
	FROM "user"
	WHERE link = $1
	`

	var avatarKeyPtr *string
	err := r.pool.QueryRow(ctx, query, userLink).Scan(&avatarKeyPtr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", common.ErrorNonexistentUser
		}

		return "", fmt.Errorf("pool.QueryRow: %w", err)
	}

	if avatarKeyPtr == nil {
		return "", nil
	}

	return *avatarKeyPtr, nil
}

func (r *Repository) UploadAvatarS3(ctx context.Context, file io.Reader, pathFile, contentType string) (string, error) {
	key, err := r.avatars.Put(ctx, file, pathFile, contentType)
	if err != nil {
		return "", fmt.Errorf("avatars.Put: %w", err)
	}

	return key, nil
}

func (r *Repository) UploadURLAvatar(ctx context.Context, userLink uuid.UUID, objectKey string) error {
	updateAvatar := `
	UPDATE "user"
	SET avatar_key = $1,
	updated_at = NOW()
	WHERE link = $2
	`

	countModifies, err := r.pool.Exec(ctx, updateAvatar, objectKey, userLink)
	if err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}

	if countModifies.RowsAffected() == 0 {
		return common.ErrorNonexistentUser
	}

	return nil
}

func (r *Repository) DeleteAvatarS3(ctx context.Context, key string) error {
	err := r.avatars.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("avatars.Delete: %w", err)
	}

	return nil
}

func (r *Repository) DeleteURLAvatar(ctx context.Context, userLink uuid.UUID) error {
	query := `
	UPDATE "user"
	SET avatar_key = $1,
	updated_at = NOW()
	WHERE link = $2
	`

	countModifies, err := r.pool.Exec(ctx, query, "", userLink)
	if err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}

	if countModifies.RowsAffected() == 0 {
		return common.ErrorNonexistentUser
	}

	return nil
}
