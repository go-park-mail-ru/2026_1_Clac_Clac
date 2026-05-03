package delivery

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/delivery/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/models"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/service/dto"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/boardRbac"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/card/v1"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

var (
	ErrInvalidCardLink         = errors.New("invalid card link")
	ErrCannotGetCard           = errors.New("cannot get card")
	ErrCannotDeleteCard        = errors.New("cannot delete card")
	ErrInvalidExecutorLink     = errors.New("invalid executor link")
	ErrCardNameIsTooBig        = errors.New("card name is too big")
	ErrCardDescriptionIsTooBig = errors.New("card description is too big")
	ErrCannotUpdateCard        = errors.New("cannot update card")
	ErrInvalidSectionLink      = errors.New("invalid section link")
	ErrCannotReorderCards      = errors.New("cannot reorder cards")
	ErrInvalidAuthorLink       = errors.New("invalid author link")
	ErrCannotCreateCard        = errors.New("cannot create card")

	ErrCannotGetComments   = errors.New("cannot get comments")
	ErrCannotDeleteComment = errors.New("cannot delete comment")
	ErrInvalidParentLink   = errors.New("invalid parent link")
	ErrCannotCreateComment = errors.New("cannot create comment")
	ErrCannotUpdateComment = errors.New("cannot update comment")

	ErrInvalidUserLink = errors.New("invalid user link")

	ErrInvalidSubtaskLink  = errors.New("invalid subtask link")
	ErrCannotCreateSubtask = errors.New("cannot create subtask")
	ErrCannotDeleteSubtask = errors.New("cannot delete subtask")
	ErrCannotUpdateSubtask = errors.New("cannot upadate subtask")
)

//go:generate mockery --name=CardService --output mock_card_srv
type CardService interface {
	GetCard(ctx context.Context, linkCard uuid.UUID, userLink uuid.UUID) (serviceDto.InfoCard, error)
	DeleteCard(ctx context.Context, linkCard uuid.UUID, userLink uuid.UUID) error
	UpdateCardDetails(ctx context.Context, updatedCard serviceDto.UpdatingCardDetails, userLink uuid.UUID) error
	ReorderCard(ctx context.Context, updatingPlaceCard serviceDto.PlaceCard, userLink uuid.UUID) error
	CreateCard(ctx context.Context, newCard serviceDto.NewCard) (serviceDto.PlaceCard, error)
	GetComments(ctx context.Context, cardLink uuid.UUID, userLink uuid.UUID) ([]serviceDto.CommentInfo, error)
	CreateComment(ctx context.Context, createCardInfo serviceDto.CreateCommentInfo) (serviceDto.CommentInfo, error)
	DeleteComment(ctx context.Context, commentLink uuid.UUID, userLink uuid.UUID) error
	UpdateComment(ctx context.Context, updateCommentInfo serviceDto.UpdateCommentInfo) error
	CreateSubtask(ctx context.Context, createInfo serviceDto.CreateSubtaskInfo, userLink uuid.UUID) (models.SubtaskInfo, error)
	DeleteSubtask(ctx context.Context, deleteInfo serviceDto.DeleteSubtask, userLink uuid.UUID) error
	UpdateSubtask(ctx context.Context, updateInfo serviceDto.UpdateSubtask, userLink uuid.UUID) error
}

type Config struct {
	MaxLenTitle       int
	MaxLenDescription int
}

type CardHandler struct {
	pb.UnimplementedCardServiceServer

	srv CardService
	cnf Config
}

func NewHandler(srv CardService, cnf Config) *CardHandler {
	return &CardHandler{
		srv: srv,
		cnf: cnf,
	}
}

