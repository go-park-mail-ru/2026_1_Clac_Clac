package dto

import (
	"time"

	"github.com/google/uuid"
)

type EntityAppealRequest struct {
	Mail        string
	Category    string
	Description string
	DisplayName string
}

type Appeal struct {
	AppealLink  uuid.UUID
	AppelID     int
	Category    string
	Description string
	CreatedAt   time.Time
}
type CardsSection struct {
	Cards []Appeal `json:"appeal"`
}
