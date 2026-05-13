package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/models"
	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/service/dto"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/boardRbac"
	"github.com/google/uuid"
)

//go:generate mockery --name=CardRepository --output=mock_card_rep
type CardRepository interface {
	GetCard(ctx context.Context, linkCard uuid.UUID) (repositoryDto.InfoCard, error)
	DeleteCard(ctx context.Context, linkCard uuid.UUID) error
	UpdateCardDetails(ctx context.Context, updatedCard repositoryDto.UpdatingCardDetails) error
	ReorderCard(ctx context.Context, updatingPlaceCard repositoryDto.PlaceCard) error
	CreateCard(ctx context.Context, newCard repositoryDto.NewCard) (int, error)
	GetComments(ctx context.Context, cardLink uuid.UUID) ([]repositoryDto.CommentInfo, error)
	CreateComment(ctx context.Context, createCardInfo repositoryDto.CreateCommentInfo) (repositoryDto.CommentInfo, error)
	IsCommentAuthor(ctx context.Context, commentLink uuid.UUID, userLink uuid.UUID) (bool, error)
	DeleteComment(ctx context.Context, commentLink uuid.UUID) error
	UpdateComment(ctx context.Context, updateCommentInfo repositoryDto.UpdateCommentInfo) error
	CreateSubtask(ctx context.Context, createInfo repositoryDto.CreateSubtaskInfo) (models.SubtaskInfo, error)
	DeleteSubtask(ctx context.Context, deleteInfo repositoryDto.DeleteSubtask) error
	UpdateSubtask(ctx context.Context, updateInfo repositoryDto.UpdateSubtask) error
	UploadAttachment(ctx context.Context, uploadInfo repositoryDto.UploadAttachment) (string, error)
	CreateAttachment(ctx context.Context, createInfo repositoryDto.CreateAttachment) (models.AttachmentInfo, error)
	DeleteAttachmentFromDB(ctx context.Context, attachmentLink uuid.UUID) (string, error)
	DeleteAttachmentFromS3(ctx context.Context, key string) error
}

type Config struct {
	BaseURLAttachment string
}

type Service struct {
	rep               CardRepository
	permissionChecker rbac.Service
	cfg               Config
}

func NewService(rep CardRepository, permissionChecker rbac.Service, cfg Config) *Service {
	return &Service{
		rep:               rep,
		permissionChecker: permissionChecker,
		cfg:               cfg,
	}
}

func (s *Service) GetCard(ctx context.Context, cardLink uuid.UUID, userLink uuid.UUID) (dto.InfoCard, error) {
	err := s.permissionChecker.CheckPermissionOnCard(ctx, cardLink, userLink, rbac.Actions.View)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return dto.InfoCard{}, rbac.ErrActionDenied
		}

		return dto.InfoCard{}, fmt.Errorf("CardService.CheckPermissionOnCard: %w", err)
	}

	card, err := s.rep.GetCard(ctx, cardLink)
	if err != nil {
		return dto.InfoCard{}, fmt.Errorf("rep.GetCard: %w", err)
	}

	return dto.InfoCard{
		Description:  card.Description,
		Title:        card.Title,
		ExecutorLink: card.ExecutorLink,
		DataDeadLine: card.DataDeadLine,
		Subtasks:     card.Subtasks,
		Position:     card.Position,
		Attachments:  card.Attachments,
	}, nil
}

func (s *Service) DeleteCard(ctx context.Context, cardLink uuid.UUID, userLink uuid.UUID) error {
	err := s.permissionChecker.CheckPermissionOnCard(ctx, cardLink, userLink, rbac.Actions.Edit)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return rbac.ErrActionDenied
		}

		return fmt.Errorf("CardService.CheckPermissionOnCard: %w", err)
	}

	err = s.rep.DeleteCard(ctx, cardLink)
	if err != nil {
		return fmt.Errorf("rep.DeleteCard: %w", err)
	}

	return nil
}

