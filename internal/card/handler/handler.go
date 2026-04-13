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
	"github.com/rs/zerolog"
)

const (
	failFindCard    = "can not find card"
	failGetCard     = "can not get info card"
	failDeleteCard  = "can not delete card"
	failUpdateCard  = "can not update card"
	failReorderCard = "can not reorder card"
	failCreateCard  = "can not create new card"
	failFindSection = "can not find section"
	failNullValue   = "can not use null element"

	incorrectMoveCard   = "can not skip mandatory section"
	incorrectUniqCard   = "link card must be unique"
	incorrectReferences = "incorrect foreign key"
	invalidCardData     = "invalid card data"

	cardLinkKey = "link"
)

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

type Handler struct {
	srv CardService
	cnf Config
}

func NewHandler(srv CardService, cnf Config) *Handler {
	return &Handler{
		srv: srv,
		cnf: cnf,
	}
}

// GetCard godoc
// @Summary      Получение информации о карточке
// @Description  Возвращает детали карточки (название, описание, дедлайн, исполнитель) по ее уникальной ссылке (UUID).
// @Tags         cards
// @Produce      json
// @Param        link path string true "UUID карточки"
// @Success      200 {object} dto.InfoCard "Успешное получение информации о карточке"
// @Failure      400 {object} api.ErrorResponse "Некорректный UUID в пути"
// @Failure      404 {object} api.ErrorResponse "Карточка не найдена"
// @Failure      500 {object} api.ErrorResponse "Внутренняя ошибка сервера"
// @Security     CookieAuth
// @Router       /cards/{link} [get]
func (h *Handler) GetCard(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())
	vars := mux.Vars(r)
	linkParam := vars[cardLinkKey]

	linkCard, err := uuid.Parse(linkParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectPath)
		return
	}

	card, err := h.srv.GetCard(r.Context(), linkCard)
	if err != nil {
		if errors.Is(err, common.ErrorNotExistingCard) {
			api.RespondError(w, http.StatusNotFound, failFindCard)
			return
		}

		logger.Error().Err(err).Msg("CardHandler.GetCard")
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

// DeleteCard godoc
// @Summary      Удаление карточки
// @Description  Удаляет карточку по ее уникальной ссылке (UUID).
// @Tags         cards
// @Produce      json
// @Param        link path string true "UUID карточки"
// @Success      200 {object} api.Response "Карточка успешно удалена"
// @Failure      400 {object} api.ErrorResponse "Некорректный UUID в пути"
// @Failure      404 {object} api.ErrorResponse "Карточка не найдена"
// @Failure      500 {object} api.ErrorResponse "Внутренняя ошибка сервера"
// @Security     CookieAuth
// @Router       /cards/{link} [delete]
func (h *Handler) DeleteCard(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())
	vars := mux.Vars(r)
	linkParam := vars[cardLinkKey]

	linkCard, err := uuid.Parse(linkParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectPath)
		return
	}

	err = h.srv.DeleteCard(r.Context(), linkCard)
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

