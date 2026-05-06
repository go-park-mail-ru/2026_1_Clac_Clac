package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/dto"
	handlerCommon "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/middleware"
	sentryLogger "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/logger"
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
	cardLinkKey    = "link"
	commentLinkKey = "comment_link"
	subtaskLinkKey = "subtask_link"

	msgCardNotFound         = "card not found"
	msgSectionNotFound      = "section not found"
	msgCommentNotFound      = "comment not found"
	msgSubtaskNotFound      = "subtask not found"
	msgPermissionDenied     = "permission denied"
	msgCardAlreadyExists    = "card already exists"
	msgTaskLimitReached     = "task limit reached"
	msgMissMandatorySection = "miss mandatory section"
	msgInvalidInput         = "invalid input"

	msgFailGetCard       = "cannot get card"
	msgFailDeleteCard    = "cannot delete card"
	msgFailUpdateCard    = "cannot update card"
	msgFailReorderCards  = "cannot reorder cards"
	msgFailCreateCard    = "cannot create card"
	msgFailGetComments   = "cannot get comments"
	msgFailCreateComment = "cannot create comment"
	msgFailDeleteComment = "cannot delete comment"
	msgFailUpdateComment = "cannot update comment"
	msgFailCreateSubtask = "cannot create subtask"
	msgFailUpdateSubtask = "cannot update subtask"
	msgFailDeleteSubtask = "cannot delete subtask"
)

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

// GetCard возвращает карточку по UUID
//
//	@Summary		Получить карточку
//	@Description	Возвращает полную информацию о карточке по её UUID: заголовок, описание, дедлайн, подзадачи.
//	@Tags			Cards
//	@Security		sessionCookie
//	@Produce		json
//	@Param			link	path		string								true	"UUID карточки"
//	@Success		200		{object}	api.OkResponse[dto.CardResponse]	"Карточка"
//	@Failure		400		{object}	api.ErrorResponse					"Некорректный UUID"
//	@Failure		401		{object}	api.ErrorResponse					"Пользователь не авторизован"
//	@Failure		403		{object}	api.ErrorResponse					"Нет прав доступа"
//	@Failure		404		{object}	api.ErrorResponse					"Карточка не найдена"
//	@Failure		500		{object}	api.ErrorResponse					"Внутренняя ошибка сервера"
//	@Router			/cards/{link} [get]
func (c *Card) GetCard(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	cardLink, ok := parseCardLink(w, r)
	if !ok {
		return
	}

	userLink, ok := getUserLink(w, r)
	if !ok {
		return
	}

	card, err := c.card.GetCard(r.Context(), domain.GetCardRequest{
		UserLink: userLink,
		CardLink: cardLink,
	})
	if err != nil {
		if errors.Is(err, common.ErrorCardNotFound) {
			api.RespondError(w, http.StatusNotFound, msgCardNotFound)
			return
		}
		if errors.Is(err, common.ErrorPermissionDenied) {
			api.RespondError(w, http.StatusForbidden, msgPermissionDenied)
			return
		}
		errLog := fmt.Errorf("card.GetCard: %w", err)
		logger.Error().Err(errLog).Msg("card.GetCard failed")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "GetCard", map[string]interface{}{
			"user_link": userLink,
			"card_link": cardLink,
			"action":    "get_card",
		})
		api.RespondError(w, http.StatusInternalServerError, msgFailGetCard)
		return
	}

	api.HandleError(api.RespondOk(w, convertToCardResponse(cardLink, card)))
}

