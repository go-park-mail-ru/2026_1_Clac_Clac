package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"

	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/service/dto"
	"github.com/google/uuid"
)

//go:generate mockery --name=ProfileRepository --output=mock_profile_rep --outpkg=mockProfileRep
type ProfileRepository interface {
	GetProfile(ctx context.Context, userLink uuid.UUID) (repositoryDto.UserInfoEntity, error)
	UpdateProfile(ctx context.Context, updatedInfo repositoryDto.UpdatedInfo) error
	GetAvatarKey(ctx context.Context, userLink uuid.UUID) (string, error)
	UploadAvatarS3(ctx context.Context, file io.Reader, pathFile, contentType string) (string, error)
	UploadURLAvatar(ctx context.Context, userLink uuid.UUID, objectKey string) error
	DeleteAvatarS3(ctx context.Context, key string) error
	DeleteURLAvatar(ctx context.Context, userLink uuid.UUID) error
}

type Config struct {
	GenerateAvatarKey func() (string, error)
	BaseURLAvatar     string
}

type Service struct {
	rep ProfileRepository
	cnf Config
}

func NewService(rep ProfileRepository, cnf Config) *Service {
	return &Service{
		rep: rep,
		cnf: cnf,
	}
}

func (s *Service) GetProfileUser(ctx context.Context, userLink uuid.UUID) (dto.UserInfo, error) {
	repositoryUser, err := s.rep.GetProfile(ctx, userLink)
	if err != nil {
		return dto.UserInfo{}, fmt.Errorf("rep.GetProfile: %w", err)
	}

	var avatarUrl string
	if repositoryUser.AvatarKey != "" {
		avatarUrl, err = url.JoinPath(s.cnf.BaseURLAvatar, repositoryUser.AvatarKey)
		if err != nil {
			return dto.UserInfo{}, fmt.Errorf("url.JoinPath: %w", err)
		}
	}

	user := dto.UserInfo{
		Link:        repositoryUser.Link,
		DisplayName: repositoryUser.DisplayName,
		Description: repositoryUser.DescriptionUser,
		Email:       repositoryUser.Email,
		AvatarURL:   avatarUrl,
	}

	return user, nil
}

func (s *Service) UpdateProfile(ctx context.Context, updatedInfo dto.UpdatedUserInfo) error {
	err := s.rep.UpdateProfile(ctx, repositoryDto.UpdatedInfo{
		Link:            updatedInfo.Link,
		NameUser:        updatedInfo.DisplayName,
		DescriptionUser: updatedInfo.Description,
	})
	if err != nil {
		return fmt.Errorf("rep.UpdateProfile: %w", err)
	}

	return nil
}

func (s *Service) UpdateAvatar(ctx context.Context, avatar dto.UpdatedAvatar) (string, error) {
	var format string
	switch avatar.MimeType {
	case "image/jpg":
		format = ".jpg"
	case "image/jpeg":
		format = ".jpeg"
	case "image/png":
		format = ".png"
	case "image/webp":
		format = ".webp"
	}

	key, err := s.cnf.GenerateAvatarKey()
	if err != nil {
		return "", fmt.Errorf("cannot generate key: %w", err)
	}

	pathFile := fmt.Sprintf("%s/%s%s", avatar.UserLink.String(), key, format)

	objectKey, err := s.rep.UploadAvatarS3(ctx, avatar.File, pathFile, avatar.MimeType)
	if err != nil {
		return "", fmt.Errorf("UploadAvatar: %w", err)
	}

	errUploadDB := s.rep.UploadURLAvatar(ctx, avatar.UserLink, objectKey)
	if errUploadDB != nil {
		resultError := fmt.Errorf("rep.UploadAvatarURL: %w", errUploadDB)

		errDelete := s.rep.DeleteAvatarS3(ctx, objectKey)
		if errDelete != nil {
			resultError = errors.Join(resultError, errDelete)
		}

		return "", resultError
	}

	fullKey, err := url.JoinPath(s.cnf.BaseURLAvatar, objectKey)
	if err != nil {
		return "", fmt.Errorf("url.JoinPath: %w", err)
	}

	return fullKey, nil
}

func (s *Service) DeleteAvatar(ctx context.Context, userLink uuid.UUID) error {
	avatarKey, err := s.rep.GetAvatarKey(ctx, userLink)
	if err != nil {
		return fmt.Errorf("rep.GetAvatarKey: %w", err)
	}

	if avatarKey == "" {
		return nil
	}

	err = s.rep.DeleteAvatarS3(ctx, avatarKey)
	if err != nil {
		return fmt.Errorf("rep.DeleteAvatar: %w", err)
	}

	err = s.rep.DeleteURLAvatar(ctx, userLink)
	if err != nil {
		return fmt.Errorf("rep.DeleteAvatarURL: %w", err)
	}

	return nil
}
