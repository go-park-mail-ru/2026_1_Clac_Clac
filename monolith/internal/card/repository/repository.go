package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	dto "github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/card/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/common"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog"
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

func (r *Repository) GetCard(ctx context.Context, linkCard uuid.UUID) (dto.InfoCard, error) {
	query := `
	SELECT
		t.title,
		t.description,
		t.due_date,
		u.display_name
	FROM task_actual AS t
	LEFT JOIN "user" u ON t.executer_link = u.link
	WHERE t.task_link = $1
	`

	var infoCard dto.InfoCard

	err := r.pool.QueryRow(ctx, query, linkCard).Scan(
		&infoCard.Title,
		&infoCard.Description,
		&infoCard.DataDeadLine,
		&infoCard.NameExecuter,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dto.InfoCard{}, common.ErrorNotExistingCard
		}

		return dto.InfoCard{}, fmt.Errorf("rep.QueryRow: %w", err)
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
		return common.ErrorNotExistingCard
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
	RETURNING section_link, position;
	`

	var oldSectionLink uuid.UUID
	var oldPosition int

	err = tx.QueryRow(ctx, queryClose, updatingCard.LinkCard).Scan(&oldSectionLink, &oldPosition)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return common.ErrorNotExistingCard
		}
		return fmt.Errorf("tx.QueryRow: %w", err)
	}

	queryInsert := `
		INSERT INTO task_version (
			task_link, section_link, executer_link, title, description, position, due_date
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`

	_, err = tx.Exec(ctx, queryInsert,
		updatingCard.LinkCard,
		oldSectionLink,
		updatingCard.LinkExecuter,
		updatingCard.Title,
		updatingCard.Description,
		oldPosition,
		updatingCard.DataDeadLine,
	)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			switch pgError.Code {
			case pgerrcode.ForeignKeyViolation:
				return common.ErrorInvalidReferenceCardData
			case pgerrcode.CheckViolation:
				return common.ErrorInvalidCardData
			case pgerrcode.NotNullViolation:
				return common.ErrorMissingRequiredField
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
			if errRollBack != nil {
				logger.Error().Msg("invalid rall back")
			}
		}
	}()

	queryClose := `
		UPDATE task_version
		SET valid_to = NOW()
		WHERE task_link = $1 AND valid_to IS NULL
		RETURNING section_link, position, title, description, executer_link, due_date;
	`
	var oldSectionLink uuid.UUID
	var position int
	var title, description string
	var executerLink *uuid.UUID
	var dueDate *time.Time

	err = tx.QueryRow(ctx, queryClose, updatingPlaceCard.LinkCard).Scan(
		&oldSectionLink, &position, &title, &description, &executerLink, &dueDate,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return common.ErrorNotExistingCard
		}

		return fmt.Errorf("tx.QueryRow: %w", err)
	}

	if updatingPlaceCard.LinkSection != oldSectionLink {
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
			return common.ErrorSkipMandatorySection
		}

		queryPos := `
			SELECT COALESCE(MAX(position), 0) + 1
			FROM task_version
			WHERE section_link = $1 AND valid_to IS NULL;
		`
		err = tx.QueryRow(ctx, queryPos, updatingPlaceCard.LinkSection).Scan(&position)
		if err != nil {
			return fmt.Errorf("tx.QueryRow: %w", err)
		}
	} else if updatingPlaceCard.Position != 0 {
		position = updatingPlaceCard.Position
	}

	queryInsert := `
		INSERT INTO task_version (
			task_link, section_link, position, title, description, executer_link, due_date
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`

	_, err = tx.Exec(ctx, queryInsert,
		updatingPlaceCard.LinkCard,
		updatingPlaceCard.LinkSection,
		position,
		title,
		description,
		executerLink,
		dueDate,
	)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			switch pgError.Code {
			case pgerrcode.ForeignKeyViolation:
				return common.ErrorInvalidReferenceCardData
			case pgerrcode.CheckViolation:
				return common.ErrorInvalidCardData
			case pgerrcode.NotNullViolation:
				return common.ErrorMissingRequiredField
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
				return -1, common.ErrorCardAlreadyExist
			case pgerrcode.NotNullViolation:
				return -1, common.ErrorMissingRequiredField
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
			task_link, section_link, executer_link, title, description, position, due_date
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`
	_, err = tx.Exec(ctx, queryVersion,
		newCard.LinkCard,
		newCard.LinkSection,
		newCard.LinkExecuter,
		newCard.Title,
		newCard.Description,
		position,
		newCard.DataDeadLine,
	)
	if err != nil {
		if strings.Contains(err.Error(), "fk_version_section") {
			return -1, common.ErrorNotExistingSection
		}

		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			if pgError.Code == pgerrcode.ForeignKeyViolation && pgError.ConstraintName == "fk_version_section" {
				return -1, common.ErrorNotExistingSection
			}
			switch pgError.Code {
			case pgerrcode.ForeignKeyViolation:
				return -1, common.ErrorInvalidReferenceCardData
			case pgerrcode.CheckViolation:
				return -1, common.ErrorInvalidCardData
			case pgerrcode.NotNullViolation:
				return -1, common.ErrorMissingRequiredField

			}

			return -1, fmt.Errorf("tx.Exec: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return -1, fmt.Errorf("tx.Commit: %w", err)
	}

	return position, nil
}