// DeleteCard удаляет карточку
//
//	@Summary		Удалить карточку
//	@Description	Удаляет карточку по UUID. Требует прав на удаление.
//	@Tags			Cards
//	@Security		sessionCookie
//	@Produce		json
//	@Param			link	path		string				true	"UUID карточки"
//	@Success		200		{object}	api.Response		"Карточка удалена"
//	@Failure		400		{object}	api.ErrorResponse	"Некорректный UUID"
//	@Failure		401		{object}	api.ErrorResponse	"Пользователь не авторизован"
//	@Failure		403		{object}	api.ErrorResponse	"Нет прав доступа"
//	@Failure		404		{object}	api.ErrorResponse	"Карточка не найдена"
//	@Failure		500		{object}	api.ErrorResponse	"Внутренняя ошибка сервера"
//	@Router			/cards/{link} [delete]
func (c *Card) DeleteCard(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	cardLink, ok := parseCardLink(w, r)
	if !ok {
		return
	}

	userLink, ok := getUserLink(w, r)
	if !ok {
		return
	}

	err := c.card.DeleteCard(r.Context(), domain.DeleteCardRequest{
		UserLink: userLink,
		CardLink: cardLink,
	})
	if err != nil {
		if errors.Is(err, common.ErrorCardNotFound) {
			api.RespondError(w, http.StatusNotFound, msgCardNotFound)
			return
		}
		if errors.Is(err, common.ErrorPermissionDenied) {
			api.RespondError(w, http.StatusForbidden, msgPermissionDenied)
			return
		}
		errLog := fmt.Errorf("card.DeleteCard: %w", err)
		logger.Error().Err(errLog).Msg("card.DeleteCard failed")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "DeleteCard", map[string]interface{}{
			"user_link": userLink,
			"card_link": cardLink,
			"action":    "delete_card",
		})
		api.RespondError(w, http.StatusInternalServerError, msgFailDeleteCard)
		return
	}

	api.HandleError(api.RespondOk(w, api.StatusOK))
}

// UpdateCard обновляет карточку
//
//	@Summary		Обновить карточку
//	@Description	Обновляет заголовок, описание, исполнителя и дедлайн карточки.
//	@Tags			Cards
//	@Security		sessionCookie
//	@Accept			json
//	@Produce		json
//	@Param			link	path		string					true	"UUID карточки"
//	@Param			input	body		dto.UpdateCardRequest	true	"Данные для обновления"
//	@Success		200		{object}	api.Response			"Карточка обновлена"
//	@Failure		400		{object}	api.ErrorResponse		"Некорректные данные или превышена длина"
//	@Failure		401		{object}	api.ErrorResponse		"Пользователь не авторизован"
//	@Failure		403		{object}	api.ErrorResponse		"Нет прав доступа"
//	@Failure		404		{object}	api.ErrorResponse		"Карточка не найдена"
//	@Failure		409		{object}	api.ErrorResponse		"Карточка уже существует"
//	@Failure		500		{object}	api.ErrorResponse		"Внутренняя ошибка сервера"
//	@Router			/cards/{link} [put]
func (c *Card) UpdateCard(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	cardLink, ok := parseCardLink(w, r)
	if !ok {
		return
	}

	userLink, ok := getUserLink(w, r)
	if !ok {
		return
	}

	var req dto.UpdateCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	req.Sanitize()

	if err := common.ValidateTextInfo(req.Title, c.cfg.MaxLenTitle); err != nil {
		api.RespondError(w, http.StatusBadRequest, fmt.Sprintf("incorrect title: %s", err.Error()))
		return
	}

	if err := common.ValidateTextInfo(req.Description, c.cfg.MaxLenDescription); err != nil {
		api.RespondError(w, http.StatusBadRequest, fmt.Sprintf("incorrect description: %s", err.Error()))
		return
	}

	var executorLink *uuid.UUID
	if req.ExecutorLink != nil && *req.ExecutorLink != "" {
		parsed, err := uuid.Parse(*req.ExecutorLink)
		if err != nil {
			api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
			return
		}
		executorLink = &parsed
	}

	err := c.card.UpdateCard(r.Context(), domain.UpdateCardRequest{
		UserLink:     userLink,
		CardLink:     cardLink,
		ExecutorLink: executorLink,
		Title:        req.Title,
		Description:  req.Description,
		Deadline:     req.Deadline,
	})
	if err != nil {
		if errors.Is(err, common.ErrorCardNotFound) {
			api.RespondError(w, http.StatusNotFound, msgCardNotFound)
			return
		}
		if errors.Is(err, common.ErrorPermissionDenied) {
			api.RespondError(w, http.StatusForbidden, msgPermissionDenied)
			return
		}
		if errors.Is(err, common.ErrorCardAlreadyExists) {
			api.RespondError(w, http.StatusConflict, msgCardAlreadyExists)
			return
		}
		if errors.Is(err, common.ErrorInvalidInput) {
			api.RespondError(w, http.StatusBadRequest, msgInvalidInput)
			return
		}
		errLog := fmt.Errorf("card.UpdateCard: %w", err)
		logger.Error().Err(errLog).Msg("card.UpdateCard failed")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "UpdateCard", map[string]interface{}{
			"user_link": userLink,
			"card_link": cardLink,
			"action":    "update_card",
		})
		api.RespondError(w, http.StatusInternalServerError, msgFailUpdateCard)
		return
	}

	api.HandleError(api.RespondOk(w, api.StatusOK))
}

