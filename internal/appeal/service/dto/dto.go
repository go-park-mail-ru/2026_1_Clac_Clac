package dto

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/common"
	"github.com/google/uuid"
)

type EntityAppeal struct {
	UserLink    uuid.UUID
	Mail        string
	Category    common.Category
	Description string
	DisplayName string
}

type Appeal struct {
	AppealLink  uuid.UUID
	AppelID     int
	Category    common.Category
	Description string
	CreatedAt   time.Time
}
type Appeals struct {
	Appeals []Appeal `json:"appeal"`
}
