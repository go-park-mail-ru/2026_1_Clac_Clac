package dto

import (
	"time"

	"github.com/google/uuid"
)

type ResetTokenEntity struct {
	ResetTokenKey string
	UserLink      uuid.UUID
	LifeTime      time.Duration
}
