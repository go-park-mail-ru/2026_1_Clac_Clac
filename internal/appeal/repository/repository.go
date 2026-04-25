package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/repository/dto"
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

func (r *Repository) CreateAppeal(ctx context.Context, info dto.CreateAppealInfo) (dto.AppealEntry, error) {
	return dto.AppealEntry{}, nil
}

func (r *Repository) GetUserAppeals(ctx context.Context, userLink uuid.UUID) ([]dto.AppealEntry, error) {
	appeals := make([]dto.AppealEntry, 0)
	return appeals, nil
}

func (r *Repository) GetOpenAppeals(ctx context.Context) ([]dto.AppealEntry, error) {
	appeals := make([]dto.AppealEntry, 0)
	return appeals, nil
}

func (r *Repository) DeleteAppeal(ctx context.Context, appealLink uuid.UUID) error {
	return nil
}

func (r *Repository) ChangeAppealStatus(ctx context.Context, info dto.ChangeAppealStatusInfo) error {
	return nil
}

func (r *Repository) GetUserRole(ctx context.Context, userLink uuid.UUID) (common.Role, error) {
	getUserRoleQuery := `
		SELECT support_link, role FROM support
		WHERE user_link = $1;
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
