package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/models"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/google/uuid"
)

var (
	ErrorNotExistingSession = errors.New("session not found or expired")
	ErrorSeesionExpired     = errors.New("time life session expired")
)

type DBEngine interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type Repository struct {
	pool DBEngine
}

func NewRepository(pool DBEngine) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) GetBoards(ctx context.Context, userLink uuid.UUID) ([]models.Board, error) {
	checkExistingQuery := `SELECT EXISTS(SELECT 1 FROM "user" WHERE link = $1)`
	var userExist bool
	err := r.pool.QueryRow(ctx, checkExistingQuery, userLink).Scan(&userExist)
	if err != nil {
		return []models.Board{}, fmt.Errorf("pool.QueryRow: %w", err)
	}

	if !userExist {
		return []models.Board{}, common.ErrorNonexistentUser
	}

	getBoardQuery := `
		SELECT b.link, b.created_at
		FROM board b
		JOIN member_board mb ON b.link = mb.board_link
		WHERE mb.user_link = $1
	`

	rows, err := r.pool.Query(ctx, getBoardQuery, userLink)

	boards := make([]models.Board, 0)
	for rows.Next() {
		var board models.Board

		err := rows.Scan(&board.Link, &board.Created_at)
		if err != nil {
			return []models.Board{}, fmt.Errorf("rows.Scan: %w", err)
		}

		boards = append(boards, board)
	}

	defer rows.Close()

	return boards, nil
}

func (r *Repository) AddEmptyBoard(ctx context.Context, emptyBoard models.Board, userLink uuid.UUID) error {
	addEmptyBoardQuery := `INSERT INTO board (link)
	VALUES ($1)`

	_, err := r.pool.Exec(ctx, addEmptyBoardQuery, emptyBoard.Link)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == common.CodeUniqError {
				return common.ErrorExistingBoard
			}

			return fmt.Errorf("pool.Exec: %w", err)
		}
	}

	addMemberBoardQuery := `INSERT INTO member_board (board_link, user_link)
	VALUES ($1, $2)`

	_, err = r.pool.Exec(ctx, addMemberBoardQuery, emptyBoard.Link, userLink)
	if err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}

	return nil
}
