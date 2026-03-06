package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	"github.com/google/uuid"
)

var (
	ErrorCreateHash    = errors.New("failed to create hash")
	ErrorWrongPassword = errors.New("write wrong password")
)

type Database interface {
	AddUser(ctx context.Context, user models.User) error
	AddSession(ctx context.Context, userID uuid.UUID, sessionID string) error
	GetUser(ctx context.Context, email string) (models.User, error)
	DeleteSession(ctx context.Context, sessionID string) error
	GetUserIDBySession(ctx context.Context, sessionID string) (uuid.UUID, error)
}

type AuthUserService struct {
	repo        Database
	hasher      func(password string) (string, error)
	checker     func(string, string) error
	generatorID func() (string, error)
}

func NewAuthService(repo Database, hasher func(password string) (string, error), checker func(string, string) error, generatorID func() (string, error)) *AuthUserService {
	return &AuthUserService{
		repo:        repo,
		hasher:      hasher,
		checker:     checker,
		generatorID: generatorID,
	}
}

func (a *AuthUserService) LogIn(ctx context.Context, email, password string) (models.User, string, error) {
	user, err := a.repo.GetUser(ctx, email)
	if err != nil {
		return models.User{}, "", fmt.Errorf("repo.GetUser: %w", err)
	}

	err = a.checker(password, user.PasswordHash)
	if err != nil {
		return models.User{}, "", fmt.Errorf("repo.CheckPassword: %w", err)
	}

	sessionID, err := a.generatorID()
	if err != nil {
		return models.User{}, "", fmt.Errorf("GenerateID: %w", err)
	}

	err = a.repo.AddSession(ctx, user.ID, sessionID)
	if err != nil {
		return models.User{}, "", fmt.Errorf("repo.AddSession: %w", err)
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
	}

	err = a.repo.AddUser(ctx, user)
	if err != nil {
		return models.User{}, "", fmt.Errorf("repo.AddUser: %w", err)
	}

	sessionID, err := a.generatorID()
	if err != nil {
		return models.User{}, "", fmt.Errorf("GenerateID: %w", err)
	}

	err = a.repo.AddSession(ctx, user.ID, sessionID)
	if err != nil {
		return models.User{}, "", fmt.Errorf("repo.AddSession: %w", err)
	}

	return user, sessionID, nil
}

func (a *AuthUserService) LogOut(ctx context.Context, sessionID string) error {
	err := a.repo.DeleteSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("repo.DeleteSession: %w", err)
	}

	return nil
}

func (a *AuthUserService) GetUserID(ctx context.Context, sessionID string) (uuid.UUID, error) {
	userID, err := a.repo.GetUserIDBySession(ctx, sessionID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("repo.GetUserIDBySession: %w", err)
	}

	return userID, nil
}
