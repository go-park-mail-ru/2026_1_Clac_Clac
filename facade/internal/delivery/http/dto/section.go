package dto

import (
	"time"

	"github.com/google/uuid"
)

type ListSectionLink struct {
	List []uuid.UUID `json:"list_links" example:"123e4567-e89b-12d3-a456-426614174001,123e4567-e89b-12d3-a456-426614174002"`
}

type SectionInfo struct {
	Link        uuid.UUID `json:"link" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name        string    `json:"name" example:"To Do"`
	Position    int64     `json:"position" example:"1"`
	IsMandatory bool      `json:"is_mandatory" example:"true"`
	Color       string    `json:"color" example:"red"`
	MaxTasks    *int64    `json:"max_tasks" example:"10"`
}

type CreateSectionRequest struct {
	BoardLink   uuid.UUID `json:"board_link" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string    `json:"name" example:"To Do"`
	IsMandatory bool      `json:"is_mandatory" example:"false"`
	Color       string    `json:"color" example:"red"`
	MaxTasks    *int64    `json:"max_tasks" example:"10"`
}

type SectionsResponse struct {
	Sections []SectionInfo `json:"sections"`
}

// Card represents card info in section.
//
//	@Description	Card info in section
type Card struct {
	Link          uuid.UUID     `json:"link" example:"123e4567-e89b-12d3-a456-426614174000"`
	ExecutorLink  *uuid.UUID    `json:"executor_link" example:"123e4567-e89b-12d3-a456-426614174000"`
	Title         string        `json:"title" example:"Fix bug on frontend"`
	Description   string        `json:"description" example:"Card description"`
	Deadline      *time.Time    `json:"deadline" example:"2026-04-12T14:35:00Z"`
	Subtasks      []SubtaskInfo `json:"subtasks"`
}

type SubtaskInfo struct {
	Link        uuid.UUID `json:"link" example:"123e4567-e89b-12d3-a456-426614174000"`
	Description string    `json:"description" example:"Subtask description"`
	IsDone      bool      `json:"is_done" example:"false"`
	Position    int64     `json:"position" example:"1"`
}

type CardsResponse struct {
	Cards []Card `json:"cards"`
}
