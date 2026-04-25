package repository

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/s3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog"
)

type DBEngine interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type Repository struct {
	pool        DBEngine
	attachments s3.S3Bucket
}

func NewRepository(pool DBEngine, attachments s3.S3Bucket) *Repository {
	return &Repository{pool: pool, attachments: attachments}
}

func (r *Repository) CreateAppeal(ctx context.Context, info dto.CreateAppealInfo) (uuid.UUID, error) {
	query := `
		INSERT INTO appeal (user_link, mail, display_name, category, description, attachment_key)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING appeal_link;
	`

	var appealLink uuid.UUID
	err := r.pool.QueryRow(ctx, query,
		info.UserLink,
		info.Email,
		info.DisplayName,
		info.Category,
		info.Description,
		info.AttachmentKey,
	).Scan(&appealLink)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "22P02" {
				return uuid.UUID{}, fmt.Errorf("create appeal: %w", common.ErrInvalidCategory)
			}
		}

		return uuid.UUID{}, fmt.Errorf("create appeal: %w", err)
	}

	return appealLink, nil
}

func (r *Repository) GetUserAppeals(ctx context.Context, userLink uuid.UUID) ([]dto.AppealEntry, error) {
	query := `
		SELECT appeal_id, appeal_link, mail, display_name, status, category, description, attachment_key, created_at
		FROM appeal
		WHERE user_link = $1
		ORDER BY created_at DESC;
	`

	rows, err := r.pool.Query(ctx, query, userLink)
	if err != nil {
		return nil, fmt.Errorf("query user appeals: %w", err)
	}
	defer rows.Close()

	appeals := make([]dto.AppealEntry, 0)
	for rows.Next() {
		var a dto.AppealEntry
		err := rows.Scan(
			&a.AppealID,
			&a.AppealLink,
			&a.Email,
			&a.DisplayName,
			&a.Status,
			&a.Category,
			&a.Description,
			&a.AttachmentKey,
			&a.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan user appeal: %w", err)
		}
		appeals = append(appeals, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user appeals: %w", err)
	}

	return appeals, nil
}

func (r *Repository) GetOpenAppeals(ctx context.Context, supportLink uuid.UUID) ([]dto.AppealEntry, error) {
	query := `
		SELECT appeal_id, appeal_link, mail, display_name, status, category, description, attachment_key, created_at
		FROM appeal
		WHERE status = $1
		   OR supporter_link = $2
		ORDER BY created_at ASC;
	`

	rows, err := r.pool.Query(ctx, query, common.Statuses.Open, supportLink)
	if err != nil {
		return nil, fmt.Errorf("query open appeals: %w", err)
	}
	defer rows.Close()

	appeals := make([]dto.AppealEntry, 0)
	for rows.Next() {
		var a dto.AppealEntry
		err := rows.Scan(
			&a.AppealID,
			&a.AppealLink,
			&a.Email,
			&a.DisplayName,
			&a.Status,
			&a.Category,
			&a.Description,
			&a.AttachmentKey,
			&a.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan open appeal: %w", err)
		}
		appeals = append(appeals, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate open appeals: %w", err)
	}

	return appeals, nil
}

func (r *Repository) DeleteAppeal(ctx context.Context, appealLink uuid.UUID) error {
	query := `
		DELETE FROM appeal
		WHERE appeal_link = $1;
	`

	_, err := r.pool.Exec(ctx, query, appealLink)
	if err != nil {
		return fmt.Errorf("delete appeal: %w", err)
	}

	return nil
}

func (r *Repository) ChangeAppealStatus(ctx context.Context, info dto.ChangeAppealStatusInfo) error {
	query := `
		UPDATE appeal
		SET status = $1,
		    supporter_link = $2
		WHERE appeal_link = $3;
	`

	_, err := r.pool.Exec(ctx, query, info.Status, info.SupporterLink, info.AppealLink)
	if err != nil {
		return fmt.Errorf("change appeal status: %w", err)
	}

	return nil
}

func (r *Repository) GetStats(ctx context.Context) (dto.AppealStats, error) {
	query := `
		SELECT
			COUNT(*) FILTER (WHERE status = $1) AS open_count,
			COUNT(*) FILTER (WHERE status = $2) AS in_work_count,
			COUNT(*) FILTER (WHERE status = $3) AS closed_count
		FROM appeal;
	`

	var stats dto.AppealStats
	err := r.pool.QueryRow(ctx, query,
		common.Statuses.Open,
		common.Statuses.InWork,
		common.Statuses.Close,
	).Scan(&stats.Open, &stats.InWork, &stats.Close)

	if err != nil {
		return dto.AppealStats{}, fmt.Errorf("get appeal stats: %w", err)
	}

	return stats, nil
}

func (r *Repository) GetUserRole(ctx context.Context, userLink uuid.UUID) (common.Role, error) {
	getUserRoleQuery := `
		SELECT role FROM support
		WHERE user_link = $1;
	`
	row := r.pool.QueryRow(ctx, getUserRoleQuery, userLink)

	var role common.Role
	err := row.Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return common.Roles.None, nil
		}

		return common.Roles.None, fmt.Errorf("get user role on board: %w", err)
	}

	return role, nil
}

func (r *Repository) UploadAttachment(ctx context.Context, source io.Reader, filename, contentType string) (string, error) {
	logger := zerolog.Ctx(ctx)
	logger.Info().Str("filename", filename).Str("content_type", contentType).Msg("s3 upload attachment start")

	key, err := r.attachments.Put(ctx, source, filename, contentType)
	if err != nil {
		logger.Error().Err(err).Str("filename", filename).Msg("s3 upload attachment failed")
		return "", fmt.Errorf("s3 cannot upload attachment: %w", err)
	}

	logger.Info().Str("key", key).Msg("s3 upload attachment success")
	return key, nil
}

func (r *Repository) UpdateAttachmentKey(ctx context.Context, key string, appealLink uuid.UUID) error {
	query := `UPDATE appeal SET attachment_key = $1 WHERE appeal_link = $2`

	tag, err := r.pool.Exec(ctx, query, key, appealLink)
	if err != nil {
		return fmt.Errorf("update appeal attachment_key: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return common.ErrorAppealNotFound
	}

	return nil
}
