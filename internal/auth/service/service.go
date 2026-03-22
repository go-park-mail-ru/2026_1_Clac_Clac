package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/models"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const (
	SessiondIdKey   = "session_id"
	SessionLifetime = 24 * time.Hour

	CountRetries = 3
)

var (
	ErrorCreateHash    = errors.New("failed to create hash")
	ErrorWrongPassword = errors.New("write wrong password")
)

type SenderLetters interface {
	SendLetter(to string, subject string, htmlBody string) error
}

type AuthRepository interface {
	AddUser(ctx context.Context, user models.User) error
	AddSession(ctx context.Context, session dto.Session) error
	GetUser(ctx context.Context, enail string) (models.User, error)
	DeleteSession(ctx context.Context, sessionID string) error
	GetUserIDBySession(ctx context.Context, sessionID string) (string, error)
	GetUserLinkByResetToken(ctx context.Context, tokenID string) (string, error)
	DeleteResetToken(ctx context.Context, tokenID string) error
	AddResetToken(ctx context.Context, token dto.ResetToken) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, newPasswordHash string) error
}

type Service struct {
	rep               AuthRepository
	sender            SenderLetters
	hasher            func(password string) (string, error)
	checker           func(string, string) error
	generatorID       func() (string, error)
	generateResetCode func() (string, error)
}

func NewService(rep AuthRepository, sender SenderLetters, hasher func(password string) (string, error), checker func(string, string) error, generatorID func() (string, error), generateResetCode func() (string, error)) *Service {
	return &Service{
		rep:               rep,
		sender:            sender,
		hasher:            hasher,
		checker:           checker,
		generatorID:       generatorID,
		generateResetCode: generateResetCode,
	}
}

func (a *Service) LogIn(ctx context.Context, requestUser dto.LoginInfoRequest) (dto.UserInfoResponce, string, error) {
	user, err := a.rep.GetUser(ctx, requestUser.Email)
	if err != nil {
		return dto.UserInfoResponce{}, "", fmt.Errorf("rep.GetUser: %w", err)
	}

	err = a.checker(requestUser.Password, user.PasswordHash)
	if err != nil {
		return dto.UserInfoResponce{}, "", fmt.Errorf("rep.CheckPassword: %w", err)
	}

	sessionID, err := a.generatorID()
	if err != nil {
		return dto.UserInfoResponce{}, "", fmt.Errorf("GenerateID: %w", err)
	}

	session := dto.Session{
		SessionID: sessionID,
		UserLink:  user.Link,
		LifeTime:  24 * time.Hour,
	}

	err = a.rep.AddSession(ctx, session)
	if err != nil {
		return dto.UserInfoResponce{}, "", fmt.Errorf("rep.AddSession: %w", err)
	}

	return dto.UserInfoResponce{
		Link:        user.Link,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Avatar:      user.Avatar,
	}, sessionID, nil

}

func (a *Service) CreateSessionForUser(ctx context.Context, link uuid.UUID) (string, error) {
	sessionID, err := a.generatorID()
	if err != nil {
		return "", fmt.Errorf("GenerateID: %w", err)
	}

	session := dto.Session{
		SessionID: sessionID,
		UserLink:  link,
		LifeTime:  SessionLifetime,
	}

	err = a.rep.AddSession(ctx, session)
	if err != nil {
		return "", fmt.Errorf("rep.AddSession: %w", err)
	}

	return sessionID, nil
}

