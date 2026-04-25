package dto

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/common"
	"github.com/google/uuid"
)

type EntityAppealRequest struct {
	Mail        string          `json:"mail"`
	Category    common.Category `json:"category"`
	Description string          `json:"description"`
	DisplayName string          `json:"display_name"`
}

type Appeal struct {
	AppealLink  uuid.UUID       `json:"appeal_link"`
	AppealID    int             `json:"appeal_id,omitempty"`
	Category    common.Category `json:"category"`
	Status      common.Status   `json:"status"`
	Description string          `json:"description"`
	CreatedAt   time.Time       `json:"created_at"`
}

type CardsSection struct {
	Cards []Appeal `json:"appeals"`
}
