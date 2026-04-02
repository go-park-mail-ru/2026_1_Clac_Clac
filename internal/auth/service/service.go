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
	SessiondIdKey   = "session_id"
	SessionLifetime = 24 * time.Hour

	CountRetries = 3
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

type AuthRepository interface {
	AddUser(ctx context.Context, user repositoryDto.UserInitialize) error
	AddSession(ctx context.Context, session repositoryDto.SessionEntity) error
	GetUser(ctx context.Context, email string) (repositoryDto.UserEntity, error)
	GetUserLink(ctx context.Context, email string) (uuid.UUID, error)
	DeleteSession(ctx context.Context, sessionID string) error
	GetUserIDBySession(ctx context.Context, sessionID string) (string, error)
	GetUserLinkByResetToken(ctx context.Context, tokenID string) (string, error)
	DeleteResetToken(ctx context.Context, tokenID string) error
	AddResetToken(ctx context.Context, token repositoryDto.ResetTokenEntity) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, newPasswordHash string) error
}

// Было лень править тесты из-за нового поля csrfSecret
// Поэтому написал конфиг для создания сервиса из него
type AuthServiceConfig struct {
	AuthRepository     AuthRepository
	EmailSender        SenderLetters
	Hasher             func(password string) (string, error)
	Checker            func(string, string) error
	IdGenerator        func() (string, error)
	ResetCodeGenerator func() (string, error)
	CSRFSecret         string
}

type Service struct {
	rep               AuthRepository
	sender            SenderLetters
	hasher            func(password string) (string, error)
	checker           func(string, string) error
	generatorID       func() (string, error)
	generateResetCode func() (string, error)
	csrfSecret        string
}

// Метод не поддерживает передачу секрета для генерации CSRF токена
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

func NewFromConfig(conf AuthServiceConfig) *Service {
	return &Service{
		rep:               conf.AuthRepository,
		sender:            conf.EmailSender,
		hasher:            conf.Hasher,
		checker:           conf.Checker,
		generatorID:       conf.IdGenerator,
		generateResetCode: conf.ResetCodeGenerator,
		csrfSecret:        conf.CSRFSecret,
	}
}

func (a *Service) LogIn(ctx context.Context, requestUser dto.LogInUser) (dto.UserInfo, string, error) {
	user, err := a.rep.GetUser(ctx, requestUser.Email)
	if err != nil {
		return dto.UserInfo{}, "", fmt.Errorf("rep.GetUser: %w", err)
	}

	err = a.checker(requestUser.Password, user.PasswordHash)
	if err != nil {
		return dto.UserInfo{}, "", fmt.Errorf("rep.CheckPassword: %w", err)
	}

	sessionID, err := a.generatorID()
	if err != nil {
		return dto.UserInfo{}, "", fmt.Errorf("GenerateID: %w", err)
	}

	session := repositoryDto.SessionEntity{
		SessionID: sessionID,
		UserLink:  user.Link,
		LifeTime:  24 * time.Hour,
	}

	err = a.rep.AddSession(ctx, session)
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

func (a *Service) CreateSessionForUser(ctx context.Context, link uuid.UUID) (string, error) {
	sessionID, err := a.generatorID()
	if err != nil {
		return "", fmt.Errorf("GenerateID: %w", err)
	}

	session := repositoryDto.SessionEntity{
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

func (a *Service) Register(ctx context.Context, userInfo dto.RegistrationUser) (dto.UserInfo, string, error) {
	hashedPassword, err := a.hasher(userInfo.Password)
	if err != nil {
		return dto.UserInfo{}, "", fmt.Errorf("HashPassword: %w", err)
	}

	user := repositoryDto.UserInitialize{
		Link:         uuid.New(),
		DisplayName:  userInfo.DisplayName,
		PasswordHash: hashedPassword,
		Email:        userInfo.Email,
	}

	err = a.rep.AddUser(ctx, user)
	if err != nil {
		return dto.UserInfo{}, "", fmt.Errorf("rep.AddUser: %w", err)
	}

	sessionID, err := a.generatorID()
	if err != nil {
		return dto.UserInfo{}, "", fmt.Errorf("GenerateID: %w", err)
	}

	session := repositoryDto.SessionEntity{
		SessionID: sessionID,
		UserLink:  user.Link,
		LifeTime:  24 * time.Hour,
	}

	err = a.rep.AddSession(ctx, session)
	if err != nil {
		return dto.UserInfo{}, "", fmt.Errorf("rep.AddSession: %w", err)
	}

	return dto.UserInfo{
		Link:        user.Link,
		DisplayName: userInfo.DisplayName,
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

func (a *Service) GetUserByEmail(ctx context.Context, email string) (dto.UserInfo, error) {
	repositoryUser, err := a.rep.GetUser(ctx, email)
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

func (a *Service) SendRecoveryCode(ctx context.Context, email string) error {
	logger := zerolog.Ctx(ctx)

	userLink, err := a.rep.GetUserLink(ctx, email)
	if err != nil {
		return fmt.Errorf("rep.GetUser: %w", err)
	}

	resetCode, err := a.generateResetCode()
	if err != nil {
		return fmt.Errorf("generateResetCode: %w", err)
	}

	resetToken := repositoryDto.ResetTokenEntity{
		ResetTokenID: resetCode,
		UserLink:     userLink,
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

	return nil
}

func (a *Service) EnsureUserByEmail(ctx context.Context, info dto.RegistrationUser) (dto.UserInfo, error) {
	const randomPasswordLength = 32

	user, err := a.GetUserByEmail(ctx, info.Email)
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
			user, _, err = a.Register(ctx, registerUserInfo)
			if err != nil {
				return dto.UserInfo{}, fmt.Errorf("authService.Register: %w", err)
			}

			return user, nil
		}

		return dto.UserInfo{}, fmt.Errorf("authService.GetUserByEmail: %w", err)
	}

	return user, nil
}

func (a *Service) SaveRefreshTokenFroUser(ctx context.Context, info dto.UserInfo, token string) error {
	// TODO: реализовать сохранение в redis
	return nil
}

func (a *Service) GetCSRFTokenExpireTime(ctx context.Context) (time.Time, error) {
	const expireInHours = 24
	return time.Now().Add(expireInHours * time.Hour), nil
}

func (a *Service) GenerateCSRFToken(ctx context.Context, sessionId string, expireTime int64) (string, error) {
	const intConvertationBase = 10

	h := hmac.New(sha256.New, []byte(a.csrfSecret))
	data := fmt.Sprintf("%s:%d", sessionId, expireTime)
	h.Write([]byte(data))

	token := fmt.Sprintf("%s:%s", hex.EncodeToString(h.Sum(nil)), strconv.FormatInt(expireTime, intConvertationBase))

	return token, nil
}

func (a *Service) CheckCSRFToken(ctx context.Context, sessionId string, token string) error {
	const requiredTokenDataLength = 2
	const intConvertationBase = 10
	const intConvertationSize = 64

	tokenData := strings.Split(token, ":")
	if len(tokenData) != requiredTokenDataLength {
		return ErrInvalidCSRFToken
	}

	expireTime, err := strconv.ParseInt(tokenData[1], intConvertationBase, intConvertationSize)
	if err != nil {
		return ErrCannotParseExpireTimeCSRFToken
	}

	if expireTime < time.Now().Unix() {
		return ErrCSRFTokenExpired
	}

	h := hmac.New(sha256.New, []byte(a.csrfSecret))
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
