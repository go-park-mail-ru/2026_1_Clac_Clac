package dto

import "github.com/google/uuid"

// UserInfoResponse содержит публичную информацию о профиле пользователя.
// @Description Полные данные профиля пользователя, возвращаемые клиенту
type UserInfoResponse struct {
	Link            uuid.UUID `json:"link" example:"123e4567-e89b-12d3-a456-426614174000"`
	DisplayName     string    `json:"display_name" example:"Ivan Ivanov"`
	DescriptionUser string    `json:"description_user" example:"Hello! I am SEO"`
	Email           string    `json:"email" example:"ivan@mail.com"`
	AvatarURL       string    `json:"avatar_url,omitempty" example:"https://example.com/avatar.jpg"`
}

// UpdatedInfo используется для запросов на обновление текстовой информации профиля.
// @Description Данные для изменения имени и описания пользователя
type UpdatedInfo struct {
	DisplayName     string `json:"display_name" example:"Ivan Ivanov"`
	DescriptionUser string `json:"description_user" example:"Hello! I am SEO"`
}

// AvatarResponse содержит ссылку на обновленный аватар пользователя.
// @Description Ответ с новой ссылкой на загруженный аватар после успешного обновления
type AvatarResponse struct {
	AvatarURL string `json:"avatar_url" example:"https://example.com/images/avatars/user123.jpg"`
}
