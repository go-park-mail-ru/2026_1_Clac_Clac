package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/section/repository/dto"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBEngine interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type Repository struct {
	pool DBEngine
}

func NewRepository(pool DBEngine) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) GetSectionInfo(ctx context.Context, link uuid.UUID) (dto.FullSectionInfo, error) {
	query := `
	SELECT
        section_name,
        position,
        is_mandatory,
        color,
        max_tasks
    FROM section_actual 
    WHERE section_link = $1
	`

	var infoSection dto.FullSectionInfo
	err := r.pool.QueryRow(ctx, query, link).Scan(
		&infoSection.SectionName,
		&infoSection.Position,
		&infoSection.IsMandatory,
		&infoSection.Color,
		&infoSection.MaxTasks,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dto.FullSectionInfo{}, common.ErrorNotExistingSection
		}

		return dto.FullSectionInfo{}, fmt.Errorf("QueryRow: %w", err)
	}

	return infoSection, nil
}

func (r *Repository) CreateSection(ctx context.Context, newSection dto.CreatingSection) (dto.FullSectionInfo, error) {
	query := `
	INSERT INTO section_actual 
	(
		section_link, 
		board_link, 
		section_name, 
		is_mandatory, 
		color, 
		max_tasks
	)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING section_name, position, is_mandatory, color, max_tasks;
	`
	var infoSection dto.FullSectionInfo

	err := r.pool.QueryRow(ctx, query, newSection.SectionLink, newSection.BoardLink, newSection.SectionName,
		newSection.IsMandatory, newSection.Color, newSection.MaxTasks).Scan(
		&infoSection.SectionName,
		&infoSection.Position,
		&infoSection.IsMandatory,
		&infoSection.Color,
		&infoSection.MaxTasks,
	)

	if err != nil {
		return dto.FullSectionInfo{}, fmt.Errorf("QueryRow: %w", err)
	}

	return infoSection, nil
}

func (r *Repository) DeleteSection(ctx context.Context, linksSection uuid.UUID) error {
	query := `
		WITH target_info AS (
			SELECT board_link FROM section WHERE section_link = $1
		),
		backlog_info AS (
			SELECT s.section_link 
			FROM section s
			JOIN section_version v ON s.section_link = v.section_link
			WHERE s.board_link = (SELECT board_link FROM target_info)
			  AND v.position = 1
			  AND v.valid_to IS NULL
			  AND s.deleted_at IS NULL
		),
		move_tasks AS (
            UPDATE task_actual 
            SET section_link = (SELECT section_link FROM backlog_info)
            WHERE section_link = $1
              AND (SELECT section_link FROM backlog_info) IS NOT NULL
        )
		UPDATE section 
		SET deleted_at = NOW() 
		WHERE section_link = $1 
		  AND deleted_at IS NULL
	`

	commandTag, err := r.pool.Exec(ctx, query, linksSection)
	if err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return common.ErrorNotExistingSection
	}
	return nil
}

func (r *Repository) ReorderSection(ctx context.Context, linkBoard uuid.UUID, linksSection []uuid.UUID) error {
	var values []string
	var args []any
	argID := 1

	for i, link := range linksSection {
		values = append(values, fmt.Sprintf("($%d::uuid, $%d::int)", argID, argID+1))
		args = append(args, link, i+1)
		argID += 2
	}

	valuesString := strings.Join(values, ", ")

	query := fmt.Sprintf(`
		UPDATE section_actual sa
		SET position = data.new_pos
		FROM (VALUES %s) AS data(section_link, new_pos)
		WHERE sa.section_link = data.section_link
		  AND sa.board_link = $%d;
	`, valuesString, argID)

	args = append(args, linkBoard)

	commandTag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}

	if commandTag.RowsAffected() != int64(len(linksSection)) {
		return common.ErrorNotFindAllLinks
	}

	return nil
}

func (r *Repository) UpdateSection(ctx context.Context, updatingSection dto.FullSectionInfo) error {
	query := `
	UPDATE section_actual 
	SET section_name = $1, 
		is_mandatory = $2, 
		color = $3, 
		max_tasks = $4
	WHERE section_link = $5
	`

	commandTag, err := r.pool.Exec(ctx, query, updatingSection.SectionName, updatingSection.IsMandatory,
		updatingSection.Color, updatingSection.MaxTasks, updatingSection.SectionLink)
	if err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return common.ErrorNotExistingSection
	}

	return nil
}

func (r *Repository) GetAllSections(ctx context.Context, boarderLink uuid.UUID) ([]dto.FullSectionInfo, error) {
	query := `
		SELECT 
			section_link, 
			section_name, 
			position, 
			is_mandatory, 
			color, 
			max_tasks
		FROM section_actual
		WHERE board_link = $1
		ORDER BY position ASC;
	`

	rows, err := r.pool.Query(ctx, query, boarderLink)
	if err != nil {
		return []dto.FullSectionInfo{}, fmt.Errorf("pool.Query: %w", err)
	}

	defer rows.Close()

	sections := []dto.FullSectionInfo{}

	for rows.Next() {
		var section dto.FullSectionInfo

		err := rows.Scan(
			&section.SectionLink,
			&section.SectionName,
			&section.Position,
			&section.IsMandatory,
			&section.Color,
			&section.MaxTasks,
		)
		if err != nil {
			return []dto.FullSectionInfo{}, fmt.Errorf("rows.Scan: %w", err)
		}

		sections = append(sections, section)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return sections, nil
}
