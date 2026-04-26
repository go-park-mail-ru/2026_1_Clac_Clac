package domain

import "github.com/google/uuid"

type User struct {
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

type NewUser struct {
	DisplayName      string
	Email            string
	Password         string
	RepeatedPassword string
}

type EntryUserInfo struct {
	Email    string
	Password string
}
