package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/repository/dto"
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

// mockery --name=RedisEngine --output=mock_redis_engine --outpkg=mockRedisEngine
type RedisEngine interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	Pipeline() redis.Pipeliner
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	TTL(ctx context.Context, key string) *redis.DurationCmd
}

type Deps struct {
	Pool        DBEngine
	RedisClient RedisEngine
}

type Repository struct {
	deps Deps
}

func NewRepository(deps Deps) *Repository {
	return &Repository{
		deps: deps,
	}
}

func (r *Repository) AddUser(ctx context.Context, user dto.UserInitialize) error {
	addUserQuery := `
		INSERT INTO "user" (link, display_name, password_hash, email)
		VALUES ($1, $2, $3, $4) 
	`

	_, err := r.deps.Pool.Exec(ctx, addUserQuery,
		user.Link,
		user.DisplayName,
		user.PasswordHash,
		user.Email,
	)

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

func (r *Repository) AddSession(ctx context.Context, session dto.SessionEntity) error {
	err := r.deps.RedisClient.Set(ctx, session.SessionKey, session.UserLink.String(), session.LifeTime).Err()
	if err != nil {
		return fmt.Errorf("client.Set: %w", err)
	}

	return nil
}

func (r *Repository) ExtendSession(ctx context.Context, session dto.ExtendedSession) error {
	err := r.deps.RedisClient.Expire(ctx, session.Key, session.Expiration).Err()
	if err != nil {
		return fmt.Errorf("redisClient.Expire: %w", err)
	}

	return nil
}

func (r *Repository) SetCooldown(ctx context.Context, config dto.CoolDownConfig) (bool, time.Duration, error) {
	isSet, err := r.deps.RedisClient.SetNX(ctx, config.Key, "", config.Expiration).Result()
	if err != nil {
		return false, 0, fmt.Errorf("redisClient.SetNX: %w", err)
	}

	if isSet {
		return true, 0, nil
	}

	ttl, err := r.deps.RedisClient.TTL(ctx, config.Key).Result()
	if err != nil {
		return false, 0, fmt.Errorf("redisClient.TTL: %w", err)
	}

	if ttl < 0 {
		ttl = 0
	}

	return false, ttl, nil
}

func (r *Repository) CheckLimit(ctx context.Context, configLimiter dto.RateLimiterConfig) (int64, error) {
	now := time.Now().UnixNano()
	bucket := now / configLimiter.Window.Nanoseconds()
	fullKey := fmt.Sprintf("rl:%s:%s:%d", configLimiter.Action, configLimiter.UserIP, bucket)

	pipe := r.deps.RedisClient.Pipeline()
	size := pipe.Incr(ctx, fullKey)
	pipe.Expire(ctx, fullKey, configLimiter.Window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("pipe.Exec: %w", err)
	}

	return size.Val(), nil
}

func (r *Repository) GetUserIDBySession(ctx context.Context, sessionKey string) (string, error) {
	userLink, err := r.deps.RedisClient.Get(ctx, sessionKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", common.ErrorNotExistingSession
		}

		return "", fmt.Errorf("client.Get: %w", err)
	}

	return userLink, nil
}

func (r *Repository) DeleteSession(ctx context.Context, sessionKey string) error {
	err := r.deps.RedisClient.Del(ctx, sessionKey).Err()
	if err != nil {
		return fmt.Errorf("client.Del: %w", err)
	}

	return nil
}

func (r *Repository) GetUser(ctx context.Context, email string) (dto.UserEntity, error) {
	getUserQuery := `
		SELECT link, display_name, password_hash, email, avatar
		FROM "user"
		WHERE email = $1
	`
	var user dto.UserEntity
	err := r.deps.Pool.QueryRow(ctx, getUserQuery, email).Scan(
		&user.Link,
		&user.DisplayName,
		&user.PasswordHash,
		&user.Email,
		&user.Avatar,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dto.UserEntity{}, common.ErrorNonexistentEmail
		}
		return dto.UserEntity{}, fmt.Errorf("pool.QueryRow: %w", err)
	}

	return user, nil
}

func (r *Repository) GetUserLink(ctx context.Context, email string) (uuid.UUID, error) {
	query := `
	SELECT link 
	FROM "user"
	WHERE email = $1`

	var link uuid.UUID

	err := r.deps.Pool.QueryRow(ctx, query, email).Scan(&link)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, common.ErrorNonexistentEmail
		}
		return uuid.Nil, fmt.Errorf("pool.QueryRow: %w", err)
	}

	return link, nil
}

func (r *Repository) AddResetToken(ctx context.Context, token dto.ResetTokenEntity) error {
	err := r.deps.RedisClient.Set(ctx, token.ResetTokenKey, token.UserLink.String(), token.LifeTime).Err()
	if err != nil {
		return fmt.Errorf("client.Set: %w", err)
	}

	return nil
}

func (r *Repository) GetUserLinkByResetToken(ctx context.Context, tokenKey string) (string, error) {
	userLink, err := r.deps.RedisClient.Get(ctx, tokenKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", common.ErrorNotExistingResetToken
		}

		return "", fmt.Errorf("client.Get: %w", err)
	}

	return userLink, nil
}

func (r *Repository) DeleteResetToken(ctx context.Context, tokenKey string) error {
	err := r.deps.RedisClient.Del(ctx, tokenKey).Err()
	if err != nil {
		return fmt.Errorf("client.Del: %w", err)
	}

	return nil
}

func (r *Repository) UpdatePassword(ctx context.Context, link uuid.UUID, newPasswordHash string) error {
	updatePasswordQuery := `
	UPDATE "user" 
	SET password_hash = $1,
	updated_at = NOW()
	WHERE link = $2
	`

	countModifies, err := r.deps.Pool.Exec(ctx, updatePasswordQuery, newPasswordHash, link)
	if err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}

	if countModifies.RowsAffected() == 0 {
		return common.ErrorNonexistentUser
	}

	return nil
}
