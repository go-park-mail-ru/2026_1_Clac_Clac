package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service/dto"
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

// mockery --name=AuthRepository --output=mock_auth_rep --outpkg=mockAuthRep
type AuthRepository interface {
	AddUser(ctx context.Context, user repositoryDto.UserInitialize) error
	AddSession(ctx context.Context, session repositoryDto.SessionEntity) error
	ExtendSession(ctx context.Context, sessionID string, sessionLifetime time.Duration) error
	DeleteSession(ctx context.Context, sessionID string) error
	GetUser(ctx context.Context, email string) (repositoryDto.UserEntity, error)
	CheckLimit(ctx context.Context, configLimiter repositoryDto.RateLimiterConfig) (int64, error)
	GetUserLink(ctx context.Context, email string) (uuid.UUID, error)
	GetUserIDBySession(ctx context.Context, sessionID string) (string, error)
	GetUserLinkByResetToken(ctx context.Context, tokenID string) (string, error)
	SetCooldown(ctx context.Context, config repositoryDto.CoolDownConfig) (bool, time.Duration, error)
	DeleteResetToken(ctx context.Context, tokenID string) error
	AddResetToken(ctx context.Context, token repositoryDto.ResetTokenEntity) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, newPasswordHash string) error
}

type Service struct {
	rep                AuthRepository
	sender             SenderLetters
	hasher             func(password string) (string, error)
	checker            func(string, string) error
	generatorID        func() (string, error)
	generatorResetCode func() (string, error)
	createrResetKey    func(string) string
	createrSessionKey  func(string) string
}

func NewService(rep AuthRepository, sender SenderLetters,
	hasher func(password string) (string, error), checker func(string, string) error,
	generatorID func() (string, error), generateResetCode func() (string, error),
	createrResetKey func(string) string, createrSessionKey func(string) string) *Service {
	return &Service{
		rep:                rep,
		sender:             sender,
		hasher:             hasher,
		checker:            checker,
		generatorID:        generatorID,
		generatorResetCode: generateResetCode,
		createrResetKey:    createrResetKey,
		createrSessionKey:  createrSessionKey,
	}
}

func (s *Service) LogIn(ctx context.Context, requestUser dto.LogInUser) (dto.UserInfo, string, error) {
	user, err := s.rep.GetUser(ctx, requestUser.Email)
	if err != nil {
		return dto.UserInfo{}, "", fmt.Errorf("rep.GetUser: %w", err)
	}

	err = s.checker(requestUser.Password, user.PasswordHash)
	if err != nil {
		return dto.UserInfo{}, "", fmt.Errorf("rep.CheckPassword: %w", err)
	}

	sessionID, err := s.generatorID()
	if err != nil {
		return dto.UserInfo{}, "", fmt.Errorf("GenerateID: %w", err)
	}

	session := repositoryDto.SessionEntity{
		SessionKey: sessionID,
		UserLink:   user.Link,
		LifeTime:   24 * time.Hour,
	}

	err = s.rep.AddSession(ctx, session)
	if err != nil {
		return dto.UserInfo{}, "", fmt.Errorf("rep.AddSession: %w", err)
	}

	return dto.UserInfo{
		Link:        user.Link,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Avatar:      user.Avatar,
	}, sessionID, nil
}

func (s *Service) CreateSessionForUser(ctx context.Context, link uuid.UUID) (string, error) {
	sessionID, err := s.generatorID()
	if err != nil {
		return "", fmt.Errorf("GenerateID: %w", err)
	}

	session := repositoryDto.SessionEntity{
		SessionKey: sessionID,
		UserLink:   link,
		LifeTime:   SessionLifetime,
	}

	err = s.rep.AddSession(ctx, session)
	if err != nil {
		return "", fmt.Errorf("rep.AddSession: %w", err)
	}

	return sessionID, nil
}

func (s *Service) RefreshSession(ctx context.Context, sessionID string) error {
	err := s.rep.ExtendSession(ctx, sessionID, SessionLifetime)
	if err != nil {
		return fmt.Errorf("rep.UpdateExpirationSession: %w", err)
	}

	return nil
}

func (s *Service) UpdateCountRequests(ctx context.Context, configRateLimiter dto.RateLimiterConfig) (bool, error) {
	size, err := s.rep.CheckLimit(ctx, repositoryDto.RateLimiterConfig{
		UserIP: configRateLimiter.UserIP,
		Action: configRateLimiter.Action,
		Window: configRateLimiter.Window,
	})
	if err != nil {
		return false, fmt.Errorf("rep.CheckLimit: %w", err)
	}

	if size > int64(configRateLimiter.Limit) {
		return true, nil
	}

	return false, nil
}

func (s *Service) Register(ctx context.Context, userInfo dto.RegistrationUser) (dto.UserInfo, string, error) {
	hashedPassword, err := s.hasher(userInfo.Password)
	if err != nil {
		return dto.UserInfo{}, "", fmt.Errorf("HashPassword: %w", err)
	}

	user := repositoryDto.UserInitialize{
		Link:         uuid.New(),
		DisplayName:  userInfo.DisplayName,
		PasswordHash: hashedPassword,
		Email:        userInfo.Email,
	}

	err = s.rep.AddUser(ctx, user)
	if err != nil {
		return dto.UserInfo{}, "", fmt.Errorf("rep.AddUser: %w", err)
	}

	sessionID, err := s.generatorID()
	if err != nil {
		return dto.UserInfo{}, "", fmt.Errorf("GenerateID: %w", err)
	}

	session := repositoryDto.SessionEntity{
		SessionKey: sessionID,
		UserLink:   user.Link,
		LifeTime:   24 * time.Hour,
	}

	err = s.rep.AddSession(ctx, session)
	if err != nil {
		return dto.UserInfo{}, "", fmt.Errorf("rep.AddSession: %w", err)
	}

	return dto.UserInfo{
		Link:        user.Link,
		DisplayName: userInfo.DisplayName,
		Email:       user.Email,
	}, sessionID, nil
}

