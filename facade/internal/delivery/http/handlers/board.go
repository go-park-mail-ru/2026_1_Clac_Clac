package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
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
	ErrCannotGetBoards        = errors.New("cannot get boards")
	ErrCannotCreateBoard      = errors.New("cannot create board")
	ErrCannotDeleteBoard      = errors.New("cannot delete board")
	ErrCannotUpdateBoard      = errors.New("cannot update board")
	ErrBoardLinkMissing       = errors.New("board link missing")
	ErrInvalidBoardLink       = errors.New("invalid board link")
	ErrParseMultipartForm     = errors.New("file too large or invalid form")
	ErrCannotFindBackground   = errors.New("cannot find 'background' key")
	ErrCannotUpdateBackground = errors.New("cannot update background")
	ErrCannotGetMembers       = errors.New("cannot get members")
)

const (
	boardLinkKey = "link"
)

//go:generate --name=BoardUsecase --output=mock_board_use_case
type BoardUsecase interface {
	GetBoards(ctx context.Context, userLink uuid.UUID) ([]domain.BoardInfo, error)
	GetBoard(ctx context.Context, boardInfo domain.GetBoardRequest) (domain.BoardInfo, error)
	CreateBoard(ctx context.Context, boardInfo domain.CreateBoardRequest) (domain.BoardInfo, error)
	DeleteBoard(ctx context.Context, boardInfo domain.GetBoardRequest) error
	UpdateBoard(ctx context.Context, boardInfo domain.UpdateBoardRequest) error
	UploadBackground(ctx context.Context, backgroundInfo domain.UploadBackgroundRequest, image io.Reader) (domain.UploadBackgroundResponse, error)
	GetMembers(ctx context.Context, membersInfo domain.GetMembersRequest) (domain.GetMembersResponse, error)
}

type BoardConfig struct {
	MultipartBackgroundFileKey string
	MaxBackgroundSize          int64
}

type Board struct {
	srv  BoardUsecase
	conf BoardConfig
}

func NewBoard(srv BoardUsecase, conf BoardConfig) *Board {
	return &Board{
		srv:  srv,
		conf: conf,
	}
}

func boardInfoToDTO(b domain.BoardInfo) dto.BoardInfo {
	return dto.BoardInfo{
		Link:        b.Link,
		Name:        b.Name,
		Description: b.Description,
		Background:  b.Background,
	}
}

// @Summary		Получить список досок пользователя
// @Description	Возвращает все доски, к которым у авторизованного пользователя есть доступ
// @Tags			boards
// @Produce		json
// @Success		200	{object}	api.OkResponse[[]dto.BoardInfo]
// @Failure		401	{object}	api.ErrorResponse	"unauthorized"
// @Failure		500	{object}	api.ErrorResponse	"cannot get boards"
// @Router			/boards [get]
func (h *Board) GetBoards(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	userLink, ok := r.Context().Value(middleware.UserContextLink{}).(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}

	boards, err := h.srv.GetBoards(r.Context(), userLink)
	if err != nil {
		logger.Error().Err(err).Msg("board usecase GetBoards")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotGetBoards.Error())
		return
	}

	response := make([]dto.BoardInfo, 0, len(boards))
	for _, b := range boards {
		response = append(response, boardInfoToDTO(b))
	}

	api.HandleError(api.RespondOk(w, response))
}

// @Summary		Получить информацию о доске
// @Description	Возвращает информацию о доске по её UUID ссылке
// @Tags			boards
// @Produce		json
// @Param			link	path		string	true	"UUID доски"	Format(uuid)
// @Success		200		{object}	api.OkResponse[dto.BoardInfo]
// @Failure		400		{object}	api.ErrorResponse	"invalid board link / board link missing"
// @Failure		401		{object}	api.ErrorResponse	"unauthorized"
// @Failure		403		{object}	api.ErrorResponse	"action denied"
// @Failure		404		{object}	api.ErrorResponse	"board not found"
// @Failure		500		{object}	api.ErrorResponse	"cannot get boards"
// @Router			/boards/{link} [get]
func (h *Board) GetBoard(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	userLink, ok := r.Context().Value(middleware.UserContextLink{}).(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}

	rawBoardLink, ok := mux.Vars(r)[boardLinkKey]
	if !ok {
		api.RespondError(w, http.StatusBadRequest, ErrBoardLinkMissing.Error())
		return
	}

	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidBoardLink.Error())
		return
	}

	board, err := h.srv.GetBoard(r.Context(), domain.GetBoardRequest{
		UserLink:  userLink,
		BoardLink: boardLink,
	})
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, err.Error())
			return
		}
		logger.Error().Err(err).Msg("board usecase GetBoard")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotGetBoards.Error())
		return
	}

	api.HandleError(api.RespondOk(w, boardInfoToDTO(board)))
}

