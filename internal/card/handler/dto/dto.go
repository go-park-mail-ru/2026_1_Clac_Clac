package dto

import (
	"time"

	"github.com/google/uuid"
)

type InfoCard struct {
	LinkCard     uuid.UUID  `json:"link_card"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	NameExecuter *string    `json:"name_executer"`
	DataDeadLine *time.Time `json:"data_dead_line"`
}

type NewCard struct {
	LinkAuthor   uuid.UUID  `json:"link_author"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	LinkExecuter *uuid.UUID `json:"link_executer"`
	DataDeadLine *time.Time `json:"data_dead_line"`
	LinkSection  uuid.UUID  `json:"link_section"`
}

type UpdatingCardDetails struct {
	LinkCard     uuid.UUID  `json:"link_card"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	LinkExecuter *uuid.UUID `json:"link_executer"`
	DataDeadLine *time.Time `json:"data_dead_line"`
}

type PlaceCard struct {
	LinkCard    uuid.UUID `json:"link_card"`
	LinkSection uuid.UUID `json:"link_section"`
	Position    int       `json:"position"`
}
