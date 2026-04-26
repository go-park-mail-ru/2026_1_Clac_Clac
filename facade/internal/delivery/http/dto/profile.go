package dto

import "github.com/google/uuid"

// ProfileResponse содержит полный профиль пользователя.
type ProfileResponse struct {
	Link        uuid.UUID `json:"link"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description_user"`
	Email       string    `json:"email"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
}

// UpdateProfileRequest содержит данные для обновления профиля.
type UpdateProfileRequest struct {
	DisplayName string `json:"display_name"`
	Description string `json:"description_user"`
}

// AvatarResponse возвращает URL обновлённого аватара.
type AvatarResponse struct {
	AvatarURL string `json:"avatar_url"`
}
