package dto

import "github.com/google/uuid"

type UserInfoResponce struct {
	Link        uuid.UUID `json:"link"`
	DisplayName string    `json:"display_name"         example:"Ivan Ivanov"`
	Email       string    `json:"email"                example:"ivan@mail.com"`
	Avatar      string    `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
}
