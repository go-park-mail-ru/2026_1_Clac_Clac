package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/common"
	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/service/dto"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/boardRbac"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/brokerEvents"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/pubsub"
	"github.com/rs/zerolog"

	"github.com/google/uuid"
)

//go:generate mockery --name=BoardRepository --output mock_board_rep
type BoardRepository interface {
	GetBoards(ctx context.Context, userLink uuid.UUID) ([]repositoryDto.BoardEntry, error)
	GetBoard(ctx context.Context, boardLink uuid.UUID) (repositoryDto.BoardEntry, error)
	CreateBoard(ctx context.Context, boardInfo repositoryDto.NewBoardInfo, authorLink uuid.UUID) (repositoryDto.BoardEntry, error)
	DeleteBoard(ctx context.Context, boardLink uuid.UUID) error
	UpdateBoard(ctx context.Context, boardInfo repositoryDto.UpdateBoardInfo) error
	UploadBackground(ctx context.Context, source io.Reader, filename string, contentType string) (string, error)
	UpdateBackground(ctx context.Context, background string, boardLink uuid.UUID) error
	GetUsersOfBoard(ctx context.Context, boardLink uuid.UUID) ([]repositoryDto.MemberEntry, error)

	CreateInvite(ctx context.Context, inviteInfo repositoryDto.NewInviteInfo) (repositoryDto.InviteEntry, error)
	GetInviteByLink(ctx context.Context, inviteLink uuid.UUID) (repositoryDto.InviteEntry, error)
	AddMemberToBoard(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID, role rbac.Role) error
	CloseInvite(ctx context.Context, inviteLink uuid.UUID) error
	CloseInviteByBoardForUser(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID) error

	UpdateMemberRole(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID, role rbac.Role) error
	RemoveMemberFromBoard(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID) error
	GetActiveInvitesByBoard(ctx context.Context, boardLink uuid.UUID) ([]repositoryDto.InviteEntry, error)
}

type Service struct {
	rep               BoardRepository
	permissionChecker rbac.Service
	pollStore         *PollStore
	pub               pubsub.Publisher[brokerEvents.BoardUpdateEvent]
}

func NewService(rep BoardRepository, permissionChecker rbac.Service, pollStore *PollStore, pub pubsub.Publisher[brokerEvents.BoardUpdateEvent]) *Service {
	return &Service{
		rep:               rep,
		permissionChecker: permissionChecker,
		pollStore:         pollStore,
		pub:               pub,
	}
}

func (s *Service) GetBoards(ctx context.Context, userLink uuid.UUID) ([]dto.BoardInfo, error) {
	// Нет проверки прав, так как репозиторий возвращает список досок для данного пользователя
	// Он и так не получит досок, доступа к которым не имеет
	entries, err := s.rep.GetBoards(ctx, userLink)
	if err != nil {
		return []dto.BoardInfo{}, fmt.Errorf("rep.GetBoards: %w", err)
	}

	boards := make([]dto.BoardInfo, 0)
	for _, entry := range entries {
		boards = append(boards, dto.BoardInfoFromEntry(entry))
	}

	return boards, nil
}

func (s *Service) GetBoard(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID) (dto.BoardInfo, error) {
	err := s.permissionChecker.CheckPermissionOnBoard(ctx, boardLink, userLink, rbac.Actions.View)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return dto.BoardInfo{}, rbac.ErrActionDenied
		}

		return dto.BoardInfo{}, fmt.Errorf("service.CheckPermission: %w", err)
	}

	entry, err := s.rep.GetBoard(ctx, boardLink)
	if err != nil {
		if errors.Is(err, common.ErrBoardNotFound) {
			return dto.BoardInfo{}, common.ErrBoardNotFound
		}

		return dto.BoardInfo{}, fmt.Errorf("rep.GetBoard: %w", err)
	}

	return dto.BoardInfoFromEntry(entry), nil
}

func (s *Service) CreateBoard(ctx context.Context, boardInfo dto.NewBoardInfo, authorLink uuid.UUID) (dto.BoardInfo, error) {
	entry, err := s.rep.CreateBoard(ctx, dto.ToNewBoardInfo(boardInfo), authorLink)
	if err != nil {
		return dto.BoardInfo{}, fmt.Errorf("rep.CreateBoard: %w", err)
	}

	return dto.BoardInfoFromEntry(entry), nil
}

