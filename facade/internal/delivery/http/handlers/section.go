package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/dto"
	handlerCommon "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

var (
	ErrCannotGetSections     = errors.New("cannot get sections")
	ErrCannotGetSection      = errors.New("cannot get section")
	ErrCannotGetCards        = errors.New("cannot get cards")
	ErrCannotCreateSection   = errors.New("cannot create section")
	ErrCannotDeleteSection   = errors.New("cannot delete section")
	ErrCannotReorderSections = errors.New("cannot reorder sections")
	ErrCannotUpdateSection   = errors.New("cannot update section")
	ErrSectionLinkMissing    = errors.New("section link missing")
	ErrInvalidSectionLink    = errors.New("invalid section link")
)

const (
	sectionLinkKey      = "link"
	sectionBoardLinkKey = "board_link"
)

//go:generate mockery --name=SectionUsecase --output=mock_section_use_case
type SectionUsecase interface {
	GetSections(ctx context.Context, sectionReq domain.GetSectionsRequest) ([]domain.SectionInfo, error)
	GetSection(ctx context.Context, sectionReq domain.GetSectionRequest) (domain.SectionInfo, error)
	GetCards(ctx context.Context, cardReq domain.GetCardsRequest) ([]domain.CardInfo, error)
	CreateSection(ctx context.Context, sectionInfo domain.CreateSectionRequest) (domain.SectionInfo, error)
	DeleteSection(ctx context.Context, sectionInfo domain.DeleteSectionRequest) error
	ReorderSection(ctx context.Context, sectionInfo domain.ReorderSectionRequest) error
	UpdateSection(ctx context.Context, sectionInfo domain.UpdateSectionRequest) error
}

type Section struct {
	srv SectionUsecase
}

func NewSection(srv SectionUsecase) *Section {
	return &Section{
		srv: srv,
	}
}

func sectionInfoToDTO(s domain.SectionInfo) dto.SectionInfo {
	return dto.SectionInfo{
		Link:        s.Link,
		Name:        s.Name,
		Position:    s.Position,
		IsMandatory: s.IsMandatory,
		Color:       s.Color,
		MaxTasks:    s.MaxTasks,
	}
}

func cardInfoToDTO(c domain.CardInfo) dto.Card {
	subtasks := make([]dto.SubtaskInfo, 0, len(c.Subtasks))
	for _, st := range c.Subtasks {
		subtasks = append(subtasks, dto.SubtaskInfo{
			Link:        st.SubtaskLink,
			Description: st.Description,
			IsDone:      st.IsDone,
			Position:    int64(st.Position),
		})
	}

	return dto.Card{
		Link:         c.CardLink,
		ExecutorName: c.ExecutorName,
		Title:        c.Title,
		Description:  c.Description,
		Deadline:     c.Deadline,
		Subtasks:     subtasks,
	}
}

// @Summary		Получить все секции доски
// @Description	Возвращает массив всех секций, привязанных к конкретной доске
// @Tags			sections
// @Produce		json
// @Param			board_link	path		string	true	"UUID доски"	Format(uuid)
// @Success		200	{object}	api.OkResponse[[]dto.SectionInfo]
// @Failure		400	{object}	api.ErrorResponse	"invalid board link"
// @Failure		401	{object}	api.ErrorResponse	"unauthorized"
// @Failure		500	{object}	api.ErrorResponse	"cannot get sections"
// @Router			/boards/{board_link}/sections [get]
func (h *Section) GetSections(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	userLink, ok := r.Context().Value(middleware.UserContextLink{}).(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}

	rawBoardLink, ok := mux.Vars(r)[sectionBoardLinkKey]
	if !ok {
		api.RespondError(w, http.StatusBadRequest, ErrBoardLinkMissing.Error())
		return
	}

	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidSectionLink.Error())
		return
	}

	sections, err := h.srv.GetSections(r.Context(), domain.GetSectionsRequest{
		UserLink:  userLink,
		BoardLink: boardLink,
	})
	if err != nil {
		logger.Error().Err(err).Msg("section usecase GetSections")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotGetSections.Error())
		return
	}

	response := make([]dto.SectionInfo, 0, len(sections))
	for _, s := range sections {
		response = append(response, sectionInfoToDTO(s))
	}

	api.HandleError(api.RespondOk(w, response))
}

