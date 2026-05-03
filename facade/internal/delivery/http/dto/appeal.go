package dto

import (
	"time"

	"github.com/google/uuid"
)

type AppealInfo struct {
	AppealID      int64     `json:"appeal_id"`
	AppealLink    uuid.UUID `json:"appeal_link"`
	Email         string    `json:"email"`
	Category      string    `json:"category"`
	Status        string    `json:"status"`
	DisplayName   string    `json:"display_name"`
	Description   string    `json:"description"`
	AttachmentURL string    `json:"attachment_url"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateAppealRequest struct {
	Email       string `json:"email"`
	Category    string `json:"category"`
	Description string `json:"description"`
	DisplayName string `json:"display_name"`
}

type CreateAppealResponse struct {
	AppealLink uuid.UUID `json:"appeal_link"`
}

type GetAppealsResponse struct {
	Role    string       `json:"role"`
	Appeals []AppealInfo `json:"appeals"`
}

type UploadAttachmentInfo struct {
	AppealLink uuid.UUID `json:"appeal_link"`
	Filename   string    `json:"filename"`
}

type UploadAttachmentResponse struct {
	AttachmentURL string `json:"attachment_url"`
}

type DeleteInfo struct {
	AppealLink uuid.UUID `json:"appeal_link"`
}

type AppealsStats struct {
	OpenAppeals   int64 `json:"open_appeals"`
	InWorkAppeals int64 `json:"in_work_appeals"`
	CloseAppeals  int64 `json:"close_appeals"`
}

type ChangeAppealStatusInfo struct {
	AppealLink uuid.UUID `json:"appeal_link"`
	NewStatus  string    `json:"new_status"`
}
