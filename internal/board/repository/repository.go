package repository

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/s3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog"

	"github.com/google/uuid"
)

type DBEngine interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type Repository struct {
	pool        DBEngine
	backgrounds s3.S3Bucket
}

func NewRepository(pool DBEngine, s3Client s3.S3Client, conf config.S3) *Repository {
	return &Repository{
		pool:        pool,
		backgrounds: s3Client.NewBucket(conf.BoardsBackgroundsBucket, conf.BoardsBackgroundsPrefix, s3.ACL.PublicRead),
	}
}

func (r *Repository) GetBoards(ctx context.Context, userLink uuid.UUID) ([]dto.BoardEntry, error) {
	getBoardsQuery := `
		SELECT b.link, b.name, b.description, b.background, b.created_at
		FROM board_actual b
		JOIN member_board mb ON b.link = mb.board_link
		WHERE mb.user_link = $1
	`

	rows, err := r.pool.Query(ctx, getBoardsQuery, userLink)
	if err != nil {
		return []dto.BoardEntry{}, fmt.Errorf("pool.Query: %w", err)
	}

	boards := make([]dto.BoardEntry, 0)
	for rows.Next() {
		var board dto.BoardEntry

		err := rows.Scan(&board.Link, &board.Name, &board.Description, &board.Background, &board.CreatedAt)
		if err != nil {
			return []dto.BoardEntry{}, fmt.Errorf("rows.Scan: %w", err)
		}

		boards = append(boards, board)
	}
	defer rows.Close()

	return boards, nil
}

func (r *Repository) GetBoard(ctx context.Context, boardLink uuid.UUID) (dto.BoardEntry, error) {
	getBoardQuery := `
		SELECT link, name, description, background, created_at
		FROM board_actual
		WHERE link = $1
	`

	row := r.pool.QueryRow(ctx, getBoardQuery, boardLink)

	var board dto.BoardEntry

	err := row.Scan(&board.Link, &board.Name, &board.Description, &board.Background, &board.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dto.BoardEntry{}, common.ErrBoardNotFound
		}

		return dto.BoardEntry{}, fmt.Errorf("pool.Query: %w", err)
	}

	return board, nil
}

func (r *Repository) CreateBoard(ctx context.Context, boardInfo dto.NewBoardInfo, authorLink uuid.UUID) (dto.BoardEntry, error) {
	logger := zerolog.Ctx(ctx)

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return dto.BoardEntry{}, fmt.Errorf("pool.BeginTx: %w", err)
	}
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback(ctx)

			if !errors.Is(rollbackErr, pgx.ErrTxClosed) {
				logger.Error().Err(rollbackErr).Msg("BoardRepository transaction rollback")
			}
		}
	}()

	var boardId int
	var boardLink uuid.UUID
	var createdAt time.Time

	createBoardQuery := `INSERT INTO board DEFAULT VALUES RETURNING board_id, link, created_at`
	err = tx.QueryRow(ctx, createBoardQuery).Scan(&boardId, &boardLink, &createdAt)
	if err != nil {
		return dto.BoardEntry{}, fmt.Errorf("create board: %w", err)
	}

	createBoardVersionQuery := `
		INSERT INTO board_version (board_id, board_name, description_board, url_path_background)
        VALUES ($1, $2, $3, $4)
    `
	_, err = tx.Exec(ctx, createBoardVersionQuery, boardId, boardInfo.Name, boardInfo.Description, boardInfo.Background)
	if err != nil {
		return dto.BoardEntry{}, fmt.Errorf("create board version: %w", err)
	}

	const defaultUserLevel = "creator"

	createBoardMemberQuery := `
		INSERT INTO member_board (board_link, user_link, level_member)
        VALUES ($1, $2, $3::user_level)
    `
	_, err = tx.Exec(ctx, createBoardMemberQuery, boardLink, authorLink, defaultUserLevel)
	if err != nil {
		return dto.BoardEntry{}, fmt.Errorf("create board member: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.BoardEntry{}, fmt.Errorf("create board commit transaction: %w", err)
	}

	return dto.BoardEntry{
		Link:        boardLink,
		Name:        boardInfo.Name,
		Description: boardInfo.Description,
		Background:  boardInfo.Background,
		CreatedAt:   createdAt,
	}, nil
}

func (r *Repository) DeleteBoard(ctx context.Context, boardLink uuid.UUID) error {
	deleteBoardQuery := `DELETE FROM board WHERE board.link = $1`

	tag, err := r.pool.Exec(ctx, deleteBoardQuery, boardLink)
	if err != nil {
		return fmt.Errorf("delete board: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return common.ErrBoardNotFound
	}

	return nil
}

func (r *Repository) UpdateBoard(ctx context.Context, boardInfo dto.UpdateBoardInfo) error {
	// На стороне БД простой замещающий триггер (он правда небольшой)
	updateBoardQuery := `
		UPDATE board_actual
		SET name = $1, description = $2, background = $3
		WHERE link = $4
	`

	tag, err := r.pool.Exec(ctx, updateBoardQuery, boardInfo.Name, boardInfo.Description, boardInfo.Background, boardInfo.Link)
	if err != nil {
		return fmt.Errorf("update board: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return common.ErrBoardNotFound
	}

	return nil
}

func (r *Repository) GetUserRoleOnBoard(ctx context.Context, userLink uuid.UUID, boardLink uuid.UUID) (common.Role, error) {
	getUserRoleQuery := `
		SELECT level_member FROM member_board
		WHERE board_link = $1 AND user_link = $2;
	`

	row := r.pool.QueryRow(ctx, getUserRoleQuery, boardLink, userLink)

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

func (r *Repository) UploadBackground(ctx context.Context, source io.Reader, filename string, contentType string) (string, error) {
	key, err := r.backgrounds.Put(ctx, source, filename, contentType)
	if err != nil {
		return "", fmt.Errorf("s3 service cannot upload background: %w", err)
	}

	return key, nil
}

func (r *Repository) UpdateBackground(ctx context.Context, background string, boardLink uuid.UUID) error {
	updateBoardQuery := `
		UPDATE board_actual
		SET background = $1
		WHERE link = $2
	`

	tag, err := r.pool.Exec(ctx, updateBoardQuery, background, boardLink)
	if err != nil {
		return fmt.Errorf("update board: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return common.ErrBoardNotFound
	}

	return nil
}
