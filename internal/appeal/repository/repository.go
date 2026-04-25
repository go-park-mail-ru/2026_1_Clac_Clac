package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/common"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBEngine interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type Repository struct {
	pool DBEngine
}

func (r *Repository) GetUserRole(ctx context.Context, userLink uuid.UUID) (common.Role, error) {
	getUserRoleQuery := `
		SELECT level_member FROM member_board
		WHERE board_link = $1 AND user_link = $2;
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
