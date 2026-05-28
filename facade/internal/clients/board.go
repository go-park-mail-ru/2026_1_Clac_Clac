package clients

import (
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

const (
	oneMegaByte = 1024 * 1024
)

var boardBufferPool = sync.Pool{
	New: func() any {
		buffer := make([]byte, oneMegaByte)
		return &buffer
	},
}

type ConfigBoard struct {
	MaxBackgroundBytesSize int
	ChunkSize              int
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
	stream, err := b.client.UploadBackground(ctx)
	if err != nil {
		return domain.UploadBackgroundResponse{}, fmt.Errorf("can not open stream client.UploadBackground: %w", err)
	}

	req := &pb.UploadBackgroundRequest{
		Request: &pb.UploadBackgroundRequest_Metadata{
			Metadata: &pb.MetadataUploadBackground{
				UserLink:  backgroundInfo.UserLink.String(),
				BoardLink: backgroundInfo.BoardLink.String(),
				Filename:  backgroundInfo.Filename,
			},
		},
	}

	if err := stream.Send(req); err != nil {
		return domain.UploadBackgroundResponse{}, fmt.Errorf("metadata background stream.Send: %w", err)
	}

	bufPtr := boardBufferPool.Get().(*[]byte)
	buffer := *bufPtr

	defer boardBufferPool.Put(bufPtr)

	for {
		n, err := image.Read(buffer)
		if err != nil && err != io.EOF {
			return domain.UploadBackgroundResponse{}, fmt.Errorf("image.Read: %w", err)
		}

		if n > 0 {
			chunkReq := &pb.UploadBackgroundRequest{
				Request: &pb.UploadBackgroundRequest_Image{
					Image: buffer[:n],
				},
			}

			if err := stream.Send(chunkReq); err != nil {
				return domain.UploadBackgroundResponse{}, fmt.Errorf("stream.Send: %w", err)
			}
		}

		if err == io.EOF {
			break
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		return domain.UploadBackgroundResponse{}, fmt.Errorf("stream.CloseAndRecv: %w", convertBoardGRPCError(err))
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

	members := make([]domain.MemberInfo, 0, len(res.Members))
	for _, m := range res.Members {
		link, err := uuid.Parse(m.Link)
		if err != nil {
			return domain.GetMembersResponse{}, common.ErrorParseLink
		}
		members = append(members, domain.MemberInfo{
			Link: link,
			Role: m.Role,
		})
	}

	return domain.GetMembersResponse{
		Members: members,
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

func (b *Board) UpdateMemberRole(ctx context.Context, req domain.UpdateMemberRoleRequest) error {
	resp := &pb.UpdateMemberRoleRequest{
		UserLink:       req.UserLink.String(),
		BoardLink:      req.BoardLink.String(),
		TargetUserLink: req.TargetUserLink.String(),
		NewRole:        req.NewRole,
	}

	_, err := b.client.UpdateMemberRole(ctx, resp)
	if err != nil {
		return fmt.Errorf("client.UpdateMemberRole: %w", convertBoardGRPCError(err))
	}

	return nil
}

func (b *Board) RemoveMemberFromBoard(ctx context.Context, req domain.RemoveMemberRequest) error {
	resp := &pb.RemoveMemberFromBoardRequest{
		UserLink:       req.UserLink.String(),
		BoardLink:      req.BoardLink.String(),
		TargetUserLink: req.TargetUserLink.String(),
	}

	_, err := b.client.RemoveMemberFromBoard(ctx, resp)
	if err != nil {
		return fmt.Errorf("client.RemoveMemberFromBoard: %w", convertBoardGRPCError(err))
	}

	return nil
}

func (b *Board) GetActiveInvites(ctx context.Context, userLink, boardLink uuid.UUID) ([]domain.InviteInfo, error) {
	resp := &pb.GetActiveInvitesRequest{
		UserLink:  userLink.String(),
		BoardLink: boardLink.String(),
	}

	res, err := b.client.GetActiveInvites(ctx, resp)
	if err != nil {
		return nil, fmt.Errorf("client.GetActiveInvites: %w", convertBoardGRPCError(err))
	}

	invites := make([]domain.InviteInfo, 0, len(res.Invites))
	for _, inv := range res.Invites {
		info := domain.InviteInfo{
			InviteLink:  inv.InviteLink,
			BoardLink:   inv.BoardLink,
			DefaultRole: inv.DefaultRole,
			Status:      inv.Status,
			CreatedAt:   inv.CreatedAt,
		}

		if inv.TargetUserLink != nil {
			info.TargetUserLink = inv.TargetUserLink
		}

		if inv.ExpireAt != 0 {
			info.ExpireAt = &inv.ExpireAt
		}

		invites = append(invites, info)
	}

	return invites, nil
}
