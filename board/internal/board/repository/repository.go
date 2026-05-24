package repository

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/repository/dto"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/boardRbac"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/s3"

	"github.com/jackc/pgerrcode"
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

type Config struct {
	CreateBoardDefaultUserRole string
}

type Repository struct {
	pool        DBEngine
	backgrounds s3.S3Bucket
	cnf         Config
}

func NewRepository(pool DBEngine, backgrounds s3.S3Bucket, cnf Config) *Repository {
	return &Repository{
		pool:        pool,
		backgrounds: backgrounds,
		cnf:         cnf,
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

	defer rows.Close()

	boards := make([]dto.BoardEntry, 0)
	for rows.Next() {
		var board dto.BoardEntry

		err := rows.Scan(&board.Link, &board.Name, &board.Description, &board.Background, &board.CreatedAt)
		if err != nil {
			return []dto.BoardEntry{}, fmt.Errorf("rows.Scan: %w", err)
		}

		boards = append(boards, board)
	}

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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.NotNullViolation:
				return dto.BoardEntry{}, common.ErrNotNullValue
			case pgerrcode.CheckViolation:
				return dto.BoardEntry{}, common.ErrInvalidBoardData
			case pgerrcode.ForeignKeyViolation:
				return dto.BoardEntry{}, common.ErrInvalidBoardReference
			}
		}

		return dto.BoardEntry{}, fmt.Errorf("create board version: %w", err)
	}

	createBoardMemberQuery := `
		INSERT INTO member_board (board_link, user_link, level_member)
        VALUES ($1, $2, $3::user_level)
    `
	_, err = tx.Exec(ctx, createBoardMemberQuery, boardLink, authorLink, r.cnf.CreateBoardDefaultUserRole)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.NotNullViolation:
				return dto.BoardEntry{}, common.ErrNotNullValue
			case pgerrcode.ForeignKeyViolation:
				return dto.BoardEntry{}, common.ErrInvalidBoardReference
			case pgerrcode.UniqueViolation:
				return dto.BoardEntry{}, common.ErrUserAlreadyMember
			}
		}

		return dto.BoardEntry{}, fmt.Errorf("create board member: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.NotNullViolation:
				return common.ErrNotNullValue
			case pgerrcode.CheckViolation:
				return common.ErrInvalidBoardData
			}
		}
		return fmt.Errorf("update board: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return common.ErrBoardNotFound
	}

	return nil
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

func (r *Repository) GetUsersOfBoard(ctx context.Context, boardLink uuid.UUID) ([]dto.MemberEntry, error) {
	getUsersOfBoardQuery := `SELECT user_link, level_member FROM member_board WHERE board_link = $1;`

	rows, err := r.pool.Query(ctx, getUsersOfBoardQuery, boardLink)
	if err != nil {
		return []dto.MemberEntry{}, fmt.Errorf("pool.Query: %w", err)
	}

	defer rows.Close()

	members := make([]dto.MemberEntry, 0)
	for rows.Next() {
		var member dto.MemberEntry

		err := rows.Scan(&member.Link, &member.Role)
		if err != nil {
			return []dto.MemberEntry{}, fmt.Errorf("rows.Scan: %w", err)
		}

		members = append(members, member)
	}

	if len(members) == 0 {
		return []dto.MemberEntry{}, common.ErrBoardNotFound
	}

	return members, nil
}

func (r *Repository) CreateInvite(ctx context.Context, inviteInfo dto.NewInviteInfo) (dto.InviteEntry, error) {
	createInviteQuery := `
		INSERT INTO invite (board_link, user_link, default_role, expire_time)
		VALUES ($1, $2, $3::user_level, $4)
		RETURNING invite_link, board_link, user_link, default_role, expire_time, status, created_at
	`

	row := r.pool.QueryRow(ctx, createInviteQuery,
		inviteInfo.BoardLink,
		inviteInfo.UserLink,
		inviteInfo.DefaultRole,
		inviteInfo.ExpireTime,
	)

	var entry dto.InviteEntry
	err := row.Scan(
		&entry.InviteLink,
		&entry.BoardLink,
		&entry.UserLink,
		&entry.DefaultRole,
		&entry.ExpireTime,
		&entry.Status,
		&entry.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.ForeignKeyViolation:
				return dto.InviteEntry{}, common.ErrInvalidBoardReference
			case pgerrcode.NotNullViolation:
				return dto.InviteEntry{}, common.ErrNotNullValue
			}
		}
		return dto.InviteEntry{}, fmt.Errorf("create invite: %w", err)
	}

	if entry.UserLink != nil && *entry.UserLink == uuid.Nil {
		entry.UserLink = nil
	}

	return entry, nil
}

