package service

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/common"
	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/service/dto"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/boardRbac"

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
	GetUsersOfBoard(ctx context.Context, boardLink uuid.UUID) ([]uuid.UUID, error)
}

type Service struct {
	rep               BoardRepository
	permissionChecker rbac.Service
}

func NewService(rep BoardRepository, permissionChecker rbac.Service) *Service {
	return &Service{
		rep:               rep,
		permissionChecker: permissionChecker,
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

func (s *Service) GetUsersOfBoard(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID) ([]uuid.UUID, error) {
	err := s.permissionChecker.CheckPermissionOnBoard(ctx, boardLink, userLink, rbac.Actions.View)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return nil, rbac.ErrActionDenied
		}

		return nil, fmt.Errorf("service.CheckPermission: %w", err)
	}

	usersLinks, err := s.rep.GetUsersOfBoard(ctx, boardLink)
	if err != nil {
		return []uuid.UUID{}, fmt.Errorf("rep.GetUsersOfBoard: %w", err)
	}

	return usersLinks, nil
}
