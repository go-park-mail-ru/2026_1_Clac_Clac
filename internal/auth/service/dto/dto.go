package dto

import (
	"time"

	"github.com/google/uuid"
)

type UserInfo struct {
	Link        uuid.UUID
	DisplayName string
	Email       string
	Avatar      string
}

type RegistrationUser struct {
	DisplayName string
	Email       string
	Password    string
}

type LogInUser struct {
	Email    string
	Password string
}

type RateLimiterConfig struct {
	UserIP string
	Limit  int64
	Action string
	Window time.Duration
}

type CoolDownConfig struct {
	Name       string
	Email      string
	Expiration time.Duration
}
