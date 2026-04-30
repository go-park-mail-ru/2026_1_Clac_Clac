package usecase

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/google/uuid"
)

type MailSenderClient interface {
	SendRecoveryCode(ctx context.Context, recoveryInfo domain.RecoveryCode) error
	CheckRecoveryCode(ctx context.Context, check domain.RecoveryCodeCheck) error
	ExchangeTokenForUser(ctx context.Context, resetToken domain.ResetToken) (uuid.UUID, error)
}

type MailSender struct {
	mail MailSenderClient
}

func NewMailSender(mail MailSenderClient) *MailSender {
	return &MailSender{
		mail: mail,
	}
}

func (ms *MailSender) SendRecoveryCode(ctx context.Context, userLink uuid.UUID, email string) error {
	err := ms.mail.SendRecoveryCode(ctx, domain.RecoveryCode{
		UserLink: userLink,
		Email:    email,
	})
	if err != nil {
		return fmt.Errorf("mail.SendRecoveryCode: %w", err)
	}

	return nil
}

func (ms *MailSender) CheckRecoveryCode(ctx context.Context, code string) error {
	err := ms.mail.CheckRecoveryCode(ctx, domain.RecoveryCodeCheck{
		Code: code,
	})
	if err != nil {
		return fmt.Errorf("mail.CheckRecoveryCode: %w", err)
	}

	return nil
}

func (ms *MailSender) ExchangeTokenForUser(ctx context.Context, resetToken domain.ResetToken) (uuid.UUID, error) {
	userLink, err := ms.mail.ExchangeTokenForUser(ctx, resetToken)
	if err != nil {
		return uuid.Nil, fmt.Errorf("ms.ExchangeTokenForUser: %w", err)
	}

	return userLink, nil
}
