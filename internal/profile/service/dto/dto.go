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
