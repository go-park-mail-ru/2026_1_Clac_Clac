package domain

import (
	"time"

	"github.com/google/uuid"
)

type SectionInfo struct {
	Link        uuid.UUID
	Name        string
	Position    int64
	IsMandatory bool
	Color       string
	MaxTasks    *int64
}

type SubtaskInfo struct {
	SubtaskLink uuid.UUID
	Description string
	IsDone      bool
	Position    int64
}

type CardInfo struct {
	Link         uuid.UUID
	ExecutorName *string
	Title        string
	Deadline     *time.Time
	Subtasks     []SubtaskInfo
}

type GetSectionsRequest struct {
	UserLink  uuid.UUID
	BoardLink uuid.UUID
}

type GetSectionRequest struct {
	UserLink    uuid.UUID
	SectionLink uuid.UUID
}

type GetCardsRequest struct {
	UserLink    uuid.UUID
	SectionLink uuid.UUID
}

type CreateSectionRequest struct {
	UserLink    uuid.UUID
	BoardLink   uuid.UUID
	Name        string
	IsMandatory bool
	Color       string
	MaxTasks    *int64
}

type DeleteSectionRequest struct {
	UserLink    uuid.UUID
	SectionLink uuid.UUID
}

type ReorderSectionRequest struct {
	UserLink  uuid.UUID
	BoardLink uuid.UUID
	LinksList []uuid.UUID
}

type UpdateSectionRequest struct {
	UserLink    uuid.UUID
	SectionLink uuid.UUID
	Name        string
	IsMandatory bool
	Color       string
	MaxTasks    *int64
}