// ReorderCards изменяет позицию карточки
//
//	@Summary		Переместить карточку
//	@Description	Изменяет позицию карточки внутри секции.
//	@Tags			Cards
//	@Security		sessionCookie
//	@Accept			json
//	@Produce		json
//	@Param			link	path		string					true	"UUID карточки"
//	@Param			input	body		dto.ReorderCardsRequest	true	"Данные для перемещения"
//	@Success		200		{object}	api.Response			"Карточка перемещена"
//	@Failure		400		{object}	api.ErrorResponse		"Некорректные данные"
//	@Failure		401		{object}	api.ErrorResponse		"Пользователь не авторизован"
//	@Failure		403		{object}	api.ErrorResponse		"Нет прав доступа"
//	@Failure		404		{object}	api.ErrorResponse		"Карточка или секция не найдены"
//	@Failure		500		{object}	api.ErrorResponse		"Внутренняя ошибка сервера"
//	@Router			/cards/{link}/reorder [patch]
func (c *Card) ReorderCards(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	cardLink, ok := parseCardLink(w, r)
	if !ok {
		return
	}

	userLink, ok := getUserLink(w, r)
	if !ok {
		return
	}

	var req dto.ReorderCardsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	sectionLink, err := uuid.Parse(req.SectionLink)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	err = c.card.ReorderCards(r.Context(), domain.ReorderCardsRequest{
		UserLink:    userLink,
		CardLink:    cardLink,
		SectionLink: sectionLink,
		Position:    req.Position,
	})
	if err != nil {
		if errors.Is(err, common.ErrorCardNotFound) {
			api.RespondError(w, http.StatusNotFound, msgCardNotFound)
			return
		}
		if errors.Is(err, common.ErrorSectionNotFound) {
			api.RespondError(w, http.StatusNotFound, msgSectionNotFound)
			return
		}
		if errors.Is(err, common.ErrorPermissionDenied) {
			api.RespondError(w, http.StatusForbidden, msgPermissionDenied)
			return
		}
		if errors.Is(err, common.ErrorInvalidInput) {
			api.RespondError(w, http.StatusBadRequest, msgInvalidInput)
			return
		}
		if errors.Is(err, common.ErrorTaskLimitReached) {
			api.RespondError(w, http.StatusBadRequest, msgTaskLimitReached)
			return
		}
		if errors.Is(err, common.ErrCannotSkipMandatorySection) {
			api.RespondError(w, http.StatusBadRequest, msgMissMandatorySection)
			return
		}

		errLog := fmt.Errorf("card.ReorderCards: %w", err)
		logger.Error().Err(errLog).Msg("card.ReorderCards failed")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "ReorderCards", map[string]interface{}{
			"user_link": userLink,
			"card_link": cardLink,
			"action":    "reorder_cards",
		})
		api.RespondError(w, http.StatusInternalServerError, msgFailReorderCards)
		return
	}

	api.HandleError(api.RespondOk(w, api.StatusOK))
}