// @Summary		Создать новую доску
// @Description	Создает новую доску на основе переданных данных
// @Tags			boards
// @Accept			json
// @Produce		json
// @Param			request	body		dto.CreateBoardRequest	true	"DTO для создания доски"
// @Success		201		{object}	api.OkResponse[dto.BoardInfo]
// @Failure		400		{object}	api.ErrorResponse	"invalid request schema"
// @Failure		401		{object}	api.ErrorResponse	"unauthorized"
// @Failure		500		{object}	api.ErrorResponse	"cannot create board"
// @Router			/boards [post]
func (h *Board) CreateBoard(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	userLink, ok := r.Context().Value(middleware.UserContextLink{}).(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}

	var req dto.CreateBoardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	if req.Name == "" {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	board, err := h.srv.CreateBoard(r.Context(), domain.CreateBoardRequest{
		UserLink:    userLink,
		Name:        req.Name,
		Description: req.Description,
		Background:  req.Background,
	})
	if err != nil {
		logger.Error().Err(err).Msg("board usecase CreateBoard")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotCreateBoard.Error())
		return
	}

	api.HandleError(api.RespondCreated(w, boardInfoToDTO(board)))
}

// @Summary		Удалить доску
// @Description	Удаляет доску по её UUID ссылке
// @Tags			boards
// @Produce		json
// @Param			link	path		string				true	"UUID доски для удаления"	Format(uuid)
// @Success		200		{object}	api.Response		"status ok"
// @Failure		400		{object}	api.ErrorResponse	"invalid board link / board link missing"
// @Failure		401		{object}	api.ErrorResponse	"unauthorized"
// @Failure		403		{object}	api.ErrorResponse	"action denied"
// @Failure		404		{object}	api.ErrorResponse	"board not found"
// @Failure		500		{object}	api.ErrorResponse	"cannot delete board"
// @Router			/boards/{link} [delete]
func (h *Board) DeleteBoard(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	userLink, ok := r.Context().Value(middleware.UserContextLink{}).(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}

	rawBoardLink, ok := mux.Vars(r)[boardLinkKey]
	if !ok {
		api.RespondError(w, http.StatusBadRequest, ErrBoardLinkMissing.Error())
		return
	}

	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidBoardLink.Error())
		return
	}

	err = h.srv.DeleteBoard(r.Context(), domain.GetBoardRequest{
		UserLink:  userLink,
		BoardLink: boardLink,
	})
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, err.Error())
			return
		}
		logger.Error().Err(err).Msg("board usecase DeleteBoard")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotDeleteBoard.Error())
		return
	}

	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}

// @Summary		Обновить информацию о доске
// @Description	Обновляет метаданные доски (имя, описание, фон)
// @Tags			boards
// @Accept			json
// @Produce		json
// @Param			link	path		string					true	"UUID доски"	Format(uuid)
// @Param			request	body		dto.UpdateBoardRequest	true	"DTO с новыми данными для обновления"
// @Success		200		{object}	api.Response			"status ok"
// @Failure		400		{object}	api.ErrorResponse		"invalid request schema"
// @Failure		401		{object}	api.ErrorResponse		"unauthorized"
// @Failure		403		{object}	api.ErrorResponse		"action denied"
// @Failure		404		{object}	api.ErrorResponse		"board not found"
// @Failure		500		{object}	api.ErrorResponse		"cannot update board"
// @Router			/boards/{link} [put]
func (h *Board) UpdateBoard(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	userLink, ok := r.Context().Value(middleware.UserContextLink{}).(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}

	rawBoardLink, ok := mux.Vars(r)[boardLinkKey]
	if !ok {
		api.RespondError(w, http.StatusBadRequest, ErrBoardLinkMissing.Error())
		return
	}

	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidBoardLink.Error())
		return
	}

	var req dto.UpdateBoardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	err = h.srv.UpdateBoard(r.Context(), domain.UpdateBoardRequest{
		UserLink:    userLink,
		BoardLink:   boardLink,
		Name:        req.Name,
		Description: req.Description,
		Background:  req.Background,
	})
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, err.Error())
			return
		}
		logger.Error().Err(err).Msg("board usecase UpdateBoard")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotUpdateBoard.Error())
		return
	}

	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}

