package usecase

import (
	"context"
	"fmt"
	"io"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/google/uuid"
)

//go:generate mockery --name=BoardClient --output=mock_board_client
type BoardClient interface {
	GetBoards(ctx context.Context, userLink uuid.UUID) ([]domain.BoardInfo, error)
	GetBoard(ctx context.Context, boardInfo domain.GetBoardRequest) (domain.BoardInfo, error)
	CreateBoard(ctx context.Context, boardInfo domain.CreateBoardRequest) (domain.BoardInfo, error)
	DeleteBoard(ctx context.Context, boardInfo domain.GetBoardRequest) error
	UpdateBoard(ctx context.Context, boardInfo domain.UpdateBoardRequest) error
	UploadBackground(ctx context.Context, backgroundInfo domain.UploadBackgroundRequest, image io.Reader) (domain.UploadBackgroundResponse, error)
	GetMembers(ctx context.Context, membersInfo domain.GetMembersRequest) (domain.GetMembersResponse, error)

	CreateInvite(ctx context.Context, inviteInfo domain.CreateInviteRequest) (domain.CreateInviteResponse, error)
	AcceptInvite(ctx context.Context, inviteInfo domain.AcceptInviteRequest) (string, string, error)
	CloseInvite(ctx context.Context, inviteInfo domain.CloseInviteRequest) error
}

type Board struct {
	client BoardClient
}

func NewBoard(client BoardClient) *Board {
	return &Board{
		client: client,
	}
}

func (b *Board) GetBoards(ctx context.Context, userLink uuid.UUID) ([]domain.BoardInfo, error) {
	boards, err := b.client.GetBoards(ctx, userLink)
	if err != nil {
		return nil, fmt.Errorf("board.GetBoards: %w", err)
	}

	return boards, nil
}

func (b *Board) GetBoard(ctx context.Context, boardInfo domain.GetBoardRequest) (domain.BoardInfo, error) {
	board, err := b.client.GetBoard(ctx, boardInfo)
	if err != nil {
		return domain.BoardInfo{}, fmt.Errorf("board.GetBoard: %w", err)
	}

	return board, nil
}

func (b *Board) CreateBoard(ctx context.Context, boardInfo domain.CreateBoardRequest) (domain.BoardInfo, error) {
	board, err := b.client.CreateBoard(ctx, boardInfo)
	if err != nil {
		return domain.BoardInfo{}, fmt.Errorf("board.CreateBoard: %w", err)
	}

	return board, nil
}

func (b *Board) DeleteBoard(ctx context.Context, boardInfo domain.GetBoardRequest) error {
	err := b.client.DeleteBoard(ctx, boardInfo)
	if err != nil {
		return fmt.Errorf("board.DeleteBoard: %w", err)
	}

	return nil
}

func (b *Board) UpdateBoard(ctx context.Context, boardInfo domain.UpdateBoardRequest) error {
	err := b.client.UpdateBoard(ctx, boardInfo)
	if err != nil {
		return fmt.Errorf("board.UpdateBoard: %w", err)
	}

	return nil
}

func (b *Board) UploadBackground(ctx context.Context, backgroundInfo domain.UploadBackgroundRequest, image io.Reader) (domain.UploadBackgroundResponse, error) {
	resp, err := b.client.UploadBackground(ctx, backgroundInfo, image)
	if err != nil {
		return domain.UploadBackgroundResponse{}, fmt.Errorf("board.UploadBackground: %w", err)
	}

	return resp, nil
}

func (b *Board) GetMembers(ctx context.Context, membersInfo domain.GetMembersRequest) (domain.GetMembersResponse, error) {
	resp, err := b.client.GetMembers(ctx, membersInfo)
	if err != nil {
		return domain.GetMembersResponse{}, fmt.Errorf("board.GetMembers: %w", err)
	}

	return resp, nil
}

func (b *Board) CreateInvite(ctx context.Context, inviteInfo domain.CreateInviteRequest) (domain.CreateInviteResponse, error) {
	resp, err := b.client.CreateInvite(ctx, inviteInfo)
	if err != nil {
		return domain.CreateInviteResponse{}, fmt.Errorf("board.CreateInvite: %w", err)
	}

	return resp, nil
}

func (b *Board) AcceptInvite(ctx context.Context, inviteInfo domain.AcceptInviteRequest) (string, string, error) {
	boardLink, role, err := b.client.AcceptInvite(ctx, inviteInfo)
	if err != nil {
		return "", "", fmt.Errorf("board.AcceptInvite: %w", err)
	}

	return boardLink, role, nil
}

func (b *Board) CloseInvite(ctx context.Context, inviteInfo domain.CloseInviteRequest) error {
	err := b.client.CloseInvite(ctx, inviteInfo)
	if err != nil {
		return fmt.Errorf("board.CloseInvite: %w", err)
	}

	return nil
}
