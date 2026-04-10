package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/card/handler/dto"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/card/service/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	failFindCard    = "can not find card"
	failGetCard     = "can not get info card"
	failDeleteCard  = "can not delete card"
	failUpdateCard  = "can not update card"
	failReorderCard = "can not reorder card"
	failCreateCard  = "can not create new card"
	failFindSection = "can not find section"

	incorectMoveCard = "can not skip mandatory section"

	cardLinkKey = "link"
)

type CardService interface {
	GetCard(ctx context.Context, linkCard uuid.UUID) (serviceDto.InfoCard, error)
	DeleteCard(ctx context.Context, linkCard uuid.UUID) error
	UpdateCardDetails(ctx context.Context, updatedCard serviceDto.UpdatingCardDetails) error
	ReorderCard(ctx context.Context, updatingPlaceCard serviceDto.PlaceCard) error
	CreateCard(ctx context.Context, newCard serviceDto.NewCard) (serviceDto.PlaceCard, error)
}

type Deps struct {
	Srv CardService

	MaxLenTitle       int
	MaxLenDescription int
}

type Handler struct {
	deps Deps
}

func NewHandler(deps Deps) *Handler {
	return &Handler{
		deps: deps,
	}
}

func (h *Handler) GetCard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	linkParam := vars[cardLinkKey]

	linkCard, err := uuid.Parse(linkParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectPath)
		return
	}

	card, err := h.deps.Srv.GetCard(r.Context(), linkCard)
	if err != nil {
		if errors.Is(err, common.ErrorNotExistingCard) {
			api.RespondError(w, http.StatusNotFound, failFindCard)
			return
		}

		api.RespondError(w, http.StatusInternalServerError, failGetCard)
		return
	}

	api.HandleError(api.RespondOk(w, dto.InfoCard{
		LinkCard:     linkCard,
		Title:        card.Title,
		Description:  card.Description,
		NameExecuter: card.NameExecuter,
		DataDeadLine: card.DataDeadLine,
	}))
}

func (h *Handler) DeleteCard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	linkParam := vars[cardLinkKey]

	linkCard, err := uuid.Parse(linkParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectPath)
		return
	}

	err = h.deps.Srv.DeleteCard(r.Context(), linkCard)
	if err != nil {
		if errors.Is(err, common.ErrorNotExistingCard) {
			api.RespondError(w, http.StatusNotFound, failFindCard)
			return
		}

		api.RespondError(w, http.StatusInternalServerError, failDeleteCard)
		return
	}

	api.RespondOk(w, api.StatusOK)
}

func (h *Handler) UpdateCardDetails(w http.ResponseWriter, r *http.Request) {
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

	err = common.ValidateTextInfo(updatingInfo.Title, h.deps.MaxLenTitle)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, fmt.Sprintf("max len title is %d", h.deps.MaxLenTitle))
		return
	}

	err = common.ValidateTextInfo(updatingInfo.Description, h.deps.MaxLenDescription)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, fmt.Sprintf("max len description is %d", h.deps.MaxLenDescription))
		return
	}

	err = h.deps.Srv.UpdateCardDetails(r.Context(), serviceDto.UpdatingCardDetails{
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

		api.RespondError(w, http.StatusInternalServerError, failUpdateCard)
		return
	}

	api.RespondOk(w, api.StatusOK)
}

func (h *Handler) ReorderCard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	linkCardParam := vars[cardLinkKey]

	linkCard, err := uuid.Parse(linkCardParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectPath)
		return
	}

	var updatingPlaceCard dto.PlaceCard

	err = json.NewDecoder(r.Body).Decode(&updatingPlaceCard)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectRequest)
		return
	}

	err = h.deps.Srv.ReorderCard(r.Context(), serviceDto.PlaceCard{
		LinkCard:    linkCard,
		LinkSection: updatingPlaceCard.LinkSection,
		Position:    updatingPlaceCard.Position,
	})
	if err != nil {
		if errors.Is(err, common.ErrorSkipMandatorySection) {
			api.RespondError(w, http.StatusBadRequest, incorectMoveCard)
			return
		}

		if errors.Is(err, common.ErrorNotExistingCard) {
			api.RespondError(w, http.StatusNotFound, failFindCard)
			return
		}

		api.RespondError(w, http.StatusInternalServerError, failReorderCard)
		return
	}

	api.RespondOk(w, api.StatusOK)
}

func (h *Handler) CreateCard(w http.ResponseWriter, r *http.Request) {
	var newCard dto.NewCard
	err := json.NewDecoder(r.Body).Decode(&newCard)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectRequest)
		return
	}

	err = common.ValidateTextInfo(newCard.Title, h.deps.MaxLenTitle)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, fmt.Sprintf("max len title is %d", h.deps.MaxLenTitle))
		return
	}

	err = common.ValidateTextInfo(newCard.Description, h.deps.MaxLenDescription)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, fmt.Sprintf("max len description is %d", h.deps.MaxLenDescription))
		return
	}

	authorLink, ok := r.Context().Value(middleware.UserContextLink{}).(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, common.FailAuthorized)
		return
	}

	card, err := h.deps.Srv.CreateCard(r.Context(), serviceDto.NewCard{
		LinkAuthor:   authorLink,
		Title:        newCard.Title,
		Description:  newCard.Description,
		LinkExecuter: newCard.LinkExecuter,
		DataDeadLine: newCard.DataDeadLine,
		LinkSection:  newCard.LinkSection,
	})

	if err != nil {
		if errors.Is(err, common.ErrorNotExistingSection) {
			api.RespondError(w, http.StatusNotFound, failFindSection)
			return
		}

		api.RespondError(w, http.StatusInternalServerError, failCreateCard)
		return
	}

	api.HandleError(api.RespondOk(w, dto.PlaceCard{
		LinkCard:    card.LinkCard,
		LinkSection: card.LinkSection,
		Position:    card.Position,
	}))
}