// @Summary		Получить секцию
// @Description	Возвращает информацию о конкретной секции по её UUID
// @Tags			sections
// @Produce		json
// @Param			link	path		string	true	"UUID секции"	Format(uuid)
// @Success		200	{object}	api.OkResponse[dto.SectionInfo]
// @Failure		400	{object}	api.ErrorResponse	"invalid section link"
// @Failure		401	{object}	api.ErrorResponse	"unauthorized"
// @Failure		404	{object}	api.ErrorResponse	"section not found"
// @Failure		500	{object}	api.ErrorResponse	"cannot get section"
// @Router			/sections/{link} [get]
func (h *Section) GetSection(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	userLink, ok := r.Context().Value(middleware.UserContextLink{}).(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}

	rawSectionLink, ok := mux.Vars(r)[sectionLinkKey]
	if !ok {
		api.RespondError(w, http.StatusBadRequest, ErrSectionLinkMissing.Error())
		return
	}

	sectionLink, err := uuid.Parse(rawSectionLink)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidSectionLink.Error())
		return
	}

	section, err := h.srv.GetSection(r.Context(), domain.GetSectionRequest{
		UserLink:    userLink,
		SectionLink: sectionLink,
	})
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, err.Error())
			return
		}
		logger.Error().Err(err).Msg("section usecase GetSection")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotGetSection.Error())
		return
	}

	api.HandleError(api.RespondOk(w, sectionInfoToDTO(section)))
}

// @Summary		Получить карточки секции
// @Description	Возвращает список всех карточек, находящихся в указанной секции
// @Tags			cards
// @Produce		json
// @Param			link	path		string	true	"UUID секции"	Format(uuid)
// @Success		200	{object}	api.OkResponse[dto.CardsResponse]
// @Failure		400	{object}	api.ErrorResponse	"invalid section link"
// @Failure		401	{object}	api.ErrorResponse	"unauthorized"
// @Failure		404	{object}	api.ErrorResponse	"section not found"
// @Failure		500	{object}	api.ErrorResponse	"cannot get cards"
// @Router			/sections/{link}/cards [get]
func (h *Section) GetCards(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	userLink, ok := r.Context().Value(middleware.UserContextLink{}).(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}

	rawSectionLink, ok := mux.Vars(r)[sectionLinkKey]
	if !ok {
		api.RespondError(w, http.StatusBadRequest, ErrSectionLinkMissing.Error())
		return
	}

	sectionLink, err := uuid.Parse(rawSectionLink)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidSectionLink.Error())
		return
	}

	cards, err := h.srv.GetCards(r.Context(), domain.GetCardsRequest{
		UserLink:    userLink,
		SectionLink: sectionLink,
	})
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, err.Error())
			return
		}
		logger.Error().Err(err).Msg("section usecase GetCards")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotGetCards.Error())
		return
	}

	response := make([]dto.Card, 0, len(cards))
	for _, c := range cards {
		response = append(response, cardInfoToDTO(c))
	}

	api.HandleError(api.RespondOk(w, dto.CardsResponse{Cards: response}))
}

// @Summary		Создать секцию
// @Description	Создает новую секцию на доске
// @Tags			sections
// @Accept		json
// @Produce		json
// @Param		request	body		dto.CreateSectionRequest	true	"Данные для создания секции"
// @Success		201	{object}	api.OkResponse[dto.SectionInfo]
// @Failure		400	{object}	api.ErrorResponse	"invalid request schema"
// @Failure		401	{object}	api.ErrorResponse	"unauthorized"
// @Failure		500	{object}	api.ErrorResponse	"cannot create section"
// @Router			/sections [post]
func (h *Section) CreateSection(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	userLink, ok := r.Context().Value(middleware.UserContextLink{}).(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}

	var req dto.CreateSectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	if req.Name == "" || req.BoardLink == uuid.Nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	section, err := h.srv.CreateSection(r.Context(), domain.CreateSectionRequest{
		UserLink:    userLink,
		BoardLink:   req.BoardLink,
		Name:        req.Name,
		IsMandatory: req.IsMandatory,
		Color:       req.Color,
		MaxTasks:    req.MaxTasks,
	})
	if err != nil {
		logger.Error().Err(err).Msg("section usecase CreateSection")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotCreateSection.Error())
		return
	}

	api.HandleError(api.RespondCreated(w, sectionInfoToDTO(section)))
}

