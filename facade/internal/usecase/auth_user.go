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
	ProcessUserWithVK(ctx context.Context, accessToken string, email string) (uuid.UUID, error)
	RessetPassword(ctx context.Context, updatedPassword domain.UpdatedPassoword) error
	GetUserLink(ctx context.Context, email string) (uuid.UUID, error)
}

type AuthClient interface {
	CreateSession(ctx context.Context, userLink uuid.UUID) (string, error)
	DeleteSession(ctx context.Context, sessionID string) error
	ExchangeVKCode(ctx context.Context, code string) (accessToken string, email string, err error)
}

type MailSenderClient interface {
	SendRecoveryCode(ctx context.Context, recoveryInfo domain.RecoveryCode) error
	CheckRecoveryCode(ctx context.Context, check domain.RecoveryCodeCheck) error
	ExchangeTokenForUser(ctx context.Context, resetToken domain.ResetToken) (uuid.UUID, error)
}

type RateLimiterClient interface {
	SetCooldown(ctx context.Context, cooldown domain.Cooldown) (domain.CooldownResult, error)
}

type AuthUser struct {
	user UserClient
	auth AuthClient
	mail MailSenderClient
}

func NewAuthUser(user UserClient, auth AuthClient, mail MailSenderClient) *AuthUser {
	return &AuthUser{
		user: user,
		auth: auth,
		mail: mail,
	}
}

func (au *AuthUser) Login(ctx context.Context, cred domain.Credentials) (domain.UserInfo, string, error) {
	user, err := au.user.GetUser(ctx, cred)
	if err != nil {
		return domain.UserInfo{}, "", fmt.Errorf("user.GetUser: %w", err)
	}

	sessionID, err := au.auth.CreateSession(ctx, user.UserLink)
	if err != nil {
		return domain.UserInfo{}, "", fmt.Errorf("auth.CreateSession: %w", err)
	}

	return domain.UserInfo{
		Link:        user.UserLink,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Avatar:      user.AvatarURL,
	}, sessionID, nil
}

func (au *AuthUser) Register(ctx context.Context, cred domain.NewCredentialsUser) (domain.UserInfo, string, error) {
	user, err := au.user.CreateUser(ctx, cred)
	if err != nil {
		return domain.UserInfo{}, "", fmt.Errorf("user.CreateUser: %w", err)
	}

	sessionID, err := au.auth.CreateSession(ctx, user.UserLink)
	if err != nil {
		return domain.UserInfo{}, "", fmt.Errorf("auth.CreateSession: %w", err)
	}

	return domain.UserInfo{
		Link:        user.UserLink,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Avatar:      user.AvatarURL,
	}, sessionID, nil
}

func (au *AuthUser) LoginWithVK(ctx context.Context, code string) (domain.UserInfo, string, error) {
	accessToken, email, err := au.auth.ExchangeVKCode(ctx, code)
	if err != nil {
		return domain.UserInfo{}, "", fmt.Errorf("auth.ExchangeVKCode: %w", err)
	}

	userLink, err := au.user.ProcessUserWithVK(ctx, accessToken, email)
	if err != nil {
		return domain.UserInfo{}, "", fmt.Errorf("user.ProcessUserWithVK: %w", err)
	}

	sessionID, err := au.auth.CreateSession(ctx, userLink)
	if err != nil {
		return domain.UserInfo{}, "", fmt.Errorf("auth.CreateSession: %w", err)
	}

	return domain.UserInfo{
		Link:  userLink,
		Email: email,
	}, sessionID, nil
}

func (au *AuthUser) Logout(ctx context.Context, sessionID string) error {
	err := au.auth.DeleteSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("auth.DeleteSession: %w", err)
	}

	return nil
}

func (au *AuthUser) SendRecoveryCode(ctx context.Context, email string) error {
	userLink, err := au.user.GetUserLink(ctx, email)
	if err != nil {
		return fmt.Errorf("user.GetUserLink: %w", err)
	}

	err = au.mail.SendRecoveryCode(ctx, domain.RecoveryCode{
		UserLink: userLink,
		Email:    email,
	})
	if err != nil {
		return fmt.Errorf("mail.SendRecoveryCode: %w", err)
	}

	return nil
}

func (au *AuthUser) CheckRecoveryCode(ctx context.Context, code string) error {
	err := au.mail.CheckRecoveryCode(ctx, domain.RecoveryCodeCheck{
		Code: code,
	})
	if err != nil {
		return fmt.Errorf("mail.CheckRecoveryCode: %w", err)
	}

	return nil
}

func (au *AuthUser) ResetPassword(ctx context.Context, tokenID, newPassword string) error {
	userLink, err := au.mail.ExchangeTokenForUser(ctx, domain.ResetToken{Token: tokenID})
	if err != nil {
		return fmt.Errorf("mail.ExchangeTokenForUser: %w", err)
	}

	err = au.user.RessetPassword(ctx, domain.UpdatedPassoword{
		UserLink:         userLink,
		Password:         newPassword,
		RepeatedPassword: newPassword,
	})
	if err != nil {
		return fmt.Errorf("user.RessetPassword: %w", err)
	}

	return nil
}
