package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/dto"
	handlerCommon "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/middleware"
	sentryLogger "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/logger"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"
	"github.com/rs/zerolog"
)

type PollUsecase interface {
	CreatePoll(ctx context.Context, boardLink, adminLink uuid.UUID, cards []uuid.UUID, invitees []uuid.UUID) error
	DeletePoll(ctx context.Context, boardLink, userLink uuid.UUID) error
	NextPollCard(ctx context.Context, boardLink, userLink uuid.UUID) error
	VotePoll(ctx context.Context, boardLink, userLink uuid.UUID, points int) error
}

const (
	pollBoardLinkKey = "board_link"
)

type PollHandler struct {
	poll PollUsecase
}

func NewPollHandler(poll PollUsecase) *PollHandler {
	return &PollHandler{
		poll: poll,
	}
}

// CreatePoll создаёт покер-комнату
//
//	@Summary		Создать покер-комнату
//	@Description	Создаёт комнату для оценки задач методом Planning Poker. Требует прав Admin/Creator.
//	@Tags			Polls
//	@Security		sessionCookie
//	@Accept			json
//	@Produce		json
//	@Param			board_link	path		string					true	"UUID доски" Format(uuid)
//	@Param			request		body		dto.CreatePollRequest	true	"Карточки и приглашённые участники"
//	@Success		201			{object}	api.Response			"poll created"
//	@Failure		400			{object}	api.ErrorResponse		"invalid board link / invalid request schema"
//	@Failure		401			{object}	api.ErrorResponse		"unauthorized"
//	@Failure		403			{object}	api.ErrorResponse		"action denied (not admin/creator)"
//	@Failure		409			{object}	api.ErrorResponse		"poll already exists"
//	@Failure		500			{object}	api.ErrorResponse		"cannot create poll"
//	@Router			/boards/{board_link}/polls [post]
func (h *PollHandler) CreatePoll(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	boardLink, ok := parsePollBoardLink(w, r)
	if !ok {
		return
	}

	userLink, ok := getPollUserLink(w, r)
	if !ok {
		return
	}

	var req dto.CreatePollRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	cards := make([]uuid.UUID, 0, len(req.CardLinks))
	for _, raw := range req.CardLinks {
		cardLink, err := uuid.Parse(raw)
		if err != nil {
			api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
			return
		}
		cards = append(cards, cardLink)
	}

	invitees := make([]uuid.UUID, 0, len(req.Invitees))
	for _, raw := range req.Invitees {
		uid, err := uuid.Parse(raw)
		if err != nil {
			api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
			return
		}
		invitees = append(invitees, uid)
	}

	err := h.poll.CreatePoll(r.Context(), boardLink, userLink, cards, invitees)
	if err != nil {
		if errors.Is(err, common.ErrorBoardPermissionDenied) {
			api.RespondError(w, http.StatusForbidden, msgPermissionDenied)
			return
		}
		errLog := fmt.Errorf("poll.CreatePoll: %w", err)
		logger.Error().Err(errLog).Msg("poll.CreatePoll failed")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "CreatePoll", map[string]interface{}{
			"user_link":  userLink,
			"board_link": boardLink,
			"action":     "create_poll",
		})
		api.RespondError(w, http.StatusInternalServerError, "cannot create poll")
		return
	}

	api.HandleError(api.Respond(w, http.StatusCreated, api.StatusOK))
}

// Vote проголосовать за текущую карточку
//
//	@Summary		Проголосовать
//	@Description	Приглашённый участник ставит оценку текущей активной карточке
//	@Tags			Polls
//	@Security		sessionCookie
//	@Accept			json
//	@Produce		json
//	@Param			board_link	path		string					true	"UUID доски" Format(uuid)
//	@Param			request		body		dto.VotePollRequest		true	"Оценка"
//	@Success		200			{object}	api.Response			"vote accepted"
//	@Failure		400			{object}	api.ErrorResponse		"invalid board link / invalid request schema"
//	@Failure		401			{object}	api.ErrorResponse		"unauthorized"
//	@Failure		403			{object}	api.ErrorResponse		"user not invited"
//	@Failure		404			{object}	api.ErrorResponse		"poll not found"
//	@Failure		500			{object}	api.ErrorResponse		"cannot vote"
//	@Router			/boards/{board_link}/polls [put]
func (h *PollHandler) Vote(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	boardLink, ok := parsePollBoardLink(w, r)
	if !ok {
		return
	}

	userLink, ok := getPollUserLink(w, r)
	if !ok {
		return
	}

	var req dto.VotePollRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	err := h.poll.VotePoll(r.Context(), boardLink, userLink, req.Points)
	if err != nil {
		if errors.Is(err, common.ErrorBoardPermissionDenied) {
			api.RespondError(w, http.StatusForbidden, "user not invited")
			return
		}
		errLog := fmt.Errorf("poll.VotePoll: %w", err)
		logger.Error().Err(errLog).Msg("poll.VotePoll failed")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "VotePoll", map[string]interface{}{
			"user_link":  userLink,
			"board_link": boardLink,
			"action":     "vote_poll",
		})
		api.RespondError(w, http.StatusInternalServerError, "cannot vote")
		return
	}

	api.HandleError(api.RespondOk(w, api.StatusOK))
}

