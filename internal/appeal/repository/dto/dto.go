package dto

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/common"
	"github.com/google/uuid"
)

<<<<<<< HEAD
type AppealEntry struct {
	AppealLink    uuid.UUID
	Email         string
	DisplayName   string
	Status        common.Status
	Category      common.Category
	Description   string
	AttachmentKey string
	CreatedAt     time.Time
}

type CreateAppealInfo struct {
	UserLink      *uuid.UUID
	Email         string
	DisplayName   string
	Category      common.Category
	Description   string
	AttachmentKey string
=======
type AppealEntry struct{}

type CreateAppealInfo struct {
	UserLink    uuid.UUID
	Mail        string
	Category    string
	Description string
	DisplayName string
>>>>>>> 66ebeec (feat/update rep)
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
