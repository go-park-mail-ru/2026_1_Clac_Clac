package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/dto"
	handlerCommon "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type CardUsecase interface {
	GetCard(ctx context.Context, infoCard domain.GetCardRequest) (domain.CardInfo, error)
	DeleteCard(ctx context.Context, infoCard domain.DeleteCardRequest) error
	UpdateCard(ctx context.Context, infoCard domain.UpdateCardRequest) error
	ReorderCards(ctx context.Context, infoCard domain.ReorderCardsRequest) error
	CreateCard(ctx context.Context, infoCard domain.CreateCardRequest) (domain.CreateCardResponse, error)
	GetComments(ctx context.Context, infoComments domain.GetCommentsRequest) (domain.GetCommentsResponse, error)
	CreateComment(ctx context.Context, infoComment domain.CreateCommentRequest) (domain.CreateCommentResponse, error)
	DeleteComment(ctx context.Context, infoComment domain.DeleteCommentRequest) error
	UpdateComment(ctx context.Context, infoComment domain.UpdateCommentRequest) error
	CreateSubtask(ctx context.Context, infoSubtask domain.CreateSubtaskRequest) (domain.SubtaskInfo, error)
	UpdateSubtask(ctx context.Context, infoSubtask domain.UpdateSubtaskRequest) error
	DeleteSubtask(ctx context.Context, infoSubtask domain.DeleteSubtask) error
}

const (
	cardLinkKey = "link"
)

const ()

type CardConfig struct {
	MaxLenTitle       int
	MaxLenDescription int
}

type Card struct {
	card CardUsecase
	cfg  CardConfig
}

func NewCard(card CardUsecase, cfg CardConfig) *Card {
	return &Card{
		card: card,
		cfg:  cfg,
	}
}

func (c *Card) GetCard(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())
	vars := mux.Vars(r)
	linkParam := vars[cardLinkKey]

	cardLink, err := uuid.Parse(linkParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	value := r.Context().Value(middleware.UserContextLink{})
	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, msgUnauthorized)
		return
	}

	card, err := c.card.GetCard(r.Context(), domain.GetCardRequest{
		UserLink: userLink,
		CardLink: cardLink,
	})
	if err != nil {
		if errors.Is(err, handlerCommon.ErrResetTokenNotExistOrExpired) {
			api.RespondError(w, http.StatusNotFound, "dds")
			return
		}

		logger.Error().Err(err).Msg("CardHandler.GetCard")
		api.RespondError(w, http.StatusInternalServerError, "sad")
		return
	}

	subtasks := make([]dto.SubtaskResponse, 0, len(card.Subtasks))

	for _, subtask := range card.Subtasks {
		subtasks = append(subtasks, dto.SubtaskResponse{
			SubtaskLink: subtask.SubtaskLink,
			Description: subtask.Description,
			IsDone:      subtask.IsDone,
			Position:    subtask.Position,
		})
	}

	api.HandleError(api.RespondOk(w, dto.CardResponse{
		CardLink:     cardLink,
		ExecutorName: card.ExecutorName,
		Title:        card.Title,
		Description:  card.Description,
		Deadline:     card.Deadline,
		Subtasks:     subtasks,
	}))
}

func (c *Card) DeleteCard(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())
	vars := mux.Vars(r)
	linkParam := vars[cardLinkKey]

	cardLink, err := uuid.Parse(linkParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectPath)
		return
	}

	value := r.Context().Value(middleware.UserContextLink{})
	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, msgUnauthorized)
		return
	}

	err = c.card.DeleteCard(r.Context(), domain.DeleteCardRequest{
		UserLink: userLink,
		CardLink: cardLink,
	})
	if err != nil {
		if errors.Is(err, common.ErrorNotExistingCard) {
			api.RespondError(w, http.StatusNotFound, failFindCard)
			return
		}

		logger.Error().Err(err).Msg("CardHandler.DeleteCard")
		api.RespondError(w, http.StatusInternalServerError, failDeleteCard)
		return
	}

	api.RespondOk(w, api.StatusOK)
}

func (c *Card) UpdateCardDetails(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())
	vars := mux.Vars(r)
	linkCardParam := vars[cardLinkKey]

	linkCard, err := uuid.Parse(linkCardParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectPath)
		return
	}

	var updatingInfo dto.UpdatingCardDetails

	err = json.NewDecoder(r.Body).Decode(&updatingInfo)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectRequest)
		return
	}

	err = common.ValidateTextInfo(updatingInfo.Title, c.cfg.MaxLenTitle)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, fmt.Sprintf("max len title is %d", h.cnf.MaxLenTitle))
		return
	}

	err = common.ValidateTextInfo(updatingInfo.Description, h.cnf.MaxLenDescription)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, fmt.Sprintf("max len description is %d", h.cnf.MaxLenDescription))
		return
	}

	err = c.card.UpdateCard(r.Context(), serviceDto.UpdatingCardDetails{
		LinkCard:     linkCard,
		Description:  updatingInfo.Description,
		Title:        updatingInfo.Title,
		LinkExecuter: updatingInfo.LinkExecuter,
		DataDeadLine: updatingInfo.DataDeadLine,
	})

	if err != nil {
		if errors.Is(err, common.ErrorNotExistingCard) {
			api.RespondError(w, http.StatusNotFound, failFindCard)
			return
		}

		if errors.Is(err, common.ErrorInvalidReferenceCardData) {
			api.RespondError(w, http.StatusBadRequest, incorrectReferences)
			return
		}

		if errors.Is(err, common.ErrorInvalidCardData) {
			api.RespondError(w, http.StatusBadRequest, invalidCardData)
			return
		}

		if errors.Is(err, common.ErrorMissingRequiredField) {
			api.RespondError(w, http.StatusBadRequest, failNullValue)
			return
		}

		logger.Error().Err(err).Msg("CardHandler.UpdateCardDetails")
		api.RespondError(w, http.StatusInternalServerError, failUpdateCard)
		return
	}

	api.RespondOk(w, api.StatusOK)
}
