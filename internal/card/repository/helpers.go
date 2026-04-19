package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

func checkLimitTasks(ctx context.Context, tx pgx.Tx, linkSection uuid.UUID) error {
	queryCheckLimits := `
		WITH count_tasks AS (
			SELECT COUNT(section_link) AS count
			FROM task_version
			WHERE section_link = $1 AND valid_to IS NULL
		)
		SELECT c.count, s.max_tasks
		FROM section s
		CROSS JOIN count_tasks c
		WHERE section_link = $1
		FOR UPDATE OF s
		`

	logger := zerolog.Ctx(ctx)

	var sizeSection int
	var maxTasks *int

	err := tx.QueryRow(ctx, queryCheckLimits, linkSection).Scan(&sizeSection, &maxTasks)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Info().Msg("Не нашёл секцию")
			return common.ErrorNotExistingSection
		}

		logger.Err(err).Msg(err.Error())
		return fmt.Errorf("tx.QueryRow: %w", err)
	}

	if maxTasks != nil && sizeSection+1 > *maxTasks {
		logger.Info().Msg("Достиггли лимита")
		return common.ErrorRichLimitTasks
	}

	return nil
}
