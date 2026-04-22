package dto

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
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

func (r *RegistrationUser) Sanitize(sanitizer *bluemonday.Policy) {
	r.DisplayName = sanitizer.Sanitize(strings.TrimSpace(r.DisplayName))
	r.Email = sanitizer.Sanitize(strings.TrimSpace(r.Email))
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
