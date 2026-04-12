package repository

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/s3"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBEngine interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Deps struct {
	Pool    DBEngine
	Avatars s3.S3Bucket
}

type Repository struct {
	deps Deps
}

func NewRepository(deps Deps) *Repository {
	return &Repository{
		deps: deps,
	}
}

func (r *Repository) GetProfile(ctx context.Context, userLink uuid.UUID) (dto.UserInfoEntity, error) {
	getProfileQuery := `SELECT link, display_name, description_user, email, avatar_key
	FROM "user"
	WHERE link = $1
	`

	var avatarKeyPtr *string

	var userInfo dto.UserInfoEntity
	err := r.deps.Pool.QueryRow(ctx, getProfileQuery, userLink).Scan(
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

func (r *Repository) GetProfileByLink(ctx context.Context, userLink uuid.UUID) (dto.UserInfoEntity, error) {
	return r.GetProfile(ctx, userLink)
}

func (r *Repository) UpdateProfile(ctx context.Context, updatedInfo dto.UpdatedInfo) error {
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

	_, err := r.deps.Pool.Exec(ctx, query, updatedInfo.NameUser, updatedInfo.DescriptionUser, updatedInfo.Link)
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
	err := r.deps.Pool.QueryRow(ctx, query, userLink).Scan(&avatarKeyPtr)
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
	key, err := r.deps.Avatars.Put(ctx, file, pathFile, contentType)
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

	countModifies, err := r.deps.Pool.Exec(ctx, updateAvatar, objectKey, userLink)
	if err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}

	if countModifies.RowsAffected() == 0 {
		return common.ErrorNonexistentUser
	}

	return nil
}

func (r *Repository) DeleteAvatarS3(ctx context.Context, key string) error {
	err := r.deps.Avatars.Delete(ctx, key)
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

	countModifies, err := r.deps.Pool.Exec(ctx, query, nil, userLink)
	if err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}

	if countModifies.RowsAffected() == 0 {
		return common.ErrorNonexistentUser
	}

	return nil
}