func (h *CardHandler) GetCard(ctx context.Context, req *pb.GetCardRequest) (*pb.GetCardResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawCardLink := req.GetCardLink()
	cardLink, err := uuid.Parse(rawCardLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidCardLink.Error())
	}

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidUserLink.Error())
	}

	card, err := h.srv.GetCard(ctx, cardLink, userLink)
	if err != nil {
		switch {
		case errors.Is(err, rbac.ErrActionDenied):
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		case errors.Is(err, common.ErrCardNotFound):
			return nil, status.Error(codes.NotFound, common.ErrCardNotFound.Error())
		}

		logger.Error().Err(err).Msg("CardHandler.GetCard")
		return nil, status.Error(codes.Internal, ErrCannotGetCard.Error())
	}

	var deadline *timestamppb.Timestamp
	if card.DataDeadLine != nil {
		deadline = timestamppb.New(*card.DataDeadLine)
	}

	subtasks := make([]*pb.SubtaskInfo, 0, len(card.Subtasks))

	for _, subtask := range card.Subtasks {
		subtasks = append(subtasks, &pb.SubtaskInfo{
			SubtaskLink: subtask.SubtaskLink.String(),
			Description: subtask.Description,
			IsDone:      subtask.IsDone,
			Position:    int64(subtask.Position),
		})
	}

	var executorLink *string
	if card.ExecutorLink != nil {
		s := card.ExecutorLink.String()
		executorLink = &s
	}

	return &pb.GetCardResponse{
		CardInfo: &pb.CardInfo{
			Link:         cardLink.String(),
			Title:        card.Title,
			Description:  card.Description,
			ExecutorLink: executorLink,
			Deadline:     deadline,
			Subtasks:    subtasks,
			Position:   int64(card.Position),
		},
	}, nil
}

func (h *CardHandler) DeleteCard(ctx context.Context, req *pb.DeleteCardRequest) (*pb.DeleteCardResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawCardLink := req.GetCardLink()
	cardLink, err := uuid.Parse(rawCardLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidCardLink.Error())
	}

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidUserLink.Error())
	}

	err = h.srv.DeleteCard(ctx, cardLink, userLink)
	if err != nil {
		switch {
		case errors.Is(err, rbac.ErrActionDenied):
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		case errors.Is(err, common.ErrCardNotFound):
			return nil, status.Error(codes.NotFound, common.ErrCardNotFound.Error())
		}

		logger.Error().Err(err).Msg("CardHandler.DeleteCard")
		return nil, status.Error(codes.Internal, ErrCannotDeleteCard.Error())
	}

	return &pb.DeleteCardResponse{}, nil
}

func (h *CardHandler) UpdateCard(ctx context.Context, req *pb.UpdateCardRequest) (*pb.UpdateCardResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawCardLink := req.GetCardLink()
	cardLink, err := uuid.Parse(rawCardLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidCardLink.Error())
	}

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidUserLink.Error())
	}

	var executorLink *uuid.UUID
	if req.ExecutorLink != nil {
		rawExecutorLink := req.GetExecutorLink()
		parsed, err := uuid.Parse(rawExecutorLink)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, ErrInvalidExecutorLink.Error())
		}
		executorLink = &parsed
	}

	var deadline *time.Time
	if req.Deadline != nil {
		d := req.GetDeadline().AsTime()
		deadline = &d
	}

	updatingInfo := dto.UpdatingCardDetails{
		LinkCard:     cardLink,
		Title:        req.GetTitle(),
		Description:  req.GetDescription(),
		LinkExecutor: executorLink,
		DataDeadLine: deadline,
	}

	if !common.CheckCardNameLength(updatingInfo.Title, h.cnf.MaxLenTitle) {
		return nil, status.Error(codes.InvalidArgument, ErrCardNameIsTooBig.Error())
	}

	if !common.CheckCardDescriptionLength(updatingInfo.Description, h.cnf.MaxLenDescription) {
		return nil, status.Error(codes.InvalidArgument, ErrCardDescriptionIsTooBig.Error())
	}

	err = h.srv.UpdateCardDetails(ctx, serviceDto.UpdatingCardDetails{
		LinkCard:     cardLink,
		Description:  updatingInfo.Description,
		Title:        updatingInfo.Title,
		LinkExecutor: updatingInfo.LinkExecutor,
		DataDeadLine: updatingInfo.DataDeadLine,
	}, userLink)
	if err != nil {
		switch {
		case errors.Is(err, rbac.ErrActionDenied):
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		case errors.Is(err, common.ErrCardNotFound):
			return nil, status.Error(codes.NotFound, common.ErrCardNotFound.Error())
		case errors.Is(err, common.ErrInvalidReferenceCardData):
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidReferenceCardData.Error())
		case errors.Is(err, common.ErrInvalidCardData):
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidCardData.Error())
		case errors.Is(err, common.ErrMissingRequiredField):
			return nil, status.Error(codes.InvalidArgument, common.ErrMissingRequiredField.Error())
		}

		logger.Error().Err(err).Msg("CardHandler.UpdateCardDetails")
		return nil, status.Error(codes.Internal, ErrCannotUpdateCard.Error())
	}

	return &pb.UpdateCardResponse{}, nil
}