// DeletePoll завершает покер-комнату
//
//	@Summary		Завершить покер-комнату
//	@Description	Администратор завершает голосование и удаляет комнату
//	@Tags			Polls
//	@Security		sessionCookie
//	@Produce		json
//	@Param			board_link	path		string				true	"UUID доски" Format(uuid)
//	@Success		200			{object}	api.Response		"poll deleted"
//	@Failure		400			{object}	api.ErrorResponse	"invalid board link"
//	@Failure		401			{object}	api.ErrorResponse	"unauthorized"
//	@Failure		403			{object}	api.ErrorResponse	"not poll admin"
//	@Failure		404			{object}	api.ErrorResponse	"poll not found"
//	@Failure		500			{object}	api.ErrorResponse	"cannot delete poll"
//	@Router			/boards/{board_link}/polls [delete]
func (h *PollHandler) DeletePoll(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	boardLink, ok := parsePollBoardLink(w, r)
	if !ok {
		return
	}

	userLink, ok := getPollUserLink(w, r)
	if !ok {
		return
	}

	err := h.poll.DeletePoll(r.Context(), boardLink, userLink)
	if err != nil {
		if errors.Is(err, common.ErrorBoardPermissionDenied) {
			api.RespondError(w, http.StatusForbidden, "not poll admin")
			return
		}
		errLog := fmt.Errorf("poll.DeletePoll: %w", err)
		logger.Error().Err(errLog).Msg("poll.DeletePoll failed")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "DeletePoll", map[string]interface{}{
			"user_link":  userLink,
			"board_link": boardLink,
			"action":     "delete_poll",
		})
		api.RespondError(w, http.StatusInternalServerError, "cannot delete poll")
		return
	}

	api.HandleError(api.RespondOk(w, api.StatusOK))
}

// NextCard переходит к следующей карточке
//
//	@Summary		Следующая карточка
//	@Description	Администратор переходит к следующей карточке в очереди. Если карточек больше нет — комната автоматически удаляется.
//	@Tags			Polls
//	@Security		sessionCookie
//	@Produce		json
//	@Param			board_link	path		string				true	"UUID доски" Format(uuid)
//	@Success		200			{object}	api.Response		"moved to next card or poll finished"
//	@Failure		400			{object}	api.ErrorResponse	"invalid board link"
//	@Failure		401			{object}	api.ErrorResponse	"unauthorized"
//	@Failure		403			{object}	api.ErrorResponse	"not poll admin"
//	@Failure		404			{object}	api.ErrorResponse	"poll not found"
//	@Failure		500			{object}	api.ErrorResponse	"cannot advance poll"
//	@Router			/boards/{board_link}/polls/next [post]
func (h *PollHandler) NextCard(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	boardLink, ok := parsePollBoardLink(w, r)
	if !ok {
		return
	}

	userLink, ok := getPollUserLink(w, r)
	if !ok {
		return
	}

	err := h.poll.NextPollCard(r.Context(), boardLink, userLink)
	if err != nil {
		if errors.Is(err, common.ErrorBoardPermissionDenied) {
			api.RespondError(w, http.StatusForbidden, "not poll admin")
			return
		}
		errLog := fmt.Errorf("poll.NextPollCard: %w", err)
		logger.Error().Err(errLog).Msg("poll.NextPollCard failed")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "NextPollCard", map[string]interface{}{
			"user_link":  userLink,
			"board_link": boardLink,
			"action":     "next_poll_card",
		})
		api.RespondError(w, http.StatusInternalServerError, "cannot advance poll")
		return
	}

	api.HandleError(api.RespondOk(w, api.StatusOK))
}

func parsePollBoardLink(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	linkParam := mux.Vars(r)[pollBoardLinkKey]
	boardLink, err := uuid.Parse(linkParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return uuid.Nil, false
	}
	return boardLink, true
}

func getPollUserLink(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	value := r.Context().Value(middleware.UserContextLink{})
	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, msgUnauthorized)
		return uuid.Nil, false
	}
	return userLink, true
}