// CreateCard создаёт карточку
//
//	@Summary		Создать карточку
//	@Description	Создаёт новую карточку в указанной секции.
//	@Tags			Cards
//	@Security		sessionCookie
//	@Accept			json
//	@Produce		json
//	@Param			input	body		dto.CreateCardRequest					true	"Данные карточки"
//	@Success		200		{object}	api.OkResponse[dto.CreateCardResponse]	"Карточка создана"
//	@Failure		400		{object}	api.ErrorResponse						"Некорректные данные или лимит задач"
//	@Failure		401		{object}	api.ErrorResponse						"Пользователь не авторизован"
//	@Failure		403		{object}	api.ErrorResponse						"Нет прав доступа"
//	@Failure		404		{object}	api.ErrorResponse						"Секция не найдена"
//	@Failure		409		{object}	api.ErrorResponse						"Карточка уже существует"
//	@Failure		500		{object}	api.ErrorResponse						"Внутренняя ошибка сервера"
//	@Router			/cards [post]
func (c *Card) CreateCard(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	userLink, ok := getUserLink(w, r)
	if !ok {
		return
	}

	var req dto.CreateCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	req.Sanitize()

	if err := common.ValidateTextInfo(req.Title, c.cfg.MaxLenTitle); err != nil {
		api.RespondError(w, http.StatusBadRequest, fmt.Sprintf("incorrect title: %s", err.Error()))
		return
	}

	if err := common.ValidateTextInfo(req.Description, c.cfg.MaxLenDescription); err != nil {
		api.RespondError(w, http.StatusBadRequest, fmt.Sprintf("incorrect description: %s", err.Error()))
		return
	}

	sectionLink, err := uuid.Parse(req.SectionLink)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	var executorLink *uuid.UUID
	if req.ExecutorLink != nil && *req.ExecutorLink != "" {
		parsed, err := uuid.Parse(*req.ExecutorLink)
		if err != nil {
			api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
			return
		}
		executorLink = &parsed
	}

	created, err := c.card.CreateCard(r.Context(), domain.CreateCardRequest{
		UserLink:     userLink,
		SectionLink:  sectionLink,
		ExecutorLink: executorLink,
		Title:        req.Title,
		Description:  req.Description,
		Deadline:     req.Deadline,
	})
	if err != nil {
		if errors.Is(err, common.ErrorSectionNotFound) {
			api.RespondError(w, http.StatusNotFound, msgSectionNotFound)
			return
		}
		if errors.Is(err, common.ErrorPermissionDenied) {
			api.RespondError(w, http.StatusForbidden, msgPermissionDenied)
			return
		}
		if errors.Is(err, common.ErrorCardAlreadyExists) {
			api.RespondError(w, http.StatusConflict, msgCardAlreadyExists)
			return
		}
		if errors.Is(err, common.ErrorTaskLimitReached) {
			api.RespondError(w, http.StatusBadRequest, msgTaskLimitReached)
			return
		}
		if errors.Is(err, common.ErrorInvalidInput) {
			api.RespondError(w, http.StatusBadRequest, msgInvalidInput)
			return
		}
		errLog := fmt.Errorf("card.CreateCard: %w", err)
		logger.Error().Err(errLog).Msg("card.CreateCard failed")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "CreateCard", map[string]interface{}{
			"user_link":    userLink,
			"section_link": sectionLink,
			"action":       "create_card",
		})
		api.RespondError(w, http.StatusInternalServerError, msgFailCreateCard)
		return
	}

	api.HandleError(api.RespondOk(w, convertToCreateCardResponse(created)))
}

// GetComments возвращает список комментариев к карточке
//
//	@Summary		Получить комментарии
//	@Description	Возвращает все комментарии к указанной карточке.
//	@Tags			Cards
//	@Security		sessionCookie
//	@Produce		json
//	@Param			link	path		string									true	"UUID карточки"
//	@Success		200		{object}	api.OkResponse[dto.CommentsResponse]	"Список комментариев"
//	@Failure		400		{object}	api.ErrorResponse						"Некорректный UUID"
//	@Failure		401		{object}	api.ErrorResponse						"Пользователь не авторизован"
//	@Failure		403		{object}	api.ErrorResponse						"Нет прав доступа"
//	@Failure		404		{object}	api.ErrorResponse						"Карточка не найдена"
//	@Failure		500		{object}	api.ErrorResponse						"Внутренняя ошибка сервера"
//	@Router			/cards/{link}/comments [get]
func (c *Card) GetComments(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	cardLink, ok := parseCardLink(w, r)
	if !ok {
		return
	}

	userLink, ok := getUserLink(w, r)
	if !ok {
		return
	}

	comments, err := c.card.GetComments(r.Context(), domain.GetCommentsRequest{
		UserLink: userLink,
		CardLink: cardLink,
	})
	if err != nil {
		if errors.Is(err, common.ErrorCardNotFound) {
			api.RespondError(w, http.StatusNotFound, msgCardNotFound)
			return
		}
		if errors.Is(err, common.ErrorPermissionDenied) {
			api.RespondError(w, http.StatusForbidden, msgPermissionDenied)
			return
		}
		errLog := fmt.Errorf("card.GetComments: %w", err)
		logger.Error().Err(errLog).Msg("card.GetComments failed")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "GetComments", map[string]interface{}{
			"user_link": userLink,
			"card_link": cardLink,
			"action":    "get_comments",
		})
		api.RespondError(w, http.StatusInternalServerError, msgFailGetComments)
		return
	}

	api.HandleError(api.RespondOk(w, convertToCommentsResponse(comments)))
}