func (h *CardHandler) ReorderCards(ctx context.Context, req *pb.ReorderCardsRequest) (*pb.ReorderCardsResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawCardLink := req.GetCardLink()
	cardLink, err := uuid.Parse(rawCardLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidCardLink.Error())
	}

	rawSectionLink := req.GetSectionLink()
	sectionLink, err := uuid.Parse(rawSectionLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidSectionLink.Error())
	}

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidUserLink.Error())
	}

	err = h.srv.ReorderCard(ctx, serviceDto.PlaceCard{
		LinkCard:    cardLink,
		LinkSection: sectionLink,
		Position:    int(req.GetPosition()),
	}, userLink)
	if err != nil {
		switch {
		case errors.Is(err, rbac.ErrActionDenied):
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		case errors.Is(err, common.ErrCannotSkipMandatorySection):
			return nil, status.Error(codes.InvalidArgument, common.ErrCannotSkipMandatorySection.Error())
		case errors.Is(err, common.ErrCardNotFound):
			return nil, status.Error(codes.NotFound, common.ErrCardNotFound.Error())
		case errors.Is(err, common.ErrInvalidReferenceCardData):
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidReferenceCardData.Error())
		case errors.Is(err, common.ErrInvalidCardData):
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidCardData.Error())
		case errors.Is(err, common.ErrMissingRequiredField):
			return nil, status.Error(codes.InvalidArgument, common.ErrMissingRequiredField.Error())
		case errors.Is(err, common.ErrTaskLimitReached):
			return nil, status.Error(codes.InvalidArgument, common.ErrTaskLimitReached.Error())
		}

		logger.Error().Err(err).Msg("CardHandler.ReorderCard")
		return nil, status.Error(codes.Internal, ErrCannotReorderCards.Error())
	}

	return &pb.ReorderCardsResponse{}, nil
}

func (h *CardHandler) CreateCard(ctx context.Context, req *pb.CreateCardRequest) (*pb.CreateCardResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidUserLink.Error())
	}

	rawSectionLink := req.GetSectionLink()
	sectionLink, err := uuid.Parse(rawSectionLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidSectionLink.Error())
	}

	title := req.GetTitle()
	if !common.CheckCardNameLength(title, h.cnf.MaxLenTitle) {
		return nil, status.Error(codes.InvalidArgument, ErrCardNameIsTooBig.Error())
	}

	description := req.GetDescription()
	if !common.CheckCardDescriptionLength(description, h.cnf.MaxLenDescription) {
		return nil, status.Error(codes.InvalidArgument, ErrCardDescriptionIsTooBig.Error())
	}

	var executorLink *uuid.UUID
	if req.ExecutorLink != nil {
		rawExecutorLink := req.GetExecutorLink()
		parsed, err := uuid.Parse(rawExecutorLink)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, ErrInvalidExecutorLink.Error())
		}
		executorLink = &parsed
	}

	var deadline *time.Time
	if req.Deadline != nil {
		d := req.GetDeadline().AsTime()
		deadline = &d
	}

	card, err := h.srv.CreateCard(ctx, serviceDto.NewCard{
		LinkAuthor:   userLink,
		Title:        title,
		Description:  description,
		LinkExecutor: executorLink,
		DataDeadLine: deadline,
		LinkSection:  sectionLink,
	})
	if err != nil {
		switch {
		case errors.Is(err, rbac.ErrActionDenied):
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		case errors.Is(err, common.ErrSectionNotFound):
			return nil, status.Error(codes.NotFound, common.ErrSectionNotFound.Error())
		case errors.Is(err, common.ErrCardAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, common.ErrCardAlreadyExists.Error())
		case errors.Is(err, common.ErrInvalidReferenceCardData):
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidReferenceCardData.Error())
		case errors.Is(err, common.ErrInvalidCardData):
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidCardData.Error())
		case errors.Is(err, common.ErrMissingRequiredField):
			return nil, status.Error(codes.InvalidArgument, common.ErrMissingRequiredField.Error())
		case errors.Is(err, common.ErrTaskLimitReached):
			return nil, status.Error(codes.InvalidArgument, common.ErrTaskLimitReached.Error())
		}

		logger.Error().Err(err).Msg("CardHandler.CreateCard")
		return nil, status.Error(codes.Internal, ErrCannotCreateCard.Error())
	}

	return &pb.CreateCardResponse{
		CardLink:    card.LinkCard.String(),
		SectionLink: card.LinkSection.String(),
		Position:    int64(card.Position),
	}, nil
}

