package dto

import (
	"time"

	"github.com/google/uuid"
)

type UserInfoResponse struct {
	Link        uuid.UUID `json:"link"`
	DisplayName string    `json:"display_name"         example:"Ivan Ivanov"`
	Email       string    `json:"email"                example:"ivan@mail.com"`
	Avatar      string    `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
}

type Session struct {
	SessionID string
	UserLink  uuid.UUID
	LifeTime  time.Duration
}

type ResetToken struct {
	ResetTokenID string
	UserLink     uuid.UUID
	LifeTime     time.Duration
}