func (s *Service) UpdateCardDetails(ctx context.Context, updatingCard dto.UpdatingCardDetails, userLink uuid.UUID) error {
	err := s.permissionChecker.CheckPermissionOnCard(ctx, updatingCard.LinkCard, userLink, rbac.Actions.Edit)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return rbac.ErrActionDenied
		}

		return fmt.Errorf("CardService.CheckPermissionOnCard: %w", err)
	}

	err = s.rep.UpdateCardDetails(ctx, repositoryDto.UpdatingCardDetails{
		LinkCard:     updatingCard.LinkCard,
		Description:  updatingCard.Description,
		Title:        updatingCard.Title,
		LinkExecutor: updatingCard.LinkExecutor,
		DataDeadLine: updatingCard.DataDeadLine,
	})
	if err != nil {
		return fmt.Errorf("rep.UpdateCardDetails: %w", err)
	}

	return nil
}

func (s *Service) ReorderCard(ctx context.Context, updatedCard dto.PlaceCard, userLink uuid.UUID) error {
	err := s.permissionChecker.CheckPermissionOnSection(ctx, updatedCard.LinkSection, userLink, rbac.Actions.Edit)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return rbac.ErrActionDenied
		}

		return fmt.Errorf("CardService.CheckPermissionOnSection: %w", err)
	}

	err = s.rep.ReorderCard(ctx, repositoryDto.PlaceCard{
		LinkCard:    updatedCard.LinkCard,
		LinkSection: updatedCard.LinkSection,
		Position:    updatedCard.Position,
	})
	if err != nil {
		return fmt.Errorf("rep.ReoorderCard: %w", err)
	}

	return nil
}

func (s *Service) CreateCard(ctx context.Context, newCard dto.NewCard) (dto.PlaceCard, error) {
	err := s.permissionChecker.CheckPermissionOnSection(ctx, newCard.LinkSection, newCard.LinkAuthor, rbac.Actions.Edit)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return dto.PlaceCard{}, rbac.ErrActionDenied
		}

		return dto.PlaceCard{}, fmt.Errorf("CardService.CheckPermissionOnSection: %w", err)
	}

	cardLink := uuid.New()

	position, err := s.rep.CreateCard(ctx, repositoryDto.NewCard{
		LinkAuthor:   newCard.LinkAuthor,
		LinkCard:     cardLink,
		LinkSection:  newCard.LinkSection,
		Description:  newCard.Description,
		Title:        newCard.Title,
		LinkExecutor: newCard.LinkExecutor,
		DataDeadLine: newCard.DataDeadLine,
	})
	if err != nil {
		return dto.PlaceCard{}, fmt.Errorf("rep.CreateCard: %w", err)
	}

	return dto.PlaceCard{
		LinkCard:    cardLink,
		LinkSection: newCard.LinkSection,
		Position:    position,
	}, nil
}

func (s *Service) GetComments(ctx context.Context, cardLink uuid.UUID, userLink uuid.UUID) ([]dto.CommentInfo, error) {
	err := s.permissionChecker.CheckPermissionOnCard(ctx, cardLink, userLink, rbac.Actions.View)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return []dto.CommentInfo{}, rbac.ErrActionDenied
		}

		return []dto.CommentInfo{}, fmt.Errorf("CardService.CheckPermissionOnCard: %w", err)
	}

	comments, err := s.rep.GetComments(ctx, cardLink)
	if err != nil {
		return []dto.CommentInfo{}, fmt.Errorf("CardRepository.GetComments: %w", err)
	}

	commentsInfo := make([]dto.CommentInfo, 0)
	for _, comment := range comments {
		commentsInfo = append(commentsInfo, dto.CommentInfo{
			Link:       comment.Link,
			ParentLink: comment.ParentLink,
			AuthorLink: comment.AuthorLink,
			Text:       comment.Text,
			CreatedAt:  comment.CreatedAt,
		})
	}

	return commentsInfo, nil
}

