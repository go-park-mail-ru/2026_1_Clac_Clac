package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/section/handler/dto"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/section/service/dto"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

const (
	failFindSection     = "can not find section"
	failDeleteBacklog   = "can not delete backlog section"
	failUpdateBacklog   = "can not update backlog section"
	failGetSection      = "can not get info section"
	failGetAllSections  = "can not get all info section"
	failCreateSection   = "can not create new section"
	failDeleteSection   = "can not delete section"
	failReorderSections = "can not reorder sections"
	failUpdateSection   = "can not update section"
	failGetCards        = "can not get cards in section"

	incorrectTypeColor   = "color can be white, grey, red, orange, blue, green, purple, pink"
	incorrectUniqSection = "section already exists"
	incorrectReferences  = "incorrect foreign key"
	failNullValue        = "can not use null element"
	invalidSectionData   = "invalid section data"
	invalidCardData      = "invalid card data"

	sectionLinkKey = "link"
	boardLinkKey   = "board_link"
)

// mockery --name=SectionService --output=mock_section_service --outpkg=mockSectionService
type SectionService interface {
	GetSectionInfo(ctx context.Context, linkSection uuid.UUID) (serviceDto.FullSectionInfo, error)
	GetAllSections(ctx context.Context, boarderLink uuid.UUID) ([]serviceDto.FullSectionInfo, error)
	GetCards(ctx context.Context, linkSection uuid.UUID) ([]serviceDto.Card, error)
	CreateSection(ctx context.Context, newSection serviceDto.CreatingSection) (serviceDto.EntitySection, error)
	DeleteSection(ctx context.Context, linkSection uuid.UUID) error
	ReorderSection(ctx context.Context, linkBoard uuid.UUID, listLinkSection []uuid.UUID) error
	UpdateSection(ctx context.Context, updatingSection serviceDto.FullSectionInfo) error
}

type Config struct {
	MaxQuantityTasks  int
	MinQuantityTasks  int
	MaxLenNameSection int
}

type Handler struct {
	srv SectionService
	cnf Config
}

func NewHandler(srv SectionService, cnf Config) *Handler {
	return &Handler{
		srv: srv,
		cnf: cnf,
	}
}

// GetSection godoc
// @Summary      Получение секции
// @Description  Возвращает информацию о конкретной секции (колонке) по её UUID.
// @Tags         sections
// @Produce      json
// @Param        link path string true "UUID секции"
// @Success      200 {object} api.OkResponse[dto.FullSectionInfo] "Успешное получение данных секции"
// @Failure      400 {object} api.ErrorResponse "Некорректный UUID в пути"
// @Failure      404 {object} api.ErrorResponse "Секция не найдена"
// @Failure      500 {object} api.ErrorResponse "Внутренняя ошибка сервера"
// @Security     CookieAuth
// @Router       /sections/{link} [get]
func (h *Handler) GetSection(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())
	vars := mux.Vars(r)
	linkParam := vars[sectionLinkKey]

	linkSection, err := uuid.Parse(linkParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectPath)
		return
	}

	result, err := h.srv.GetSectionInfo(r.Context(), linkSection)
	if err != nil {
		if errors.Is(err, common.ErrorNotExistingSection) {
			api.RespondError(w, http.StatusNotFound, failFindSection)
			return
		}
		logger.Error().Err(err).Msg("SectionService.GetSectionInfo")
		api.RespondError(w, http.StatusInternalServerError, failGetSection)
		return
	}

	sectionInfo := dto.FullSectionInfo{
		SectionLink: result.SectionLink,
		SectionName: result.SectionName,
		Position:    result.Position,
		IsMandatory: result.IsMandatory,
		Color:       result.Color,
		MaxTasks:    result.MaxTasks,
	}

	api.HandleError(api.RespondOk(w, sectionInfo))
}

