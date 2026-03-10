package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	dbConnection "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/db_connection"
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
	AddSession(ctx context.Context, session dbConnection.Session) error
	GetUser(ctx context.Context, enail string) (models.User, error)
	DeleteSession(ctx context.Context, sessionID string) error
	GetUserIDBySession(ctx context.Context, sessionID string) (uuid.UUID, error)
	GetResetToken(ctx context.Context, tokenID string) (dbConnection.ResetToken, error)
	DeleteResetToken(ctx context.Context, tokenID string) error
	AddResetToken(ctx context.Context, token dbConnection.ResetToken) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, newPasswordHash string) error
}

type AuthUserService struct {
	rep               AuthRepository
	sender            SenderLetters
	hasher            func(password string) (string, error)
	checker           func(string, string) error
	generatorID       func() (string, error)
	generateResetCode func() (string, error)
}

func NewAuthService(rep AuthRepository, sender SenderLetters, hasher func(password string) (string, error), checker func(string, string) error, generatorID func() (string, error), generateResetCode func() (string, error)) *AuthUserService {
	return &AuthUserService{
		rep:               rep,
		sender:            sender,
		hasher:            hasher,
		checker:           checker,
		generatorID:       generatorID,
		generateResetCode: generateResetCode,
	}
}

func (a *AuthUserService) LogIn(ctx context.Context, email, password string) (models.User, string, error) {
	user, err := a.rep.GetUser(ctx, email)
	if err != nil {
		return models.User{}, "", fmt.Errorf("rep.GetUser: %w", err)
	}

	err = a.checker(password, user.PasswordHash)
	if err != nil {
		return models.User{}, "", fmt.Errorf("rep.CheckPassword: %w", err)
	}

	sessionID, err := a.generatorID()
	if err != nil {
		return models.User{}, "", fmt.Errorf("GenerateID: %w", err)
	}

	session := dbConnection.Session{
		SessionID: sessionID,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	err = a.rep.AddSession(ctx, session)
	if err != nil {
		return models.User{}, "", fmt.Errorf("rep.AddSession: %w", err)
	}

	return user, sessionID, nil

}

func (a *AuthUserService) Register(ctx context.Context, name, password, email string) (models.User, string, error) {
	hashedPassword, err := a.hasher(password)
	if err != nil {
		return models.User{}, "", fmt.Errorf("HashPassword: %w", err)
	}

	user := models.User{
		ID:           uuid.New(),
		DisplayName:  name,
		PasswordHash: hashedPassword,
		Email:        email,
		Boards:       make([]models.Board, 0),
	}

	err = a.rep.AddUser(ctx, user)
	if err != nil {
		return models.User{}, "", fmt.Errorf("rep.AddUser: %w", err)
	}

	sessionID, err := a.generatorID()
	if err != nil {
		return models.User{}, "", fmt.Errorf("GenerateID: %w", err)
	}

	session := dbConnection.Session{
		SessionID: sessionID,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	err = a.rep.AddSession(ctx, session)
	if err != nil {
		return models.User{}, "", fmt.Errorf("rep.AddSession: %w", err)
	}

	return user, sessionID, nil
}

func (a *AuthUserService) LogOut(ctx context.Context, sessionID string) error {
	err := a.rep.DeleteSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("rep.DeleteSession: %w", err)
	}

	return nil
}

func (a *AuthUserService) GetUserID(ctx context.Context, sessionID string) (uuid.UUID, error) {
	userID, err := a.rep.GetUserIDBySession(ctx, sessionID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("rep.GetUserIDBySession: %w", err)
	}

	return userID, nil
}

func (a *AuthUserService) SendRecoveryCode(ctx context.Context, email string) error {
	logger := zerolog.Ctx(ctx)

	user, err := a.rep.GetUser(ctx, email)
	if err != nil {
		return fmt.Errorf("rep.GetUser: %w", err)
	}

	resetCode, err := a.generateResetCode()
	if err != nil {
		return fmt.Errorf("generateResetCode: %w", err)
	}

	resetToken := dbConnection.ResetToken{
		ResetTokenID: resetCode,
		UserID:       user.ID,
		ExpiresAt:    time.Now().Add(time.Minute * 15),
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

func (a *AuthUserService) CheckRecoveryCode(ctx context.Context, tokenID string) error {
	resetToken, err := a.rep.GetResetToken(ctx, tokenID)
	if err != nil {
		return fmt.Errorf("rep.GetResetToken: %w", err)
	}

	if time.Now().After(resetToken.ExpiresAt) {
		err := a.rep.DeleteResetToken(ctx, resetToken.ResetTokenID)
		if err != nil {
			return fmt.Errorf("rep.DeleteResetToken: %w", err)
		}

		return common.ErrorResetTokenExpired
	}

	return nil
}

func (a *AuthUserService) ResetPassword(ctx context.Context, tokenID, newPassword string) error {
	resetToken, err := a.rep.GetResetToken(ctx, tokenID)
	if err != nil {
		return fmt.Errorf("rep.GetResetToken: %w", err)
	}

	newHashPassword, err := a.hasher(newPassword)
	if err != nil {
		return fmt.Errorf("hasher: %w", err)
	}

	err = a.rep.UpdatePassword(ctx, resetToken.UserID, newHashPassword)
	if err != nil {
		return fmt.Errorf("rep.UpdatePassword: %w", err)
	}

	err = a.rep.DeleteResetToken(ctx, resetToken.ResetTokenID)
	if err != nil {
		return fmt.Errorf("rep.DeleteResetToken: %w", err)
	}

	return nil
}
