package clients

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/card/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type Card struct {
	client pb.CardServiceClient
}

func NewCardClient(connection *grpc.ClientConn) *Card {
	return &Card{
		client: pb.NewCardServiceClient(connection),
	}
}

func (c *Card) GetCard(ctx context.Context, infoCard domain.GetCardRequest) (domain.CardInfo, error) {
	req := &pb.GetCardRequest{
		UserLink: infoCard.UserLink.String(),
		CardLink: infoCard.CardLink.String(),
	}

	resp, err := c.client.GetCard(ctx, req)

	if err != nil {
		return domain.CardInfo{}, fmt.Errorf("CardClient.GetCard: %w", err)
	}

	subtasks := make([]domain.SubtaskInfo, 0, len(resp.CardInfo.Subtasks))

	for _, subtask := range resp.CardInfo.Subtasks {
		subTaskLink, err := uuid.Parse(subtask.SubtaskLink)
		if err != nil {
			return domain.CardInfo{}, common.ErrorParseLink
		}

		subtasks = append(subtasks, domain.SubtaskInfo{
			SubtaskLink: subTaskLink,
			Description: subtask.Description,
			IsDone:      subtask.IsDone,
			Position:    int(subtask.Position),
		})
	}

	return domain.CardInfo{
		CardLink:     infoCard.CardLink,
		ExecutorName: resp.CardInfo.ExecutorName,
		Title:        resp.CardInfo.Title,
		Description:  resp.CardInfo.Description,
		Deadline:     convertTimestamppbToTime(resp.CardInfo.Deadline),
		Subtasks:     subtasks,
	}, nil
}

func (c *Card) DeleteCard(ctx context.Context, infoCard domain.DeleteCardRequest) error {
	req := &pb.DeleteCardRequest{
		UserLink: infoCard.UserLink.String(),
		CardLink: infoCard.CardLink.String(),
	}

	_, err := c.client.DeleteCard(ctx, req)
	if err != nil {
		return fmt.Errorf("CardClient: %w", err)
	}

	return nil
}

func (c *Card) UpdateCard(ctx context.Context, infoCard domain.UpdateCardRequest) error

func (c *Card) ReorderCards(ctx context.Context, infoCard domain.ReorderCardsRequest) error

func (c *Card) CreateCard(ctx context.Context, infoCard domain.CreateCardRequest) (domain.CreateCardResponse, error)

func (c *Card) GetComments(ctx context.Context, infoComments domain.GetCommentsRequest) (domain.GetCommentsResponse, error)

func (c *Card) CreateComment(ctx context.Context, infoComment domain.CreateCommentRequest) (domain.CreateCommentResponse, error)

func (c *Card) DeleteComment(ctx context.Context, infoComment domain.DeleteCommentRequest) error

func (c *Card) UpdateComment(ctx context.Context, infoComment domain.UpdateCommentRequest) error

func (c *Card) CreateSubtask(ctx context.Context, infoSubtask domain.CreateSubtaskRequest) (domain.SubtaskInfo, error)

func (c *Card) UpdateSubtask(ctx context.Context, infoSubtask domain.UpdateSubtaskRequest) error

func (c *Card) DeleteSubtask(ctx context.Context, infoSubtask domain.DeleteSubtask) error
