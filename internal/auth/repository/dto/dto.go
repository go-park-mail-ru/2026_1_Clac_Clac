package dto

import (
	"time"

	"github.com/google/uuid"
)

type UserInitialize struct {
	Link         uuid.UUID
	DisplayName  string
	Email        string
	PasswordHash string
}

type UserEntity struct {
	Link         uuid.UUID
	DisplayName  string
	Email        string
	PasswordHash string
	Avatar       string
}

type SessionEntity struct {
	SessionID string
	UserLink  uuid.UUID
	LifeTime  time.Duration
}

type ResetTokenEntity struct {
	ResetTokenID string
	UserLink     uuid.UUID
	LifeTime     time.Duration
}

type RateLimiterConfig struct {
	UserIP string
	Action string
	Window time.Duration
}

type CoolDownConfig struct {
	Name       string
	Email      string
	Expiration time.Duration
}
