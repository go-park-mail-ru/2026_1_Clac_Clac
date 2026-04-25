package dto

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/common"
	"github.com/google/uuid"
)

type EntityAppealRequest struct {
	Mail        string
	Category    string
	Description string
	DisplayName string
}

type Appeal struct {
	AppealLink uuid.UUID
	AppelID    int
	Category   common.Category

	Description string
	CreatedAt   time.Time
}
type CardsSection struct {
	Cards []Appeal `json:"appeal"`
}
