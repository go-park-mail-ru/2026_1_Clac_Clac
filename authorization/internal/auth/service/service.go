package service

import (
	"context"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/auth/repository/dto"
	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/auth/repository/dto"
	"github.com/google/uuid"
)

type AuthRepository interface {
	AddSession(ctx context.Context, session repositoryDto.SessionEntity) error
	GetUserLink(ctx context.Context, sessionKey string) (string, error)
	DeleteSession(ctx context.Context, sessionKey string) error
	ExtendSession(ctx context.Context, session repositoryDto.ExtendedSession) error
}

type Tools struct {
	GeneratorSessionID func() (string, error)
	CreateSessionKey   func(string) string
}

type Config struct {
	SessionLifetime time.Duration
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

func (s *Service) CreateSession(ctx context.Context, userLink uuid.UUID) (string, error) {
	sessionKey, err := s.tools.GeneratorSessionID()
	if err != nil {
		return "", fmt.Errorf("tools.generatorSessionID: %w", err)
	}

	err = s.rep.AddSession(ctx, repositoryDto.SessionEntity{
		SessionKey: s.tools.CreateSessionKey(sessionKey),
		UserLink:   userLink,
		LifeTime:   s.cfg.SessionLifetime,
	})
	if err != nil {
		return "", fmt.Errorf("rep.AddSession: %w", err)
	}

	return sessionKey, nil
}

func (s *Service) GetUserLink(ctx context.Context, sessionID string) (string, error) {
	sessionKey := s.tools.CreateSessionKey(sessionID)
	userLink, err := s.rep.GetUserLink(ctx, sessionKey)
	if err != nil {
		return "", fmt.Errorf("rep.GetUserLink: %w", err)
	}

	return userLink, nil
}

func (s *Service) DeleteSession(ctx context.Context, sessionID string) error {
	key := s.tools.CreateSessionKey(sessionID)
	if err := s.rep.DeleteSession(ctx, key); err != nil {
		return fmt.Errorf("rep.DeleteSession: %w", err)
	}

	return nil
}

func (s *Service) ExtendSession(ctx context.Context, sessionID string) error {
	err := s.rep.ExtendSession(ctx, dto.ExtendedSession{
		SessionKey: s.tools.CreateSessionKey(sessionID),
		Expiration: s.cfg.SessionLifetime,
	})
	if err != nil {
		return fmt.Errorf("rep.ExtendSession: %w", err)
	}

	return nil
}
