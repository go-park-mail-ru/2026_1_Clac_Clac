package dto

import (
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
