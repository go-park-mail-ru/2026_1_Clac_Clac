package rbac

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Service interface {
	CheckPermissionOnBoard(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID, action Action) error
	CheckPermissionOnSection(ctx context.Context, sectionLink uuid.UUID, userLink uuid.UUID, action Action) error
	CheckPermissionOnCard(ctx context.Context, cardLink uuid.UUID, userLink uuid.UUID, action Action) error
	CheckPermissionOnComment(ctx context.Context, commentLink uuid.UUID, userLink uuid.UUID, action Action) error
	CheckPermissionOnSubtask(ctx context.Context, subtaskLink uuid.UUID, userLink uuid.UUID, action Action) error
	CheckPermissionOnAttachment(ctx context.Context, attachmentLink uuid.UUID, userLink uuid.UUID, action Action) error
}

type service struct {
	rep Repository
}

func NewService(rep Repository) Service {
	return &service{
		rep: rep,
	}
}

func (s *service) CheckPermissionOnBoard(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID, action Action) error {
	role, err := s.rep.GetUserRoleByBoardLink(ctx, boardLink, userLink)
	if err != nil {
		return fmt.Errorf("rep.GetUserRoleByBoardLink: %w", err)
	}

	if !IsActionAllowed(role, action) {
		return ErrActionDenied
	}

	return nil
}

func (s *service) CheckPermissionOnSection(ctx context.Context, sectionLink uuid.UUID, userLink uuid.UUID, action Action) error {
	role, _, err := s.rep.GetUserRoleBySectionLink(ctx, sectionLink, userLink)
	if err != nil {
		return fmt.Errorf("rep.GetUserRoleBySectionLink: %w", err)
	}

	if !IsActionAllowed(role, action) {
		return ErrActionDenied
	}

	return nil
}

func (s *service) CheckPermissionOnCard(ctx context.Context, cardLink uuid.UUID, userLink uuid.UUID, action Action) error {
	role, _, err := s.rep.GetUserRoleByCardLink(ctx, cardLink, userLink)
	if err != nil {
		return fmt.Errorf("rep.GetUserRoleByCardLink: %w", err)
	}

	if !IsActionAllowed(role, action) {
		return ErrActionDenied
	}

	return nil
}

func (s *service) CheckPermissionOnComment(ctx context.Context, commentLink uuid.UUID, userLink uuid.UUID, action Action) error {
	role, _, err := s.rep.GetUserRoleByCommentLink(ctx, commentLink, userLink)
	if err != nil {
		return fmt.Errorf("rep.GetUserRoleByCommentLink: %w", err)
	}

	if !IsActionAllowed(role, action) {
		return ErrActionDenied
	}

	return nil
}

func (s *service) CheckPermissionOnSubtask(ctx context.Context, subtaskLink uuid.UUID, userLink uuid.UUID, action Action) error {
	role, _, err := s.rep.GetUserRoleBySubtaskLink(ctx, subtaskLink, userLink)
	if err != nil {
		return fmt.Errorf("rep.GetUserRoleBySubtaskLink: %w", err)
	}

	if !IsActionAllowed(role, action) {
		return ErrActionDenied
	}

	return nil
}

func (s *service) CheckPermissionOnAttachment(ctx context.Context, attchmentLink uuid.UUID, userLink uuid.UUID, action Action) error {
	role, _, err := s.rep.GetUserRoleByAttachmentLink(ctx, attchmentLink, userLink)
	if err != nil {
		return fmt.Errorf("rep.GetUserRoleByAttachmentLink: %w", err)
	}

	if !IsActionAllowed(role, action) {
		return ErrActionDenied
	}

	return nil
}
