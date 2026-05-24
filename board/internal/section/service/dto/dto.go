package dto

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/models"
	"github.com/google/uuid"
)

type SectionRequest struct {
	Link uuid.UUID
}

type FullSectionInfo struct {
	SectionLink uuid.UUID
	SectionName string
	Position    int
	IsMandatory bool
	Color       string
	MaxTasks    *int
}

type CreatingSection struct {
	BoardLink   uuid.UUID
	SectionName string
	IsMandatory bool
	Color       string
	MaxTasks    *int
}

type EntitySection struct {
	SectionLink uuid.UUID
	SectionName string
	Position    int
	IsMandatory bool
	Color       string
	MaxTasks    *int
}

type SectionsInfo struct {
	Sections []FullSectionInfo
}

type Card struct {
	CardLink     uuid.UUID
	ExecutorLink *uuid.UUID
	Title        string
	DeadLine     *time.Time
	Start        *time.Time
	Subtasks     []models.SubtaskInfo
	Position     int
	Status       bool
}
