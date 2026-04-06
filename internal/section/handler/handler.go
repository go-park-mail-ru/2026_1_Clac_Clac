package handler

import (
	"context"
	"encoding/json"
	"errors"
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

	incorrectRequest        = "get incorrect format of request"
	incorrectPath           = "get incorrect format of path"
	incorrectQuantityTasks  = "max quantity tasks is 100"
	incorrectLenNameSection = "max len name is 128"
	incorrectTypeColor      = "color can be white, grey, red, orange, blue, green, purple, pink"

	maxQuantityTasks = 100
	minQuantityTasks = 00

	maxLenNameSection = 128

	sectionLinkKey = "link"
	boardLinkKey   = "board_link"
)

// mockery --name=SectionService --output=mock_section_service --outpkg=mockSectionService
type SectionService interface {
	GetSectionInfo(ctx context.Context, linkSection uuid.UUID) (serviceDto.FullSectionInfo, error)
	GetAllSections(ctx context.Context, boarderLink uuid.UUID) ([]serviceDto.FullSectionInfo, error)
	CreateSection(ctx context.Context, newSection serviceDto.CreatingSection) (serviceDto.EntitySection, error)
	DeleteSection(ctx context.Context, linkSection uuid.UUID) error
	ReorderSection(ctx context.Context, linkBoard uuid.UUID, listLinkSection []uuid.UUID) error
	UpdateSection(ctx context.Context, updatingSection serviceDto.FullSectionInfo) error
}

type Handler struct {
	srv SectionService
}

func NewHandler(srv SectionService) *Handler {
	return &Handler{
		srv: srv,
	}
}

func (h *Handler) GetSection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	linkParam := vars["link"]

	linkSection, err := uuid.Parse(linkParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, incorrectRequest)
		return
	}

	result, err := h.srv.GetSectionInfo(r.Context(), linkSection)
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

func (h *Handler) CreateSection(w http.ResponseWriter, r *http.Request) {
	var newSection dto.CreatingSection

	err := json.NewDecoder(r.Body).Decode(&newSection)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, incorrectPath)
		return
	}

	if newSection.MaxTasks != nil && *newSection.MaxTasks > maxQuantityTasks {
		api.RespondError(w, http.StatusBadRequest, incorrectRequest)
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

func (h *Handler) DeleteSection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	linkParam := vars[sectionLinkKey]

	linkSection, err := uuid.Parse(linkParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, incorrectPath)
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

		api.RespondError(w, http.StatusInternalServerError, failDeleteSection)
		return
	}

	api.RespondOk(w, api.StatusOK)
}

func (h *Handler) ReorderSection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	boardParam := vars[boardLinkKey]

	linkBoard, err := uuid.Parse(boardParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, incorrectPath)
		return
	}

	listSectionLinks := dto.ListSectionLink{}

	err = json.NewDecoder(r.Body).Decode(&listSectionLinks)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, incorrectRequest)
		return
	}

	err = h.srv.ReorderSection(r.Context(), linkBoard, listSectionLinks.List)
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

func (h *Handler) UpdateSection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sectionParam := vars[sectionLinkKey]

	sectionLink, err := uuid.Parse(sectionParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, incorrectPath)
		return
	}

	var sectionInfo dto.FullSectionInfo

	err = json.NewDecoder(r.Body).Decode(&sectionInfo)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, incorrectRequest)
		return
	}

	err = common.ValidateTextInfo(sectionInfo.SectionName, maxLenNameSection)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, incorrectLenNameSection)
		return
	}

	err = common.ValidateColor(sectionInfo.Color)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, incorrectTypeColor)
		return
	}

	if sectionInfo.MaxTasks != nil {
		if err := common.ValidateNumberInfo(*sectionInfo.MaxTasks, maxQuantityTasks, minQuantityTasks); err != nil {
			api.RespondError(w, http.StatusBadRequest, incorrectQuantityTasks)
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

		api.RespondError(w, http.StatusInternalServerError, failUpdateSection)
		return
	}

	api.RespondOk(w, api.StatusOK)
}

func (h *Handler) GetAllSections(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	boarderParam := vars[boardLinkKey]

	boarderLink, err := uuid.Parse(boarderParam)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, incorrectPath)
		return
	}

	sections, err := h.srv.GetAllSections(r.Context(), boarderLink)
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
