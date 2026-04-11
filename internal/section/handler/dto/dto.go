package dto

import "github.com/google/uuid"

type ListSectionLink struct {
	List []uuid.UUID `json:"list_links"`
}

type FullSectionInfo struct {
	SectionLink uuid.UUID `json:"section_link"`
	SectionName string    `json:"section_name" example:"To Do"`
	Position    int       `json:"position"     example:"1"`
	IsMandatory bool      `json:"is_mandatory" example:"true"`
	Color       string    `json:"color"        example:"red"`
	MaxTasks    *int      `json:"max_tasks"    example:"10"`
}

type CreatingSection struct {
	BoardLink   uuid.UUID `json:"board_link"`
	SectionName string    `json:"section_name" example:"To Do"`
	IsMandatory bool      `json:"is_mandatory" example:"true"`
	Color       string    `json:"color"        example:"red"`
	MaxTasks    *int      `json:"max_tasks"    example:"10"`
}

type SectionsResponse struct {
	Sections []FullSectionInfo `json:"sections"`
}
