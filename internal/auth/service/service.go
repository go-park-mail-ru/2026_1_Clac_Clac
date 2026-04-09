package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const (
	SessiondIdKey = "session_id"
)

var (
	ErrorCreateHash    = errors.New("failed to create hash")
	ErrorWrongPassword = errors.New("write wrong password")

	ErrInvalidCSRFToken               = errors.New("invalid csrf token")
	ErrCannotParseExpireTimeCSRFToken = errors.New("cannot parse expire time csrf token")
	ErrCSRFTokenExpired               = errors.New("csrf token expired")
	ErrCannotDecodeRecievedCSRFToken  = errors.New("cannot decode recieved csrf token")
	ErrCSRFTokensDoNotEqual           = errors.New("csrf tokens do not equal")
)

type SenderLetters interface {
	SendLetter(to string, subject string, htmlBody string) error
}

// mockery --name=AuthRepository --output=mock_auth_rep --outpkg=mockAuthRep
type AuthRepository interface {
	AddUser(ctx context.Context, user repositoryDto.UserInitialize) error
	AddSession(ctx context.Context, session repositoryDto.SessionEntity) error
	ExtendSession(ctx context.Context, session repositoryDto.ExtendedSession) error
	DeleteSession(ctx context.Context, sessionKey string) error
	GetUser(ctx context.Context, email string) (repositoryDto.UserEntity, error)
	CheckLimit(ctx context.Context, configLimiter repositoryDto.RateLimiterConfig) (int64, error)
	GetUserLink(ctx context.Context, email string) (uuid.UUID, error)
	GetUserIDBySession(ctx context.Context, sessionKey string) (string, error)
	GetUserLinkByResetToken(ctx context.Context, tokenID string) (string, error)
	SetCooldown(ctx context.Context, config repositoryDto.CoolDownConfig) (bool, time.Duration, error)
	DeleteResetToken(ctx context.Context, tokenKey string) error
	AddResetToken(ctx context.Context, token repositoryDto.ResetTokenEntity) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, newPasswordHash string) error
}

type Deps struct {
	Rep                AuthRepository
	Sender             SenderLetters
	Hasher             func(password string) (string, error)
	Checker            func(string, string) error
	GeneratorID        func() (string, error)
	GeneratorResetCode func() (string, error)
	CreaterResetKey    func(string) string
	CreaterSessionKey  func(string) string
	CsrfSecret         string

	SessionLifetime time.Duration
	CountRetries    int
}

type Service struct {
	deps Deps
}

func NewService(deps Deps) *Service {
	return &Service{
		deps: deps,
	}
}

func (s *Service) LogIn(ctx context.Context, requestUser dto.LogInUser) (dto.UserInfo, string, error) {
	user, err := s.deps.Rep.GetUser(ctx, requestUser.Email)
	if err != nil {
		return dto.UserInfo{}, "", fmt.Errorf("rep.GetUser: %w", err)
	}

	err = s.deps.Checker(requestUser.Password, user.PasswordHash)
	if err != nil {
		return dto.UserInfo{}, "", fmt.Errorf("rep.CheckPassword: %w", err)
	}

	sessionID, err := s.deps.GeneratorID()
	if err != nil {
		return dto.UserInfo{}, "", fmt.Errorf("GenerateID: %w", err)
	}

	session := repositoryDto.SessionEntity{
		SessionKey: s.deps.CreaterSessionKey(sessionID),
		UserLink:   user.Link,
		LifeTime:   s.deps.SessionLifetime,
	}

	err = s.deps.Rep.AddSession(ctx, session)
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
	sessionID, err := s.deps.GeneratorID()
	if err != nil {
		return "", fmt.Errorf("GenerateID: %w", err)
	}

	session := repositoryDto.SessionEntity{
		SessionKey: s.deps.CreaterSessionKey(sessionID),
		UserLink:   link,
		LifeTime:   s.deps.SessionLifetime,
	}

	err = s.deps.Rep.AddSession(ctx, session)
	if err != nil {
		return "", fmt.Errorf("rep.AddSession: %w", err)
	}

	return sessionID, nil
}

func (s *Service) RefreshSession(ctx context.Context, sessionID string) error {
	err := s.deps.Rep.ExtendSession(ctx, repositoryDto.ExtendedSession{
		Key:        s.deps.CreaterSessionKey(sessionID),
		Expiration: s.deps.SessionLifetime,
	})
	if err != nil {
		return fmt.Errorf("rep.UpdateExpirationSession: %w", err)
	}

	return nil
}

