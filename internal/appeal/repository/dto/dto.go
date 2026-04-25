package dto

import (
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/common"
	"github.com/google/uuid"
)

type AppealEntry struct {
}

type CreateAppealInfo struct {
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