func (r *Repository) GetInviteByLink(ctx context.Context, inviteLink uuid.UUID) (dto.InviteEntry, error) {
	getInviteQuery := `
		SELECT invite_link, board_link, user_link, default_role, expire_time, status, created_at
		FROM invite
		WHERE invite_link = $1
	`

	row := r.pool.QueryRow(ctx, getInviteQuery, inviteLink)

	var entry dto.InviteEntry
	err := row.Scan(
		&entry.InviteLink,
		&entry.BoardLink,
		&entry.UserLink,
		&entry.DefaultRole,
		&entry.ExpireTime,
		&entry.Status,
		&entry.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dto.InviteEntry{}, common.ErrInviteNotFound
		}
		return dto.InviteEntry{}, fmt.Errorf("get invite by link: %w", err)
	}

	if entry.UserLink != nil && *entry.UserLink == uuid.Nil {
		entry.UserLink = nil
	}

	return entry, nil
}

func (r *Repository) AddMemberToBoard(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID, role rbac.Role) error {
	addMemberQuery := `
		INSERT INTO member_board (board_link, user_link, level_member)
		VALUES ($1, $2, $3::user_level)
	`

	_, err := r.pool.Exec(ctx, addMemberQuery, boardLink, userLink, role)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return common.ErrUserAlreadyMember
			case pgerrcode.ForeignKeyViolation:
				return common.ErrInvalidBoardReference
			}
		}
		return fmt.Errorf("add member to board: %w", err)
	}

	return nil
}

func (r *Repository) CloseInvite(ctx context.Context, inviteLink uuid.UUID) error {
	closeInviteQuery := `UPDATE invite SET status = 'closed' WHERE invite_link = $1 AND status = 'active'`

	tag, err := r.pool.Exec(ctx, closeInviteQuery, inviteLink)
	if err != nil {
		return fmt.Errorf("close invite: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return common.ErrInviteNotFound
	}

	return nil
}

func (r *Repository) CloseInviteByBoardForUser(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID) error {
	closeInviteQuery := `
		UPDATE invite SET status = 'closed'
		WHERE board_link = $1 AND user_link = $2 AND status = 'active'
	`

	_, err := r.pool.Exec(ctx, closeInviteQuery, boardLink, userLink)
	if err != nil {
		return fmt.Errorf("close invite by board and user: %w", err)
	}

	return nil
}

func (r *Repository) UpdateMemberRole(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID, role rbac.Role) error {
	boardExists, err := r.boardExists(ctx, boardLink)
	if err != nil {
		return fmt.Errorf("check board exists: %w", err)
	}
	if !boardExists {
		return common.ErrBoardNotFound
	}

	updateRoleQuery := `
		UPDATE member_board SET level_member = $3::user_level
		WHERE board_link = $1 AND user_link = $2
	`

	tag, err := r.pool.Exec(ctx, updateRoleQuery, boardLink, userLink, role)
	if err != nil {
		return fmt.Errorf("update member role: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return common.ErrUserNotFound
	}

	return nil
}

func (r *Repository) RemoveMemberFromBoard(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID) error {
	boardExists, err := r.boardExists(ctx, boardLink)
	if err != nil {
		return fmt.Errorf("check board exists: %w", err)
	}
	if !boardExists {
		return common.ErrBoardNotFound
	}

	deleteMemberQuery := `DELETE FROM member_board WHERE board_link = $1 AND user_link = $2`

	tag, err := r.pool.Exec(ctx, deleteMemberQuery, boardLink, userLink)
	if err != nil {
		return fmt.Errorf("remove member from board: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return common.ErrUserNotFound
	}

	return nil
}

func (r *Repository) boardExists(ctx context.Context, boardLink uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM board WHERE link = $1)`
	err := r.pool.QueryRow(ctx, query, boardLink).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("query board exists: %w", err)
	}
	return exists, nil
}

func (r *Repository) GetActiveInvitesByBoard(ctx context.Context, boardLink uuid.UUID) ([]dto.InviteEntry, error) {
	getInvitesQuery := `
		SELECT invite_link, board_link, user_link, default_role, expire_time, status, created_at
		FROM invite
		WHERE board_link = $1 AND status = 'active'
		  AND (expire_time IS NULL OR expire_time > now())
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, getInvitesQuery, boardLink)
	if err != nil {
		return []dto.InviteEntry{}, fmt.Errorf("pool.Query: %w", err)
	}
	defer rows.Close()

	invites := make([]dto.InviteEntry, 0)
	for rows.Next() {
		var entry dto.InviteEntry
		err := rows.Scan(
			&entry.InviteLink,
			&entry.BoardLink,
			&entry.UserLink,
			&entry.DefaultRole,
			&entry.ExpireTime,
			&entry.Status,
			&entry.CreatedAt,
		)
		if err != nil {
			return []dto.InviteEntry{}, fmt.Errorf("rows.Scan: %w", err)
		}

		if entry.UserLink != nil && *entry.UserLink == uuid.Nil {
			entry.UserLink = nil
		}

		invites = append(invites, entry)
	}

	return invites, nil
}