// CreateComment создаёт комментарий к карточке
//
//	@Summary		Создать комментарий
//	@Description	Добавляет комментарий к карточке. Может быть ответом на другой комментарий.
//	@Tags			Cards
//	@Security		sessionCookie
//	@Accept			json
//	@Produce		json
//	@Param			link	path		string										true	"UUID карточки"
//	@Param			input	body		dto.CreateCommentRequest					true	"Данные комментария"
//	@Success		200		{object}	api.OkResponse[dto.CreateCommentResponse]	"Комментарий создан"
//	@Failure		400		{object}	api.ErrorResponse							"Некорректные данные"
//	@Failure		401		{object}	api.ErrorResponse							"Пользователь не авторизован"
//	@Failure		403		{object}	api.ErrorResponse							"Нет прав доступа"
//	@Failure		404		{object}	api.ErrorResponse							"Карточка не найдена"
//	@Failure		500		{object}	api.ErrorResponse							"Внутренняя ошибка сервера"
//	@Router			/cards/{link}/comments [post]
func (c *Card) CreateComment(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	cardLink, ok := parseCardLink(w, r)
	if !ok {
		return
	}

	userLink, ok := getUserLink(w, r)
	if !ok {
		return
	}

	var req dto.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	req.Sanitize()

	var parentLink *uuid.UUID
	if req.ParentLink != nil && *req.ParentLink != "" {
		parsed, err := uuid.Parse(*req.ParentLink)
		if err != nil {
			api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
			return
		}
		parentLink = &parsed
	}

	created, err := c.card.CreateComment(r.Context(), domain.CreateCommentRequest{
		UserLink:   userLink,
		CardLink:   cardLink,
		ParentLink: parentLink,
		Text:       req.Text,
	})
	if err != nil {
		if errors.Is(err, common.ErrorCardNotFound) {
			api.RespondError(w, http.StatusNotFound, msgCardNotFound)
			return
		}
		if errors.Is(err, common.ErrorPermissionDenied) {
			api.RespondError(w, http.StatusForbidden, msgPermissionDenied)
			return
		}
		if errors.Is(err, common.ErrorInvalidInput) {
			api.RespondError(w, http.StatusBadRequest, msgInvalidInput)
			return
		}
		errLog := fmt.Errorf("card.CreateComment: %w", err)
		logger.Error().Err(errLog).Msg("card.CreateComment failed")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "CreateComment", map[string]interface{}{
			"user_link": userLink,
			"card_link": cardLink,
			"action":    "create_comment",
		})
		api.RespondError(w, http.StatusInternalServerError, msgFailCreateComment)
		return
	}

	api.HandleError(api.RespondOk(w, convertToCreateCommentResponse(created)))
}

// DeleteComment удаляет комментарий
//
//	@Summary		Удалить комментарий
//	@Description	Удаляет комментарий по UUID.
//	@Tags			Cards
//	@Security		sessionCookie
//	@Produce		json
//	@Param			comment_link	path		string				true	"UUID комментария"
//	@Success		200				{object}	api.Response		"Комментарий удалён"
//	@Failure		400				{object}	api.ErrorResponse	"Некорректный UUID"
//	@Failure		401				{object}	api.ErrorResponse	"Пользователь не авторизован"
//	@Failure		403				{object}	api.ErrorResponse	"Нет прав доступа"
//	@Failure		404				{object}	api.ErrorResponse	"Комментарий не найден"
//	@Failure		500				{object}	api.ErrorResponse	"Внутренняя ошибка сервера"
//	@Router			/comments/{comment_link} [delete]
func (c *Card) DeleteComment(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	commentLinkParam := mux.Vars(r)[commentLinkKey]
	commentLink, err := uuid.Parse(commentLinkParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	userLink, ok := getUserLink(w, r)
	if !ok {
		return
	}

	err = c.card.DeleteComment(r.Context(), domain.DeleteCommentRequest{
		UserLink:    userLink,
		CommentLink: commentLink,
	})
	if err != nil {
		if errors.Is(err, common.ErrorCommentNotFound) {
			api.RespondError(w, http.StatusNotFound, msgCommentNotFound)
			return
		}
		if errors.Is(err, common.ErrorPermissionDenied) {
			api.RespondError(w, http.StatusForbidden, msgPermissionDenied)
			return
		}
		errLog := fmt.Errorf("card.DeleteComment: %w", err)
		logger.Error().Err(errLog).Msg("card.DeleteComment failed")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "DeleteComment", map[string]interface{}{
			"user_link":    userLink,
			"comment_link": commentLink,
			"action":       "delete_comment",
		})
		api.RespondError(w, http.StatusInternalServerError, msgFailDeleteComment)
		return
	}

	api.HandleError(api.RespondOk(w, api.StatusOK))
}

