package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/section/handler/dto"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/section/service/dto"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	failFindSection     = "can not find section"
	failDeleteBacklog   = "can not delete backlog section"
	failUpdateBacklog   = "can not update backlog section"
	failGetSection      = "can not get info section"
	failGetAllSections  = "can not get all info section"
	failCreateSection   = "can not create new section"
	failDeleteSection   = "can not delete new"
	failReorderSections = "can not reorder sections"
	failUpdateSection   = "can not update section"
	failGetCards        = "can not get cards in section"

	incorrectTypeColor = "color can be white, grey, red, orange, blue, green, purple, pink"

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

type Deps struct {
	Srv               SectionService
	MaxQuantityTasks  int
	MinQuantityTasks  int
	MaxLenNameSection int
}

type Handler struct {
	deps Deps
}

func NewHandler(deps Deps) *Handler {
	return &Handler{
		deps: deps,
	}
}

// GetSection godoc
// @Summary      Получение секции
// @Description  Возвращает информацию о конкретной секции (колонке) по её UUID.
// @Tags         sections
// @Produce      json
// @Param        link path string true "UUID секции"
// @Success      200 {object} dto.FullSectionInfo "Успешное получение данных секции"
// @Failure      400 {object} api.ErrorResponse "Некорректный UUID или секция не найдена"
// @Failure      500 {object} api.ErrorResponse "Внутренняя ошибка сервера"
// @Security     CookieAuth
// @Router       /sections/{link} [get]
func (h *Handler) GetSection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	linkParam := vars[sectionLinkKey]

	linkSection, err := uuid.Parse(linkParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectPath)
		return
	}

	result, err := h.deps.Srv.GetSectionInfo(r.Context(), linkSection)
	if err != nil {
		if errors.Is(err, common.ErrorNotExistingSection) {
			api.RespondError(w, http.StatusBadRequest, failFindSection)
			return
		}
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
	vars := mux.Vars(r)
	linkParam := vars[sectionLinkKey]

	linkSection, err := uuid.Parse(linkParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectPath)
		return
	}

	cards, err := h.deps.Srv.GetCards(r.Context(), linkSection)
	if err != nil {
		if errors.Is(err, common.ErrorNotExistingSection) {
			api.RespondError(w, http.StatusNotFound, failFindSection)
			return
		}

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
// @Success      200 {object} dto.FullSectionInfo "Секция успешно создана"
// @Failure      400 {object} api.ErrorResponse "Некорректный запрос или превышены лимиты задач"
// @Failure      500 {object} api.ErrorResponse "Внутренняя ошибка сервера"
// @Security     CookieAuth
// @Router       /sections [post]
func (h *Handler) CreateSection(w http.ResponseWriter, r *http.Request) {
	var newSection dto.CreatingSection

	err := json.NewDecoder(r.Body).Decode(&newSection)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectPath)
		return
	}

	if newSection.MaxTasks != nil && (*newSection.MaxTasks > h.deps.MaxQuantityTasks ||
		*newSection.MaxTasks < h.deps.MinQuantityTasks) {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectRequest)
		return
	}

	result, err := h.deps.Srv.CreateSection(r.Context(), serviceDto.CreatingSection{
		BoardLink:   newSection.BoardLink,
		SectionName: newSection.SectionName,
		IsMandatory: newSection.IsMandatory,
		Color:       newSection.Color,
		MaxTasks:    newSection.MaxTasks,
	})
	if err != nil {
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
// @Failure      400 {object} api.ErrorResponse "Попытка удалить Backlog или некорректный путь"
// @Failure      404 {object} api.ErrorResponse "Секция не найдена"
// @Failure      500 {object} api.ErrorResponse "Внутренняя ошибка сервера"
// @Security     CookieAuth
// @Router       /sections/{link} [delete]
func (h *Handler) DeleteSection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	linkParam := vars[sectionLinkKey]

	linkSection, err := uuid.Parse(linkParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectPath)
		return
	}

	err = h.deps.Srv.DeleteSection(r.Context(), linkSection)
	if err != nil {
		if errors.Is(err, common.ErrorNotExistingSection) {
			api.RespondError(w, http.StatusNotFound, failFindSection)
			return
		}

		if errors.Is(err, common.ErrorDeleteBacklog) {
			api.RespondError(w, http.StatusBadRequest, failDeleteBacklog)
			return
		}

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
// @Failure      400 {object} api.ErrorResponse "Некорректный запрос или UUID доски"
// @Failure      404 {object} api.ErrorResponse "Не все переданные секции найдены"
// @Failure      500 {object} api.ErrorResponse "Внутренняя ошибка сервера"
// @Security     CookieAuth
// @Router       /boards/{board_link}/sections/reorder [patch]
func (h *Handler) ReorderSection(w http.ResponseWriter, r *http.Request) {
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

	err = h.deps.Srv.ReorderSection(r.Context(), linkBoard, listSectionLinks.List)
	if err != nil {
		if errors.Is(err, common.ErrorNotFindAllLinks) {
			api.RespondError(w, http.StatusNotFound, failFindSection)
			return
		}

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
// @Failure      400 {object} api.ErrorResponse "Ошибка валидации, попытка изменить Backlog или неверный цвет"
// @Failure      404 {object} api.ErrorResponse "Секция не найдена"
// @Failure      500 {object} api.ErrorResponse "Внутренняя ошибка сервера"
// @Security     CookieAuth
// @Router       /sections/{link} [put]
func (h *Handler) UpdateSection(w http.ResponseWriter, r *http.Request) {
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

	err = common.ValidateTextInfo(sectionInfo.SectionName, h.deps.MaxLenNameSection)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, fmt.Sprintf("max len name is %d", h.deps.MaxLenNameSection))
		return
	}

	err = common.ValidateColor(sectionInfo.Color)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, incorrectTypeColor)
		return
	}

	if sectionInfo.MaxTasks != nil {
		if err := common.ValidateNumberInfo(*sectionInfo.MaxTasks, h.deps.MaxQuantityTasks, h.deps.MinQuantityTasks); err != nil {
			api.RespondError(w, http.StatusBadRequest, fmt.Sprintf("max quantity tasks is %d", h.deps.MaxQuantityTasks))
			return
		}
	}

	err = h.deps.Srv.UpdateSection(r.Context(), serviceDto.FullSectionInfo{
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
// @Success      200 {object} dto.SectionsResponse "Успешное получение списка секций"
// @Failure      400 {object} api.ErrorResponse "Некорректный UUID доски"
// @Failure      500 {object} api.ErrorResponse "Внутренняя ошибка сервера"
// @Security     CookieAuth
// @Router       /boards/{board_link}/sections [get]
func (h *Handler) GetAllSections(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	boarderParam := vars[boardLinkKey]

	boarderLink, err := uuid.Parse(boarderParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, common.IncorrectPath)
		return
	}

	sections, err := h.deps.Srv.GetAllSections(r.Context(), boarderLink)
	if err != nil {
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
