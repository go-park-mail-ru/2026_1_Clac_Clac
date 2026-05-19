package clients

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/board/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

var boardBufferPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

type ConfigBoard struct {
	MaxBackgroundBytesSize int
}

type Board struct {
	client pb.BoardServiceClient
	cfg    ConfigBoard
}

func NewBoardClient(connection *grpc.ClientConn, cfg ConfigBoard) *Board {
	return &Board{
		client: pb.NewBoardServiceClient(connection),
		cfg:    cfg,
	}
}

func (b *Board) GetBoards(ctx context.Context, userLink uuid.UUID) ([]domain.BoardInfo, error) {
	req := &pb.GetBoardsRequest{
		UserLink: userLink.String(),
	}

	res, err := b.client.GetBoards(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("client.GetBoards: %w", convertBoardGRPCError(err))
	}

	boards := make([]domain.BoardInfo, 0, len(res.BoardsInfo))
	for _, bi := range res.BoardsInfo {
		link, err := uuid.Parse(bi.Link)
		if err != nil {
			return nil, common.ErrorParseLink
		}
		boards = append(boards, domain.BoardInfo{
			Link:        link,
			Name:        bi.Name,
			Description: bi.Description,
			Background:  bi.Background,
		})
	}

	return boards, nil
}

func (b *Board) GetBoard(ctx context.Context, boardInfo domain.GetBoardRequest) (domain.BoardInfo, error) {
	req := &pb.GetBoardRequest{
		UserLink:  boardInfo.UserLink.String(),
		BoardLink: boardInfo.BoardLink.String(),
	}

	res, err := b.client.GetBoard(ctx, req)
	if err != nil {
		return domain.BoardInfo{}, fmt.Errorf("client.GetBoard: %w", convertBoardGRPCError(err))
	}

	link, err := uuid.Parse(res.BoardInfo.Link)
	if err != nil {
		return domain.BoardInfo{}, common.ErrorParseLink
	}

	return domain.BoardInfo{
		Link:        link,
		Name:        res.BoardInfo.Name,
		Description: res.BoardInfo.Description,
		Background:  res.BoardInfo.Background,
	}, nil
}

func (b *Board) CreateBoard(ctx context.Context, boardInfo domain.CreateBoardRequest) (domain.BoardInfo, error) {
	req := &pb.CreateBoardRequest{
		UserLink:    boardInfo.UserLink.String(),
		Name:        boardInfo.Name,
		Description: boardInfo.Description,
		Background:  boardInfo.Background,
	}

	res, err := b.client.CreateBoard(ctx, req)
	if err != nil {
		return domain.BoardInfo{}, fmt.Errorf("client.CreateBoard: %w", convertBoardGRPCError(err))
	}

	link, err := uuid.Parse(res.BoardInfo.Link)
	if err != nil {
		return domain.BoardInfo{}, common.ErrorParseLink
	}

	return domain.BoardInfo{
		Link:        link,
		Name:        res.BoardInfo.Name,
		Description: res.BoardInfo.Description,
		Background:  res.BoardInfo.Background,
	}, nil
}

func (b *Board) DeleteBoard(ctx context.Context, boardInfo domain.GetBoardRequest) error {
	req := &pb.DeleteBoardRequest{
		UserLink:  boardInfo.UserLink.String(),
		BoardLink: boardInfo.BoardLink.String(),
	}

	_, err := b.client.DeleteBoard(ctx, req)
	if err != nil {
		return fmt.Errorf("client.DeleteBoard: %w", convertBoardGRPCError(err))
	}

	return nil
}

func (b *Board) UpdateBoard(ctx context.Context, boardInfo domain.UpdateBoardRequest) error {
	req := &pb.UpdateBoardRequest{
		UserLink:    boardInfo.UserLink.String(),
		BoardLink:   boardInfo.BoardLink.String(),
		Name:        boardInfo.Name,
		Description: boardInfo.Description,
		Background:  boardInfo.Background,
	}

	_, err := b.client.UpdateBoard(ctx, req)
	if err != nil {
		return fmt.Errorf("client.UpdateBoard: %w", convertBoardGRPCError(err))
	}

	return nil
}

