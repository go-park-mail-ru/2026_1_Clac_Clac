package dto

import (
	"time"

	"github.com/google/uuid"
)

type FullSectionInfo struct {
	SectionLink uuid.UUID
	SectionName string
	Position    int
	IsMandatory bool
	Color       string
	MaxTasks    *int
}

type CreatingSection struct {
	SectionLink uuid.UUID
	BoardLink   uuid.UUID
	SectionName string
	IsMandatory bool
	Color       string
	MaxTasks    *int
}

type ListSectionLink struct {
	ListLinks []uuid.UUID
}

type Card struct {
	CardLink     uuid.UUID
	ExecutorName *string
	Title        string
	DeadLine     *time.Time
}