// UpdateComment обновляет текст комментария
//
//	@Summary		Обновить комментарий
//	@Description	Изменяет текст существующего комментария.
//	@Tags			Cards
//	@Security		sessionCookie
//	@Accept			json
//	@Produce		json
//	@Param			comment_link	path		string						true	"UUID комментария"
//	@Param			input			body		dto.UpdateCommentRequest	true	"Новый текст"
//	@Success		200				{object}	api.Response				"Комментарий обновлён"
//	@Failure		400				{object}	api.ErrorResponse			"Некорректные данные"
//	@Failure		401				{object}	api.ErrorResponse			"Пользователь не авторизован"
//	@Failure		403				{object}	api.ErrorResponse			"Нет прав доступа"
//	@Failure		404				{object}	api.ErrorResponse			"Комментарий не найден"
//	@Failure		500				{object}	api.ErrorResponse			"Внутренняя ошибка сервера"
//	@Router			/comments/{comment_link} [put]
func (c *Card) UpdateComment(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	commentLinkParam := mux.Vars(r)[commentLinkKey]
	commentLink, err := uuid.Parse(commentLinkParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	userLink, ok := getUserLink(w, r)
	if !ok {
		return
	}

	var req dto.UpdateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	req.Sanitize()

	err = c.card.UpdateComment(r.Context(), domain.UpdateCommentRequest{
		UserLink:    userLink,
		CommentLink: commentLink,
		Text:        req.Text,
	})
	if err != nil {
		if errors.Is(err, common.ErrorCommentNotFound) {
			api.RespondError(w, http.StatusNotFound, msgCommentNotFound)
			return
		}
		if errors.Is(err, common.ErrorPermissionDenied) {
			api.RespondError(w, http.StatusForbidden, msgPermissionDenied)
			return
		}
		if errors.Is(err, common.ErrorInvalidInput) {
			api.RespondError(w, http.StatusBadRequest, msgInvalidInput)
			return
		}
		errLog := fmt.Errorf("card.UpdateComment: %w", err)
		logger.Error().Err(errLog).Msg("card.UpdateComment failed")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "UpdateComment", map[string]interface{}{
			"user_link":    userLink,
			"comment_link": commentLink,
			"action":       "update_comment",
		})
		api.RespondError(w, http.StatusInternalServerError, msgFailUpdateComment)
		return
	}

	api.HandleError(api.RespondOk(w, api.StatusOK))
}

