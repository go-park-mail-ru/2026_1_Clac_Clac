package dto

import (
	"time"

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
	ExecuterName *string
	Title        string
	DeadLine     *time.Time
}
