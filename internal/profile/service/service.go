package service

import (
	"context"
	"fmt"

	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/service/dto"
	"github.com/google/uuid"
)

type ProfileRepository interface {
	GetProfile(ctx context.Context, link uuid.UUID) (repositoryDto.UserInfoEntity, error)
}

type Service struct {
	rep ProfileRepository
}

func NewService(rep ProfileRepository) *Service {
	return &Service{
		rep: rep,
	}
}

func (s *Service) GetProfileUser(ctx context.Context, userID uuid.UUID) (dto.UserInfo, error) {
	repositoryUser, err := s.rep.GetProfile(ctx, userID)
	if err != nil {
		return dto.UserInfo{}, fmt.Errorf("rep.GetProfile: %w", err)
	}

	user := dto.UserInfo{
		Link:        repositoryUser.Link,
		DisplayName: repositoryUser.DisplayName,
		Email:       repositoryUser.Email,
		Avatar:      repositoryUser.Avatar,
	}

	return user, nil
}
