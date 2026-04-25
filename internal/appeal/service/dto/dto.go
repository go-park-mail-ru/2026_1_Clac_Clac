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
	AppelID       int
	AppealLink    uuid.UUID
	Email         string
	DisplayName   string
	Status        common.Status
	Category      common.Category
	Description   string
	AttachmentKey string
	CreatedAt     time.Time
}

type Appeals struct {
	Role    common.Role
	Appeals []Appeal
}

type AppealInfo struct {
	AppealLink    uuid.UUID
	Email         string
	DisplayName   string
	Status        common.Status
	Category      common.Category
	Description   string
	AttachmentKey string
	CreatedAt     time.Time
}

type ChangeAppealStatusInfo struct {
	SupporterLink uuid.UUID
	AppealLink    uuid.UUID
	Status        common.Status
}

type AppealStats struct {
	Open   int
	InWork int
	Close  int
}