// GetCards godoc
// @Summary      Получение карточек секции
// @Description  Возвращает список всех актуальных карточек, находящихся в указанной секции, отсортированных по позиции.
// @Tags         cards
// @Produce      json
// @Param        link path string true "UUID секции"
// @Success      200 {object} dto.CardsSection "Успешное получение списка карточек"
// @Failure      400 {object} api.ErrorResponse "Некорректный формат UUID секции"
// @Failure      404 {object} api.ErrorResponse "Секция не найдена"
// @Failure      500 {object} api.ErrorResponse "Внутренняя ошибка сервера"
// @Security     CookieAuth
// @Router       /sections/{link}/cards [get]
func (h *Handler) GetCards(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())
	vars := mux.Vars(r)
	linkParam := vars[sectionLinkKey]

	linkSection, err := uuid.Parse(linkParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectPath)
		return
	}

	cards, err := h.srv.GetCards(r.Context(), linkSection)
	if err != nil {
		if errors.Is(err, common.ErrorNotExistingSection) {
			api.RespondError(w, http.StatusNotFound, failFindSection)
			return
		}

		logger.Error().Err(err).Msg("SectionService.GetCards")
		api.RespondError(w, http.StatusInternalServerError, failGetCards)
		return
	}

	cardsResponse := make([]dto.Card, 0, len(cards))

	for _, card := range cards {
		cardsResponse = append(cardsResponse, dto.Card{
			CardLink:     card.CardLink,
			ExecuterName: card.ExecuterName,
			Title:        card.Title,
			DeadLine:     card.DeadLine,
		})
	}

	api.HandleError(api.RespondOk(w, dto.CardsSection{Cards: cardsResponse}))
}

// CreateSection godoc
// @Summary      Создание секции
// @Description  Создает новую секцию (колонку) на доске.
// @Tags         sections
// @Accept       json
// @Produce      json
// @Param        request body dto.CreatingSection true "Данные для создания секции"
// @Success      200 {object} api.OkResponse[dto.FullSectionInfo] "Секция успешно создана"
// @Failure      400 {object} api.ErrorResponse "Некорректный запрос, превышены лимиты задач, неверный внешний ключ или отсутствуют обязательные поля"
// @Failure      409 {object} api.ErrorResponse "Секция с таким названием/ссылкой уже существует"
// @Failure      500 {object} api.ErrorResponse "Внутренняя ошибка сервера"
// @Security     CookieAuth
// @Router       /sections [post]
func (h *Handler) CreateSection(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	var newSection dto.CreatingSection

	err := json.NewDecoder(r.Body).Decode(&newSection)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectPath)
		return
	}

	if newSection.MaxTasks != nil && (*newSection.MaxTasks > h.cnf.MaxQuantityTasks ||
		*newSection.MaxTasks < h.cnf.MinQuantityTasks) {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectRequest)
		return
	}

	result, err := h.srv.CreateSection(r.Context(), serviceDto.CreatingSection{
		BoardLink:   newSection.BoardLink,
		SectionName: newSection.SectionName,
		IsMandatory: newSection.IsMandatory,
		Color:       newSection.Color,
		MaxTasks:    newSection.MaxTasks,
	})

	if err != nil {
		if errors.Is(err, common.ErrorSectionAlreadyExist) {
			api.RespondError(w, http.StatusConflict, incorrectUniqSection)
			return
		}

		if errors.Is(err, common.ErrorInvalidReferenceSectionData) {
			api.RespondError(w, http.StatusBadRequest, incorrectReferences)
			return
		}

		if errors.Is(err, common.ErrorInvalidSectionData) {
			api.RespondError(w, http.StatusBadRequest, invalidSectionData)
			return
		}

		if errors.Is(err, common.ErrorMissingRequiredField) {
			api.RespondError(w, http.StatusBadRequest, failNullValue)
			return
		}

		logger.Error().Err(err).Msg("SectionService.CreateSection")
		api.RespondError(w, http.StatusInternalServerError, failCreateSection)
		return
	}

	api.HandleError(api.RespondOk(w, dto.FullSectionInfo{
		SectionLink: result.SectionLink,
		SectionName: result.SectionName,
		IsMandatory: result.IsMandatory,
		Position:    result.Position,
		Color:       result.Color,
		MaxTasks:    result.MaxTasks,
	}))
}

