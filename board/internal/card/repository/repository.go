package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/models"
	dto "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/s3"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog"
)

const (
	msgInvalidUnmarshalSubtasks    = "can not unmarshal subtasks"
	msgInvalidUnmarshalAttachments = "can not unmarshal attachments"
)

type Config struct {
	MaxAttachments  int
	MaxNestingDepth int
}

type DBEngine interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Repository struct {
	pool        DBEngine
	attachments s3.S3Bucket
	cfg         Config
}

func NewRepository(pool DBEngine, attachments s3.S3Bucket, cfg Config) *Repository {
	return &Repository{
		pool:        pool,
		attachments: attachments,
		cfg:         cfg,
	}
}

type rawSubtask struct {
	SubtaskLink string `json:"subtask_link"`
	Description string `json:"description"`
	IsDone      bool   `json:"is_done"`
	Position    int    `json:"position"`
}

type rawAttachment struct {
	AttachmentLink string `json:"attachment_link"`
	Name           string `json:"attachment_name"`
	Path           string `json:"attachment_path"`
	Position       int    `json:"position"`
}

func (r *Repository) GetCard(ctx context.Context, linkCard uuid.UUID) (dto.InfoCard, error) {
	query := `
	SELECT
		t.title,
		t.description,
		t.due_date,
		t.executer_link,
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
		) AS subtasks,
		(
			SELECT COALESCE(jsonb_agg(
				jsonb_build_object(
					'attachment_link', COALESCE(a.attachment_link, '00000000-0000-0000-0000-000000000000'::uuid),
					'attachment_path', a.attachment_path,
					'attachment_name', a.attachment_name,
					'position', a.position
				)
			), '[]'::jsonb)
			FROM attachment a
			WHERE a.task_link = t.task_link
		) AS attachments
	FROM task_actual AS t
	WHERE t.task_link = $1;
	`

	var infoCard dto.InfoCard
	var subtasks []byte
	var attachments []byte

	err := r.pool.QueryRow(ctx, query, linkCard).Scan(
		&infoCard.Title,
		&infoCard.Description,
		&infoCard.DataDeadLine,
		&infoCard.ExecutorLink,
		&infoCard.Position,
		&infoCard.DataStart,
		&infoCard.Status,
		&infoCard.Points,
		&subtasks,
		&attachments,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dto.InfoCard{}, common.ErrCardNotFound
		}

		return dto.InfoCard{}, fmt.Errorf("rep.QueryRow: %w", err)
	}

	var rawSubtasks []rawSubtask
	if err := json.Unmarshal(subtasks, &rawSubtasks); err != nil {
		return dto.InfoCard{}, errors.New(msgInvalidUnmarshalSubtasks)
	}

	var rawAttachments []rawAttachment
	if err := json.Unmarshal(attachments, &rawAttachments); err != nil {
		return dto.InfoCard{}, errors.New(msgInvalidUnmarshalAttachments)
	}

	infoCard.Subtasks = make([]models.SubtaskInfo, 0, len(rawSubtasks))
	for _, rs := range rawSubtasks {
		link, err := uuid.Parse(rs.SubtaskLink)
		if err != nil {
			return dto.InfoCard{}, fmt.Errorf("SubtaskLink uuid.Parse: %w", err)
		}

		infoCard.Subtasks = append(infoCard.Subtasks, models.SubtaskInfo{
			SubtaskLink: link,
			Description: rs.Description,
			IsDone:      rs.IsDone,
			Position:    rs.Position,
		})
	}

	infoCard.Attachments = make([]models.AttachmentInfo, 0, len(rawAttachments))
	for _, ra := range rawAttachments {
		link, err := uuid.Parse(ra.AttachmentLink)
		if err != nil {
			return dto.InfoCard{}, fmt.Errorf("AttachmentLink uuid.Parse: %w", err)
		}
		infoCard.Attachments = append(infoCard.Attachments, models.AttachmentInfo{
			AttachmentLink: link,
			Path:           ra.Path,
			Name:           ra.Name,
			Position:       ra.Position,
		})
	}

	return infoCard, nil
}

func (r *Repository) DeleteCard(ctx context.Context, linkCard uuid.UUID) error {
	query := `DELETE FROM task WHERE task_link = $1;`

	commandTag, err := r.pool.Exec(ctx, query, linkCard)
	if err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return common.ErrCardNotFound
	}

	return nil
}

