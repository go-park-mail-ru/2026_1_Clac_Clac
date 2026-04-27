package usecase

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/google/uuid"
)

type ProfileClient interface {
	GetProfile(ctx context.Context, userLink uuid.UUID) (domain.FullInfoUser, error)
	UpdateProfile(ctx context.Context, updatedInfo domain.UpdatedInfo) error
	UpdateAvatar(ctx context.Context, avatarInfo domain.AvatarInfo) (string, error)
	DeleteAvatar(ctx context.Context, userLink uuid.UUID) error
}

type Profile struct {
	user ProfileClient
}

func NewProfile(user ProfileClient) *Profile {
	return &Profile{
		user: user,
	}
}

func (p *Profile) GetProfile(ctx context.Context, userLink uuid.UUID) (domain.FullInfoUser, error) {
	user, err := p.user.GetProfile(ctx, userLink)
	if err != nil {
		return domain.FullInfoUser{}, fmt.Errorf("user.GetProfile: %w", err)
	}

	return user, nil
}

func (p *Profile) UpdateProfile(ctx context.Context, info domain.UpdatedInfo) error {
	err := p.user.UpdateProfile(ctx, info)
	if err != nil {
		return fmt.Errorf("user.UpdateProfile: %w", err)
	}

	return nil
}

func (p *Profile) UpdateAvatar(ctx context.Context, info domain.AvatarInfo) (string, error) {
	avatarURL, err := p.user.UpdateAvatar(ctx, info)
	if err != nil {
		return "", fmt.Errorf("user.UpdateAvatar: %w", err)
	}

	return avatarURL, nil
}

func (p *Profile) DeleteAvatar(ctx context.Context, userLink uuid.UUID) error {
	err := p.user.DeleteAvatar(ctx, userLink)
	if err != nil {
		return fmt.Errorf("user.DeleteAvatar: %w", err)
	}

	return nil
}
