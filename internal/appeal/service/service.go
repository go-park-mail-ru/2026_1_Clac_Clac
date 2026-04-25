package service

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/service/dto"
)

type AppealRepository interface {
}

type Service struct {
	rep AppealRepository
}

func (s *Service) CreateAppeal(ctx context.Context, appeal dto.EntityAppeal) error {

}