func (r *Repository) UpdateCardDetails(ctx context.Context, updatingCard dto.UpdatingCardDetails) (err error) {
	logger := zerolog.Ctx(ctx)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("pool.Begin: %w", err)
	}

	defer func() {
		if err != nil {
			errRollBack := tx.Rollback(ctx)
			if errRollBack != nil {
				logger.Error().Msg("invalid rall back")
			}
		}
	}()

	queryClose := `
	UPDATE task_version
	SET valid_to = NOW()
	WHERE task_link = $1 AND valid_to IS NULL
	RETURNING section_link, position, start, status;
	`

	var oldSectionLink uuid.UUID
	var oldPosition int
	var oldStart time.Time
	var oldStatus bool

	err = tx.QueryRow(ctx, queryClose, updatingCard.LinkCard).Scan(&oldSectionLink, &oldPosition, &oldStart, &oldStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return common.ErrCardNotFound
		}
		return fmt.Errorf("tx.QueryRow: %w", err)
	}

	dataStart := &oldStart
	if updatingCard.DataStart != nil {
		dataStart = updatingCard.DataStart
	}

	queryInsert := `
		INSERT INTO task_version (
			task_link, section_link, executer_link, title, description, position, due_date, start, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
	`

	_, err = tx.Exec(ctx, queryInsert,
		updatingCard.LinkCard,
		oldSectionLink,
		updatingCard.LinkExecutor,
		updatingCard.Title,
		updatingCard.Description,
		oldPosition,
		updatingCard.DataDeadLine,
		dataStart,
		oldStatus,
	)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			switch pgError.Code {
			case pgerrcode.ForeignKeyViolation:
				return common.ErrInvalidReferenceCardData
			case pgerrcode.CheckViolation:
				return common.ErrInvalidCardData
			case pgerrcode.NotNullViolation:
				return common.ErrMissingRequiredField
			}
		}

		return fmt.Errorf("tx.Exec: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

func (r *Repository) ReorderCard(ctx context.Context, updatingPlaceCard dto.PlaceCard) (err error) {
	logger := zerolog.Ctx(ctx)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("pool.Begin: %w", err)
	}

	defer func() {
		if err != nil {
			errRollBack := tx.Rollback(ctx)
			if errRollBack != nil && !errors.Is(errRollBack, pgx.ErrTxClosed) {
				logger.Error().Err(errRollBack).Msg("failed to rollback transaction")
			}
		}
	}()

	queryClose := `
		UPDATE task_version
		SET valid_to = NOW()
		WHERE task_link = $1 AND valid_to IS NULL
		RETURNING section_link, position, title, description, executer_link, due_date, start, status;
	`
	var oldSectionLink uuid.UUID
	var position int
	var title, description string
	var executorLink *uuid.UUID
	var dueDate *time.Time
	var oldStart time.Time
	var oldStatus bool

	err = tx.QueryRow(ctx, queryClose, updatingPlaceCard.LinkCard).Scan(
		&oldSectionLink, &position, &title, &description, &executorLink, &dueDate, &oldStart, &oldStatus,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return common.ErrCardNotFound
		}

		return fmt.Errorf("tx.QueryRow: %w", err)
	}

	if updatingPlaceCard.LinkSection != oldSectionLink {
		err := CheckTaskLimit(ctx, tx, updatingPlaceCard.LinkSection)
		if err != nil {
			return fmt.Errorf("CheckTaskLimit: %w", err)
		}

		queryCheckMandatory := `
			WITH positions AS (
				SELECT
					(SELECT position FROM section_actual WHERE section_link = $1) AS old_pos,
					(SELECT position FROM section_actual WHERE section_link = $2) AS new_pos,
					(SELECT board_link FROM section_actual WHERE section_link = $1) AS b_link
			)
			SELECT EXISTS (
				SELECT 1
				FROM section_actual sa
				CROSS JOIN positions p
				WHERE sa.board_link = p.b_link
				  AND sa.is_mandatory = true
				  AND p.old_pos < p.new_pos
				  AND sa.position > p.old_pos
				  AND sa.position < p.new_pos
			);
		`
		var hasMandatorySkipped bool
		err = tx.QueryRow(ctx, queryCheckMandatory, oldSectionLink, updatingPlaceCard.LinkSection).Scan(&hasMandatorySkipped)
		if err != nil {
			return fmt.Errorf("tx.QueryRow: %w", err)
		}

		if hasMandatorySkipped {
			return common.ErrCannotSkipMandatorySection
		}

		queryDownPos := `
			UPDATE task_version 
			SET position = position - 1
			WHERE section_link = $1 AND valid_to IS NULL AND position > $2;
		`
		_, err = tx.Exec(ctx, queryDownPos, oldSectionLink, position)
		if err != nil {
			return fmt.Errorf("tx.Exec: %w", err)
		}

		queryUpPos := `
			UPDATE task_version 
			SET position = position + 1
			WHERE section_link = $1 AND valid_to IS NULL AND position >= $2;
		`
		_, err = tx.Exec(ctx, queryUpPos, updatingPlaceCard.LinkSection, updatingPlaceCard.Position)
		if err != nil {
			return fmt.Errorf("tx.Exec: %w", err)
		}

		position = updatingPlaceCard.Position
	} else if updatingPlaceCard.Position != position {
		if updatingPlaceCard.Position > position {
			queryDownPos := `
				UPDATE task_version
				SET position = position - 1
				WHERE section_link = $1 AND valid_to IS NULL
				AND position > $2 AND position <= $3
			`
			_, err = tx.Exec(ctx, queryDownPos, updatingPlaceCard.LinkSection, position, updatingPlaceCard.Position)
			if err != nil {
				return fmt.Errorf("tx.Exec: %w", err)
			}
		} else {
			queryUpPos := `
				UPDATE task_version
				SET position = position + 1
				WHERE section_link = $1 AND valid_to IS NULL
				AND position < $2 AND position >= $3
			`
			_, err = tx.Exec(ctx, queryUpPos, updatingPlaceCard.LinkSection, position, updatingPlaceCard.Position)
			if err != nil {
				return fmt.Errorf("tx.Exec: %w", err)
			}
		}
	}

	position = updatingPlaceCard.Position

	queryInsert := `
		INSERT INTO task_version (
			task_link, section_link, position, title, description, executer_link, due_date, start, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
	`

	_, err = tx.Exec(ctx, queryInsert,
		updatingPlaceCard.LinkCard,
		updatingPlaceCard.LinkSection,
		position,
		title,
		description,
		executorLink,
		dueDate,
		&oldStart,
		oldStatus,
	)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			switch pgError.Code {
			case pgerrcode.ForeignKeyViolation:
				return common.ErrInvalidReferenceCardData
			case pgerrcode.CheckViolation:
				return common.ErrInvalidCardData
			case pgerrcode.NotNullViolation:
				return common.ErrMissingRequiredField
			}
		}
		return fmt.Errorf("tx.Exec: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

func (r *Repository) CreateCard(ctx context.Context, newCard dto.NewCard) (int, error) {
	logger := zerolog.Ctx(ctx)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return -1, fmt.Errorf("pool.Begin: %w", err)
	}
	defer func() {
		if err != nil {
			errRollBack := tx.Rollback(ctx)
			if errRollBack != nil {
				logger.Error().Msg("invalid rall back")
			}
		}
	}()

	err = CheckTaskLimit(ctx, tx, newCard.LinkSection)
	if err != nil {
		return -1, fmt.Errorf("CheckTaskLimit: %w", err)
	}

	queryTask := `
		INSERT INTO task (task_link, author_link)
		VALUES ($1, $2);
	`
	_, err = tx.Exec(ctx, queryTask, newCard.LinkCard, newCard.LinkAuthor)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			switch pgError.Code {
			case pgerrcode.UniqueViolation:
				return -1, common.ErrCardAlreadyExists
			case pgerrcode.NotNullViolation:
				return -1, common.ErrMissingRequiredField
			}
		}
		return -1, fmt.Errorf("tx.Exec insert task: %w", err)
	}

	queryLock := `
		SELECT 1
		FROM task_version
		WHERE section_link = $1 AND valid_to IS NULL
		FOR NO KEY UPDATE;
	`
	_, _ = tx.Exec(ctx, queryLock, newCard.LinkSection)

	queryPos := `
		SELECT COALESCE(MAX(position), 0) + 1
		FROM task_version
		WHERE section_link = $1 AND valid_to IS NULL;
	`
	var position int
	err = tx.QueryRow(ctx, queryPos, newCard.LinkSection).Scan(&position)
	if err != nil {
		return -1, fmt.Errorf("tx.QueryRow: %w", err)
	}
	queryVersion := `
		INSERT INTO task_version (
			task_link, section_link, executer_link, title, description, position, due_date, start, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, COALESCE($8, NOW()), $9);
	`
	_, err = tx.Exec(ctx, queryVersion,
		newCard.LinkCard,
		newCard.LinkSection,
		newCard.LinkExecutor,
		newCard.Title,
		newCard.Description,
		position,
		newCard.DataDeadLine,
		newCard.DataStart,
		false,
	)
	if err != nil {
		if strings.Contains(err.Error(), "fk_version_section") {
			return -1, common.ErrCardNotFound
		}

		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			if pgError.Code == pgerrcode.ForeignKeyViolation && pgError.ConstraintName == "fk_version_section" {
				return -1, common.ErrSectionNotFound
			}
			switch pgError.Code {
			case pgerrcode.ForeignKeyViolation:
				return -1, common.ErrInvalidReferenceCardData
			case pgerrcode.CheckViolation:
				return -1, common.ErrInvalidCardData
			case pgerrcode.NotNullViolation:
				return -1, common.ErrMissingRequiredField

			}

			return -1, fmt.Errorf("tx.Exec: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return -1, fmt.Errorf("tx.Commit: %w", err)
	}

	return position, nil
}