func (s *Service) DeleteBoard(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID) error {
	err := s.permissionChecker.CheckPermissionOnBoard(ctx, boardLink, userLink, rbac.Actions.Delete)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return rbac.ErrActionDenied
		}

		return fmt.Errorf("service.CheckPermission: %w", err)
	}

	err = s.rep.DeleteBoard(ctx, boardLink)
	if err != nil {
		if errors.Is(err, common.ErrBoardNotFound) {
			return common.ErrBoardNotFound
		}

		return fmt.Errorf("rep.DeleteBoard: %w", err)
	}

	return nil
}

func (s *Service) UpdateBoard(ctx context.Context, boardInfo dto.UpdateBoardInfo, userLink uuid.UUID) error {
	err := s.permissionChecker.CheckPermissionOnBoard(ctx, boardInfo.Link, userLink, rbac.Actions.Edit)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return rbac.ErrActionDenied
		}

		return fmt.Errorf("service.CheckPermission: %w", err)
	}

	err = s.rep.UpdateBoard(ctx, dto.ToUpdateBoardInfo(boardInfo))
	if err != nil {
		if errors.Is(err, common.ErrBoardNotFound) {
			return common.ErrBoardNotFound
		}

		return fmt.Errorf("rep.UpdateBoard: %w", err)
	}

	return nil
}

func (s *Service) UpdateBackground(ctx context.Context, file io.Reader, contentType string, extension string, boardLink uuid.UUID, userLink uuid.UUID) (string, error) {
	err := s.permissionChecker.CheckPermissionOnBoard(ctx, boardLink, userLink, rbac.Actions.Edit)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return "", rbac.ErrActionDenied
		}

		return "", fmt.Errorf("service.CheckPermission: %w", err)
	}

	filename := uuid.New().String() + extension

	key, err := s.rep.UploadBackground(ctx, file, filename, contentType)
	if err != nil {
		return "", fmt.Errorf("rep.UploadBackground: %w", err)
	}

	err = s.rep.UpdateBackground(ctx, key, boardLink)
	if err != nil {
		if errors.Is(err, common.ErrBoardNotFound) {
			return "", common.ErrBoardNotFound
		}

		return "", fmt.Errorf("rep.UpdateBackground: %w", err)
	}

	return key, nil
}

func (s *Service) GetUsersOfBoard(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID) ([]dto.MemberInfo, error) {
	err := s.permissionChecker.CheckPermissionOnBoard(ctx, boardLink, userLink, rbac.Actions.View)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return nil, rbac.ErrActionDenied
		}

		return nil, fmt.Errorf("service.CheckPermission: %w", err)
	}

	members, err := s.rep.GetUsersOfBoard(ctx, boardLink)
	if err != nil {
		return []dto.MemberInfo{}, fmt.Errorf("rep.GetUsersOfBoard: %w", err)
	}

	result := make([]dto.MemberInfo, 0, len(members))
	for _, m := range members {
		result = append(result, dto.MemberInfo{
			Link: m.Link,
			Role: m.Role,
		})
	}

	return result, nil
}

func (s *Service) CreateInvite(ctx context.Context, inviteInfo dto.NewInviteInfo, creatorLink uuid.UUID) (dto.InviteInfo, error) {
	err := s.permissionChecker.CheckPermissionOnBoard(ctx, inviteInfo.BoardLink, creatorLink, rbac.Actions.Invite)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return dto.InviteInfo{}, rbac.ErrActionDenied
		}
		return dto.InviteInfo{}, fmt.Errorf("service.CheckPermission: %w", err)
	}

	entry, err := s.rep.CreateInvite(ctx, repositoryDto.NewInviteInfo{
		BoardLink:   inviteInfo.BoardLink,
		UserLink:    inviteInfo.UserLink,
		DefaultRole: inviteInfo.DefaultRole,
		ExpireTime:  inviteInfo.ExpireTime,
	})
	if err != nil {
		return dto.InviteInfo{}, fmt.Errorf("rep.CreateInvite: %w", err)
	}

	return dto.InviteInfo{
		InviteLink:  entry.InviteLink,
		BoardLink:   entry.BoardLink,
		TargetUser:  entry.UserLink,
		DefaultRole: entry.DefaultRole,
		Status:      entry.Status,
		ExpireAt:    entry.ExpireTime,
		CreatedAt:   entry.CreatedAt,
	}, nil
}