// @Summary		Загрузить фон для доски
// @Description	Загружает изображение (multipart/form-data) и устанавливает его как фон доски
// @Tags			boards
// @Accept			multipart/form-data
// @Produce		json
// @Param			link		path		string	true	"UUID доски"	Format(uuid)
// @Param			background	formData	file	true	"Файл изображения (например, PNG/JPEG)"
// @Success		200			{object}	api.OkResponse[dto.UploadBackgroundResponse]
// @Failure		400			{object}	api.ErrorResponse	"invalid board link / invalid content type / cannot find background key"
// @Failure		401			{object}	api.ErrorResponse	"unauthorized"
// @Failure		404			{object}	api.ErrorResponse	"board not found"
// @Failure		500			{object}	api.ErrorResponse	"cannot update background / cannot read file"
// @Router			/boards/{link}/background [put]
func (h *Board) UploadBackground(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	userLink, ok := r.Context().Value(middleware.UserContextLink{}).(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}

	rawBoardLink, ok := mux.Vars(r)[boardLinkKey]
	if !ok {
		api.RespondError(w, http.StatusBadRequest, ErrBoardLinkMissing.Error())
		return
	}

	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidBoardLink.Error())
		return
	}

	if err := r.ParseMultipartForm(h.conf.MaxBackgroundSize); err != nil {
		logger.Error().Err(err).Msg("parse multipart form")
		api.RespondError(w, http.StatusBadRequest, ErrParseMultipartForm.Error())
		return
	}

	file, header, err := r.FormFile(h.conf.MultipartBackgroundFileKey)
	if err != nil {
		logger.Error().Err(err).Msg("cannot find background key")
		api.RespondError(w, http.StatusBadRequest, ErrCannotFindBackground.Error())
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Error().Err(err).Msg("BoardHandler.UploadBackground close file")
		}
	}()

	resp, err := h.srv.UploadBackground(r.Context(), domain.UploadBackgroundRequest{
		UserLink:  userLink,
		BoardLink: boardLink,
		Filename:  header.Filename,
	}, file)
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, err.Error())
			return
		}
		logger.Error().Err(err).Msg("board usecase UploadBackground")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotUpdateBackground.Error())
		return
	}

	api.HandleError(api.RespondOk(w, dto.UploadBackgroundResponse{
		BackgroundKey: resp.BackgroundKey,
	}))
}

// @Summary		Получить пользователей доски
// @Description	Возвращает массив UUID всех пользователей, имеющих доступ к доске
// @Tags			boards
// @Produce		json
// @Param			link	path		string						true	"UUID доски"	Format(uuid)
// @Success		200		{object}	api.OkResponse[dto.GetMembersResponse]
// @Failure		400		{object}	api.ErrorResponse			"invalid board link / board link missing"
// @Failure		401		{object}	api.ErrorResponse			"unauthorized"
// @Failure		404		{object}	api.ErrorResponse			"board not found"
// @Failure		500		{object}	api.ErrorResponse			"cannot get members"
// @Router			/boards/{link}/users [get]
func (h *Board) GetMembers(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	userLink, ok := r.Context().Value(middleware.UserContextLink{}).(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}

	rawBoardLink, ok := mux.Vars(r)[boardLinkKey]
	if !ok {
		api.RespondError(w, http.StatusBadRequest, ErrBoardLinkMissing.Error())
		return
	}

	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidBoardLink.Error())
		return
	}

	members, err := h.srv.GetMembers(r.Context(), domain.GetMembersRequest{
		UserLink:  userLink,
		BoardLink: boardLink,
	})
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, err.Error())
			return
		}
		logger.Error().Err(err).Msg("board usecase GetMembers")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotGetMembers.Error())
		return
	}

	api.HandleError(api.RespondOk(w, dto.GetMembersResponse{
		UserLinks: members.UserLinks,
	}))
}
