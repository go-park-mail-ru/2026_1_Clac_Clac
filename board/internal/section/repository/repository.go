package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/models"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/repository/dto"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	msgInvalidUnmarshalSubtasks = "can not unmarshal subtasks"
)

type DBEngine interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Repository struct {
	pool DBEngine
}

func NewRepository(pool DBEngine) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) GetSection(ctx context.Context, link uuid.UUID) (dto.FullSectionInfo, error) {
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
			return dto.FullSectionInfo{}, common.ErrSectionNotFound
		}

		return dto.FullSectionInfo{}, fmt.Errorf("QueryRow: %w", err)
	}

	return infoSection, nil
}

type rawSubtask struct {
	SubtaskLink string `json:"subtask_link"`
	Description string `json:"description"`
	IsDone      bool   `json:"is_done"`
	Position    int    `json:"position"`
}

func (r *Repository) GetCards(ctx context.Context, linkSection uuid.UUID) ([]dto.Card, error) {
	query := `
	SELECT
		t.task_link,
		COALESCE(t.executer_link, '00000000-0000-0000-0000-000000000000'::uuid) as executer_link,
		t.title,
		t.due_date,
		t.position,
		t.start,
		t.status,
		t.points,
		(
			SELECT COALESCE(jsonb_agg(
				jsonb_build_object(
					'subtask_link', COALESCE(s.subtask_link, '00000000-0000-0000-0000-000000000000'::uuid),
                    'description', s.description,
                    'is_done', s.is_done,
                    'position', s.position
				)
			), '[]'::jsonb)
			FROM subtask s
			WHERE s.task_link = t.task_link
		) AS subtasks
	FROM task_actual AS t
	WHERE t.section_link = $1
	ORDER BY t.position ASC;
	`

	rows, err := r.pool.Query(ctx, query, linkSection)
	if err != nil {
		return []dto.Card{}, fmt.Errorf("pool.Query: %w", err)
	}

	defer rows.Close()

	cards := make([]dto.Card, 0)

	for rows.Next() {
		var card dto.Card
		var subtasks json.RawMessage
		var execLink uuid.UUID

		err := rows.Scan(
			&card.CardLink,
			&execLink,
			&card.Title,
			&card.DeadLine,
			&card.Position,
			&card.Start,
			&card.Status,
			&card.Points,
			&subtasks,
		)

		if err != nil {
			return []dto.Card{}, fmt.Errorf("rows.Scan: %w", err)
		}

		if execLink == uuid.Nil {
			card.ExecutorLink = nil
		} else {
			card.ExecutorLink = &execLink
		}

		var rawSubtasks []rawSubtask
		if err := json.Unmarshal(subtasks, &rawSubtasks); err != nil {
			return []dto.Card{}, fmt.Errorf(msgInvalidUnmarshalSubtasks)
		}

		card.Subtasks = make([]models.SubtaskInfo, 0, len(rawSubtasks))
		for _, rs := range rawSubtasks {
			link, _ := uuid.Parse(rs.SubtaskLink)
			card.Subtasks = append(card.Subtasks, models.SubtaskInfo{
				SubtaskLink: link,
				Description: rs.Description,
				IsDone:      rs.IsDone,
				Position:    rs.Position,
			})
		}

		cards = append(cards, card)
	}

	if rows.Err() != nil {
		return []dto.Card{}, rows.Err()
	}

	return cards, nil
}

