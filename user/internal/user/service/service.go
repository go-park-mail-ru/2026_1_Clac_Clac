package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/url"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/common"
	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/user/repository/dto"
	dto "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/user/service/dto"
	"github.com/google/uuid"
)

const (
	SessiondIdKey = "session_id"
)

// mockery --name=AuthRepository --output=mock_auth_rep --outpkg=mockAuthRep
type AuthRepository interface {
	AddUser(ctx context.Context, user repositoryDto.UserInitialize) error
	GetUser(ctx context.Context, email string) (repositoryDto.UserEntity, error)
	GetUserLink(ctx context.Context, email string) (uuid.UUID, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, newPasswordHash string) error

	GetProfile(ctx context.Context, userLink uuid.UUID) (repositoryDto.UserInfoEntity, error)
	UpdateProfile(ctx context.Context, updatedInfo repositoryDto.UpdatedInfo) error
	GetAvatarKey(ctx context.Context, userLink uuid.UUID) (string, error)
	UploadAvatarS3(ctx context.Context, file io.Reader, pathFile, contentType string) (string, error)
	UploadURLAvatar(ctx context.Context, userLink uuid.UUID, objectKey string) error
	DeleteAvatarS3(ctx context.Context, key string) error
	DeleteURLAvatar(ctx context.Context, userLink uuid.UUID) error
}

type Config struct {
	BaseURLAvatar string
}

type Tools struct {
	Hasher            func(password string) (string, error)
	Checker           func(string, string) error
	GenerateAvatarKey func() (string, error)
}

type Service struct {
	rep   AuthRepository
	cfg   Config
	tools Tools
}

func NewService(rep AuthRepository, cfg Config, tools Tools) *Service {
	return &Service{
		rep:   rep,
		cfg:   cfg,
		tools: tools,
	}
}

func (s *Service) LogIn(ctx context.Context, requestUser dto.LogInUser) (dto.UserInfo, error) {
	user, err := s.rep.GetUser(ctx, requestUser.Email)
	if err != nil {
		return dto.UserInfo{}, fmt.Errorf("rep.GetUser: %w", err)
	}

	err = s.tools.Checker(requestUser.Password, user.PasswordHash)
	if err != nil {
		return dto.UserInfo{}, fmt.Errorf("rep.CheckPassword: %w", err)
	}

	return dto.UserInfo{
		Link:        user.Link,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		AvatarURL:   user.Avatar,
	}, nil
}

func (s *Service) Register(ctx context.Context, userInfo dto.RegistrationUser) (dto.UserInfo, error) {
	hashedPassword, err := s.tools.Hasher(userInfo.Password)
	if err != nil {
		return dto.UserInfo{}, fmt.Errorf("HashPassword: %w", err)
	}

	user := repositoryDto.UserInitialize{
		Link:         uuid.New(),
		DisplayName:  userInfo.DisplayName,
		PasswordHash: hashedPassword,
		Email:        userInfo.Email,
	}

	err = s.rep.AddUser(ctx, user)
	if err != nil {
		return dto.UserInfo{}, fmt.Errorf("rep.AddUser: %w", err)
	}

	return dto.UserInfo{
		Link:        user.Link,
		DisplayName: userInfo.DisplayName,
		Email:       user.Email,
	}, nil
}

func (s *Service) ResetPassword(ctx context.Context, passwordInfo dto.ResetPasswordInfo) error {
	parseUserLink, err := uuid.Parse(passwordInfo.UserLink)
	if err != nil {
		return fmt.Errorf("uuid.Parse: %w", err)
	}

	newHashPassword, err := s.tools.Hasher(passwordInfo.NewPassword)
	if err != nil {
		return fmt.Errorf("hasher: %w", err)
	}

	err = s.rep.UpdatePassword(ctx, parseUserLink, newHashPassword)
	if err != nil {
		return fmt.Errorf("rep.UpdatePassword: %w", err)
	}

	return nil
}

func (s *Service) GetUserLink(ctx context.Context, email string) (string, error) {
	userLink, err := s.rep.GetUserLink(ctx, email)
	if err != nil {
		return "", fmt.Errorf("rep.GetUser: %w", err)
	}

	return userLink.String(), nil
}

func (s *Service) GetUserByEmail(ctx context.Context, email string) (dto.UserInfo, error) {
	repositoryUser, err := s.rep.GetUser(ctx, email)
	if err != nil {
		return dto.UserInfo{}, fmt.Errorf("rep.GetUser: %w", err)
	}

	user := dto.UserInfo{
		Link:        repositoryUser.Link,
		DisplayName: repositoryUser.DisplayName,
		Email:       repositoryUser.Email,
		AvatarURL:   repositoryUser.Avatar,
	}

	return user, nil
}

func (s *Service) EnsureUserByEmail(ctx context.Context, info dto.RegistrationUser) (string, error) {
	const randomPasswordLength = 32

	user, err := s.GetUserByEmail(ctx, info.Email)
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentEmail) {
			b := make([]byte, randomPasswordLength)
			if _, err := rand.Read(b); err != nil {
				return "", fmt.Errorf("generate random password: %w", err)
			}

			password := base64.URLEncoding.EncodeToString(b)

			registerUserInfo := dto.RegistrationUser{
				DisplayName: info.DisplayName,
				Email:       info.Email,
				Password:    password,
			}
			newUser, err := s.Register(ctx, registerUserInfo)
			if err != nil {
				return "", fmt.Errorf("authService.Register: %w", err)
			}

			return newUser.Link.String(), nil
		}

		return "", fmt.Errorf("authService.GetUserByEmail: %w", err)
	}

	return user.Link.String(), nil
}

func (s *Service) GetProfile(ctx context.Context, userLink uuid.UUID) (dto.UserInfo, error) {
	repositoryUser, err := s.rep.GetProfile(ctx, userLink)
	if err != nil {
		return dto.UserInfo{}, fmt.Errorf("rep.GetProfile: %w", err)
	}

	var avatarUrl string
	if repositoryUser.AvatarKey != "" {
		avatarUrl, err = url.JoinPath(s.cfg.BaseURLAvatar, repositoryUser.AvatarKey)
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

	key, err := s.tools.GenerateAvatarKey()
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

	fullKey, err := url.JoinPath(s.cfg.BaseURLAvatar, objectKey)
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

	err = s.rep.DeleteURLAvatar(ctx, userLink)
	if err != nil {
		return fmt.Errorf("rep.DeleteAvatarURL: %w", err)
	}

	err = s.rep.DeleteAvatarS3(ctx, avatarKey)
	if err != nil {
		return fmt.Errorf("rep.DeleteAvatar: %w", err)
	}

	return nil
}
