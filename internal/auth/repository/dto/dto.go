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
