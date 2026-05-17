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
	GetUserRoleByBoardLink(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID) (Role, error)
	GetUserRoleBySectionLink(ctx context.Context, sectionLink uuid.UUID, userLink uuid.UUID) (Role, uuid.UUID, error)
	GetUserRoleByCardLink(ctx context.Context, cardLink uuid.UUID, userLink uuid.UUID) (Role, uuid.UUID, error)
	GetUserRoleByCommentLink(ctx context.Context, commentLink uuid.UUID, userLink uuid.UUID) (Role, uuid.UUID, error)
	GetUserRoleBySubtaskLink(ctx context.Context, subtaskLink uuid.UUID, userLink uuid.UUID) (Role, uuid.UUID, error)
	GetUserRoleByAttachmentLink(ctx context.Context, attachmentLink uuid.UUID, userLink uuid.UUID) (Role, uuid.UUID, error)
}

type repository struct {
	pool DBEngine
}

func NewRepository(pool DBEngine) Repository {
	return &repository{
		pool: pool,
	}
}

func (r *repository) GetUserRoleByBoardLink(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID) (Role, error) {
	query := `
		SELECT level_member
		FROM member_board
		WHERE board_link = $1
			AND user_link = $2
			AND is_archive = false
	`

	row := r.pool.QueryRow(ctx, query, boardLink, userLink)

	var role Role
	err := row.Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Roles.None, nil
		}

		return Roles.None, fmt.Errorf("get user role on board: %w", err)
	}

	return role, nil
}

func (r *repository) GetUserRoleBySectionLink(ctx context.Context, sectionLink uuid.UUID, userLink uuid.UUID) (Role, uuid.UUID, error) {
	query := `
		SELECT m.level_member, s.board_link
		FROM member_board m
		JOIN section s ON m.board_link = s.board_link
		WHERE s.section_link = $1
			AND m.user_link = $2
			AND s.deleted_at IS NULL
			AND m.is_archive = false
	`

	row := r.pool.QueryRow(ctx, query, sectionLink, userLink)

	var role Role
	var boardLink uuid.UUID
	err := row.Scan(&role, &boardLink)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Roles.None, uuid.Nil, nil
		}

		return Roles.None, uuid.Nil, fmt.Errorf("get user role on section: %w", err)
	}

	return role, boardLink, nil
}

func (r *repository) GetUserRoleByCardLink(ctx context.Context, cardLink uuid.UUID, userLink uuid.UUID) (Role, uuid.UUID, error) {
	query := `
		SELECT m.level_member, s.board_link
		FROM member_board m
		JOIN section s ON m.board_link = s.board_link
		JOIN task_actual t ON s.section_link = t.section_link
		WHERE t.task_link = $1
		  AND m.user_link = $2
		  AND s.deleted_at IS NULL
		  AND m.is_archive = false
	`

	row := r.pool.QueryRow(ctx, query, cardLink, userLink)

	var role Role
	var boardLink uuid.UUID
	err := row.Scan(&role, &boardLink)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Roles.None, uuid.Nil, nil
		}

		return Roles.None, uuid.Nil, fmt.Errorf("get user role on card: %w", err)
	}

	return role, boardLink, nil
}

func (r *repository) GetUserRoleByCommentLink(ctx context.Context, commentLink uuid.UUID, userLink uuid.UUID) (Role, uuid.UUID, error) {
	query := `
		SELECT m.level_member, s.board_link
		FROM member_board m
		JOIN section s ON m.board_link = s.board_link
		JOIN task_actual t ON s.section_link = t.section_link
		JOIN comment_task c ON c.task_id = t.task_id
		WHERE c.link = $1
		  AND m.user_link = $2
		  AND s.deleted_at IS NULL
		  AND m.is_archive = false
	`

	row := r.pool.QueryRow(ctx, query, commentLink, userLink)

	var role Role
	var boardLink uuid.UUID
	err := row.Scan(&role, &boardLink)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Roles.None, uuid.Nil, nil
		}

		return Roles.None, uuid.Nil, fmt.Errorf("get user role on comment: %w", err)
	}

	return role, boardLink, nil
}

func (r *repository) GetUserRoleBySubtaskLink(ctx context.Context, subtaskLink uuid.UUID, userLink uuid.UUID) (Role, uuid.UUID, error) {
	query := `
		SELECT m.level_member, s.board_link
		FROM member_board m
		JOIN section s ON m.board_link = s.board_link
		JOIN task_actual t ON s.section_link = t.section_link
		JOIN subtask st ON st.task_link = t.task_link
		WHERE st.subtask_link = $1
		  AND m.user_link = $2
		  AND s.deleted_at IS NULL
		  AND m.is_archive = false
	`

	row := r.pool.QueryRow(ctx, query, subtaskLink, userLink)

	var role Role
	var boardLink uuid.UUID
	err := row.Scan(&role, &boardLink)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Roles.None, uuid.Nil, nil
		}

		return Roles.None, uuid.Nil, fmt.Errorf("get user role on subtask: %w", err)
	}

	return role, boardLink, nil
}

func (r *repository) GetUserRoleByAttachmentLink(ctx context.Context, attachmentLink uuid.UUID, userLink uuid.UUID) (Role, uuid.UUID, error) {
	query := `
		SELECT m.level_member, s.board_link
		FROM member_board m
		JOIN section s ON m.board_link = s.board_link
		JOIN task_actual t ON s.section_link = t.section_link
		JOIN attachment a ON a.task_link = t.task_link
		WHERE a.attachment_link = $1
		  AND m.user_link = $2
		  AND s.deleted_at IS NULL
		  AND m.is_archive = false
	`

	row := r.pool.QueryRow(ctx, query, attachmentLink, userLink)

	var role Role
	var boardLink uuid.UUID
	err := row.Scan(&role, &boardLink)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Roles.None, uuid.Nil, nil
		}

		return Roles.None, uuid.Nil, fmt.Errorf("get user role on attachement: %w", err)
	}

	return role, boardLink, nil
}