func (s *Service) AcceptInvite(ctx context.Context, inviteLink uuid.UUID, userLink uuid.UUID) (dto.InviteInfo, error) {
	logger := zerolog.Ctx(ctx)

	entry, err := s.rep.GetInviteByLink(ctx, inviteLink)
	if err != nil {
		if errors.Is(err, common.ErrInviteNotFound) {
			return dto.InviteInfo{}, common.ErrInviteNotFound
		}
		return dto.InviteInfo{}, fmt.Errorf("rep.GetInviteByLink: %w", err)
	}

	if entry.Status == common.InviteStatuses.Closed {
		return dto.InviteInfo{}, common.ErrInviteClosed
	}

	if entry.ExpireTime != nil && entry.ExpireTime.Before(time.Now()) {
		err = s.rep.CloseInvite(ctx, inviteLink)
		if err != nil {
			// Так и задумано, даже если ошибка произошла в CloseInvtie
			// изначальная причина в том, что инвайт истек
			logger.Error().Err(err).Msg("rep.CloseInvite")
			return dto.InviteInfo{}, common.ErrInviteExpired
		}

		return dto.InviteInfo{}, common.ErrInviteExpired
	}

	if entry.UserLink != nil && *entry.UserLink != userLink {
		return dto.InviteInfo{}, common.ErrInviteNotForUser
	}

	err = s.rep.AddMemberToBoard(ctx, entry.BoardLink, userLink, entry.DefaultRole)
	if err != nil {
		if errors.Is(err, common.ErrUserAlreadyMember) {
			return dto.InviteInfo{}, common.ErrUserAlreadyMember
		}
		return dto.InviteInfo{}, fmt.Errorf("rep.AddMemberToBoard: %w", err)
	}

	if invalidateErr := s.permissionChecker.InvalidateUserBoardRole(ctx, userLink, entry.BoardLink); invalidateErr != nil {
		logger.Error().Err(invalidateErr).Msg("permissionChecker.InvalidateUserBoardRole")
	}

	if entry.UserLink != nil {
		err = s.rep.CloseInvite(ctx, inviteLink)
		if err != nil {
			return dto.InviteInfo{}, fmt.Errorf("rep.CloseInvite: %w", err)
		}
		entry.Status = common.InviteStatuses.Closed
	}

	return dto.InviteInfo{
		InviteLink:  entry.InviteLink,
		BoardLink:   entry.BoardLink,
		TargetUser:  entry.UserLink,
		DefaultRole: entry.DefaultRole,
		Status:      entry.Status,
		ExpireAt:    entry.ExpireTime,
		CreatedAt:   entry.CreatedAt,
	}, nil
}

func (s *Service) CloseInvite(ctx context.Context, inviteLink uuid.UUID, userLink uuid.UUID) error {
	entry, err := s.rep.GetInviteByLink(ctx, inviteLink)
	if err != nil {
		if errors.Is(err, common.ErrInviteNotFound) {
			return common.ErrInviteNotFound
		}
		return fmt.Errorf("rep.GetInviteByLink: %w", err)
	}

	err = s.permissionChecker.CheckPermissionOnBoard(ctx, entry.BoardLink, userLink, rbac.Actions.Invite)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return rbac.ErrActionDenied
		}
		return fmt.Errorf("service.CheckPermission: %w", err)
	}

	err = s.rep.CloseInvite(ctx, inviteLink)
	if err != nil {
		return fmt.Errorf("rep.CloseInvite: %w", err)
	}

	return nil
}

