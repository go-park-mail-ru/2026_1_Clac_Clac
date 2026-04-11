package repository

import (
	"context"
	"errors"
	"fmt"

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
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Deps struct {
	Pool DBEngine
}

type Repository struct {
	deps Deps
}

func NewRepository(deps Deps) *Repository {
	return &Repository{
		deps: deps,
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
	err := r.deps.Pool.QueryRow(ctx, query, link).Scan(
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
	tx, err := r.deps.Pool.Begin(ctx)
	if err != nil {
		return dto.FullSectionInfo{}, fmt.Errorf("pool.Begin: %w", err)
	}

	defer tx.Rollback(ctx)

	querySection := `
		INSERT INTO section (section_link, board_link) 
		VALUES ($1, $2);
	`
	_, err = tx.Exec(ctx, querySection, newSection.SectionLink, newSection.BoardLink)
	if err != nil {
		return dto.FullSectionInfo{}, fmt.Errorf("tx.Exec: %w", err)
	}

	var position int
	queryPos := `
		SELECT COALESCE(MAX(v.position), 0) + 1 
		FROM section_version v
		JOIN section s ON s.section_link = v.section_link
		WHERE s.board_link = $1 AND v.valid_to IS NULL;
	`
	err = tx.QueryRow(ctx, queryPos, newSection.BoardLink).Scan(&position)
	if err != nil {
		return dto.FullSectionInfo{}, fmt.Errorf("tx.QueryRow: %w", err)
	}

	queryVersion := `
		INSERT INTO section_version (
			section_link, section_name, position, is_mandatory, color, max_tasks
		) 
		VALUES ($1, $2, $3, $4, $5, $6);
	`
	_, err = tx.Exec(ctx, queryVersion,
		newSection.SectionLink,
		newSection.SectionName,
		position,
		newSection.IsMandatory,
		newSection.Color,
		newSection.MaxTasks,
	)
	if err != nil {
		return dto.FullSectionInfo{}, fmt.Errorf("tx.Exec: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return dto.FullSectionInfo{}, fmt.Errorf("tx.Commit: %w", err)
	}

	infoSection := dto.FullSectionInfo{
		SectionLink: newSection.SectionLink,
		SectionName: newSection.SectionName,
		Position:    position,
		IsMandatory: newSection.IsMandatory,
		Color:       newSection.Color,
		MaxTasks:    newSection.MaxTasks,
	}

	return infoSection, nil
}

func (r *Repository) DeleteSection(ctx context.Context, linksSection uuid.UUID) error {
	tx, err := r.deps.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("pool.Begin: %w", err)
	}
	defer tx.Rollback(ctx)

	var boardLink uuid.UUID
	var position int
	queryCheck := `
		SELECT s.board_link, v.position
		FROM section s
		JOIN section_version v ON s.section_link = v.section_link
		WHERE s.section_link = $1 AND s.deleted_at IS NULL AND v.valid_to IS NULL;
	`
	err = tx.QueryRow(ctx, queryCheck, linksSection).Scan(&boardLink, &position)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return common.ErrorNotExistingSection
		}
		return fmt.Errorf("tx.QueryRow check target section: %w", err)
	}

	if position == 1 {
		return common.ErrorDeleteBacklog
	}

	var backlogLink uuid.UUID
	queryBacklog := `
		SELECT s.section_link
		FROM section s
		JOIN section_version v ON s.section_link = v.section_link
		WHERE s.board_link = $1 AND v.position = 1 
		  AND s.deleted_at IS NULL AND v.valid_to IS NULL;
	`
	err = tx.QueryRow(ctx, queryBacklog, boardLink).Scan(&backlogLink)
	if err != nil {
		return fmt.Errorf("tx.QueryRow: %w", err)
	}

	queryDelete := `UPDATE section SET deleted_at = NOW() WHERE section_link = $1;`
	_, err = tx.Exec(ctx, queryDelete, linksSection)
	if err != nil {
		return fmt.Errorf("tx.Exec: %w", err)
	}

	queryMoveTasks := `
		WITH closed_tasks AS (
			UPDATE task_version
			SET valid_to = NOW()
			WHERE section_link = $1 AND valid_to IS NULL
			RETURNING task_link, executer_link, title, description, due_date
		),
		backlog_max AS (
			SELECT COALESCE(MAX(position), 0) AS max_pos
			FROM task_version
			WHERE section_link = $2 AND valid_to IS NULL
		)
		INSERT INTO task_version (
			task_link, section_link, executer_link, title, description, position, due_date
		)
		SELECT
			ct.task_link,
			$2,
			ct.executer_link,
			ct.title,
			ct.description,
			bm.max_pos + ROW_NUMBER() OVER (),
			ct.due_date
		FROM closed_tasks ct CROSS JOIN backlog_max bm;
	`
	_, err = tx.Exec(ctx, queryMoveTasks, linksSection, backlogLink)
	if err != nil {
		return fmt.Errorf("tx.Exec: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

func (r *Repository) ReorderSection(ctx context.Context, linkBoard uuid.UUID, linksSection []uuid.UUID) error {
	tx, err := r.deps.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("pool.Begin: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		WITH new_positions AS (
			SELECT link::uuid AS section_link, ord::int AS new_pos
			FROM UNNEST($1::uuid[]) WITH ORDINALITY AS t(link, ord)
		),
		closed_versions AS (
			UPDATE section_version sv
			SET valid_to = NOW()
			FROM new_positions np
			JOIN section s ON s.section_link = np.section_link
			WHERE sv.section_link = np.section_link
			  AND sv.valid_to IS NULL
			  AND s.board_link = $2
			RETURNING sv.section_link, sv.section_name, sv.is_mandatory, sv.color, sv.max_tasks, np.new_pos
		)
		INSERT INTO section_version (section_link, section_name, position, is_mandatory, color, max_tasks)
		SELECT section_link, section_name, new_pos, is_mandatory, color, max_tasks
		FROM closed_versions;
	`

	commandTag, err := tx.Exec(ctx, query, linksSection, linkBoard)
	if err != nil {
		return fmt.Errorf("tx.Exec: %w", err)
	}

	if commandTag.RowsAffected() != int64(len(linksSection)) {
		return common.ErrorNotFindAllLinks
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

func (r *Repository) UpdateSection(ctx context.Context, updatingSection dto.FullSectionInfo) error {
	tx, err := r.deps.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("pool.Begin: %w", err)
	}
	defer tx.Rollback(ctx)

	queryClose := `
		UPDATE section_version
		SET valid_to = NOW()
		WHERE section_link = $1 AND valid_to IS NULL
		RETURNING position;
	`

	var position int
	err = tx.QueryRow(ctx, queryClose, updatingSection.SectionLink).Scan(&position)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return common.ErrorNotExistingSection
		}
		return fmt.Errorf("tx.QueryRow: %w", err)
	}

	queryInsert := `
		INSERT INTO section_version (section_link, section_name, position, is_mandatory, color, max_tasks)
		VALUES ($1, $2, $3, $4, $5, $6);
	`

	_, err = tx.Exec(ctx, queryInsert,
		updatingSection.SectionLink,
		updatingSection.SectionName,
		position,
		updatingSection.IsMandatory,
		updatingSection.Color,
		updatingSection.MaxTasks,
	)
	if err != nil {
		return fmt.Errorf("tx.Exec: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
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

	rows, err := r.deps.Pool.Query(ctx, query, boarderLink)
	if err != nil {
		return nil, fmt.Errorf("pool.Query: %w", err)
	}
	defer rows.Close()

	var sections []dto.FullSectionInfo

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
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		sections = append(sections, section)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return sections, nil
}
