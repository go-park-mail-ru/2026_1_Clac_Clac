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
	SessionKey string
	UserLink   uuid.UUID
	LifeTime   time.Duration
}

type ResetTokenEntity struct {
	ResetTokenKey string
	UserLink      uuid.UUID
	LifeTime      time.Duration
}

type RateLimiterConfig struct {
	UserIP string
	Action string
	Window time.Duration
}

type CoolDownConfig struct {
	Key        string
	Expiration time.Duration
}
