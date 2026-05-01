package usecase

import (
	"context"
	"fmt"
	"io"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/google/uuid"
)

//go:generate mockery --name=AppealClient --output=mock_appeal_client
type AppealClient interface {
	CreateAppeal(ctx context.Context, newAppeal domain.CreateAppealInfo) (uuid.UUID, error)
	GetAppeal(ctx context.Context, userLink uuid.UUID) (string, []domain.AppealInfo, error)
	UploadAttachment(ctx context.Context, attachmentInfo domain.UploadAttachmentInfo, attachment io.Reader) (string, error)
	DeleteAppeal(ctx context.Context, deleteInfo domain.DeleteInfo) error
	GetStats(ctx context.Context, userLink uuid.UUID) (domain.AppealsStats, error)
	ChangeAppealStatus(ctx context.Context, changeStatusInfo domain.ChangeAppealStatusInfo) error
}

type Appeal struct {
	client AppealClient
}

func NewAppeal(client AppealClient) *Appeal {
	return &Appeal{
		client: client,
	}
}

func (a *Appeal) CreateAppeal(ctx context.Context, newAppeal domain.CreateAppealInfo) (uuid.UUID, error) {
	appealLink, err := a.client.CreateAppeal(ctx, newAppeal)
	if err != nil {
		return uuid.Nil, fmt.Errorf("client.CreateAppeal: %w", err)
	}

	return appealLink, nil
}

func (a *Appeal) GetAppeal(ctx context.Context, userLink uuid.UUID) (string, []domain.AppealInfo, error) {
	role, appeals, err := a.client.GetAppeal(ctx, userLink)
	if err != nil {
		return "", []domain.AppealInfo{}, fmt.Errorf("client.GetAppeal: %w", err)
	}

	return role, appeals, nil
}

func (a *Appeal) UploadAttachment(ctx context.Context, attachmentInfo domain.UploadAttachmentInfo, attachment io.Reader) (string, error) {
	attachmentKey, err := a.client.UploadAttachment(ctx, attachmentInfo, attachment)
	if err != nil {
		return "", fmt.Errorf("client.UploadAttachment: %w", err)
	}

	return attachmentKey, nil
}

func (a *Appeal) DeleteAppeal(ctx context.Context, deleteInfo domain.DeleteInfo) error {
	err := a.client.DeleteAppeal(ctx, deleteInfo)
	if err != nil {
		return fmt.Errorf("client.DeleteAppeal: %w", err)
	}

	return nil
}

func (a *Appeal) GetStats(ctx context.Context, userLink uuid.UUID) (domain.AppealsStats, error) {
	stats, err := a.client.GetStats(ctx, userLink)
	if err != nil {
		return domain.AppealsStats{}, fmt.Errorf("client.GetStats: %w", err)
	}

	return stats, nil
}

func (a *Appeal) ChangeAppealStatus(ctx context.Context, changeStatusInfo domain.ChangeAppealStatusInfo) error {
	err := a.client.ChangeAppealStatus(ctx, changeStatusInfo)
	if err != nil {
		return fmt.Errorf("client.ChangeAppealStatus: %w", err)
	}

	return nil
}
