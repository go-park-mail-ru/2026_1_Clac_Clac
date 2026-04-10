package dto

import (
	"time"

	"github.com/google/uuid"
)

type InfoCard struct {
	Description  string
	Title        string
	NameExecuter *string
	DataDeadLine *time.Time
}

type NewCard struct {
	LinkAuthor   uuid.UUID
	Title        string
	Description  string
	LinkExecuter *uuid.UUID
	DataDeadLine *time.Time
	LinkSection  uuid.UUID
}

type UpdatingCardDetails struct {
	LinkCard     uuid.UUID
	Title        string
	Description  string
	LinkExecuter *uuid.UUID
	DataDeadLine *time.Time
}

type PlaceCard struct {
	LinkCard    uuid.UUID
	LinkSection uuid.UUID
	Position    int
}