func (h *CardHandler) GetComments(ctx context.Context, req *pb.GetCommentsRequest) (*pb.GetCommentsResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawCardLink := req.GetCardLink()
	cardLink, err := uuid.Parse(rawCardLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidCardLink.Error())
	}

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidUserLink.Error())
	}

	comments, err := h.srv.GetComments(ctx, cardLink, userLink)
	if err != nil {
		switch {
		case errors.Is(err, rbac.ErrActionDenied):
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		case errors.Is(err, common.ErrCardNotFound):
			return nil, status.Error(codes.NotFound, common.ErrCardNotFound.Error())
		}

		logger.Error().Err(err).Msg("CardService.GetComments")
		return nil, status.Error(codes.Internal, ErrCannotGetComments.Error())
	}

	commentsInfo := make([]*pb.CommentInfo, 0)
	for _, comment := range comments {
		var parentLink string
		if comment.ParentLink != nil {
			parentLink = comment.ParentLink.String()
		}

		commentsInfo = append(commentsInfo, &pb.CommentInfo{
			CommentLink: comment.Link.String(),
			ParentLink:  &parentLink,
			AuthorLink:  comment.AuthorLink.String(),
			Text:        comment.Text,
		})
	}

	return &pb.GetCommentsResponse{
		CommentsInfo: commentsInfo,
	}, nil
}

func (h *CardHandler) CreateComment(ctx context.Context, req *pb.CreateCommentRequest) (*pb.CreateCommentResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawCardLink := req.GetCardLink()
	cardLink, err := uuid.Parse(rawCardLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidCardLink.Error())
	}

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidUserLink.Error())
	}

	var parentLink *uuid.UUID
	if req.ParentLink != nil {
		p, err := uuid.Parse(req.GetParentLink())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, ErrInvalidParentLink.Error())
		}
		parentLink = &p
	}

	comment, err := h.srv.CreateComment(ctx, serviceDto.CreateCommentInfo{
		CardLink:   cardLink,
		ParentLink: parentLink,
		AuthorLink: userLink,
		Text:       req.GetText(),
	})
	if err != nil {
		switch {
		case errors.Is(err, rbac.ErrActionDenied):
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		case errors.Is(err, common.ErrMissingRequiredField):
			return nil, status.Error(codes.InvalidArgument, common.ErrMissingRequiredField.Error())
		case errors.Is(err, common.ErrInvalidReferenceCardData):
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidReferenceCardData.Error())
		}

		logger.Error().Err(err).Msg("CardService.CreateComment")
		return nil, status.Error(codes.Internal, ErrCannotCreateComment.Error())
	}

	return &pb.CreateCommentResponse{
		CommentLink: comment.Link.String(),
	}, nil
}

func (h *CardHandler) DeleteComment(ctx context.Context, req *pb.DeleteCommentRequest) (*pb.DeleteCommentResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawCommentLink := req.GetCommentLink()
	commentLink, err := uuid.Parse(rawCommentLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidCardLink.Error())
	}

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidUserLink.Error())
	}

	err = h.srv.DeleteComment(ctx, commentLink, userLink)
	if err != nil {
		switch {
		case errors.Is(err, rbac.ErrActionDenied):
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		case errors.Is(err, common.ErrCommentNotFound):
			return nil, status.Error(codes.NotFound, common.ErrCommentNotFound.Error())
		case errors.Is(err, common.ErrPermissionDenied):
			return nil, status.Error(codes.PermissionDenied, common.ErrPermissionDenied.Error())
		}

		logger.Error().Err(err).Msg("CardService.DeleteComment")
		return nil, status.Error(codes.Internal, ErrCannotDeleteComment.Error())
	}

	return &pb.DeleteCommentResponse{}, nil
}

