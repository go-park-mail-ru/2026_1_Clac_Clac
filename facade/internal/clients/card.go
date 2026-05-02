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
		return domain.CardInfo{}, fmt.Errorf("CardClient.GetCard: %w", convertCardGRPCError(err))
	}

	subtasks := make([]domain.SubtaskInfo, 0, len(resp.CardInfo.Subtasks))
	for _, subtask := range resp.CardInfo.Subtasks {
		subtaskLink, err := uuid.Parse(subtask.SubtaskLink)
		if err != nil {
			return domain.CardInfo{}, common.ErrorParseLink
		}
		subtasks = append(subtasks, domain.SubtaskInfo{
			SubtaskLink: subtaskLink,
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
		return fmt.Errorf("CardClient.DeleteCard: %w", convertCardGRPCError(err))
	}

	return nil
}

func (c *Card) UpdateCard(ctx context.Context, infoCard domain.UpdateCardRequest) error {
	var executorLink *string
	if infoCard.ExecutorLink != nil {
		s := infoCard.ExecutorLink.String()
		executorLink = &s
	}

	req := &pb.UpdateCardRequest{
		UserLink:     infoCard.UserLink.String(),
		CardLink:     infoCard.CardLink.String(),
		ExecutorLink: executorLink,
		Title:        infoCard.Title,
		Description:  infoCard.Description,
		Deadline:     convertTimeToTimestamppb(infoCard.Deadline),
	}

	_, err := c.client.UpdateCard(ctx, req)
	if err != nil {
		return fmt.Errorf("CardClient.UpdateCard: %w", convertCardGRPCError(err))
	}

	return nil
}

func (c *Card) ReorderCards(ctx context.Context, infoCard domain.ReorderCardsRequest) error {
	req := &pb.ReorderCardsRequest{
		UserLink:    infoCard.UserLink.String(),
		CardLink:    infoCard.CardLink.String(),
		SectionLink: infoCard.SectionLink.String(),
		Position:    int64(infoCard.Position),
	}

	_, err := c.client.ReorderCards(ctx, req)
	if err != nil {
		return fmt.Errorf("CardClient.ReorderCards: %w", convertCardGRPCError(err))
	}

	return nil
}

func (c *Card) CreateCard(ctx context.Context, infoCard domain.CreateCardRequest) (domain.CreateCardResponse, error) {
	var executorLink *string
	if infoCard.ExecutorLink != nil {
		s := infoCard.ExecutorLink.String()
		executorLink = &s
	}

	req := &pb.CreateCardRequest{
		UserLink:     infoCard.UserLink.String(),
		SectionLink:  infoCard.SectionLink.String(),
		Title:        infoCard.Title,
		Description:  infoCard.Description,
		ExecutorLink: executorLink,
		Deadline:     convertTimeToTimestamppb(infoCard.Deadline),
	}

	resp, err := c.client.CreateCard(ctx, req)
	if err != nil {
		return domain.CreateCardResponse{}, fmt.Errorf("CardClient.CreateCard: %w", convertCardGRPCError(err))
	}

	cardLink, err := uuid.Parse(resp.CardLink)
	if err != nil {
		return domain.CreateCardResponse{}, common.ErrorParseLink
	}

	sectionLink, err := uuid.Parse(resp.SectionLink)
	if err != nil {
		return domain.CreateCardResponse{}, common.ErrorParseLink
	}

	return domain.CreateCardResponse{
		CardLink:    cardLink,
		SectionLink: sectionLink,
		Position:    int(resp.Position),
	}, nil
}

func (c *Card) GetComments(ctx context.Context, infoComments domain.GetCommentsRequest) (domain.GetCommentsResponse, error) {
	req := &pb.GetCommentsRequest{
		UserLink: infoComments.UserLink.String(),
		CardLink: infoComments.CardLink.String(),
	}

	resp, err := c.client.GetComments(ctx, req)
	if err != nil {
		return domain.GetCommentsResponse{}, fmt.Errorf("CardClient.GetComments: %w", convertCardGRPCError(err))
	}

	comments := make([]domain.CommentInfo, 0, len(resp.CommentsInfo))
	for _, comment := range resp.CommentsInfo {
		commentLink, err := uuid.Parse(comment.CommentLink)
		if err != nil {
			return domain.GetCommentsResponse{}, common.ErrorParseLink
		}

		authorLink, err := uuid.Parse(comment.AuthorLink)
		if err != nil {
			return domain.GetCommentsResponse{}, common.ErrorParseLink
		}

		var parentLink *uuid.UUID
		if comment.ParentLink != nil && *comment.ParentLink != "" {
			p, err := uuid.Parse(*comment.ParentLink)
			if err != nil {
				return domain.GetCommentsResponse{}, common.ErrorParseLink
			}
			parentLink = &p
		}

		comments = append(comments, domain.CommentInfo{
			CommentLink: commentLink,
			ParentLink:  parentLink,
			AuthorLink:  authorLink,
			Text:        comment.Text,
		})
	}

	return domain.GetCommentsResponse{CommentsInfo: comments}, nil
}

func (c *Card) CreateComment(ctx context.Context, infoComment domain.CreateCommentRequest) (domain.CreateCommentResponse, error) {
	var parentLink *string
	if infoComment.ParentLink != nil {
		s := infoComment.ParentLink.String()
		parentLink = &s
	}

	req := &pb.CreateCommentRequest{
		UserLink:   infoComment.UserLink.String(),
		CardLink:   infoComment.CardLink.String(),
		ParentLink: parentLink,
		Text:       infoComment.Text,
	}

	resp, err := c.client.CreateComment(ctx, req)
	if err != nil {
		return domain.CreateCommentResponse{}, fmt.Errorf("CardClient.CreateComment: %w", convertCardGRPCError(err))
	}

	commentLink, err := uuid.Parse(resp.CommentLink)
	if err != nil {
		return domain.CreateCommentResponse{}, common.ErrorParseLink
	}

	return domain.CreateCommentResponse{CommentLink: commentLink}, nil
}

func (c *Card) DeleteComment(ctx context.Context, infoComment domain.DeleteCommentRequest) error {
	req := &pb.DeleteCommentRequest{
		UserLink:    infoComment.UserLink.String(),
		CommentLink: infoComment.CommentLink.String(),
	}

	_, err := c.client.DeleteComment(ctx, req)
	if err != nil {
		return fmt.Errorf("CardClient.DeleteComment: %w", convertCardGRPCError(err))
	}

	return nil
}

func (c *Card) UpdateComment(ctx context.Context, infoComment domain.UpdateCommentRequest) error {
	req := &pb.UpdateCommentRequest{
		UserLink:    infoComment.UserLink.String(),
		CommentLink: infoComment.CommentLink.String(),
		Text:        infoComment.Text,
	}

	_, err := c.client.UpdateComment(ctx, req)
	if err != nil {
		return fmt.Errorf("CardClient.UpdateComment: %w", convertCardGRPCError(err))
	}

	return nil
}

func (c *Card) CreateSubtask(ctx context.Context, infoSubtask domain.CreateSubtaskRequest) (domain.SubtaskInfo, error) {
	req := &pb.CreateSubtaskRequest{
		UserLink:    infoSubtask.UserLink.String(),
		CardLink:    infoSubtask.CardLink.String(),
		Description: infoSubtask.Description,
	}

	resp, err := c.client.CreateSubtask(ctx, req)
	if err != nil {
		return domain.SubtaskInfo{}, fmt.Errorf("CardClient.CreateSubtask: %w", convertCardGRPCError(err))
	}

	subtaskLink, err := uuid.Parse(resp.SubtaskLink)
	if err != nil {
		return domain.SubtaskInfo{}, common.ErrorParseLink
	}

	return domain.SubtaskInfo{
		SubtaskLink: subtaskLink,
		Description: resp.Description,
		IsDone:      resp.IsDone,
		Position:    int(resp.Position),
	}, nil
}

func (c *Card) UpdateSubtask(ctx context.Context, infoSubtask domain.UpdateSubtaskRequest) error {
	req := &pb.UpdateSubtaskRequest{
		UserLink:    infoSubtask.UserLink.String(),
		SubtaskLink: infoSubtask.SubtaskLink.String(),
		IsDone:      infoSubtask.IsDone,
		Description: infoSubtask.Description,
	}

	_, err := c.client.UpdateSubtask(ctx, req)
	if err != nil {
		return fmt.Errorf("CardClient.UpdateSubtask: %w", convertCardGRPCError(err))
	}

	return nil
}

func (c *Card) DeleteSubtask(ctx context.Context, infoSubtask domain.DeleteSubtask) error {
	req := &pb.DeleteSubtaskRequest{
		UserLink:    infoSubtask.UserLink.String(),
		SubtaskLink: infoSubtask.SubtaskLink.String(),
	}

	_, err := c.client.DeleteSubtask(ctx, req)
	if err != nil {
		return fmt.Errorf("CardClient.DeleteSubtask: %w", convertCardGRPCError(err))
	}

	return nil
}