// CreateSubtask создаёт подзадачу карточки
//
//	@Summary		Создать подзадачу
//	@Description	Добавляет подзадачу к карточке.
//	@Tags			Cards
//	@Security		sessionCookie
//	@Accept			json
//	@Produce		json
//	@Param			link	path		string								true	"UUID карточки"
//	@Param			input	body		dto.CreateSubtaskRequest			true	"Данные подзадачи"
//	@Success		200		{object}	api.OkResponse[dto.SubtaskResponse]	"Подзадача создана"
//	@Failure		400		{object}	api.ErrorResponse					"Некорректные данные"
//	@Failure		401		{object}	api.ErrorResponse					"Пользователь не авторизован"
//	@Failure		403		{object}	api.ErrorResponse					"Нет прав доступа"
//	@Failure		404		{object}	api.ErrorResponse					"Карточка не найдена"
//	@Failure		500		{object}	api.ErrorResponse					"Внутренняя ошибка сервера"
//	@Router			/cards/{link}/subtasks [post]
func (c *Card) CreateSubtask(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	cardLink, ok := parseCardLink(w, r)
	if !ok {
		return
	}

	userLink, ok := getUserLink(w, r)
	if !ok {
		return
	}

	var req dto.CreateSubtaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	req.Sanitize()

	subtask, err := c.card.CreateSubtask(r.Context(), domain.CreateSubtaskRequest{
		UserLink:    userLink,
		CardLink:    cardLink,
		Description: req.Description,
	})
	if err != nil {
		if errors.Is(err, common.ErrorCardNotFound) {
			api.RespondError(w, http.StatusNotFound, msgCardNotFound)
			return
		}
		if errors.Is(err, common.ErrorPermissionDenied) {
			api.RespondError(w, http.StatusForbidden, msgPermissionDenied)
			return
		}
		if errors.Is(err, common.ErrorInvalidInput) {
			api.RespondError(w, http.StatusBadRequest, msgInvalidInput)
			return
		}
		errLog := fmt.Errorf("card.CreateSubtask: %w", err)
		logger.Error().Err(errLog).Msg("card.CreateSubtask failed")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "CreateSubtask", map[string]interface{}{
			"user_link": userLink,
			"card_link": cardLink,
			"action":    "create_subtask",
		})
		api.RespondError(w, http.StatusInternalServerError, msgFailCreateSubtask)
		return
	}

	api.HandleError(api.RespondOk(w, convertToSubtaskResponse(subtask)))
}

// UpdateSubtask обновляет подзадачу
//
//	@Summary		Обновить подзадачу
//	@Description	Изменяет описание и статус выполнения подзадачи.
//	@Tags			Cards
//	@Security		sessionCookie
//	@Accept			json
//	@Produce		json
//	@Param			subtask_link	path		string						true	"UUID подзадачи"
//	@Param			input			body		dto.UpdateSubtaskRequest	true	"Данные подзадачи"
//	@Success		200				{object}	api.Response				"Подзадача обновлена"
//	@Failure		400				{object}	api.ErrorResponse			"Некорректные данные"
//	@Failure		401				{object}	api.ErrorResponse			"Пользователь не авторизован"
//	@Failure		403				{object}	api.ErrorResponse			"Нет прав доступа"
//	@Failure		404				{object}	api.ErrorResponse			"Подзадача не найдена"
//	@Failure		500				{object}	api.ErrorResponse			"Внутренняя ошибка сервера"
//	@Router			/subtasks/{subtask_link} [put]
func (c *Card) UpdateSubtask(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	subtaskLinkParam := mux.Vars(r)[subtaskLinkKey]
	subtaskLink, err := uuid.Parse(subtaskLinkParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	userLink, ok := getUserLink(w, r)
	if !ok {
		return
	}

	var req dto.UpdateSubtaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	req.Sanitize()

	err = c.card.UpdateSubtask(r.Context(), domain.UpdateSubtaskRequest{
		UserLink:    userLink,
		SubtaskLink: subtaskLink,
		IsDone:      req.IsDone,
		Description: req.Description,
	})
	if err != nil {
		if errors.Is(err, common.ErrorSubtaskNotFound) {
			api.RespondError(w, http.StatusNotFound, msgSubtaskNotFound)
			return
		}
		if errors.Is(err, common.ErrorPermissionDenied) {
			api.RespondError(w, http.StatusForbidden, msgPermissionDenied)
			return
		}
		if errors.Is(err, common.ErrorInvalidInput) {
			api.RespondError(w, http.StatusBadRequest, msgInvalidInput)
			return
		}
		errLog := fmt.Errorf("card.UpdateSubtask: %w", err)
		logger.Error().Err(errLog).Msg("card.UpdateSubtask failed")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "UpdateSubtask", map[string]interface{}{
			"user_link":    userLink,
			"subtask_link": subtaskLink,
			"action":       "update_subtask",
		})
		api.RespondError(w, http.StatusInternalServerError, msgFailUpdateSubtask)
		return
	}

	api.HandleError(api.RespondOk(w, api.StatusOK))
}

