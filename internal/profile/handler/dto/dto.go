package dto

import "github.com/google/uuid"

type UserInfoResponse struct {
	Link            uuid.UUID `json:"link"`
	DisplayName     string    `json:"display_name"         example:"Ivan Ivanov"`
	DescriptionUser string    `json:"description_user"         example:"Hello! I am SEO"`
	Email           string    `json:"email"                example:"ivan@mail.com"`
	AvatarURL       string    `json:"avatar_url,omitempty" example:"https://example.com/avatar.jpg"`
}

type UpdatedInfo struct {
	DisplayName string `json:"display_name"         example:"Ivan Ivanov"`
	Avatar      string `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
}

type AvatarResponse struct {
	AvatarURL string `json:"avatar_url"  example:"https://example.com/images/avatars/user123.jpg"`
}
