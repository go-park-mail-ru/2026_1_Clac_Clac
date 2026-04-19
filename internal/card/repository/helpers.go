package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func checkLimitTasks(ctx context.Context, tx pgx.Tx, linkSection uuid.UUID) error {
	queryCheckLimits := `
		WITH locked_section AS (
			SELECT section_link
			FROM section
			WHERE section_link = $1 AND deleted_at IS NULL
			FOR UPDATE
		),
		count_tasks AS (
			SELECT COUNT(t.section_link) AS count
			FROM task_version t
			JOIN locked_section ls ON t.section_link = ls.section_link
			WHERE t.valid_to IS NULL
		)

		SELECT c.count, v.max_tasks
		FROM locked_section ls
		JOIN section_version v
			ON v.section_link = ls.section_link AND v.valid_to IS NULL
		CROSS JOIN count_tasks c
	`
	var sizeSection int
	var maxTasks *int

	err := tx.QueryRow(ctx, queryCheckLimits, linkSection).Scan(&sizeSection, &maxTasks)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return common.ErrorNotExistingSection
		}

		return fmt.Errorf("tx.QueryRow: %w", err)
	}

	if maxTasks != nil && sizeSection+1 > *maxTasks {
		return common.ErrorRichLimitTasks
	}

	return nil
}
