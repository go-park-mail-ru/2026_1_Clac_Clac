package models

import "github.com/google/uuid"

type LevelUser int

const (
	Viewer LevelUser = iota + 1
	Editor
	Admin
	Creater
)

// User описывает сущность пользователя в системе
//
// @Description Полная информация о пользователе
type User struct {
	ID          uuid.UUID `json:"id"                   example:"123e4567-e89b-12d3-a456-426614174000"`
	DisplayName string    `json:"display_name"         example:"Ivan Ivanov"`
	// PasswordHash не отправляется клиенту
	PasswordHash string `json:"-"                    swaggerignore:"true"`
	Email        string `json:"email"                example:"ivan@mail.com"`
	// Поле имеет тег 'avatar', так оно и будет отображаться в JSON
	Avatar *string `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
	Boards []Board `json:"boards"`
}

// Board представляет рабочую доску пользователя
//
// @Description Краткая информация о доске
type Board struct {
	ID uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
}

type MemberBoard struct {
	BoardID uuid.UUID `json:"board_id"`
	UserID  uuid.UUID `json:"user_id"`

	Level     LevelUser `json:"level"`
	IsLike    bool      `json:"is_like"`
	IsArchive bool      `json:"is_archive"`
}

type BoardTemplate struct {
	ID       uuid.UUID  `json:"id"`
	AuthorID *uuid.UUID `json:"author_id,omitempty"`

	BoardName   string `json:"board_name"`
	Description string `json:"description"`
	Background  string `json:"background"`

	TemplateName string `json:"template_name"`
}
