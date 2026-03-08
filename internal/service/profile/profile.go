package profile

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	"github.com/google/uuid"
)

type ProfileRepository interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (models.User, error)
}

type ProfileService struct {
	rep ProfileRepository
}

func NewProfileService(rep ProfileRepository) *ProfileService {
	return &ProfileService{
		rep: rep,
	}
}

func (pr *ProfileService) GetProfileUser(ctx context.Context, userID uuid.UUID) (models.User, error) {
	user, err := pr.rep.GetProfile(ctx, userID)
	if err != nil {
		return models.User{}, fmt.Errorf("rep.GetProfile: %w", err)
	}

	return user, nil
}
