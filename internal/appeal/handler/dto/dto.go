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
	AppealID      int             `json:"appeal_id"`
	AppealLink    uuid.UUID       `json:"appeal_link"`
	Email         string          `json:"email"`
	DisplayName   string          `json:"display_name"`
	Status        common.Status   `json:"status"`
	Category      common.Category `json:"category"`
	Description   string          `json:"description"`
	AttachmentKey string          `json:"attachment_key"`
	CreatedAt     time.Time       `json:"created_at"`
}

type Appeals struct {
	Appeals []Appeal `json:"appeals"`
}
