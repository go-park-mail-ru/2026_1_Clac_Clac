package dto

import "github.com/google/uuid"

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
