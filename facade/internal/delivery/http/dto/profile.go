package dto

import "github.com/google/uuid"

// ProfileResponse содержит полный профиль пользователя.
//
//	@Description	Полная информация о профиле пользователя
type ProfileResponse struct {
	Link        uuid.UUID `json:"link" example:"123e4567-e89b-12d3-a456-426614174000"`
	DisplayName string    `json:"display_name" example:"Ivan Ivanov"`
	Description string    `json:"description_user" example:"Люблю писать код на Go и проектировать микросервисы"`
	Email       string    `json:"email" example:"ivan.ivanov@example.com"`
	AvatarURL   string    `json:"avatar_url,omitempty" example:"https://storage.yoursite.com/avatars/123.jpg"`
}

// UpdateProfileRequest содержит данные для обновления профиля.
//
//	@Description	Данные для изменения текстовой информации профиля
type UpdateProfileRequest struct {
	DisplayName string `json:"display_name" example:"Ivan Ivanov"`
	Description string `json:"description_user" example:"Люблю писать код на Go и проектировать микросервисы"`
}

// AvatarResponse возвращает URL обновлённого аватара.
//
//	@Description	Ответ с новой ссылкой на успешно загруженный аватар
type AvatarResponse struct {
	AvatarURL string `json:"avatar_url" example:"https://storage.yoursite.com/avatars/123.jpg"`
}

// MeResponse содержит userLink и профиль пользователя.
//
//	@Description	Информация об авторизованном пользователе
type MeResponse struct {
	UserLink uuid.UUID       `json:"user_link" example:"123e4567-e89b-12d3-a456-426614174000"`
	Profile  ProfileResponse `json:"profile"`
}