func (r *Repository) GetBoardLinkByCard(ctx context.Context, cardLink uuid.UUID) (uuid.UUID, error) {
	query := `
		SELECT s.board_link
		FROM section s
		JOIN task_actual t ON s.section_link = t.section_link
		WHERE t.task_link = $1 AND s.deleted_at IS NULL
	`

	var boardLink uuid.UUID
	err := r.pool.QueryRow(ctx, query, cardLink).Scan(&boardLink)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, common.ErrCardNotFound
		}
		return uuid.Nil, fmt.Errorf("pool.QueryRow: %w", err)
	}
	return boardLink, nil
}

func (r *Repository) GetBoardLinkByComment(ctx context.Context, commentLink uuid.UUID) (uuid.UUID, error) {
	query := `
		SELECT s.board_link
		FROM section s
		JOIN task_actual t ON s.section_link = t.section_link
		JOIN comment_task c ON c.task_id = t.task_id
		WHERE c.link = $1 AND s.deleted_at IS NULL
	`

	var boardLink uuid.UUID
	err := r.pool.QueryRow(ctx, query, commentLink).Scan(&boardLink)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, common.ErrCommentNotFound
		}
		return uuid.Nil, fmt.Errorf("pool.QueryRow: %w", err)
	}
	return boardLink, nil
}

