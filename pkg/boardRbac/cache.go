package rbac

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	roleTTL    = 5 * time.Minute
	mappingTTL = 30 * time.Minute
)

type CachedService struct {
	rep   Repository
	redis *redis.Client
}

func NewCachedService(rep Repository, redis *redis.Client) *CachedService {
	return &CachedService{rep: rep, redis: redis}
}

func roleKey(userLink, boardLink uuid.UUID) string {
	return fmt.Sprintf("rbac:role:%s:%s", userLink, boardLink)
}

func mappingKey(resource string, link uuid.UUID) string {
	return fmt.Sprintf("rbac:res:%s:%s", resource, link)
}

func (s *CachedService) getCachedRole(ctx context.Context, userLink, boardLink uuid.UUID) (Role, bool) {
	val, err := s.redis.Get(ctx, roleKey(userLink, boardLink)).Result()
	if err != nil {
		return Roles.None, false
	}
	return Role(val), true
}

func (s *CachedService) setCachedRole(ctx context.Context, userLink, boardLink uuid.UUID, role Role) {
	s.redis.Set(ctx, roleKey(userLink, boardLink), string(role), roleTTL)
}

func (s *CachedService) getCachedMapping(ctx context.Context, resource string, link uuid.UUID) (uuid.UUID, bool) {
	val, err := s.redis.Get(ctx, mappingKey(resource, link)).Result()
	if err != nil {
		return uuid.Nil, false
	}
	boardLink, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, false
	}
	return boardLink, true
}

func (s *CachedService) setCachedMapping(ctx context.Context, resource string, link, boardLink uuid.UUID) {
	s.redis.Set(ctx, mappingKey(resource, link), boardLink.String(), mappingTTL)
}

func (s *CachedService) CheckPermissionOnBoard(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID, action Action) error {
	if role, ok := s.getCachedRole(ctx, userLink, boardLink); ok {
		if !IsActionAllowed(role, action) {
			return ErrActionDenied
		}
		return nil
	}

	role, err := s.rep.GetUserRoleByBoardLink(ctx, boardLink, userLink)
	if err != nil {
		return fmt.Errorf("rep.GetUserRoleByBoardLink: %w", err)
	}

	s.setCachedRole(ctx, userLink, boardLink, role)

	if !IsActionAllowed(role, action) {
		return ErrActionDenied
	}
	return nil
}

func (s *CachedService) CheckPermissionOnSection(ctx context.Context, sectionLink uuid.UUID, userLink uuid.UUID, action Action) error {
	if boardLink, ok := s.getCachedMapping(ctx, "section", sectionLink); ok {
		if role, ok := s.getCachedRole(ctx, userLink, boardLink); ok {
			if !IsActionAllowed(role, action) {
				return ErrActionDenied
			}
			return nil
		}
	}

	role, boardLink, err := s.rep.GetUserRoleBySectionLink(ctx, sectionLink, userLink)
	if err != nil {
		return fmt.Errorf("rep.GetUserRoleBySectionLink: %w", err)
	}

	if boardLink != uuid.Nil {
		s.setCachedMapping(ctx, "section", sectionLink, boardLink)
		s.setCachedRole(ctx, userLink, boardLink, role)
	}

	if !IsActionAllowed(role, action) {
		return ErrActionDenied
	}
	return nil
}

func (s *CachedService) CheckPermissionOnCard(ctx context.Context, cardLink uuid.UUID, userLink uuid.UUID, action Action) error {
	if boardLink, ok := s.getCachedMapping(ctx, "card", cardLink); ok {
		if role, ok := s.getCachedRole(ctx, userLink, boardLink); ok {
			if !IsActionAllowed(role, action) {
				return ErrActionDenied
			}
			return nil
		}
	}

	role, boardLink, err := s.rep.GetUserRoleByCardLink(ctx, cardLink, userLink)
	if err != nil {
		return fmt.Errorf("rep.GetUserRoleByCardLink: %w", err)
	}

	if boardLink != uuid.Nil {
		s.setCachedMapping(ctx, "card", cardLink, boardLink)
		s.setCachedRole(ctx, userLink, boardLink, role)
	}

	if !IsActionAllowed(role, action) {
		return ErrActionDenied
	}
	return nil
}

func (s *CachedService) CheckPermissionOnComment(ctx context.Context, commentLink uuid.UUID, userLink uuid.UUID, action Action) error {
	if boardLink, ok := s.getCachedMapping(ctx, "comment", commentLink); ok {
		if role, ok := s.getCachedRole(ctx, userLink, boardLink); ok {
			if !IsActionAllowed(role, action) {
				return ErrActionDenied
			}
			return nil
		}
	}

	role, boardLink, err := s.rep.GetUserRoleByCommentLink(ctx, commentLink, userLink)
	if err != nil {
		return fmt.Errorf("rep.GetUserRoleByCommentLink: %w", err)
	}

	if boardLink != uuid.Nil {
		s.setCachedMapping(ctx, "comment", commentLink, boardLink)
		s.setCachedRole(ctx, userLink, boardLink, role)
	}

	if !IsActionAllowed(role, action) {
		return ErrActionDenied
	}
	return nil
}

func (s *CachedService) CheckPermissionOnSubtask(ctx context.Context, subtaskLink uuid.UUID, userLink uuid.UUID, action Action) error {
	if boardLink, ok := s.getCachedMapping(ctx, "subtask", subtaskLink); ok {
		if role, ok := s.getCachedRole(ctx, userLink, boardLink); ok {
			if !IsActionAllowed(role, action) {
				return ErrActionDenied
			}
			return nil
		}
	}

	role, boardLink, err := s.rep.GetUserRoleBySubtaskLink(ctx, subtaskLink, userLink)
	if err != nil {
		return fmt.Errorf("rep.GetUserRoleBySubtaskLink: %w", err)
	}

	if boardLink != uuid.Nil {
		s.setCachedMapping(ctx, "subtask", subtaskLink, boardLink)
		s.setCachedRole(ctx, userLink, boardLink, role)
	}

	if !IsActionAllowed(role, action) {
		return ErrActionDenied
	}
	return nil
}

// Метод нужен, чтобы инвалидировать кеш при изменении роли пользователя
func (s *CachedService) InvalidateUserBoardRole(ctx context.Context, userLink, boardLink uuid.UUID) error {
	err := s.redis.Del(ctx, roleKey(userLink, boardLink)).Err()
	if err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("redis.Del: %w", err)
	}
	return nil
}