func (s *Service) LogOut(ctx context.Context, sessionID string) error {
	err := s.rep.DeleteSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("rep.DeleteSession: %w", err)
	}

	return nil
}

func (s *Service) GetUserLink(ctx context.Context, sessionID string) (uuid.UUID, error) {
	userLink, err := s.rep.GetUserIDBySession(ctx, sessionID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("rep.GetUserIDBySession: %w", err)
	}

	parseUserLink, err := uuid.Parse(userLink)
	if err != nil {
		return uuid.Nil, fmt.Errorf("uuid.Parse: %w", err)
	}

	return parseUserLink, nil
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
		Avatar:      repositoryUser.Avatar,
	}

	return user, nil
}

func (s *Service) CheckCoolDown(ctx context.Context, config dto.CoolDownConfig) (bool, time.Duration, error) {
	isAllowed, waitTime, err := s.rep.SetCooldown(ctx, repositoryDto.CoolDownConfig{
		Name:       config.Name,
		Email:      config.Email,
		Expiration: config.Expiration,
	})
	if err != nil {
		return false, 0, fmt.Errorf("rep.SetCooldown: %w", err)
	}

	return isAllowed, waitTime, nil
}

func (s *Service) SendRecoveryCode(ctx context.Context, email string) error {
	logger := zerolog.Ctx(ctx)

	userLink, err := s.rep.GetUserLink(ctx, email)
	if err != nil {
		return fmt.Errorf("rep.GetUser: %w", err)
	}

	resetCode, err := s.generateResetCode()
	if err != nil {
		return fmt.Errorf("generateResetCode: %w", err)
	}

	resetToken := repositoryDto.ResetTokenEntity{
		ResetTokenKey: resetCode,
		UserLink:      userLink,
		LifeTime:      time.Minute * 15,
	}

	err = s.rep.AddResetToken(ctx, resetToken)
	if err != nil {
		return fmt.Errorf("rep.AddResetToken: %w", err)
	}

	htmlBody := fmt.Sprintf(common.TemplateLetter, resetCode)

	go func(email, body string) {
		for range CountRetries {
			err := s.sender.SendLetter(email, "Code to create a new password", body)
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

func (s *Service) CheckRecoveryCode(ctx context.Context, tokenID string) error {
	_, err := s.rep.GetUserLinkByResetToken(ctx, tokenID)
	if err != nil {
		return fmt.Errorf("rep.GetResetToken: %w", err)
	}

	return nil
}

func (s *Service) ResetPassword(ctx context.Context, tokenID, newPassword string) error {
	userLink, err := s.rep.GetUserLinkByResetToken(ctx, tokenID)
	if err != nil {
		return fmt.Errorf("rep.GetResetToken: %w", err)
	}

	parseUserLink, err := uuid.Parse(userLink)
	if err != nil {
		return fmt.Errorf("uuid.Parse: %w", err)
	}

	newHashPassword, err := s.hasher(newPassword)
	if err != nil {
		return fmt.Errorf("hasher: %w", err)
	}

	err = s.rep.UpdatePassword(ctx, parseUserLink, newHashPassword)
	if err != nil {
		return fmt.Errorf("rep.UpdatePassword: %w", err)
	}

	return nil
}

func (s *Service) EnsureUserByEmail(ctx context.Context, info dto.RegistrationUser) (dto.UserInfo, error) {
	const randomPasswordLength = 32

	user, err := s.GetUserByEmail(ctx, info.Email)
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			b := make([]byte, randomPasswordLength)
			if _, err := rand.Read(b); err != nil {
				return dto.UserInfo{}, fmt.Errorf("generate random password: %w", err)
			}

			password := base64.URLEncoding.EncodeToString(b)

			// TODO: просто игнорирую сессию, но как-то это некрасиво
			// Что если нам регистрировать пользователя, а потом уже создавать сессию?
			// Или добавить usecase для этого, который уже будет сразу создавать сессию
			registerUserInfo := dto.RegistrationUser{
				DisplayName: info.DisplayName,
				Email:       info.Email,
				Password:    password,
			}
			user, _, err = s.Register(ctx, registerUserInfo)
			if err != nil {
				return dto.UserInfo{}, fmt.Errorf("authService.Register: %w", err)
			}

			return user, nil
		}

		return dto.UserInfo{}, fmt.Errorf("authService.GetUserByEmail: %w", err)
	}

	return user, nil
}

func (s *Service) SaveRefreshTokenFroUser(ctx context.Context, info dto.UserInfo, token string) error {
	// TODO: реализовать сохранение в redis
	return nil
}

func (s *Service) GenerateRandomCSRFToken(ctx context.Context) (string, error) {
	const tokenLength = 32

	b := make([]byte, tokenLength)

	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("rand.Read: %w", err)
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
