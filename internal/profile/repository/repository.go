package repository

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/s3"
	"github.com/google/uuid"
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

func NewRepository(pool DBEngine, s3Client s3.S3Client, conf config.S3) *Repository {
	return &Repository{
		pool:    pool,
		avatars: s3Client.NewBucket(conf.AvatarsBucket, conf.AvatarsPrefix, s3.ACL.PublicRead),
	}
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
	SET 
		display_name = $1, 
		description_user = $2,
		updated_at = NOW()
	WHERE link = $3 AND (
		display_name IS DISTINCT FROM $1 OR
      	description_user IS DISTINCT FROM $2
	)`

	_, err := r.pool.Exec(ctx, query, updatedInfo.NameUser, updatedInfo.DescriptionUser, updatedInfo.Link)
	if err != nil {
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
		return "", common.ErrorNonexistentUser
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

	countModifies, err := r.pool.Exec(ctx, query, nil, userLink)
	if err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}

	if countModifies.RowsAffected() == 0 {
		return common.ErrorNonexistentUser
	}

	return nil
}
