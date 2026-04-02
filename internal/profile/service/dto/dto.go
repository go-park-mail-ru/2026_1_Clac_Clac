package dto

import "github.com/google/uuid"

type UserInfo struct {
	Link            uuid.UUID
	DisplayName     string
	DescriptionUser string
	Email           string
	AvatarURL       string
}
