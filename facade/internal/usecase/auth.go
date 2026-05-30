package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type AuthClient interface {
	CreateSession(ctx context.Context, userLink uuid.UUID) (string, error)
	DeleteSession(ctx context.Context, sessionID string) error
}

type Auth struct {
	auth AuthClient
}

func NewAuth(auth AuthClient) *Auth {
	return &Auth{
		auth: auth,
	}
}

func (a *Auth) CreateSession(ctx context.Context, userLink uuid.UUID) (string, error) {
	sessionID, err := a.auth.CreateSession(ctx, userLink)
	if err != nil {
		return "", fmt.Errorf("auth.CreateSession: %w", err)
	}

	return sessionID, nil
}

func (a *Auth) DeleteSession(ctx context.Context, sessionID string) error {
	err := a.auth.DeleteSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("auth.DeleteSession: %w", err)
	}

	return nil
}