func (h *CardHandler) UpdateComment(ctx context.Context, req *pb.UpdateCommentRequest) (*pb.UpdateCommentResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawCommentLink := req.GetCommentLink()
	commentLink, err := uuid.Parse(rawCommentLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidCardLink.Error())
	}

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidUserLink.Error())
	}

	err = h.srv.UpdateComment(ctx, serviceDto.UpdateCommentInfo{
		CommentLink: commentLink,
		UserLink:    userLink,
		Text:        req.GetText(),
	})
	if err != nil {
		switch {
		case errors.Is(err, rbac.ErrActionDenied):
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		case errors.Is(err, common.ErrCommentNotFound):
			return nil, status.Error(codes.NotFound, common.ErrCommentNotFound.Error())
		case errors.Is(err, common.ErrPermissionDenied):
			return nil, status.Error(codes.PermissionDenied, common.ErrPermissionDenied.Error())
		}

		logger.Error().Err(err).Msg("CardService.UpdateComment")
		return nil, status.Error(codes.Internal, ErrCannotUpdateComment.Error())
	}

	return &pb.UpdateCommentResponse{}, nil
}

func (h *CardHandler) CreateSubtask(ctx context.Context, req *pb.CreateSubtaskRequest) (*pb.CreateSubtaskResponse, error) {
	logger := zerolog.Ctx(ctx)

	taskLink, err := uuid.Parse(req.CardLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidCardLink.Error())
	}

	userLink, err := uuid.Parse(req.UserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidSubtaskLink.Error())
	}

	logger.Info().Str("task link", taskLink.String()).Str("user link", userLink.String()).Msg("create subtask")

	subtask, err := h.srv.CreateSubtask(ctx, serviceDto.CreateSubtaskInfo{
		TaskLink:    taskLink,
		Description: req.Description,
	}, userLink)
	if err != nil {
		logger.Error().Err(err).Msg("CardService.CreateSubtask")

		switch {
		case errors.Is(err, common.ErrMissingRequiredField):
			return nil, status.Error(codes.InvalidArgument, common.ErrMissingRequiredField.Error())
		case errors.Is(err, common.ErrInvalidReferenceCardData):
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidReferenceCardData.Error())
		}

		return nil, status.Error(codes.Internal, ErrCannotCreateSubtask.Error())
	}

	return &pb.CreateSubtaskResponse{
		SubtaskLink: subtask.SubtaskLink.String(),
		Description: subtask.Description,
		IsDone:      subtask.IsDone,
		Position:    int64(subtask.Position),
	}, nil
}

func (h *CardHandler) DeleteSubtask(ctx context.Context, req *pb.DeleteSubtaskRequest) (*pb.DeleteSubtaskResponse, error) {
	logger := zerolog.Ctx(ctx)

	subtaskLink, err := uuid.Parse(req.SubtaskLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidSubtaskLink.Error())
	}

	userLink, err := uuid.Parse(req.UserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidSubtaskLink.Error())
	}

	err = h.srv.DeleteSubtask(ctx, serviceDto.DeleteSubtask{
		SubTaskLink: subtaskLink,
	}, userLink)
	if err != nil {
		if errors.Is(err, common.ErrSubtaskNotFound) {
			return nil, status.Error(codes.NotFound, common.ErrSubtaskNotFound.Error())
		}

		logger.Error().Err(err).Msg("CardService.DeleteSubtask")
		return nil, status.Error(codes.Internal, ErrCannotDeleteSubtask.Error())
	}

	return &pb.DeleteSubtaskResponse{}, nil
}

func (h *CardHandler) UpdateSubtask(ctx context.Context, req *pb.UpdateSubtaskRequest) (*pb.UpdateSubtaskResponse, error) {
	logger := zerolog.Ctx(ctx)

	subtaskLink, err := uuid.Parse(req.SubtaskLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidSubtaskLink.Error())
	}

	userLink, err := uuid.Parse(req.UserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidSubtaskLink.Error())
	}

	err = h.srv.UpdateSubtask(ctx, serviceDto.UpdateSubtask{
		SubTaskLink: subtaskLink,
		IsDone:      req.IsDone,
		Description: req.Description,
	}, userLink)
	if err != nil {
		if errors.Is(err, common.ErrSubtaskNotFound) {
			return nil, status.Error(codes.NotFound, common.ErrSubtaskNotFound.Error())
		}

		logger.Error().Err(err).Msg("CardService.UpdateSubtask")
		return nil, status.Error(codes.Internal, ErrCannotUpdateSubtask.Error())
	}

	return &pb.UpdateSubtaskResponse{}, nil
}
