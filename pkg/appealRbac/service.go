package rbac

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Service interface {
	CheckPermission(ctx context.Context, userLink uuid.UUID, action Action) error
	GetUserRole(ctx context.Context, userLink uuid.UUID) (Role, error)
}

type service struct {
	rep Repository
}

func NewService(rep Repository) Service {
	return &service{
		rep: rep,
	}
}

func (s *service) CheckPermission(ctx context.Context, userLink uuid.UUID, action Action) error {
	role, err := s.rep.GetUserRoleByLink(ctx, userLink)
	if err != nil {
		return fmt.Errorf("rep.GetUserRoleByLink: %w", err)
	}

	if !IsActionAllowed(role, action) {
		return ErrActionDenied
	}

	return nil
}

func (s *service) GetUserRole(ctx context.Context, userLink uuid.UUID) (Role, error) {
	role, err := s.rep.GetUserRoleByLink(ctx, userLink)
	if err != nil {
		return Roles.User, fmt.Errorf("rep.GetUserRoleByLink: %w", err)
	}

	return role, nil
}
