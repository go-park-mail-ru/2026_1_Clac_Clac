package dto

import (
	"time"

	"github.com/google/uuid"
)

// ListSectionLink используется для передачи нового порядка секций на доске.
//
//	@Description	Массив UUID секций в нужном порядке
type ListSectionLink struct {
	List []uuid.UUID `json:"list_links" example:"123e4567-e89b-12d3-a456-426614174001,123e4567-e89b-12d3-a456-426614174002"`
}

// FullSectionInfo содержит полную информацию о секции (колонке).
//
//	@Description	Полные данные секции
type FullSectionInfo struct {
	SectionLink uuid.UUID `json:"section_link" example:"123e4567-e89b-12d3-a456-426614174000"`
	SectionName string    `json:"section_name" example:"To Do"`
	Position    int       `json:"position"     example:"1"`
	IsMandatory bool      `json:"is_mandatory" example:"true"`
	Color       string    `json:"color"        example:"red"`
	MaxTasks    *int      `json:"max_tasks"    example:"10"`
}

// CreatingSection используется для создания новой секции на доске.
//
//	@Description	Данные для создания новой секции
type CreatingSection struct {
	BoardLink   uuid.UUID `json:"board_link"   example:"550e8400-e29b-41d4-a716-446655440000"`
	SectionName string    `json:"section_name" example:"To Do"`
	IsMandatory bool      `json:"is_mandatory" example:"false"`
	Color       string    `json:"color"        example:"red"`
	MaxTasks    *int      `json:"max_tasks"    example:"10"`
}

// SectionsResponse используется для возврата списка всех секций доски.
//
//	@Description	Ответ, содержащий массив всех секций доски
type SectionsResponse struct {
	Sections []FullSectionInfo `json:"sections"`
}

// Card используется для представления краткой информации о карточке в списке.
//
//	@Description	Краткая информация о карточке задачи
type Card struct {
	CardLink      uuid.UUID  `json:"card_link" example:"123e4567-e89b-12d3-a456-426614174000"`
	ExecutorLink  *uuid.UUID `json:"executor_link" example:"123e4567-e89b-12d3-a456-426614174000"`
	Title         string     `json:"title" example:"Починить баг на фронтенде"`
	DeadLine      *time.Time `json:"dead_line" example:"2026-04-12T14:35:00Z"`
}

// CardsSection используется для возврата списка всех карточек конкретной секции.
//
//	@Description	Ответ, содержащий массив карточек секции
type CardsSection struct {
	Cards []Card `json:"cards"`
}
