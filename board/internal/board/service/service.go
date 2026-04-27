package service

import (
	"context"
	"errors"
	"fmt"
	"io"

<<<<<<<< HEAD:board/internal/board/service/service.go
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/common"
	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/service/dto"
========
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/board/common"
	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/board/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/board/service/dto"
>>>>>>>> feat/add-facade:monolith/internal/board/service/service.go

	"github.com/google/uuid"
)

//go:generate mockery --name=BoardRepository --output mock_board_rep
type BoardRepository interface {
	GetBoards(ctx context.Context, userLink uuid.UUID) ([]repositoryDto.BoardEntry, error)
	GetBoard(ctx context.Context, boardLink uuid.UUID) (repositoryDto.BoardEntry, error)
	CreateBoard(ctx context.Context, boardInfo repositoryDto.NewBoardInfo, authorLink uuid.UUID) (repositoryDto.BoardEntry, error)
	DeleteBoard(ctx context.Context, boardLink uuid.UUID) error
	UpdateBoard(ctx context.Context, boardInfo repositoryDto.UpdateBoardInfo) error
	GetUserRoleOnBoard(ctx context.Context, userLink uuid.UUID, boardLink uuid.UUID) (common.Role, error)
	UploadBackground(ctx context.Context, source io.Reader, filename string, contentType string) (string, error)
	UpdateBackground(ctx context.Context, background string, boardLink uuid.UUID) error
	GetUsersOfBoard(ctx context.Context, boardLink uuid.UUID) ([]uuid.UUID, error)
}

type Service struct {
	rep BoardRepository
}

func NewService(rep BoardRepository) *Service {
	return &Service{
		rep: rep,
	}
}

func (s *Service) GetBoards(ctx context.Context, userLink uuid.UUID) ([]dto.BoardInfo, error) {
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
	err := s.CheckPermission(ctx, boardLink, userLink, common.Actions.View)
	if err != nil {
		if errors.Is(err, common.ErrActionDenied) {
			return dto.BoardInfo{}, common.ErrActionDenied
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
	err := s.CheckPermission(ctx, boardLink, userLink, common.Actions.Delete)
	if err != nil {
		if errors.Is(err, common.ErrActionDenied) {
			return common.ErrActionDenied
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
	err := s.CheckPermission(ctx, boardInfo.Link, userLink, common.Actions.Edit)
	if err != nil {
		if errors.Is(err, common.ErrActionDenied) {
			return common.ErrActionDenied
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

func (s *Service) CheckPermission(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID, action common.Action) error {
	role, err := s.rep.GetUserRoleOnBoard(ctx, userLink, boardLink)
	if err != nil {
		return fmt.Errorf("rep.GetUserRoleOnBoard: %w", err)
	}

	if !common.IsActionAllowed(role, action) {
		return common.ErrActionDenied
	}

	return nil
}

func (s *Service) UpdateBackground(ctx context.Context, file io.Reader, contentType string, extension string, boardLink uuid.UUID, userLink uuid.UUID) (string, error) {
	err := s.CheckPermission(ctx, boardLink, userLink, common.Actions.Edit)
	if err != nil {
		if errors.Is(err, common.ErrActionDenied) {
			return "", common.ErrActionDenied
		}

		return "", fmt.Errorf("service.CheckPermission: %w", err)
	}

	filename := uuid.New().String() + extension

	key, err := s.rep.UploadBackground(ctx, file, filename, contentType)
	if err != nil {
		return "", fmt.Errorf("board repository UploadBackground: %w", err)
	}

	err = s.rep.UpdateBackground(ctx, key, boardLink)
	if err != nil {
		if errors.Is(err, common.ErrBoardNotFound) {
			return "", common.ErrBoardNotFound
		}

		return "", fmt.Errorf("board repository UpdateBackground: %w", err)
	}

	return key, nil
}

func (s *Service) GetUsersOfBoard(ctx context.Context, boardLink uuid.UUID) ([]uuid.UUID, error) {
	usersLinks, err := s.rep.GetUsersOfBoard(ctx, boardLink)
	if err != nil {
		return []uuid.UUID{}, fmt.Errorf("BoardRepository.GetUsersOfBoard: %w", err)
	}

	return usersLinks, nil
}
