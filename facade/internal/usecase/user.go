package usecase

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/google/uuid"
)

type UserClient interface {
	GetUser(ctx context.Context, entryUser domain.Credentials) (domain.FullInfoUser, error)
	CreateUser(ctx context.Context, infoUser domain.NewCredentialsUser) (domain.FullInfoUser, error)
	ProcessUserWithVK(ctx context.Context, code, codeVerifier, state, deviceID string) (uuid.UUID, error)
	ResetPassword(ctx context.Context, updatedPassword domain.UpdatedPassword) error
	GetUserLink(ctx context.Context, email string) (uuid.UUID, error)

	GetProfile(ctx context.Context, userLink uuid.UUID) (domain.FullInfoUser, error)
	GetProfiles(ctx context.Context, links []uuid.UUID) ([]domain.FullInfoUser, error)
	UpdateProfile(ctx context.Context, updatedInfo domain.UpdatedInfo) error
	UpdateAvatar(ctx context.Context, avatarInfo domain.AvatarInfo) (string, error)
	DeleteAvatar(ctx context.Context, userLink uuid.UUID) error
}

type User struct {
	user UserClient
}

func NewUser(user UserClient) *User {
	return &User{
		user: user,
	}
}

func (u *User) ProcessUserWithVK(ctx context.Context, code, codeVerifier, state, deviceID string) (uuid.UUID, error) {
	userLink, err := u.user.ProcessUserWithVK(ctx, code, codeVerifier, state, deviceID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("user.ProcessUserWithVK: %w", err)
	}

	return userLink, nil
}

func (u *User) GetUser(ctx context.Context, cred domain.Credentials) (domain.FullInfoUser, error) {
	user, err := u.user.GetUser(ctx, cred)
	if err != nil {
		return domain.FullInfoUser{}, fmt.Errorf("user.GetUser: %w", err)
	}

	return domain.FullInfoUser{
		UserLink:    user.UserLink,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		AvatarURL:   user.AvatarURL,
	}, nil
}

func (u *User) CreateUser(ctx context.Context, cred domain.NewCredentialsUser) (domain.FullInfoUser, error) {
	user, err := u.user.CreateUser(ctx, cred)
	if err != nil {
		return domain.FullInfoUser{}, fmt.Errorf("user.CreateUser: %w", err)
	}

	return domain.FullInfoUser{
		UserLink:    user.UserLink,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		AvatarURL:   user.AvatarURL,
	}, nil
}

func (u *User) GetProfile(ctx context.Context, userLink uuid.UUID) (domain.FullInfoUser, error) {
	user, err := u.user.GetProfile(ctx, userLink)
	if err != nil {
		return domain.FullInfoUser{}, fmt.Errorf("user.GetProfile: %w", err)
	}

	return user, nil
}

func (u *User) GetProfiles(ctx context.Context, links []uuid.UUID) ([]domain.FullInfoUser, error) {
	profiles, err := u.user.GetProfiles(ctx, links)
	if err != nil {
		return nil, fmt.Errorf("user.GetProfiles: %w", err)
	}

	return profiles, nil
}

func (u *User) UpdateProfile(ctx context.Context, info domain.UpdatedInfo) error {
	err := u.user.UpdateProfile(ctx, info)
	if err != nil {
		return fmt.Errorf("user.UpdateProfile: %w", err)
	}

	return nil
}

func (u *User) UpdateAvatar(ctx context.Context, info domain.AvatarInfo) (string, error) {
	avatarURL, err := u.user.UpdateAvatar(ctx, info)
	if err != nil {
		return "", fmt.Errorf("user.UpdateAvatar: %w", err)
	}

	return avatarURL, nil
}

func (u *User) ResetPassword(ctx context.Context, updatedPassword domain.UpdatedPassword) error {
	err := u.user.ResetPassword(ctx, updatedPassword)
	if err != nil {
		return fmt.Errorf("user.ResetPassword: %w", err)
	}

	return nil
}

func (u *User) DeleteAvatar(ctx context.Context, userLink uuid.UUID) error {
	err := u.user.DeleteAvatar(ctx, userLink)
	if err != nil {
		return fmt.Errorf("user.DeleteAvatar: %w", err)
	}

	return nil
}

func (u *User) GetUserLink(ctx context.Context, email string) (uuid.UUID, error) {
	userLink, err := u.user.GetUserLink(ctx, email)
	if err != nil {
		return uuid.Nil, fmt.Errorf("user.GetUserLink: %w", err)
	}

	return userLink, nil
}
