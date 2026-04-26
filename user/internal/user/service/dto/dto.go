package dto

import (
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
)

type UserInfo struct {
	Link        uuid.UUID
	DisplayName string
	Description string
	Email       string
	AvatarURL   string
}

type EntityUser struct {
	DisplayName string
	Email       string
	Password    string
}

func (r *EntityUser) Sanitize(sanitizer *bluemonday.Policy) {
	r.DisplayName = sanitizer.Sanitize(strings.TrimSpace(r.DisplayName))
	r.Email = sanitizer.Sanitize(strings.TrimSpace(r.Email))
}

type GetUserInfo struct {
	Email    string
	Password string
}

type ResetPasswordInfo struct {
	UserLink    string
	NewPassword string
}

type UpdatedUserInfo struct {
	Link        uuid.UUID
	DisplayName string
	Description string
}

type UpdatedAvatar struct {
	UserLink uuid.UUID
	File     io.Reader
	MimeType string
}
