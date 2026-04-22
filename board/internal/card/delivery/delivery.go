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
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/service/dto"
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
)

//go:generate mockery --name=CardService --output mock_card_srv
type CardService interface {
	GetCard(ctx context.Context, linkCard uuid.UUID) (serviceDto.InfoCard, error)
	DeleteCard(ctx context.Context, linkCard uuid.UUID) error
	UpdateCardDetails(ctx context.Context, updatedCard serviceDto.UpdatingCardDetails) error
	ReorderCard(ctx context.Context, updatingPlaceCard serviceDto.PlaceCard) error
	CreateCard(ctx context.Context, newCard serviceDto.NewCard) (serviceDto.PlaceCard, error)
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

	card, err := h.srv.GetCard(ctx, cardLink)
	if err != nil {
		if errors.Is(err, common.ErrCardNotFound) {
			return nil, status.Error(codes.NotFound, common.ErrCardNotFound.Error())
		}

		logger.Error().Err(err).Msg("CardHandler.GetCard")
		return nil, status.Error(codes.Internal, ErrCannotGetCard.Error())
	}

	return &pb.GetCardResponse{
		CardInfo: &pb.CardInfo{
			Link:         cardLink.String(),
			Title:        card.Title,
			Description:  card.Description,
			ExecuterName: card.NameExecuter,
			Deadline:     timestamppb.New(*card.DataDeadLine),
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

	err = h.srv.DeleteCard(ctx, cardLink)
	if err != nil {
		if errors.Is(err, common.ErrCardNotFound) {
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

	var executorLink uuid.UUID
	if req.ExecutorLink != nil {
		rawExecutorLink := req.GetExecutorLink()
		executorLink, err = uuid.Parse(rawExecutorLink)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, ErrInvalidExecutorLink.Error())
		}
	}

	var deadline time.Time
	if req.Deadline != nil {
		deadline = req.GetDeadline().AsTime()
	}

	updatingInfo := dto.UpdatingCardDetails{
		LinkCard:     cardLink,
		Title:        req.GetTitle(),
		Description:  req.GetDescription(),
		LinkExecuter: &executorLink,
		DataDeadLine: &deadline,
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
		LinkExecuter: updatingInfo.LinkExecuter,
		DataDeadLine: updatingInfo.DataDeadLine,
	})
	if err != nil {
		if errors.Is(err, common.ErrCardNotFound) {
			return nil, status.Error(codes.NotFound, common.ErrCardNotFound.Error())
		}

		if errors.Is(err, common.ErrInvalidReferenceCardData) {
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidReferenceCardData.Error())
		}

		if errors.Is(err, common.ErrInvalidCardData) {
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidCardData.Error())
		}

		if errors.Is(err, common.ErrMissingRequiredField) {
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

	err = h.srv.ReorderCard(ctx, serviceDto.PlaceCard{
		LinkCard:    cardLink,
		LinkSection: sectionLink,
		Position:    int(req.GetPosition()),
	})
	if err != nil {
		if errors.Is(err, common.ErrCannotSkipMandatorySection) {
			return nil, status.Error(codes.InvalidArgument, common.ErrCannotSkipMandatorySection.Error())
		}

		if errors.Is(err, common.ErrCardNotFound) {
			return nil, status.Error(codes.NotFound, common.ErrCardNotFound.Error())
		}

		if errors.Is(err, common.ErrInvalidReferenceCardData) {
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidReferenceCardData.Error())
		}

		if errors.Is(err, common.ErrInvalidCardData) {
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidCardData.Error())
		}

		if errors.Is(err, common.ErrMissingRequiredField) {
			return nil, status.Error(codes.InvalidArgument, common.ErrMissingRequiredField.Error())
		}

		if errors.Is(err, common.ErrTaskLimitReached) {
			return nil, status.Error(codes.InvalidArgument, common.ErrTaskLimitReached.Error())
		}

		logger.Error().Err(err).Msg("CardHandler.ReorderCard")
		return nil, status.Error(codes.Internal, ErrCannotReorderCards.Error())
	}

	return &pb.ReorderCardsResponse{}, nil
}

func (h *CardHandler) CreateCard(ctx context.Context, req *pb.CreateCardRequest) (*pb.CreateCardResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawAuthorLink := req.GetAuthorLink()
	authorLink, err := uuid.Parse(rawAuthorLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidAuthorLink.Error())
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

	var executorLink uuid.UUID
	if req.ExecutorLink != nil {
		rawExecutorLink := req.GetExecutorLink()
		executorLink, err = uuid.Parse(rawExecutorLink)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, ErrInvalidExecutorLink.Error())
		}
	}

	var deadline time.Time
	if req.Deadline != nil {
		deadline = req.GetDeadline().AsTime()
	}

	card, err := h.srv.CreateCard(ctx, serviceDto.NewCard{
		LinkAuthor:   authorLink,
		Title:        title,
		Description:  description,
		LinkExecuter: &executorLink,
		DataDeadLine: &deadline,
		LinkSection:  sectionLink,
	})
	if err != nil {
		if errors.Is(err, common.ErrSectionNotFound) {
			return nil, status.Error(codes.NotFound, common.ErrSectionNotFound.Error())
		}

		if errors.Is(err, common.ErrCardAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, common.ErrCardAlreadyExists.Error())
		}

		if errors.Is(err, common.ErrInvalidReferenceCardData) {
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidReferenceCardData.Error())
		}

		if errors.Is(err, common.ErrInvalidCardData) {
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidCardData.Error())
		}

		if errors.Is(err, common.ErrMissingRequiredField) {
			return nil, status.Error(codes.InvalidArgument, common.ErrMissingRequiredField.Error())
		}

		if errors.Is(err, common.ErrTaskLimitReached) {
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
