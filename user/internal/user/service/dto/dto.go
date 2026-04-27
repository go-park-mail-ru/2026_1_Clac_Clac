package dto

import (
	"io"

	"github.com/google/uuid"
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
