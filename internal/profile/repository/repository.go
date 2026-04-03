package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/s3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type DBEngine interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Repository struct {
	pool    DBEngine
	avatars s3.S3Bucket
}

func NewRepository(pool DBEngine, s3Client s3.S3Client, conf *config.S3Avatars) *Repository {
	return &Repository{
		pool:    pool,
		avatars: s3Client.NewBucket(conf.Bucket, conf.Prefix, s3.ACL.PublicRead),
	}
}

func (r *Repository) GetProfile(ctx context.Context, link uuid.UUID) (dto.UserInfoEntity, error) {
	getProfileQuery := `SELECT link, display_name, email, avatar
	FROM "user"
	WHERE link = $1
	`

	var userInfo dto.UserInfoEntity
	err := r.pool.QueryRow(ctx, getProfileQuery, link).Scan(
		&userInfo.Link,
		&userInfo.DisplayName,
		&userInfo.Email,
		&userInfo.Avatar,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dto.UserInfoEntity{}, common.ErrorNonexistentUser
		}

		return dto.UserInfoEntity{}, fmt.Errorf("pool.QueryRow: %w", err)
	}

	return userInfo, nil
}
