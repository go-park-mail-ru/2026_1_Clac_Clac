package models

import "github.com/google/uuid"

// User описывает сущность пользователя в системе
//
//	@Name			auth.models.User
//	@Description	Полная информация о пользователе
type User struct {
	Link        uuid.UUID `json:"link"                   example:"123e4567-e89b-12d3-a456-426614174000"`
	DisplayName string    `json:"display_name"         example:"Ivan Ivanov"`
	// PasswordHash не отправляется клиенту
	PasswordHash string `json:"-"                    swaggerignore:"true"`
	Email        string `json:"email"                example:"ivan@mail.com"`
	// Поле имеет тег 'avatar', так оно и будет отображаться в JSON
	Avatar string `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
}

// Board представляет рабочую доску пользователя
//
//	@Name			auth.models.Board
//	@Description	Краткая информация о доске
type Board struct {
	ID uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
}

type Session struct {
}