func (r *Repository) CreateSection(ctx context.Context, newSection dto.CreatingSection) (dto.FullSectionInfo, error) {
	tx, err := r.pool.Begin(ctx)
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return dto.FullSectionInfo{}, common.ErrSectionAlreadyExists
			case pgerrcode.ForeignKeyViolation:
				return dto.FullSectionInfo{}, common.ErrInvalidReferenceSectionData
			case pgerrcode.NotNullViolation:
				return dto.FullSectionInfo{}, common.ErrMissingRequiredField
			}
		}
		return dto.FullSectionInfo{}, fmt.Errorf("tx.Exec section: %w", err)
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
		return dto.FullSectionInfo{}, fmt.Errorf("tx.QueryRow position: %w", err)
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.ForeignKeyViolation:
				return dto.FullSectionInfo{}, common.ErrInvalidReferenceSectionData
			case pgerrcode.CheckViolation:
				return dto.FullSectionInfo{}, common.ErrInvalidSectionData
			case pgerrcode.NotNullViolation:
				return dto.FullSectionInfo{}, common.ErrMissingRequiredField
			}
		}
		return dto.FullSectionInfo{}, fmt.Errorf("tx.Exec version: %w", err)
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
	tx, err := r.pool.Begin(ctx)
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
			return common.ErrSectionNotFound
		}
		return fmt.Errorf("tx.QueryRow check target section: %w", err)
	}

	if position == 1 {
		return common.ErrCannotDeleteBacklog
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
		return fmt.Errorf("tx.QueryRow backlog: %w", err)
	}

	queryDelete := `UPDATE section SET deleted_at = NOW() WHERE section_link = $1;`
	_, err = tx.Exec(ctx, queryDelete, linksSection)
	if err != nil {
		return fmt.Errorf("tx.Exec delete section: %w", err)
	}

	queryMoveTasks := `
		WITH closed_tasks AS (
			UPDATE task_version
			SET valid_to = NOW()
			WHERE section_link = $1 AND valid_to IS NULL
			RETURNING task_link, executer_link, title, description, due_date, start, status
		),
		backlog_max AS (
			SELECT COALESCE(MAX(position), 0) AS max_pos
			FROM task_version
			WHERE section_link = $2 AND valid_to IS NULL
		)
		INSERT INTO task_version (
			task_link, section_link, executer_link, title, description, position, due_date, start, status
		)
		SELECT
			ct.task_link,
			$2,
			ct.executer_link,
			ct.title,
			ct.description,
			bm.max_pos + ROW_NUMBER() OVER (),
			ct.due_date,
			ct.start,
			ct.status
		FROM closed_tasks ct CROSS JOIN backlog_max bm;
	`
	_, err = tx.Exec(ctx, queryMoveTasks, linksSection, backlogLink)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.ForeignKeyViolation:
				return common.ErrInvalidReferenceSectionData
			case pgerrcode.CheckViolation:
				return common.ErrInvalidCardData
			case pgerrcode.NotNullViolation:
				return common.ErrMissingRequiredField
			}
		}
		return fmt.Errorf("tx.Exec move tasks: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

func (r *Repository) ReorderSection(ctx context.Context, linkBoard uuid.UUID, linksSection []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.CheckViolation:
				return common.ErrInvalidSectionData
			case pgerrcode.NotNullViolation:
				return common.ErrMissingRequiredField
			case pgerrcode.ForeignKeyViolation:
				return common.ErrInvalidReferenceSectionData
			}
		}
		return fmt.Errorf("tx.Exec reorder: %w", err)
	}

	if commandTag.RowsAffected() != int64(len(linksSection)) {
		return common.ErrNotFindAllLinks
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

func (r *Repository) UpdateSection(ctx context.Context, updatingSection dto.FullSectionInfo) error {
	tx, err := r.pool.Begin(ctx)
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
			return common.ErrSectionNotFound
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.ForeignKeyViolation:
				return common.ErrInvalidReferenceSectionData
			case pgerrcode.CheckViolation:
				return common.ErrInvalidSectionData
			case pgerrcode.NotNullViolation:
				return common.ErrMissingRequiredField
			}
		}
		return fmt.Errorf("tx.Exec insert update: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

func (r *Repository) GetSections(ctx context.Context, boardLink uuid.UUID) ([]dto.FullSectionInfo, error) {
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

	rows, err := r.pool.Query(ctx, query, boardLink)
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