func (s *Service) UpdateMemberRole(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID, newRole rbac.Role, callerLink uuid.UUID) error {
	if userLink == callerLink {
		return common.ErrSelfRoleChange
	}

	err := s.permissionChecker.CheckPermissionOnBoard(ctx, boardLink, callerLink, rbac.Actions.Invite)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return rbac.ErrActionDenied
		}
		return fmt.Errorf("service.CheckPermission: %w", err)
	}

	err = s.rep.UpdateMemberRole(ctx, boardLink, userLink, newRole)
	if err != nil {
		if errors.Is(err, common.ErrBoardNotFound) {
			return common.ErrBoardNotFound
		}
		if errors.Is(err, common.ErrUserNotFound) {
			return common.ErrUserNotFound
		}
		return fmt.Errorf("rep.UpdateMemberRole: %w", err)
	}

	if invalidateErr := s.permissionChecker.InvalidateUserBoardRole(ctx, userLink, boardLink); invalidateErr != nil {
		return fmt.Errorf("permissionChecker.InvalidateUserBoardRole: %w", invalidateErr)
	}

	return nil
}

func (s *Service) RemoveMemberFromBoard(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID, callerLink uuid.UUID) error {
	if userLink != callerLink {
		err := s.permissionChecker.CheckPermissionOnBoard(ctx, boardLink, callerLink, rbac.Actions.Invite)
		if err != nil {
			if errors.Is(err, rbac.ErrActionDenied) {
				return rbac.ErrActionDenied
			}
			return fmt.Errorf("service.CheckPermission: %w", err)
		}
	}

	// Только создатель имеет право на Delete, так что если проверка прошла — это создатель
	if userLink == callerLink && s.permissionChecker.CheckPermissionOnBoard(ctx, boardLink, callerLink, rbac.Actions.Delete) == nil {
		return common.ErrCreatorCannotLeave
	}

	err := s.rep.RemoveMemberFromBoard(ctx, boardLink, userLink)
	if err != nil {
		if errors.Is(err, common.ErrBoardNotFound) {
			return common.ErrBoardNotFound
		}
		if errors.Is(err, common.ErrUserNotFound) {
			return common.ErrUserNotFound
		}
		return fmt.Errorf("rep.RemoveMemberFromBoard: %w", err)
	}

	if invalidateErr := s.permissionChecker.InvalidateUserBoardRole(ctx, userLink, boardLink); invalidateErr != nil {
		return fmt.Errorf("permissionChecker.InvalidateUserBoardRole: %w", invalidateErr)
	}

	return nil
}

func (s *Service) CanView(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID) error {
	err := s.permissionChecker.CheckPermissionOnBoard(ctx, boardLink, userLink, rbac.Actions.View)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return rbac.ErrActionDenied
		}
		return fmt.Errorf("service.CheckPermission: %w", err)
	}

	return nil
}

func (s *Service) GetActiveInvites(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID) ([]dto.InviteInfo, error) {
	err := s.permissionChecker.CheckPermissionOnBoard(ctx, boardLink, userLink, rbac.Actions.Invite)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return nil, rbac.ErrActionDenied
		}
		return nil, fmt.Errorf("service.CheckPermission: %w", err)
	}

	entries, err := s.rep.GetActiveInvitesByBoard(ctx, boardLink)
	if err != nil {
		return nil, fmt.Errorf("rep.GetActiveInvitesByBoard: %w", err)
	}

	invites := make([]dto.InviteInfo, 0, len(entries))
	for _, entry := range entries {
		invites = append(invites, dto.InviteInfo{
			InviteLink:  entry.InviteLink,
			BoardLink:   entry.BoardLink,
			TargetUser:  entry.UserLink,
			DefaultRole: entry.DefaultRole,
			Status:      entry.Status,
			ExpireAt:    entry.ExpireTime,
			CreatedAt:   entry.CreatedAt,
		})
	}

	return invites, nil
}

func (s *Service) CreatePoll(ctx context.Context, boardLink, adminLink uuid.UUID, cards []uuid.UUID, invitees []uuid.UUID) error {
	err := s.permissionChecker.CheckPermissionOnBoard(ctx, boardLink, adminLink, rbac.Actions.Invite)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return rbac.ErrActionDenied
		}
		return fmt.Errorf("service.CheckPermission: %w", err)
	}

	if err := s.pollStore.Create(boardLink, adminLink, cards, invitees); err != nil {
		return fmt.Errorf("pollStore.Create: %w", err)
	}

	if _, err := s.pub.Publish(
		ctx,
		pubsub.Channel(boardLink.String()),
		pubsub.Event[brokerEvents.BoardUpdateEvent]{
			Type: pubsub.Type("poll_start"),
			Payload: brokerEvents.BoardUpdateEvent{
				BoardLink: boardLink.String(),
				UserLink:  adminLink.String(),
				Data: struct {
					ParticipantCount int `json:"participant_count"`
				}{
					ParticipantCount: len(invitees),
				},
			},
		},
	); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("failed to publish poll_start event")
	}

	return nil
}

