package domain

import "github.com/google/uuid"

type FullInfoUser struct {
	UserLink    uuid.UUID
	Email       string
	DisplayName string
	Description string
	AvatarURL   string
}

type UpdatedInfo struct {
	UserLink    uuid.UUID
	DisplayName string
	Description string
}

type AvatarInfo struct {
	UserLink      uuid.UUID
	FileData      []byte
	ContentType   string
	FileExtension string
}

type UpdatedPassoword struct {
	UserLink         uuid.UUID
	Password         string
	RepeatedPassword string
}

type NewCredentialsUser struct {
	DisplayName string
	Email       string
	Password    string
}

type Credentials struct {
	Email    string
	Password string
}

type UserInfo struct {
	Link        uuid.UUID
	DisplayName string
	Email       string
	Avatar      string
}
