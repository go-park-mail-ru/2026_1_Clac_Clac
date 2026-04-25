package dto

import (
	"time"

	"github.com/google/uuid"
)

type EntityAppeal struct {
	UserLink uuid.UUID
	Mail     string
	Category
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
type Appeals struct {
	Appeals []Appeal `json:"appeal"`
}
