package dto

import "github.com/google/uuid"

type UserInfoEntity struct {
	Link        uuid.UUID
	DisplayName string
	Email       string
	Avatar      string
}
