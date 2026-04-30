package rbac

import (
	"context"
	"errors"
	"fmt"

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

type Repository interface {
	GetUserRoleByLink(ctx context.Context, userLink uuid.UUID) (Role, error)
}

type repository struct {
	pool DBEngine
}

func NewRepository(pool DBEngine) Repository {
	return &repository{
		pool: pool,
	}
}

func (r *repository) GetUserRoleByLink(ctx context.Context, userLink uuid.UUID) (Role, error) {
	getUserRoleQuery := `
		SELECT role FROM support
		WHERE user_link = $1;
	`
	row := r.pool.QueryRow(ctx, getUserRoleQuery, userLink)

	var role Role
	err := row.Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Roles.User, nil
		}

		return Roles.User, fmt.Errorf("support get user role: %w", err)
	}

	return role, nil
}