// DeleteSection godoc
// @Summary      Удаление секции
// @Description  Удаляет секцию (колонку) по её UUID. Нельзя удалять системную секцию Backlog.
// @Tags         sections
// @Produce      json
// @Param        link path string true "UUID секции"
// @Success      200 {object} api.Response "Секция успешно удалена"
// @Failure      400 {object} api.ErrorResponse "Попытка удалить Backlog, неверный внешний ключ, нарушены данные карточек или пустые обязательные поля"
// @Failure      404 {object} api.ErrorResponse "Секция не найдена"
// @Failure      500 {object} api.ErrorResponse "Внутренняя ошибка сервера"
// @Security     CookieAuth
// @Router       /sections/{link} [delete]
func (h *Handler) DeleteSection(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())
	vars := mux.Vars(r)
	linkParam := vars[sectionLinkKey]

	linkSection, err := uuid.Parse(linkParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectPath)
		return
	}

	err = h.srv.DeleteSection(r.Context(), linkSection)
	if err != nil {
		if errors.Is(err, common.ErrorNotExistingSection) {
			api.RespondError(w, http.StatusNotFound, failFindSection)
			return
		}

		if errors.Is(err, common.ErrorDeleteBacklog) {
			api.RespondError(w, http.StatusBadRequest, failDeleteBacklog)
			return
		}

		if errors.Is(err, common.ErrorInvalidReferenceSectionData) {
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

		logger.Error().Err(err).Msg("SectionService.DeleteSection")
		api.RespondError(w, http.StatusInternalServerError, failDeleteSection)
		return
	}

	api.RespondOk(w, api.StatusOK)
}

// ReorderSection godoc
// @Summary      Перемещение секций
// @Description  Обновляет порядок секций на доске. Передается упорядоченный массив UUID секций.
// @Tags         sections
// @Accept       json
// @Produce      json
// @Param        board_link path string true "UUID доски"
// @Param        request body dto.ListSectionLink true "Новый порядок секций (массив UUID)"
// @Success      200 {object} api.Response "Порядок секций успешно обновлен"
// @Failure      400 {object} api.ErrorResponse "Некорректный запрос, неверные данные секции, внешний ключ или пропущены поля"
// @Failure      404 {object} api.ErrorResponse "Не все переданные секции найдены для реордера"
// @Failure      500 {object} api.ErrorResponse "Внутренняя ошибка сервера"
// @Security     CookieAuth
// @Router       /boards/{board_link}/sections/reorder [patch]
func (h *Handler) ReorderSection(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())
	vars := mux.Vars(r)
	boardParam := vars[boardLinkKey]

	linkBoard, err := uuid.Parse(boardParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectPath)
		return
	}

	listSectionLinks := dto.ListSectionLink{}

	err = json.NewDecoder(r.Body).Decode(&listSectionLinks)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectRequest)
		return
	}

	err = h.srv.ReorderSection(r.Context(), linkBoard, listSectionLinks.List)
	if err != nil {
		if errors.Is(err, common.ErrorNotFindAllLinks) {
			api.RespondError(w, http.StatusNotFound, failFindSection)
			return
		}

		if errors.Is(err, common.ErrorInvalidSectionData) {
			api.RespondError(w, http.StatusBadRequest, invalidSectionData)
			return
		}

		if errors.Is(err, common.ErrorInvalidReferenceSectionData) {
			api.RespondError(w, http.StatusBadRequest, incorrectReferences)
			return
		}

		if errors.Is(err, common.ErrorMissingRequiredField) {
			api.RespondError(w, http.StatusBadRequest, failNullValue)
			return
		}

		logger.Error().Err(err).Msg("SectionService.ReorderSection")
		api.RespondError(w, http.StatusInternalServerError, failReorderSections)
		return
	}

	api.RespondOk(w, api.StatusOK)
}

