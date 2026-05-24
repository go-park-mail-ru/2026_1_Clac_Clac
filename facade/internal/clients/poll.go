package clients

import (
	"context"
	"fmt"

	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/board/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type Poll struct {
	client pb.BoardServiceClient
}

func NewPollClient(connection *grpc.ClientConn) *Poll {
	return &Poll{
		client: pb.NewBoardServiceClient(connection),
	}
}

func (p *Poll) CreatePoll(ctx context.Context, boardLink, adminLink uuid.UUID, cards []uuid.UUID, invitees []uuid.UUID) error {
	rawCards := make([]string, 0, len(cards))
	for _, c := range cards {
		rawCards = append(rawCards, c.String())
	}

	rawInvitees := make([]string, 0, len(invitees))
	for _, i := range invitees {
		rawInvitees = append(rawInvitees, i.String())
	}

	req := &pb.CreatePollRequest{
		UserLink:  adminLink.String(),
		BoardLink: boardLink.String(),
		CardLinks: rawCards,
		Invitees:  rawInvitees,
	}

	_, err := p.client.CreatePoll(ctx, req)
	if err != nil {
		return fmt.Errorf("PollClient.CreatePoll: %w", convertBoardGRPCError(err))
	}

	return nil
}

func (p *Poll) DeletePoll(ctx context.Context, boardLink, userLink uuid.UUID) error {
	req := &pb.DeletePollRequest{
		UserLink:  userLink.String(),
		BoardLink: boardLink.String(),
	}

	_, err := p.client.DeletePoll(ctx, req)
	if err != nil {
		return fmt.Errorf("PollClient.DeletePoll: %w", convertBoardGRPCError(err))
	}

	return nil
}

func (p *Poll) NextPollCard(ctx context.Context, boardLink, userLink uuid.UUID) error {
	req := &pb.NextPollCardRequest{
		UserLink:  userLink.String(),
		BoardLink: boardLink.String(),
	}

	_, err := p.client.NextPollCard(ctx, req)
	if err != nil {
		return fmt.Errorf("PollClient.NextPollCard: %w", convertBoardGRPCError(err))
	}

	return nil
}

func (p *Poll) VotePoll(ctx context.Context, boardLink, userLink uuid.UUID, points int) error {
	req := &pb.VotePollRequest{
		UserLink:  userLink.String(),
		BoardLink: boardLink.String(),
		Points:    int32(points),
	}

	_, err := p.client.VotePoll(ctx, req)
	if err != nil {
		return fmt.Errorf("PollClient.VotePoll: %w", convertBoardGRPCError(err))
	}

	return nil
}

