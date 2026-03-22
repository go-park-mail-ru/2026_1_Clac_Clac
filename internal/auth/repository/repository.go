package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/dto"
	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/models"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/redis/go-redis/v9"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBEngine interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type RedisEngine interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

type Repository struct {
	pool   DBEngine
	client RedisEngine
}

func NewRepository(pool DBEngine, client RedisEngine) *Repository {
	return &Repository{
		pool:   pool,
		client: client,
	}
}

func (r *Repository) AddUser(ctx context.Context, user models.User) error {
	addUserQuery := `
		INSERT INTO "user" (link, display_name, password_hash, email, avatar)
		VALUES ($1, $2, $3, $4, $5) 
	`

	_, err := r.pool.Exec(ctx, addUserQuery,
		user.Link,
		user.DisplayName,
		user.PasswordHash,
		user.Email,
		user.Avatar)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == common.CodeUniqError {
				return common.ErrorExistingUser
			}

			return fmt.Errorf("pool.Exec: %w", err)
		}
	}

	return nil
}

func (r *Repository) AddSession(ctx context.Context, session dto.Session) error {
	key := fmt.Sprintf("session:%s", session.SessionID)

	err := r.client.Set(ctx, key, session.UserLink.String(), session.LifeTime).Err()
	if err != nil {
		return fmt.Errorf("client.Set: %w", err)
	}

	return nil
}

func (r *Repository) GetUserIDBySession(ctx context.Context, sessionID string) (string, error) {
	key := fmt.Sprintf("session:%s", sessionID)

	userLink, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", common.ErrorNotExistingSession
		}

		return "", fmt.Errorf("client.Get: %w", err)
	}

	return userLink, nil
}

func (r *Repository) DeleteSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("client.Del: %w", err)
	}

	return nil
}

func (r *Repository) GetUser(ctx context.Context, email string) (models.User, error) {
	getUserQuery := `
		SELECT link, display_name, password_hash, email, avatar
		FROM "user"
		WHERE email = $1
	`
	var user models.User
	err := r.pool.QueryRow(ctx, getUserQuery, email).Scan(
		&user.Link,
		&user.DisplayName,
		&user.PasswordHash,
		&user.Email,
		&user.Avatar,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, common.ErrorNonexistentEmail
		}
		return models.User{}, fmt.Errorf("pool.QueryRow: %w", err)
	}

	return user, nil
}

func (r *Repository) AddResetToken(ctx context.Context, token dto.ResetToken) error {
	key := fmt.Sprintf("reset_token:%s", token.ResetTokenID)

	err := r.client.Set(ctx, key, token.UserLink.String(), token.LifeTime).Err()
	if err != nil {
		return fmt.Errorf("client.Set: %w", err)
	}

	return nil
}

func (r *Repository) GetUserLinkByResetToken(ctx context.Context, tokenID string) (string, error) {
	key := fmt.Sprintf("reset_token:%s", tokenID)

	userLink, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", common.ErrorNotExistingResetToken
		}

		return "", fmt.Errorf("client.Get: %w", err)
	}

	return userLink, nil
}

func (r *Repository) DeleteResetToken(ctx context.Context, tokenID string) error {
	key := fmt.Sprintf("reset_token:%s", tokenID)

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("client.Del: %w", err)
	}

	return nil
}

func (r *Repository) UpdatePassword(ctx context.Context, link uuid.UUID, newPasswordHash string) error {
	updatePasswordQuery := `
	UPDATE "users" 
	SET password_hash = $1,
	updated_at = NOW()
	WHERE link = $2
	`

	countModifies, err := r.pool.Exec(ctx, updatePasswordQuery, newPasswordHash, link)
	if err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}

	if countModifies.RowsAffected() == 0 {
		return common.ErrorNonexistentUser
	}

	return nil
}