func (r *Repository) GetComments(ctx context.Context, cardLink uuid.UUID) ([]dto.CommentInfo, error) {
	getCommentsQuery := `
		SELECT
			c.link AS comment_link,
			c.author_link,
			p.link AS parent_link,
			c.text,
			c.created_at
		FROM comment_task c
		LEFT JOIN comment_task p ON p.comment_id = c.parent_id
		JOIN task t ON t.task_id = c.task_id
		WHERE t.task_link = $1
		ORDER BY c.created_at ASC
	`

	rows, err := r.pool.Query(ctx, getCommentsQuery, cardLink)
	if err != nil {
		return []dto.CommentInfo{}, fmt.Errorf("pool.Query: %w", err)
	}

	defer rows.Close()

	comments := make([]dto.CommentInfo, 0)
	for rows.Next() {
		var comment dto.CommentInfo

		err := rows.Scan(&comment.Link, &comment.AuthorLink, &comment.ParentLink, &comment.Text, &comment.CreatedAt)
		if err != nil {
			return []dto.CommentInfo{}, fmt.Errorf("rows.Scan: %w", err)
		}

		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *Repository) CreateComment(ctx context.Context, createCardInfo dto.CreateCommentInfo) (dto.CommentInfo, error) {
	createCommentQuery := `
		INSERT INTO comment_task (link, task_id, parent_id, author_link, text)
		VALUES (
			$1,
			(SELECT task_id FROM task WHERE task_link = $2),
			(SELECT comment_id FROM comment_task WHERE link = $3),
			$4,
			$5
		)
	`

	_, err := r.pool.Exec(ctx, createCommentQuery,
		createCardInfo.CommentLink,
		createCardInfo.CardLink,
		createCardInfo.ParentLink,
		createCardInfo.AuthorLink,
		createCardInfo.Text,
	)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			switch pgError.Code {
			case pgerrcode.NotNullViolation:
				return dto.CommentInfo{}, common.ErrMissingRequiredField
			case pgerrcode.ForeignKeyViolation:
				return dto.CommentInfo{}, common.ErrInvalidReferenceCardData
			}
		}
		return dto.CommentInfo{}, fmt.Errorf("pool.Exec: %w", err)
	}

	return dto.CommentInfo{
		Link:       createCardInfo.CommentLink,
		ParentLink: createCardInfo.ParentLink,
		AuthorLink: createCardInfo.AuthorLink,
		Text:       createCardInfo.Text,
	}, nil
}