// @Summary		Удалить секцию
// @Description	Удаляет секцию по её UUID
// @Tags			sections
// @Produce		json
// @Param			link	path		string	true	"UUID секции"	Format(uuid)
// @Success		200	{object}	api.Response	"status ok"
// @Failure		400	{object}	api.ErrorResponse	"invalid section link"
// @Failure		401	{object}	api.ErrorResponse	"unauthorized"
// @Failure		404	{object}	api.ErrorResponse	"section not found"
// @Failure		500	{object}	api.ErrorResponse	"cannot delete section"
// @Router			/sections/{link} [delete]
func (h *Section) DeleteSection(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	userLink, ok := r.Context().Value(middleware.UserContextLink{}).(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}

	rawSectionLink, ok := mux.Vars(r)[sectionLinkKey]
	if !ok {
		api.RespondError(w, http.StatusBadRequest, ErrSectionLinkMissing.Error())
		return
	}

	sectionLink, err := uuid.Parse(rawSectionLink)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidSectionLink.Error())
		return
	}

	err = h.srv.DeleteSection(r.Context(), domain.DeleteSectionRequest{
		UserLink:    userLink,
		SectionLink: sectionLink,
	})
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, err.Error())
			return
		}
		logger.Error().Err(err).Msg("section usecase DeleteSection")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotDeleteSection.Error())
		return
	}

	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}

// @Summary		Переупорядочить секции
// @Description	Обновляет порядок секций на доске
// @Tags			sections
// @Accept		json
// @Produce		json
// @Param			board_link	path		string				true	"UUID доски"	Format(uuid)
// @Param			request		body		dto.ListSectionLink	true	"Новый порядок секций"
// @Success		200	{object}	api.Response	"status ok"
// @Failure		400	{object}	api.ErrorResponse	"invalid request schema"
// @Failure		401	{object}	api.ErrorResponse	"unauthorized"
// @Failure		500	{object}	api.ErrorResponse	"cannot reorder sections"
// @Router			/boards/{board_link}/sections/reorder [patch]
func (h *Section) ReorderSections(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	userLink, ok := r.Context().Value(middleware.UserContextLink{}).(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}

	rawBoardLink, ok := mux.Vars(r)[sectionBoardLinkKey]
	if !ok {
		api.RespondError(w, http.StatusBadRequest, ErrBoardLinkMissing.Error())
		return
	}

	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidSectionLink.Error())
		return
	}

	var req dto.ListSectionLink
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	err = h.srv.ReorderSection(r.Context(), domain.ReorderSectionRequest{
		UserLink:  userLink,
		BoardLink: boardLink,
		LinksList: req.List,
	})
	if err != nil {
		logger.Error().Err(err).Msg("section usecase ReorderSection")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotReorderSections.Error())
		return
	}

	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}

// @Summary		Обновить секцию
// @Description	Изменяет данные существующей секции
// @Tags			sections
// @Accept		json
// @Produce		json
// @Param			link	path		string					true	"UUID секции"	Format(uuid)
// @Param			request	body		dto.SectionInfo			true	"Новые данные секции"
// @Success		200	{object}	api.Response	"status ok"
// @Failure		400	{object}	api.ErrorResponse	"invalid request schema"
// @Failure		401	{object}	api.ErrorResponse	"unauthorized"
// @Failure		404	{object}	api.ErrorResponse	"section not found"
// @Failure		500	{object}	api.ErrorResponse	"cannot update section"
// @Router			/sections/{link} [put]
func (h *Section) UpdateSection(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	userLink, ok := r.Context().Value(middleware.UserContextLink{}).(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}

	rawSectionLink, ok := mux.Vars(r)[sectionLinkKey]
	if !ok {
		api.RespondError(w, http.StatusBadRequest, ErrSectionLinkMissing.Error())
		return
	}

	sectionLink, err := uuid.Parse(rawSectionLink)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidSectionLink.Error())
		return
	}

	var req dto.SectionInfo
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	err = h.srv.UpdateSection(r.Context(), domain.UpdateSectionRequest{
		UserLink:    userLink,
		SectionLink: sectionLink,
		Name:        req.Name,
		IsMandatory: req.IsMandatory,
		Color:       req.Color,
		MaxTasks:    req.MaxTasks,
	})
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, err.Error())
			return
		}
		logger.Error().Err(err).Msg("section usecase UpdateSection")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotUpdateSection.Error())
		return
	}

	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}