func (s *Service) DeletePoll(ctx context.Context, boardLink, userLink uuid.UUID) error {
	if err := s.pollStore.Delete(boardLink, userLink); err != nil {
		return fmt.Errorf("pollStore.Delete: %w", err)
	}

	if _, err := s.pub.Publish(
		ctx,
		pubsub.Channel(boardLink.String()),
		pubsub.Event[brokerEvents.BoardUpdateEvent]{
			Type: pubsub.Type("poll_end"),
			Payload: brokerEvents.BoardUpdateEvent{
				BoardLink: boardLink.String(),
				UserLink:  userLink.String(),
				Data:      struct{}{},
			},
		},
	); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("failed to publish poll_end event")
	}

	return nil
}

func (s *Service) NextPollCard(ctx context.Context, boardLink, userLink uuid.UUID) error {
	poll, err := s.pollStore.NextCard(boardLink, userLink)
	if err != nil {
		if errors.Is(err, common.ErrPollNoMoreCards) {
			if _, pubErr := s.pub.Publish(
				ctx,
				pubsub.Channel(boardLink.String()),
				pubsub.Event[brokerEvents.BoardUpdateEvent]{
					Type: pubsub.Type("poll_end"),
					Payload: brokerEvents.BoardUpdateEvent{
						BoardLink: boardLink.String(),
						UserLink:  userLink.String(),
						Data:      struct{}{},
					},
				},
			); pubErr != nil {
				zerolog.Ctx(ctx).Error().Err(pubErr).Msg("failed to publish poll_end event")
			}
			return common.ErrPollNoMoreCards
		}
		return fmt.Errorf("pollStore.NextCard: %w", err)
	}

	if _, err := s.pub.Publish(
		ctx,
		pubsub.Channel(boardLink.String()),
		pubsub.Event[brokerEvents.BoardUpdateEvent]{
			Type: pubsub.Type("next_card"),
			Payload: brokerEvents.BoardUpdateEvent{
				BoardLink: boardLink.String(),
				UserLink:  userLink.String(),
				Data: struct {
					Title string `json:"title"`
				}{
					Title: poll.Tasks[poll.CurrentIdx].Title,
				},
			},
		},
	); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("failed to publish next_card event")
	}

	return nil
}

func (s *Service) VotePoll(ctx context.Context, boardLink, userLink uuid.UUID, points int) error {
	if err := s.pollStore.Vote(boardLink, userLink, points); err != nil {
		return fmt.Errorf("pollStore.Vote: %w", err)
	}

	if _, err := s.pub.Publish(
		ctx,
		pubsub.Channel(boardLink.String()),
		pubsub.Event[brokerEvents.BoardUpdateEvent]{
			Type: pubsub.Type("new_answer"),
			Payload: brokerEvents.BoardUpdateEvent{
				BoardLink: boardLink.String(),
				UserLink:  userLink.String(),
				Data: struct {
					UserLink string `json:"user_link"`
					Points   int    `json:"points"`
				}{
					UserLink: userLink.String(),
					Points:   points,
				},
			},
		},
	); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("failed to publish new_answer event")
	}

	return nil
}

func (s *Service) GetActivePoll(ctx context.Context, boardLink, userLink uuid.UUID) (*Poll, error) {
	err := s.permissionChecker.CheckPermissionOnBoard(ctx, boardLink, userLink, rbac.Actions.View)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return nil, rbac.ErrActionDenied
		}
		return nil, fmt.Errorf("service.CheckPermission: %w", err)
	}

	poll, ok := s.pollStore.GetActivePoll(boardLink)
	if !ok {
		return nil, common.ErrPollNotFound
	}
	return poll, nil
}
