package dto

import (
	"time"

	"github.com/google/uuid"
)

type SessionEntity struct {
	SessionKey string
	UserLink   uuid.UUID
	LifeTime   time.Duration
}

type ExtendedSession struct {
	SessionKey string
	Expiration time.Duration
}