func (b *Board) UploadBackground(ctx context.Context, backgroundInfo domain.UploadBackgroundRequest, image io.Reader) (domain.UploadBackgroundResponse, error) {
	buf := boardBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer func() {
		if buf.Cap() <= b.cfg.MaxBackgroundBytesSize {
			boardBufferPool.Put(buf)
		}
	}()

	if _, err := io.Copy(buf, image); err != nil {
		return domain.UploadBackgroundResponse{}, fmt.Errorf("read image into buffer: %w", err)
	}

	req := &pb.UploadBackgroundRequest{
		UserLink:  backgroundInfo.UserLink.String(),
		BoardLink: backgroundInfo.BoardLink.String(),
		Image:     buf.Bytes(),
		Filename:  backgroundInfo.Filename,
	}

	res, err := b.client.UploadBackground(ctx, req)
	if err != nil {
		return domain.UploadBackgroundResponse{}, fmt.Errorf("client.UploadBackground: %w", convertBoardGRPCError(err))
	}

	return domain.UploadBackgroundResponse{
		BackgroundKey: res.BackgroundKey,
	}, nil
}

func (b *Board) GetMembers(ctx context.Context, membersInfo domain.GetMembersRequest) (domain.GetMembersResponse, error) {
	req := &pb.GetMembersRequest{
		UserLink:  membersInfo.UserLink.String(),
		BoardLink: membersInfo.BoardLink.String(),
	}

	res, err := b.client.GetMembers(ctx, req)
	if err != nil {
		return domain.GetMembersResponse{}, fmt.Errorf("client.GetMembers: %w", convertBoardGRPCError(err))
	}

	userLinks := make([]uuid.UUID, 0, len(res.UsersLinks))
	for _, l := range res.UsersLinks {
		link, err := uuid.Parse(l)
		if err != nil {
			return domain.GetMembersResponse{}, common.ErrorParseLink
		}
		userLinks = append(userLinks, link)
	}

	return domain.GetMembersResponse{
		UserLinks: userLinks,
	}, nil
}

func (b *Board) CreateInvite(ctx context.Context, inviteInfo domain.CreateInviteRequest) (domain.CreateInviteResponse, error) {
	req := &pb.CreateInviteRequest{
		UserLink:    inviteInfo.UserLink.String(),
		BoardLink:   inviteInfo.BoardLink.String(),
		DefaultRole: inviteInfo.DefaultRole,
	}

	if inviteInfo.TargetUserLink != nil {
		target := inviteInfo.TargetUserLink.String()
		req.TargetUserLink = &target
	}

	if inviteInfo.ExpireSeconds > 0 {
		req.ExpireSeconds = inviteInfo.ExpireSeconds
	}

	res, err := b.client.CreateInvite(ctx, req)
	if err != nil {
		return domain.CreateInviteResponse{}, fmt.Errorf("client.CreateInvite: %w", convertBoardGRPCError(err))
	}

	resp := domain.CreateInviteResponse{
		InviteLink:  res.InviteLink,
		BoardLink:   res.BoardLink,
		DefaultRole: res.DefaultRole,
		Status:      res.Status,
		CreatedAt:   res.CreatedAt,
	}

	if res.TargetUserLink != nil {
		resp.TargetUserLink = res.TargetUserLink
	}

	if res.ExpireAt != 0 {
		resp.ExpireAt = &res.ExpireAt
	}

	return resp, nil
}

func (b *Board) AcceptInvite(ctx context.Context, inviteInfo domain.AcceptInviteRequest) (string, string, error) {
	req := &pb.AcceptInviteRequest{
		InviteLink: inviteInfo.InviteLink,
		UserLink:   inviteInfo.UserLink.String(),
	}

	res, err := b.client.AcceptInvite(ctx, req)
	if err != nil {
		return "", "", fmt.Errorf("client.AcceptInvite: %w", convertBoardGRPCError(err))
	}

	return res.BoardLink, res.Role, nil
}

func (b *Board) CloseInvite(ctx context.Context, inviteInfo domain.CloseInviteRequest) error {
	req := &pb.CloseInviteRequest{
		UserLink:   inviteInfo.UserLink.String(),
		InviteLink: inviteInfo.InviteLink,
	}

	_, err := b.client.CloseInvite(ctx, req)
	if err != nil {
		return fmt.Errorf("client.CloseInvite: %w", convertBoardGRPCError(err))
	}

	return nil
}
