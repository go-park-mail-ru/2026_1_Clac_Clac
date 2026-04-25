package service

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/common"
	"github.com/google/uuid"
)

type AppealRepository interface {
	GetUserRole(ctx context.Context, userLink uuid.UUID) (common.Role, error)
}

type Service struct {
	rep AppealRepository
}
