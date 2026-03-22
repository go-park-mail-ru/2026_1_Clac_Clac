package dto

import (
	"time"

	"github.com/google/uuid"
)

type RegistraionInfoRequest struct {
	Name     string
	Email    string
	Password string
}

type UserInfoResponce struct {
	Link        uuid.UUID `json:"link"`
	DisplayName string    `json:"display_name"         example:"Ivan Ivanov"`
	Email       string    `json:"email"                example:"ivan@mail.com"`
	Avatar      string    `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
}

type LoginInfoRequest struct {
	Email    string
	Password string
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