func (a *Service) Register(ctx context.Context, userInfo dto.RegistraionInfoRequest) (dto.UserInfoResponce, string, error) {
	hashedPassword, err := a.hasher(userInfo.Password)
	if err != nil {
		return dto.UserInfoResponce{}, "", fmt.Errorf("HashPassword: %w", err)
	}

	user := models.User{
		Link:         uuid.New(),
		DisplayName:  userInfo.Name,
		PasswordHash: hashedPassword,
		Email:        userInfo.Email,
	}

	err = a.rep.AddUser(ctx, user)
	if err != nil {
		return dto.UserInfoResponce{}, "", fmt.Errorf("rep.AddUser: %w", err)
	}

	sessionID, err := a.generatorID()
	if err != nil {
		return dto.UserInfoResponce{}, "", fmt.Errorf("GenerateID: %w", err)
	}

	session := dto.Session{
		SessionID: sessionID,
		UserLink:  user.Link,
		LifeTime:  24 * time.Hour,
	}

	err = a.rep.AddSession(ctx, session)
	if err != nil {
		return dto.UserInfoResponce{}, "", fmt.Errorf("rep.AddSession: %w", err)
	}

	return dto.UserInfoResponce{
		Link:        user.Link,
		DisplayName: userInfo.Name,
		Email:       user.Email,
	}, sessionID, nil
}

func (a *Service) LogOut(ctx context.Context, sessionID string) error {
	err := a.rep.DeleteSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("rep.DeleteSession: %w", err)
	}

	return nil
}

func (a *Service) GetUserLink(ctx context.Context, sessionID string) (uuid.UUID, error) {
	userLink, err := a.rep.GetUserIDBySession(ctx, sessionID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("rep.GetUserIDBySession: %w", err)
	}

	parseUserLink, err := uuid.Parse(userLink)
	if err != nil {
		return uuid.Nil, fmt.Errorf("uuid.Parse: %w", err)
	}

	return parseUserLink, nil
}

func (a *Service) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	user, err := a.rep.GetUser(ctx, email)
	if err != nil {
		return models.User{}, fmt.Errorf("rep.GetUser: %w", err)
	}

	return user, nil
}

func (a *Service) SendRecoveryCode(ctx context.Context, email string) error {
	logger := zerolog.Ctx(ctx)

	user, err := a.rep.GetUser(ctx, email)
	if err != nil {
		return fmt.Errorf("rep.GetUser: %w", err)
	}

	resetCode, err := a.generateResetCode()
	if err != nil {
		return fmt.Errorf("generateResetCode: %w", err)
	}

	resetToken := dto.ResetToken{
		ResetTokenID: resetCode,
		UserLink:     user.Link,
		LifeTime:     time.Minute * 15,
	}

	err = a.rep.AddResetToken(ctx, resetToken)
	if err != nil {
		return fmt.Errorf("rep.AddResetToken: %w", err)
	}

	htmlBody := fmt.Sprintf(common.TemplateLetter, resetCode)

	go func(email, body string) {
		for range CountRetries {
			err := a.sender.SendLetter(email, "Code to create a new password", body)
			if err == nil {
				return
			}

			logger.Error().Msgf("mail error %v", err)

			time.Sleep(time.Second * 2)
		}

		logger.Error().Msg("all attempts to send mail failed")
	}(email, htmlBody)

	return nil
}

func (a *Service) CheckRecoveryCode(ctx context.Context, tokenID string) error {
	_, err := a.rep.GetUserLinkByResetToken(ctx, tokenID)
	if err != nil {
		return fmt.Errorf("rep.GetResetToken: %w", err)
	}

	return nil
}

func (a *Service) ResetPassword(ctx context.Context, tokenID, newPassword string) error {
	logger := zerolog.Ctx(ctx)

	userLink, err := a.rep.GetUserLinkByResetToken(ctx, tokenID)
	if err != nil {
		return fmt.Errorf("rep.GetResetToken: %w", err)
	}

	parseUserLink, err := uuid.Parse(userLink)
	if err != nil {
		return fmt.Errorf("uuid.Parse: %w", err)
	}

	newHashPassword, err := a.hasher(newPassword)
	if err != nil {
		return fmt.Errorf("hasher: %w", err)
	}

	err = a.rep.UpdatePassword(ctx, parseUserLink, newHashPassword)
	if err != nil {
		return fmt.Errorf("rep.UpdatePassword: %w", err)
	}

	err = a.rep.DeleteResetToken(ctx, tokenID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete reset token after successful password change")
	}

	return nil
}
