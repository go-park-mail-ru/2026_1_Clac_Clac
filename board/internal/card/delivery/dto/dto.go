package dto

import (
	"time"

	"github.com/google/uuid"
)

// InfoCard представляет данные карточки, которые возвращаются клиенту.
//
//	@Description	Полная информация о карточке задачи
type InfoCard struct {
	LinkCard     uuid.UUID  `json:"link_card" example:"123e4567-e89b-12d3-a456-426614174000"`
	Title        string     `json:"title" example:"Написать отчет"`
	Description  string     `json:"description" example:"Собрать метрики за Q3 и подготовить презентацию"`
	NameExecutor *string    `json:"name_executor" example:"Иван Иванов"`
	DataDeadLine *time.Time `json:"data_dead_line" example:"2026-04-15T15:04:05Z"`
	DataStart    *time.Time `json:"data_start" example:"2026-04-01T09:00:00Z"`
	Status       bool       `json:"status" example:"false"`
	Position     int        `json:"position" example:"2"`
}

// NewCard используется для запросов на создание новой карточки.
//
//	@Description	Данные для создания новой карточки
type NewCard struct {
	LinkAuthor   uuid.UUID  `json:"link_author" example:"550e8400-e29b-41d4-a716-446655440000"`
	Title        string     `json:"title" example:"Новая задача"`
	Description  string     `json:"description" example:"Описание для новой задачи"`
	LinkExecutor *uuid.UUID `json:"link_executor" example:"123e4567-e89b-12d3-a456-426614174000"`
	DataDeadLine *time.Time `json:"data_dead_line" example:"2026-05-01T12:00:00Z"`
	LinkSection  uuid.UUID  `json:"link_section" example:"987e6543-e21b-12d3-a456-426614174111"`
}

// UpdatingCardDetails используется для обновления текстовых данных и дедлайна карточки.
//
//	@Description	Данные для изменения свойств карточки
type UpdatingCardDetails struct {
	LinkCard     uuid.UUID  `json:"link_card" example:"123e4567-e89b-12d3-a456-426614174000"`
	Title        string     `json:"title" example:"Обновленный заголовок"`
	Description  string     `json:"description" example:"Дополненное описание задачи"`
	LinkExecutor *uuid.UUID `json:"link_executor" example:"550e8400-e29b-41d4-a716-446655440000"`
	DataDeadLine *time.Time `json:"data_dead_line" example:"2026-06-01T18:30:00Z"`
	DataStart    *time.Time `json:"data_start" example:"2026-04-01T09:00:00Z"`
}

// PlaceCard используется для изменения положения карточки на доске.
//
//	@Description	Данные о новом местоположении карточки (секция и позиция)
type PlaceCard struct {
	LinkCard    uuid.UUID `json:"link_card" example:"123e4567-e89b-12d3-a456-426614174000"`
	LinkSection uuid.UUID `json:"link_section" example:"987e6543-e21b-12d3-a456-426614174111"`
	Position    int       `json:"position" example:"2"`
}
