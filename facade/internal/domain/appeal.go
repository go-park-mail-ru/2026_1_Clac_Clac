package domain

import (
	"time"

	"github.com/google/uuid"
)

type AppealInfo struct {
	AppealID      int64
	AppealLink    uuid.UUID
	Email         string
	Category      string
	Status        string
	Description   string
	DisplayName   string
	AttachmentURL string
	CreatedAt     time.Time
}

type CreateAppealInfo struct {
	UserLink    uuid.UUID
	Email       string
	Category    string
	Description string
	DisplayName string
}

type UploadAttachmentInfo struct {
	UserLink   uuid.UUID
	AppealLink uuid.UUID
	Filename   string
}

type DeleteInfo struct {
	UserLink   uuid.UUID
	AppealLink uuid.UUID
}

type AppealsStats struct {
	OpenAppeals   int64
	InWorkAppeals int64
	CloseAppeals  int64
}

type ChangeAppealStatusInfo struct {
	UserLink   uuid.UUID
	AppealLink uuid.UUID
	NewStatus  string
}