// DeleteSubtask удаляет подзадачу
//
//	@Summary		Удалить подзадачу
//	@Description	Удаляет подзадачу по UUID.
//	@Tags			Cards
//	@Security		sessionCookie
//	@Produce		json
//	@Param			subtask_link	path		string				true	"UUID подзадачи"
//	@Success		200				{object}	api.Response		"Подзадача удалена"
//	@Failure		400				{object}	api.ErrorResponse	"Некорректный UUID"
//	@Failure		401				{object}	api.ErrorResponse	"Пользователь не авторизован"
//	@Failure		403				{object}	api.ErrorResponse	"Нет прав доступа"
//	@Failure		404				{object}	api.ErrorResponse	"Подзадача не найдена"
//	@Failure		500				{object}	api.ErrorResponse	"Внутренняя ошибка сервера"
//	@Router			/subtasks/{subtask_link} [delete]
func (c *Card) DeleteSubtask(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	subtaskLinkParam := mux.Vars(r)[subtaskLinkKey]
	subtaskLink, err := uuid.Parse(subtaskLinkParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	userLink, ok := getUserLink(w, r)
	if !ok {
		return
	}

	err = c.card.DeleteSubtask(r.Context(), domain.DeleteSubtask{
		UserLink:    userLink,
		SubtaskLink: subtaskLink,
	})
	if err != nil {
		if errors.Is(err, common.ErrorSubtaskNotFound) {
			api.RespondError(w, http.StatusNotFound, msgSubtaskNotFound)
			return
		}
		if errors.Is(err, common.ErrorPermissionDenied) {
			api.RespondError(w, http.StatusForbidden, msgPermissionDenied)
			return
		}
		errLog := fmt.Errorf("card.DeleteSubtask: %w", err)
		logger.Error().Err(errLog).Msg("card.DeleteSubtask failed")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "DeleteSubtask", map[string]interface{}{
			"user_link":    userLink,
			"subtask_link": subtaskLink,
			"action":       "delete_subtask",
		})
		api.RespondError(w, http.StatusInternalServerError, msgFailDeleteSubtask)
		return
	}

	api.HandleError(api.RespondOk(w, api.StatusOK))
}

func parseCardLink(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	linkParam := mux.Vars(r)[cardLinkKey]
	cardLink, err := uuid.Parse(linkParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return uuid.Nil, false
	}
	return cardLink, true
}

func getUserLink(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	value := r.Context().Value(middleware.UserContextLink{})
	userLink, ok := value.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, msgUnauthorized)
		return uuid.Nil, false
	}
	return userLink, true
}

func convertToCardResponse(cardLink uuid.UUID, card domain.CardInfo) dto.CardResponse {
	subtasks := make([]dto.SubtaskResponse, 0, len(card.Subtasks))
	for _, s := range card.Subtasks {
		subtasks = append(subtasks, dto.SubtaskResponse{
			SubtaskLink: s.SubtaskLink,
			Description: s.Description,
			IsDone:      s.IsDone,
			Position:    int(s.Position),
		})
	}
	var executorLink *string
	if card.ExecutorLink != nil {
		s := card.ExecutorLink.String()
		executorLink = &s
	}

	return dto.CardResponse{
		CardLink:     cardLink,
		ExecutorLink: executorLink,
		Title:        card.Title,
		Description:  card.Description,
		Deadline:     card.Deadline,
		Subtasks:     subtasks,
		Position:     card.Position,
	}
}

func convertToCommentsResponse(resp domain.GetCommentsResponse) dto.CommentsResponse {
	comments := make([]dto.CommentResponse, 0, len(resp.CommentsInfo))
	for _, c := range resp.CommentsInfo {
		comments = append(comments, dto.CommentResponse{
			CommentLink: c.CommentLink,
			ParentLink:  c.ParentLink,
			AuthorLink:  c.AuthorLink,
			Text:        c.Text,
			CreatedAt:   c.CreatedAt,
		})
	}
	return dto.CommentsResponse{Comments: comments}
}

func convertToSubtaskResponse(s domain.SubtaskInfo) dto.SubtaskResponse {
	return dto.SubtaskResponse{
		SubtaskLink: s.SubtaskLink,
		Description: s.Description,
		IsDone:      s.IsDone,
		Position:    int(s.Position),
	}
}

func convertToCreateCardResponse(resp domain.CreateCardResponse) dto.CreateCardResponse {
	return dto.CreateCardResponse{
		CardLink:    resp.CardLink,
		SectionLink: resp.SectionLink,
		Position:    resp.Position,
	}
}

func convertToCreateCommentResponse(resp domain.CreateCommentResponse) dto.CreateCommentResponse {
	return dto.CreateCommentResponse{
		CommentLink: resp.CommentLink,
	}
}
