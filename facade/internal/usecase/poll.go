package usecase

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/clients"
	"github.com/google/uuid"
)

type Poll struct {
	client *clients.Poll
}

func NewPoll(client *clients.Poll) *Poll {
	return &Poll{
		client: client,
	}
}

func (p *Poll) CreatePoll(ctx context.Context, boardLink, adminLink uuid.UUID, cards []uuid.UUID, invitees []uuid.UUID) error {
	if err := p.client.CreatePoll(ctx, boardLink, adminLink, cards, invitees); err != nil {
		return fmt.Errorf("poll.CreatePoll: %w", err)
	}

	return nil
}

func (p *Poll) DeletePoll(ctx context.Context, boardLink, userLink uuid.UUID) error {
	if err := p.client.DeletePoll(ctx, boardLink, userLink); err != nil {
		return fmt.Errorf("poll.DeletePoll: %w", err)
	}

	return nil
}

func (p *Poll) NextPollCard(ctx context.Context, boardLink, userLink uuid.UUID) error {
	if err := p.client.NextPollCard(ctx, boardLink, userLink); err != nil {
		return fmt.Errorf("poll.NextPollCard: %w", err)
	}

	return nil
}

func (p *Poll) VotePoll(ctx context.Context, boardLink, userLink uuid.UUID, points int) error {
	if err := p.client.VotePoll(ctx, boardLink, userLink, points); err != nil {
		return fmt.Errorf("poll.VotePoll: %w", err)
	}

	return nil
}