// UpdateSection godoc
// @Summary      Обновление секции
// @Description  Изменяет данные существующей секции (название, цвет, макс. кол-во задач). Попытка обновить параметры системного Backlog вызовет ошибку.
// @Tags         sections
// @Accept       json
// @Produce      json
// @Param        link path string true "UUID секции"
// @Param        request body dto.FullSectionInfo true "Новые данные секции"
// @Success      200 {object} api.Response "Секция успешно обновлена"
// @Failure      400 {object} api.ErrorResponse "Ошибка валидации, попытка изменить Backlog, неверный цвет, неверный внешний ключ или пустые обязательные поля"
// @Failure      404 {object} api.ErrorResponse "Секция не найдена"
// @Failure      500 {object} api.ErrorResponse "Внутренняя ошибка сервера"
// @Security     CookieAuth
// @Router       /sections/{link} [put]
func (h *Handler) UpdateSection(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())
	vars := mux.Vars(r)
	sectionParam := vars[sectionLinkKey]

	sectionLink, err := uuid.Parse(sectionParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectPath)
		return
	}

	var sectionInfo dto.FullSectionInfo

	err = json.NewDecoder(r.Body).Decode(&sectionInfo)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectPath)
		return
	}

	err = common.ValidateTextInfo(sectionInfo.SectionName, h.cnf.MaxLenNameSection)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, fmt.Sprintf("max len name is %d", h.cnf.MaxLenNameSection))
		return
	}

	err = common.ValidateColor(sectionInfo.Color)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, incorrectTypeColor)
		return
	}

	if sectionInfo.MaxTasks != nil {
		if err := common.ValidateNumberInfo(*sectionInfo.MaxTasks, h.cnf.MaxQuantityTasks, h.cnf.MinQuantityTasks); err != nil {
			api.RespondError(w, http.StatusBadRequest, fmt.Sprintf("max quantity tasks is %d", h.cnf.MaxQuantityTasks))
			return
		}
	}

	err = h.srv.UpdateSection(r.Context(), serviceDto.FullSectionInfo{
		SectionLink: sectionLink,
		SectionName: sectionInfo.SectionName,
		Position:    sectionInfo.Position,
		IsMandatory: sectionInfo.IsMandatory,
		Color:       sectionInfo.Color,
		MaxTasks:    sectionInfo.MaxTasks,
	})

	if err != nil {
		if errors.Is(err, common.ErrorNotExistingSection) {
			api.RespondError(w, http.StatusNotFound, failFindSection)
			return
		}

		if errors.Is(err, common.ErrorUpdateBacklog) {
			api.RespondError(w, http.StatusBadRequest, failUpdateBacklog)
			return
		}

		if errors.Is(err, common.ErrorInvalidReferenceSectionData) {
			api.RespondError(w, http.StatusBadRequest, incorrectReferences)
			return
		}

		if errors.Is(err, common.ErrorInvalidSectionData) {
			api.RespondError(w, http.StatusBadRequest, invalidSectionData)
			return
		}

		if errors.Is(err, common.ErrorMissingRequiredField) {
			api.RespondError(w, http.StatusBadRequest, failNullValue)
			return
		}

		logger.Error().Err(err).Msg("SectionService.UpdateSection")
		api.RespondError(w, http.StatusInternalServerError, failUpdateSection)
		return
	}

	api.RespondOk(w, api.StatusOK)
}

// GetAllSections godoc
// @Summary      Получить все секции доски
// @Description  Возвращает массив всех секций, привязанных к конкретной доске.
// @Tags         sections
// @Produce      json
// @Param        board_link path string true "UUID доски"
// @Success      200 {object} api.OkResponse[dto.SectionsResponse] "Успешное получение списка секций"
// @Failure      400 {object} api.ErrorResponse "Некорректный UUID доски"
// @Failure      500 {object} api.ErrorResponse "Внутренняя ошибка сервера"
// @Security     CookieAuth
// @Router       /boards/{board_link}/sections [get]
func (h *Handler) GetAllSections(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())
	vars := mux.Vars(r)
	boarderParam := vars[boardLinkKey]

	boarderLink, err := uuid.Parse(boarderParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectPath)
		return
	}

	sections, err := h.srv.GetAllSections(r.Context(), boarderLink)
	if err != nil {
		logger.Error().Err(err).Msg("SectionService.GetAllSections")
		api.RespondError(w, http.StatusInternalServerError, failGetAllSections)
		return
	}

	sectionsResponse := []dto.FullSectionInfo{}

	for _, section := range sections {
		sectionsResponse = append(sectionsResponse, dto.FullSectionInfo{
			SectionLink: section.SectionLink,
			SectionName: section.SectionName,
			Position:    section.Position,
			IsMandatory: section.IsMandatory,
			Color:       section.Color,
			MaxTasks:    section.MaxTasks,
		})
	}

	api.HandleError(api.RespondOk(w, dto.SectionsResponse{
		Sections: sectionsResponse,
	}))
}
