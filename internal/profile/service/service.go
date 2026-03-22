package service

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/dto"
	"github.com/google/uuid"
)

type ProfileRepository interface {
	GetProfile(ctx context.Context, link uuid.UUID) (dto.UserInfoResponce, error)
}

type Service struct {
	rep ProfileRepository
}

func NewService(rep ProfileRepository) *Service {
	return &Service{
		rep: rep,
	}
}

func (s *Service) GetProfileUser(ctx context.Context, userID uuid.UUID) (dto.UserInfoResponce, error) {
	user, err := s.rep.GetProfile(ctx, userID)
	if err != nil {
		return dto.UserInfoResponce{}, fmt.Errorf("rep.GetProfile: %w", err)
	}

	return user, nil
}