func (s *Service) UpdateCountRequests(ctx context.Context, configRateLimiter dto.RateLimiterConfig) (bool, error) {
	size, err := s.deps.Rep.CheckLimit(ctx, repositoryDto.RateLimiterConfig{
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
	hashedPassword, err := s.deps.Hasher(userInfo.Password)
	if err != nil {
		return dto.UserInfo{}, "", fmt.Errorf("HashPassword: %w", err)
	}

	user := repositoryDto.UserInitialize{
		Link:         uuid.New(),
		DisplayName:  userInfo.DisplayName,
		PasswordHash: hashedPassword,
		Email:        userInfo.Email,
	}

	err = s.deps.Rep.AddUser(ctx, user)
	if err != nil {
		return dto.UserInfo{}, "", fmt.Errorf("rep.AddUser: %w", err)
	}

	sessionID, err := s.deps.GeneratorID()
	if err != nil {
		return dto.UserInfo{}, "", fmt.Errorf("GenerateID: %w", err)
	}

	session := repositoryDto.SessionEntity{
		SessionKey: s.deps.CreaterSessionKey(sessionID),
		UserLink:   user.Link,
		LifeTime:   24 * time.Hour,
	}

	err = s.deps.Rep.AddSession(ctx, session)
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
	key := s.deps.CreaterSessionKey(sessionID)
	err := s.deps.Rep.DeleteSession(ctx, key)
	if err != nil {
		return fmt.Errorf("rep.DeleteSession: %w", err)
	}

	return nil
}

func (s *Service) GetUserLink(ctx context.Context, sessionID string) (uuid.UUID, error) {
	key := s.deps.CreaterSessionKey(sessionID)
	userLink, err := s.deps.Rep.GetUserIDBySession(ctx, key)
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
	repositoryUser, err := s.deps.Rep.GetUser(ctx, email)
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
	fullKey := fmt.Sprintf("cd:%s:%s", config.Name, config.Email)

	isAllowed, waitTime, err := s.deps.Rep.SetCooldown(ctx, repositoryDto.CoolDownConfig{
		Key:        fullKey,
		Expiration: config.Expiration,
	})
	if err != nil {
		return false, 0, fmt.Errorf("rep.SetCooldown: %w", err)
	}

	return isAllowed, waitTime, nil
}

func (s *Service) SendRecoveryCode(ctx context.Context, email string) error {
	logger := zerolog.Ctx(ctx)

	userLink, err := s.deps.Rep.GetUserLink(ctx, email)
	if err != nil {
		return fmt.Errorf("rep.GetUser: %w", err)
	}

	resetCode, err := s.deps.GeneratorResetCode()
	if err != nil {
		return fmt.Errorf("generatorResetCode: %w", err)
	}

	resetToken := repositoryDto.ResetTokenEntity{
		ResetTokenKey: s.deps.CreaterResetKey(resetCode),
		UserLink:      userLink,
		LifeTime:      time.Minute * 15,
	}

	err = s.deps.Rep.AddResetToken(ctx, resetToken)
	if err != nil {
		return fmt.Errorf("rep.AddResetToken: %w", err)
	}

	htmlBody := fmt.Sprintf(common.TemplateLetter, resetCode)

	go func(email, body string) {
		for range s.deps.CountRetries {
			err := s.deps.Sender.SendLetter(email, "Code to create a new password", body)
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
	tokenKey := s.deps.CreaterResetKey(tokenID)
	_, err := s.deps.Rep.GetUserLinkByResetToken(ctx, tokenKey)
	if err != nil {
		return fmt.Errorf("rep.GetResetToken: %w", err)
	}

	return nil
}

func (s *Service) ResetPassword(ctx context.Context, tokenID, newPassword string) error {
	tokenKey := s.deps.CreaterResetKey(tokenID)
	userLink, err := s.deps.Rep.GetUserLinkByResetToken(ctx, tokenKey)
	if err != nil {
		return fmt.Errorf("rep.GetResetToken: %w", err)
	}

	parseUserLink, err := uuid.Parse(userLink)
	if err != nil {
		return fmt.Errorf("uuid.Parse: %w", err)
	}

	newHashPassword, err := s.deps.Hasher(newPassword)
	if err != nil {
		return fmt.Errorf("hasher: %w", err)
	}

	err = s.deps.Rep.UpdatePassword(ctx, parseUserLink, newHashPassword)
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

func (a *Service) GetCSRFTokenExpireTime(ctx context.Context) (time.Time, error) {
	return time.Now().Add(csrfTokenExpireInHours * time.Hour), nil
}

func (a *Service) GenerateCSRFToken(ctx context.Context, sessionId string, expireTime int64) (string, error) {
	h := hmac.New(sha256.New, []byte(a.deps.CsrfSecret))
	data := fmt.Sprintf("%s:%d", sessionId, expireTime)
	h.Write([]byte(data))

	token := fmt.Sprintf("%s:%s", hex.EncodeToString(h.Sum(nil)), strconv.FormatInt(expireTime, csrfTokenExpireTimeConvertationBase))

	return token, nil
}

func (a *Service) CheckCSRFToken(ctx context.Context, sessionId string, token string) error {
	tokenData := strings.Split(token, ":")
	if len(tokenData) != csrfTokenPartsCount {
		return ErrInvalidCSRFToken
	}

	expireTime, err := strconv.ParseInt(tokenData[1], csrfTokenExpireTimeConvertationBase, csrfTokenExpireTimeConvertationTypeSize)
	if err != nil {
		return ErrCannotParseExpireTimeCSRFToken
	}

	if expireTime < time.Now().Unix() {
		return ErrCSRFTokenExpired
	}

	h := hmac.New(sha256.New, []byte(a.deps.CsrfSecret))
	data := fmt.Sprintf("%s:%d", sessionId, expireTime)
	h.Write([]byte(data))

	expected := h.Sum(nil)
	recieved, err := hex.DecodeString(tokenData[0])
	if err != nil {
		return ErrCannotDecodeRecievedCSRFToken
	}

	if !hmac.Equal(recieved, expected) {
		return ErrCSRFTokensDoNotEqual
	}

	return nil
}
