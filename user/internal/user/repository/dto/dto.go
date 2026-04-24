package dto

import (
	"github.com/google/uuid"
)

type UserInitialize struct {
	Link         uuid.UUID
	DisplayName  string
	Email        string
	PasswordHash string
}

type UserEntity struct {
	Link         uuid.UUID
	DisplayName  string
	Email        string
	PasswordHash string
	Avatar       string
}

type UserInfoEntity struct {
	Link            uuid.UUID
	DisplayName     string
	DescriptionUser string
	Email           string
	AvatarKey       string
}

type UpdatedInfo struct {
	Link            uuid.UUID
	NameUser        string
	DescriptionUser string
}