func (s *Service) CreateComment(ctx context.Context, createCommentInfo dto.CreateCommentInfo) (dto.CommentInfo, error) {
	// Чтобы оставить комментарий, надо иметь права на чтение доски
	err := s.permissionChecker.CheckPermissionOnCard(ctx, createCommentInfo.CardLink, createCommentInfo.AuthorLink, rbac.Actions.View)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return dto.CommentInfo{}, rbac.ErrActionDenied
		}

		return dto.CommentInfo{}, fmt.Errorf("CardService.CheckPermissionOnCard: %w", err)
	}

	newCommentLink := uuid.New()

	comment, err := s.rep.CreateComment(ctx, repositoryDto.CreateCommentInfo{
		CommentLink: newCommentLink,
		CardLink:    createCommentInfo.CardLink,
		ParentLink:  createCommentInfo.ParentLink,
		AuthorLink:  createCommentInfo.AuthorLink,
		Text:        createCommentInfo.Text,
	})
	if err != nil {
		return dto.CommentInfo{}, fmt.Errorf("CardRepository.CreateComment: %w", err)
	}

	return dto.CommentInfo{
		Link:       comment.Link,
		ParentLink: comment.ParentLink,
		AuthorLink: comment.AuthorLink,
		Text:       comment.Text,
	}, nil
}

func (s *Service) DeleteComment(ctx context.Context, commentLink uuid.UUID, userLink uuid.UUID) error {
	err := s.permissionChecker.CheckPermissionOnComment(ctx, commentLink, userLink, rbac.Actions.View)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return rbac.ErrActionDenied
		}
		return fmt.Errorf("CardService.CheckPermissionOnComment: %w", err)
	}

	isCommentAuthor, err := s.rep.IsCommentAuthor(ctx, commentLink, userLink)
	if err != nil {
		switch {
		case errors.Is(err, common.ErrCommentNotFound):
			return common.ErrCommentNotFound
		}

		return fmt.Errorf("CardRepository.IsCommentAuthor: %w", err)
	}

	if !isCommentAuthor {
		return common.ErrPermissionDenied
	}

	err = s.rep.DeleteComment(ctx, commentLink)
	if err != nil {
		return fmt.Errorf("CardService.DeleteComment: %w", err)
	}

	return nil
}

func (s *Service) UpdateComment(ctx context.Context, updateCommentInfo dto.UpdateCommentInfo) error {
	err := s.permissionChecker.CheckPermissionOnComment(ctx, updateCommentInfo.CommentLink, updateCommentInfo.UserLink, rbac.Actions.View)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return rbac.ErrActionDenied
		}
		return fmt.Errorf("CardService.CheckPermissionOnComment: %w", err)
	}

	isCommentAuthor, err := s.rep.IsCommentAuthor(ctx, updateCommentInfo.CommentLink, updateCommentInfo.UserLink)
	if err != nil {
		switch {
		case errors.Is(err, common.ErrCommentNotFound):
			return common.ErrCommentNotFound
		}

		return fmt.Errorf("CardRepository.IsCommentAuthor: %w", err)
	}

	if !isCommentAuthor {
		return common.ErrPermissionDenied
	}

	err = s.rep.UpdateComment(ctx, repositoryDto.UpdateCommentInfo{
		CommentLink: updateCommentInfo.CommentLink,
		Text:        updateCommentInfo.Text,
	})
	if err != nil {
		return fmt.Errorf("CardRepository.UpdateComment: %w", err)
	}

	return nil
}

func (s *Service) CreateSubtask(ctx context.Context, createInfo dto.CreateSubtaskInfo, userLink uuid.UUID) (models.SubtaskInfo, error) {
	err := s.permissionChecker.CheckPermissionOnCard(ctx, createInfo.TaskLink, userLink, rbac.Actions.Edit)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return models.SubtaskInfo{}, rbac.ErrActionDenied
		}

		return models.SubtaskInfo{}, fmt.Errorf("CardService.CheckPermissionOnCard: %w", err)
	}

	newSubtaskLink := uuid.New()

	subtask, err := s.rep.CreateSubtask(ctx, repositoryDto.CreateSubtaskInfo{
		TaskLink:    createInfo.TaskLink,
		SubtaskLink: newSubtaskLink,
		Description: createInfo.Description,
	})

	if err != nil {
		return models.SubtaskInfo{}, fmt.Errorf("CardRepository.CreateSubtas: %w", err)
	}

	return models.SubtaskInfo{
		SubtaskLink: subtask.SubtaskLink,
		Description: subtask.Description,
		IsDone:      subtask.IsDone,
		Position:    subtask.Position,
	}, nil
}