// UpdateCardDetails godoc
// @Summary      Обновление данных карточки
// @Description  Изменяет детали существующей карточки (название, описание, дедлайн, исполнитель).
// @Tags         cards
// @Accept       json
// @Produce      json
// @Param        link    path string true "UUID карточки"
// @Param        request body dto.UpdatingCardDetails true "Новые данные карточки"
// @Success      200 {object} api.Response "Карточка успешно обновлена"
// @Failure      400 {object} api.ErrorResponse "Некорректный запрос или превышена максимальная длина текста"
// @Failure      404 {object} api.ErrorResponse "Карточка не найдена"
// @Failure      500 {object} api.ErrorResponse "Внутренняя ошибка сервера"
// @Security     CookieAuth
// @Router       /cards/{link} [put]
func (h *Handler) UpdateCardDetails(w http.ResponseWriter, r *http.Request) {
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

	err = common.ValidateTextInfo(updatingInfo.Title, h.cnf.MaxLenTitle)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, fmt.Sprintf("max len title is %d", h.cnf.MaxLenTitle))
		return
	}

	err = common.ValidateTextInfo(updatingInfo.Description, h.cnf.MaxLenDescription)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, fmt.Sprintf("max len description is %d", h.cnf.MaxLenDescription))
		return
	}

	err = h.srv.UpdateCardDetails(r.Context(), serviceDto.UpdatingCardDetails{
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

// ReorderCard godoc
// @Summary      Перемещение карточки
// @Description  Изменяет порядок карточки в текущей секции или переносит ее в новую секцию на определенную позицию.
// @Tags         cards
// @Accept       json
// @Produce      json
// @Param        link    path string true "UUID перемещаемой карточки"
// @Param        request body dto.PlaceCard true "Новая секция и позиция для карточки"
// @Success      200 {object} api.Response "Карточка успешно перемещена"
// @Failure      400 {object} api.ErrorResponse "Некорректный запрос или нарушение логики (пропуск обязательной секции)"
// @Failure      404 {object} api.ErrorResponse "Карточка не найдена"
// @Failure      500 {object} api.ErrorResponse "Внутренняя ошибка сервера"
// @Security     CookieAuth
// @Router       /cards/{link}/reorder [patch]
func (h *Handler) ReorderCard(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())
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

	err = h.srv.ReorderCard(r.Context(), serviceDto.PlaceCard{
		LinkCard:    linkCard,
		LinkSection: updatingPlaceCard.LinkSection,
		Position:    updatingPlaceCard.Position,
	})

	if err != nil {
		if errors.Is(err, common.ErrorSkipMandatorySection) {
			api.RespondError(w, http.StatusBadRequest, incorrectMoveCard)
			return
		}

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

		logger.Error().Err(err).Msg("CardHandler.ReorderCard")
		api.RespondError(w, http.StatusInternalServerError, failReorderCard)
		return
	}

	api.RespondOk(w, api.StatusOK)
}

// CreateCard godoc
// @Summary      Создание новой карточки
// @Description  Добавляет новую карточку в указанную секцию от лица текущего авторизованного пользователя.
// @Tags         cards
// @Accept       json
// @Produce      json
// @Param        request body dto.NewCard true "Данные для создания карточки"
// @Success      200 {object} dto.PlaceCard "Карточка создана, возвращает информацию о ее расположении"
// @Failure      400 {object} api.ErrorResponse "Некорректный запрос или превышена максимальная длина текста"
// @Failure      401 {object} api.ErrorResponse "Пользователь не авторизован"
// @Failure      404 {object} api.ErrorResponse "Указанная секция не найдена"
// @Failure      409 {object} api.ErrorResponse "Конфликт данных (например, карточка уже существует)"
// @Failure      500 {object} api.ErrorResponse "Внутренняя ошибка сервера"
// @Security     CookieAuth
// @Router       /cards [post]
func (h *Handler) CreateCard(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	var newCard dto.NewCard
	err := json.NewDecoder(r.Body).Decode(&newCard)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectRequest)
		return
	}

	err = common.ValidateTextInfo(newCard.Title, h.cnf.MaxLenTitle)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, fmt.Sprintf("max len title is %d", h.cnf.MaxLenTitle))
		return
	}

	err = common.ValidateTextInfo(newCard.Description, h.cnf.MaxLenDescription)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, fmt.Sprintf("max len description is %d", h.cnf.MaxLenDescription))
		return
	}

	authorLink, ok := r.Context().Value(middleware.UserContextLink{}).(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, common.FailAuthorized)
		return
	}

	card, err := h.srv.CreateCard(r.Context(), serviceDto.NewCard{
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

		if errors.Is(err, common.ErrorCardAlreadyExist) {
			api.RespondError(w, http.StatusConflict, incorrectUniqCard)
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

		logger.Error().Err(err).Msg("CardHandler.CreateCard")
		api.RespondError(w, http.StatusInternalServerError, failCreateCard)
		return
	}

	api.HandleError(api.RespondOk(w, dto.PlaceCard{
		LinkCard:    card.LinkCard,
		LinkSection: card.LinkSection,
		Position:    card.Position,
	}))
}