func (r *Repository) IsCommentAuthor(ctx context.Context, commentLink uuid.UUID, userLink uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM comment_task
			WHERE link = $1 AND author_link = $2
		)
	`

	var isAuthor bool
	err := r.pool.QueryRow(ctx, query, commentLink, userLink).Scan(&isAuthor)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, common.ErrCommentNotFound
		}

		return false, fmt.Errorf("pool.QueryRow: %w", err)
	}

	return isAuthor, nil
}

func (r *Repository) DeleteComment(ctx context.Context, commentLink uuid.UUID) error {
	query := `DELETE FROM comment_task WHERE link = $1`

	commandTag, err := r.pool.Exec(ctx, query, commentLink)
	if err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return common.ErrCommentNotFound
	}

	return nil
}

func (r *Repository) UpdateComment(ctx context.Context, updateCommentInfo dto.UpdateCommentInfo) error {
	query := `
		UPDATE comment_task
		SET
			text = $1,
			updated_at = now()
		WHERE link = $2;
	`

	commandTag, err := r.pool.Exec(ctx, query, updateCommentInfo.Text, updateCommentInfo.CommentLink)
	if err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return common.ErrCommentNotFound
	}

	return nil
}

func (r *Repository) CreateSubtask(ctx context.Context, createInfo dto.CreateSubtaskInfo) (models.SubtaskInfo, error) {
	logger := zerolog.Ctx(ctx)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.SubtaskInfo{}, fmt.Errorf("pool.Begin: %w", err)
	}
	defer func() {
		if err != nil {
			errRollBack := tx.Rollback(ctx)
			if errRollBack != nil {
				logger.Error().Msg("invalid rall back")
			}
		}
	}()

	queryPos := `
		SELECT COALESCE(MAX(position), 0) + 1
		FROM subtask
		WHERE task_link = $1
	`
	var position int
	err = tx.QueryRow(ctx, queryPos, createInfo.TaskLink).Scan(&position)
	if err != nil {
		return models.SubtaskInfo{}, fmt.Errorf("tx.QueryRow: %w", err)
	}

	query := `
		INSERT INTO subtask (task_link, subtask_link, description, position)
		VALUES ($1, $2, $3, $4)
		RETURNING subtask_link, is_done, position
	`
	var savedSubtaskLink uuid.UUID
	var isDone bool
	err = tx.QueryRow(ctx, query, createInfo.TaskLink, createInfo.SubtaskLink, createInfo.Description, position).Scan(&savedSubtaskLink, &isDone, &position)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			switch pgError.Code {
			case pgerrcode.NotNullViolation:
				return models.SubtaskInfo{}, common.ErrMissingRequiredField
			case pgerrcode.ForeignKeyViolation:
				return models.SubtaskInfo{}, common.ErrInvalidReferenceCardData
			}
		}
		return models.SubtaskInfo{}, fmt.Errorf("tx.QueryRow: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return models.SubtaskInfo{}, fmt.Errorf("tx.Commit: %w", err)
	}

	return models.SubtaskInfo{
		SubtaskLink: savedSubtaskLink,
		Description: createInfo.Description,
		IsDone:      isDone,
		Position:    position,
	}, nil
}

func (r *Repository) DeleteSubtask(ctx context.Context, deleteInfo dto.DeleteSubtask) error {
	query := `
		DELETE FROM subtask 
		WHERE subtask_link = $1
	`

	commandTag, err := r.pool.Exec(ctx, query, deleteInfo.SubTaskLink)
	if err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return common.ErrSubtaskNotFound
	}

	return nil
}

func (r *Repository) UpdateSubtask(ctx context.Context, updateInfo dto.UpdateSubtask) error {
	query := `
	UPDATE subtask
	SET description = $1,
	is_done = $2
	WHERE subtask_link = $3
	`

	commandTag, err := r.pool.Exec(ctx, query, updateInfo.Description, updateInfo.IsDone, updateInfo.SubTaskLink)
	if err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return common.ErrSubtaskNotFound
	}

	return nil
}

func (r *Repository) UploadAttachment(ctx context.Context, uploadInfo dto.UploadAttachment) (string, error) {
	key, err := r.attachments.Put(ctx, uploadInfo.Data, uploadInfo.FilePath, uploadInfo.ContentType)
	if err != nil {
		return "", fmt.Errorf("s3 service cannot upload attachment: %w", err)
	}

	return key, nil
}

func (r *Repository) CreateAttachment(ctx context.Context, createInfo dto.CreateAttachment) (models.AttachmentInfo, error) {
	logger := zerolog.Ctx(ctx)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.AttachmentInfo{}, fmt.Errorf("CreateAttachment pool.Begin: %w", err)
	}

	var errTx error

	defer func() {
		if errTx != nil {
			if errRollBack := tx.Rollback(ctx); errRollBack != nil {
				logger.Error().Err(errRollBack).Msg("invalid RollBack during create attachment")
			}
		}
	}()

	queryCount := `
		SELECT COUNT(*)
		FROM attachment
		WHERE task_link = $1
	`

	var count int

	errTx = tx.QueryRow(ctx, queryCount, createInfo.TaskLink).Scan(&count)
	if errTx != nil {
		return models.AttachmentInfo{}, fmt.Errorf("RepositoryCard tx.QueryRow count: %w", errTx)
	}

	if count >= r.cfg.MaxAttachments {
		errTx = common.ErrAttachmentLimitReached
		return models.AttachmentInfo{}, errTx
	}

	queryPos := `
		SELECT COALESCE(MAX(position), 0) + 1
		FROM attachment
		WHERE task_link = $1
	`

	var position int

	errTx = tx.QueryRow(ctx, queryPos, createInfo.TaskLink).Scan(&position)
	if errTx != nil {
		if errors.Is(errTx, pgx.ErrNoRows) {
			return models.AttachmentInfo{}, common.ErrCardNotFound
		}

		return models.AttachmentInfo{}, fmt.Errorf("RepositoryCard tx.QueryRow: %w", errTx)
	}

	createAttachmentQuery := `
	INSERT INTO attachment (attachment_link, task_link, attachment_name, attachment_path, position)
	VALUES ($1, $2, $3, $4, $5)
	`

	_, errTx = tx.Exec(ctx, createAttachmentQuery, createInfo.AttachmentLink, createInfo.TaskLink, createInfo.Name, createInfo.Key, position)
	if errTx != nil {
		var pgError *pgconn.PgError
		if errors.As(errTx, &pgError) {
			switch pgError.Code {
			case pgerrcode.NotNullViolation:
				return models.AttachmentInfo{}, common.ErrMissingRequiredField
			case pgerrcode.ForeignKeyViolation:
				return models.AttachmentInfo{}, common.ErrInvalidReferenceCardData
			}
		}
		return models.AttachmentInfo{}, fmt.Errorf("CreateAttachment tx.Exec: %w", errTx)
	}

	errTx = tx.Commit(ctx)
	if errTx != nil {
		return models.AttachmentInfo{}, fmt.Errorf("CreateAttachment tx.Commit: %w", errTx)
	}

	return models.AttachmentInfo{
		AttachmentLink: createInfo.AttachmentLink,
		Path:           createInfo.Key,
		Name:           createInfo.Name,
		Position:       position,
	}, nil
}

func (r *Repository) DeleteAttachmentFromDB(ctx context.Context, attachmentLink uuid.UUID) (string, error) {
	deleteAttachmentQuery := `
	DELETE FROM attachment
	WHERE attachment_link = $1
	RETURNING attachment_path
	`

	var key string
	err := r.pool.QueryRow(ctx, deleteAttachmentQuery, attachmentLink).Scan(&key)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", common.ErrAttachmentNotFound
		}

		return "", fmt.Errorf("CardRepository pool.Query: %w", err)
	}

	return key, nil
}

func (r *Repository) DeleteAttachmentFromS3(ctx context.Context, key string) error {
	err := r.attachments.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("attachments.Delete: %w", err)
	}

	return nil
}

func (r *Repository) UpdateStatusTask(ctx context.Context, updateInfo dto.UpdateStatusTask) (err error) {
	logger := zerolog.Ctx(ctx)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("pool.Begin: %w", err)
	}

	defer func() {
		if err != nil {
			errRollBack := tx.Rollback(ctx)
			if errRollBack != nil {
				logger.Error().Msg("invalid roll back")
			}
		}
	}()

	queryClose := `
		UPDATE task_version
		SET valid_to = NOW()
		WHERE task_link = $1 AND valid_to IS NULL
		RETURNING section_link, executer_link, title, description, position, due_date, start;
	`

	var sectionLink uuid.UUID
	var executorLink *uuid.UUID
	var title, description string
	var position int
	var dueDate *time.Time
	var oldStart time.Time

	err = tx.QueryRow(ctx, queryClose, updateInfo.TaskLink).Scan(
		&sectionLink, &executorLink, &title, &description, &position, &dueDate, &oldStart,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return common.ErrCardNotFound
		}
		return fmt.Errorf("tx.QueryRow: %w", err)
	}

	queryInsert := `
		INSERT INTO task_version (
			task_link, section_link, executer_link, title, description, position, due_date, start, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
	`

	_, err = tx.Exec(ctx, queryInsert,
		updateInfo.TaskLink,
		sectionLink,
		executorLink,
		title,
		description,
		position,
		dueDate,
		&oldStart,
		updateInfo.Status,
	)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			switch pgError.Code {
			case pgerrcode.ForeignKeyViolation:
				return common.ErrInvalidReferenceCardData
			case pgerrcode.CheckViolation:
				return common.ErrInvalidCardData
			case pgerrcode.NotNullViolation:
				return common.ErrMissingRequiredField
			}
		}

		return fmt.Errorf("tx.Exec: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

func (r *Repository) UpdateCardPoints(ctx context.Context, dto dto.UpdateCardPoints) error {
	query := `UPDATE task SET points = $1 WHERE task_link = $2`

	commandTag, err := r.pool.Exec(ctx, query, dto.Points, dto.CardLink)
	if err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return common.ErrCardNotFound
	}

	return nil
}

func (r *Repository) UpdateTimeLine(ctx context.Context, updateInfo dto.UpdateTimeLine) (err error) {
	logger := zerolog.Ctx(ctx)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("pool.Begin: %w", err)
	}

	defer func() {
		if err != nil {
			errRollBack := tx.Rollback(ctx)
			if errRollBack != nil {
				logger.Error().Msg("invalid roll back")
			}
		}
	}()

	queryClose := `
		UPDATE task_version
		SET valid_to = NOW()
		WHERE task_link = $1 AND valid_to IS NULL
		RETURNING section_link, executer_link, title, description, position, start, status;
	`

	var sectionLink uuid.UUID
	var executorLink *uuid.UUID
	var title, description string
	var position int
	var oldStart time.Time
	var oldStatus bool

	err = tx.QueryRow(ctx, queryClose, updateInfo.TaskLink).Scan(
		&sectionLink, &executorLink, &title, &description, &position, &oldStart, &oldStatus,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return common.ErrCardNotFound
		}
		return fmt.Errorf("tx.QueryRow: %w", err)
	}

	dataStart := &oldStart
	if updateInfo.Start != nil {
		dataStart = updateInfo.Start
	}

	queryInsert := `
		INSERT INTO task_version (
			task_link, section_link, executer_link, title, description, position, due_date, start, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
	`

	_, err = tx.Exec(ctx, queryInsert,
		updateInfo.TaskLink,
		sectionLink,
		executorLink,
		title,
		description,
		position,
		updateInfo.DeadLine,
		dataStart,
		oldStatus,
	)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			switch pgError.Code {
			case pgerrcode.ForeignKeyViolation:
				return common.ErrInvalidReferenceCardData
			case pgerrcode.CheckViolation:
				return common.ErrInvalidCardData
			case pgerrcode.NotNullViolation:
				return common.ErrMissingRequiredField
			}
		}

		return fmt.Errorf("tx.Exec: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}
