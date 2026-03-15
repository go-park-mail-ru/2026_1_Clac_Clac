package service

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/models"
	"github.com/google/uuid"
)

type Repository interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (models.User, error)
}

type Service struct {
	rep Repository
}

func NewProfileService(rep Repository) *Service {
	return &Service{
		rep: rep,
	}
}

func (pr *Service) GetProfileUser(ctx context.Context, userID uuid.UUID) (models.User, error) {
	user, err := pr.rep.GetProfile(ctx, userID)
	if err != nil {
		return models.User{}, fmt.Errorf("rep.GetProfile: %w", err)
	}

	return user, nil
}