func (s *Service) DeleteSubtask(ctx context.Context, deleteInfo dto.DeleteSubtask, userLink uuid.UUID) error {
	err := s.permissionChecker.CheckPermissionOnSubtask(ctx, deleteInfo.SubTaskLink, userLink, rbac.Actions.Delete)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return rbac.ErrActionDenied
		}
		return fmt.Errorf("CardService.CheckPermissionOnSubtask: %w", err)
	}

	err = s.rep.DeleteSubtask(ctx, repositoryDto.DeleteSubtask{
		SubTaskLink: deleteInfo.SubTaskLink,
	})

	if err != nil {
		return fmt.Errorf("CardRepository.DeleteSubtask: %w", err)
	}

	return nil
}

func (s *Service) UpdateSubtask(ctx context.Context, updateInfo dto.UpdateSubtask, userLink uuid.UUID) error {
	err := s.permissionChecker.CheckPermissionOnSubtask(ctx, updateInfo.SubTaskLink, userLink, rbac.Actions.Edit)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return rbac.ErrActionDenied
		}
		return fmt.Errorf("CardService.CheckPermissionOnSubtask: %w", err)
	}

	err = s.rep.UpdateSubtask(ctx, repositoryDto.UpdateSubtask{
		SubTaskLink: updateInfo.SubTaskLink,
		Description: updateInfo.Description,
		IsDone:      updateInfo.IsDone,
	})
	if err != nil {
		return fmt.Errorf("CardRepository.UpdateSubtask: %w", err)
	}

	return nil
}

func (s *Service) CreateAttachment(ctx context.Context, createInfo dto.CreateAttachment) (dto.AttachmentInfo, error) {
	err := s.permissionChecker.CheckPermissionOnCard(ctx, createInfo.TaskLink, createInfo.UserLink, rbac.Actions.Edit)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return dto.AttachmentInfo{}, rbac.ErrActionDenied
		}

		return dto.AttachmentInfo{}, fmt.Errorf("CardRepository.CreateAttachment: %w", err)
	}

	filePath := uuid.New().String() + createInfo.Extension

	key, err := s.rep.UploadAttachment(ctx, repositoryDto.UploadAttachment{
		Data:        createInfo.Data,
		FilePath:    filePath,
		ContentType: createInfo.ContentType,
	})
	if err != nil {
		return dto.AttachmentInfo{}, fmt.Errorf("CardRepository.UploadAttachment")
	}

	attachment, err := s.rep.CreateAttachment(ctx, repositoryDto.CreateAttachment{
		AttachmentLink: uuid.New(),
		TaskLink:       createInfo.TaskLink,
		Key:            key,
		Name:           createInfo.DisplayName,
	})
	if err != nil {
		return dto.AttachmentInfo{}, fmt.Errorf("CardRepository.CreateAttachment")
	}

	fullKey, err := url.JoinPath(s.cfg.BaseURLAttachment, key)
	if err != nil {
		return dto.AttachmentInfo{}, fmt.Errorf("CardService url.JoinPath: %w", err)
	}

	return dto.AttachmentInfo{
		AttachmentLink: attachment.AttachmentLink,
		Path:           fullKey,
		Position:       attachment.Position,
		DisplayName:    attachment.Name,
	}, nil
}

func (s *Service) DeleteAttachment(ctx context.Context, deleteInfo dto.DeleteAttachment) error {
	err := s.permissionChecker.CheckPermissionOnAttachment(ctx, deleteInfo.AttachmentLink, deleteInfo.UserLink, rbac.Actions.Delete)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return rbac.ErrActionDenied
		}
		return fmt.Errorf("CardService.CheckPermissionOnAttachment: %w", err)
	}

	key, err := s.rep.DeleteAttachmentFromDB(ctx, deleteInfo.AttachmentLink)
	if err != nil {
		return fmt.Errorf("CardRepository.DeleteSubtask: %w", err)
	}

	if err = s.rep.DeleteAttachmentFromS3(ctx, key); err != nil {
		return fmt.Errorf("CardRepository.DeleteAttachmentFromS3: %w", err)
	}

	return nil
}
