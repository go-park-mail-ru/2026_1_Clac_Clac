package service

import (
	"context"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/mail_sender/internal/common"
	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/mail_sender/internal/sender/repository/dto"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type SenderLetters interface {
	SendLetter(ctx context.Context, to string, subject string, htmlBody string) error
}

type SenderRepository interface {
	DeleteResetToken(ctx context.Context, tokenKey string) error
	AddResetToken(ctx context.Context, token repositoryDto.ResetTokenEntity) error
	GetUserLinkByResetToken(ctx context.Context, tokenID string) (string, error)
}

type Config struct {
	CountRetries       int
	LifeTimeResetToken time.Duration
	SleepTime          time.Duration
}

type Tools struct {
	GeneratorResetCode func() (string, error)
	CreatorResetKey    func(string) string
}

type Service struct {
	rep    SenderRepository
	sender SenderLetters
	cfg    Config
	tools  Tools
}

func NewService(rep SenderRepository, sender SenderLetters, cfg Config, tools Tools) *Service {
	return &Service{
		rep:    rep,
		sender: sender,
		cfg:    cfg,
		tools:  tools,
	}
}

func (s *Service) SendRecoveryCode(ctx context.Context, userLink uuid.UUID, email string) error {
	logger := zerolog.Ctx(ctx)

	resetCode, err := s.tools.GeneratorResetCode()
	if err != nil {
		return fmt.Errorf("generatorResetCode: %w", err)
	}

	resetToken := repositoryDto.ResetTokenEntity{
		ResetTokenKey: s.tools.CreatorResetKey(resetCode),
		UserLink:      userLink,
		LifeTime:      s.cfg.LifeTimeResetToken,
	}

	err = s.rep.AddResetToken(ctx, resetToken)
	if err != nil {
		return fmt.Errorf("rep.AddResetToken: %w", err)
	}

	htmlBody := fmt.Sprintf(common.TemplateLetter, resetCode)

	mailContext := context.WithoutCancel(ctx)
	go func(ctx context.Context, logger *zerolog.Logger, email, body string) {
		for range s.cfg.CountRetries {

			err := s.sender.SendLetter(ctx, email, "Code to create a new password", body)
			if err == nil {
				return
			}

			logger.Error().Msgf("mail error %v", err)

			time.Sleep(s.cfg.SleepTime)
		}

		logger.Error().Msg("all attempts to send mail failed")
	}(mailContext, logger, email, htmlBody)

	return nil
}

func (s *Service) CheckRecoveryCode(ctx context.Context, tokenID string) error {
	tokenKey := s.tools.CreatorResetKey(tokenID)

	_, err := s.rep.GetUserLinkByResetToken(ctx, tokenKey)
	if err != nil {
		return fmt.Errorf("rep.GetUserLinkByResetToken: %w", err)
	}

	return nil
}

func (s *Service) GetUserLink(ctx context.Context, tokenID string) (string, error) {
	tokenKey := s.tools.CreatorResetKey(tokenID)

	userLink, err := s.rep.GetUserLinkByResetToken(ctx, tokenKey)
	if err != nil {
		return "", fmt.Errorf("rep.GetUserLinkByResetToken: %w", err)
	}

	if err := s.rep.DeleteResetToken(ctx, tokenKey); err != nil {
		return "", fmt.Errorf("rep.DeleteResetToken: %w", err)
	}

	return userLink, nil
}
