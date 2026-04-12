package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/handler/dto"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/service/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/s3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

var (
	ErrUserUnauthorized       = errors.New("unauthorized")
	ErrCannotGetBoards        = errors.New("cannot get boards")
	ErrInvalidRequestSchema   = errors.New("invalid request schema")
	ErrCannotCreateBoard      = errors.New("cannot create board")
	ErrCannotDeleteBoard      = errors.New("cannot delete board")
	ErrCannotUpdateBoard      = errors.New("cannot update board")
	ErrBoardLinkMissing       = errors.New("board link missing")
	ErrInvalidBoardLink       = errors.New("invalid board link")
	ErrParseMultipartForm     = errors.New("file too large or invalid form")
	ErrCannotFindBackground   = errors.New("cannot find 'background' key")
	ErrCannotReadFile         = errors.New("cannot read file")
	ErrInvalidContentType     = errors.New("invalid content type")
	ErrCannotOperateWithFile  = errors.New("cannot operate with file")
	ErrCannotUpdateBackground = errors.New("cannot update background")
	ErrCannotGetUsersOfBoard  = errors.New("cannot get users of board")
)

const (
	boardLinkKey = "link"
)

//go:generate mockery --name=BoardService --output mock_board_srv
type BoardService interface {
	GetBoards(ctx context.Context, userLink uuid.UUID) ([]serviceDto.BoardInfo, error)
	GetBoard(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID) (serviceDto.BoardInfo, error)
	CreateBoard(ctx context.Context, boardInfo serviceDto.NewBoardInfo, authorLink uuid.UUID) (serviceDto.BoardInfo, error)
	DeleteBoard(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID) error
	UpdateBoard(ctx context.Context, boardInfo serviceDto.UpdateBoardInfo, userLink uuid.UUID) error
	UpdateBackground(ctx context.Context, file io.Reader, contentType string, extension string, boardLink uuid.UUID, userLink uuid.UUID) (string, error)
	GetUsersOfBoard(ctx context.Context, boardLink uuid.UUID) ([]uuid.UUID, error)
}

func NewHandler(srv BoardService, conf config.BoardHandler) *BoardHandler {
	return &BoardHandler{
		srv:  srv,
		conf: conf,
	}
}

type BoardHandler struct {
	srv  BoardService
	conf config.BoardHandler
}

// @Summary		Получить список досок пользователя
// @Description	Возвращает все доски, к которым у авторизованного пользователя есть доступ
// @Tags			boards
// @Produce		json
// @Success		200	{object}	api.OkResponse[[]dto.BoardInfoResponse]
// @Failure		401	{object}	api.ErrorResponse	"unauthorized"
// @Failure		500	{object}	api.ErrorResponse	"cannot get boards"
// @Router			/boards [get]
func (h *BoardHandler) GetBoards(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	rawUserLink := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := rawUserLink.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, ErrUserUnauthorized.Error())
		return
	}

	boards, err := h.srv.GetBoards(r.Context(), userLink)
	if err != nil {
		logger.Error().Err(err).Msg("board service GetBoards")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotGetBoards.Error())
		return
	}

	boardsResponse := make([]dto.BoardInfoResponse, 0)
	for _, board := range boards {
		info := dto.BoardInfoResponseFromInfo(board)
		if strings.HasPrefix(info.Background, "backgrounds") {
			info.Background = fmt.Sprintf("%s/%s", s3.GetURL("hb.ru-msk.vkcloud-storage.ru", "nexus-boards-prod"), info.Background)
		}
		boardsResponse = append(boardsResponse, info)
	}

	api.HandleError(api.RespondOk(w, boardsResponse))
}

// @Summary		Получить информацию о доске
// @Description	Возвращает информацию о доске по её UUID ссылке
// @Tags			boards
// @Produce		json
// @Param			link	path		string	true	"UUID доски"	Format(uuid)
// @Success		200		{object}	api.OkResponse[dto.BoardInfoResponse]
// @Failure		400		{object}	api.ErrorResponse	"invalid board link / board link missing"
// @Failure		401		{object}	api.ErrorResponse	"unauthorized"
// @Failure		403		{object}	api.ErrorResponse	"action denied"
// @Failure		404		{object}	api.ErrorResponse	"board not found"
// @Failure		500		{object}	api.ErrorResponse	"cannot get boards"
// @Router			/boards/{link} [get]
func (h *BoardHandler) GetBoard(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	rawUserLink := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := rawUserLink.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, ErrUserUnauthorized.Error())
		return
	}

	vars := mux.Vars(r)

	rawBoardLink, ok := vars[boardLinkKey]
	if !ok {
		api.RespondError(w, http.StatusBadRequest, ErrBoardLinkMissing.Error())
		return
	}

	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidBoardLink.Error())
		return
	}

	board, err := h.srv.GetBoard(r.Context(), boardLink, userLink)
	if err != nil {
		if errors.Is(err, common.ErrActionDenied) {
			api.RespondError(w, http.StatusForbidden, common.ErrActionDenied.Error())
			return
		}

		if errors.Is(err, common.ErrBoardNotFound) {
			api.RespondError(w, http.StatusNotFound, common.ErrBoardNotFound.Error())
			return
		}

		logger.Error().Err(err).Msg("board service GetBoard")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotGetBoards.Error())
		return
	}

	info := dto.BoardInfoResponseFromInfo(board)
	if strings.HasPrefix(info.Background, "backgrounds") {
		info.Background = fmt.Sprintf("%s/%s", s3.GetURL("hb.ru-msk.vkcloud-storage.ru", "nexus-boards-prod"), info.Background)
	}

	api.HandleError(api.RespondOk(w, info))
}

// @Summary		Создать новую доску
// @Description	Создает новую доску на основе переданных данных
// @Tags			boards
// @Accept			json
// @Produce		json
// @Param			request	body		dto.CreateBoardRequest	true	"DTO для создания доски"
// @Success		201		{object}	api.OkResponse[dto.BoardInfoResponse]
// @Failure		400		{object}	api.ErrorResponse	"invalid request schema"
// @Failure		401		{object}	api.ErrorResponse	"unauthorized"
// @Failure		500		{object}	api.ErrorResponse	"cannot create board"
// @Router			/boards [post]
func (h *BoardHandler) CreateBoard(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	rawUserLink := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := rawUserLink.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, ErrUserUnauthorized.Error())
		return
	}

	var createRequest dto.CreateBoardRequest
	if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
		logger.Error().Err(ErrInvalidRequestSchema).Msg("decode create board request")
		api.RespondError(w, http.StatusBadRequest, ErrInvalidRequestSchema.Error())
		return
	}
	createRequest.Sanitize()

	board, err := h.srv.CreateBoard(r.Context(), dto.ToNewBoardInfo(createRequest), userLink)
	if err != nil {
		if errors.Is(err, common.ErrorNotNullValue) {
			api.RespondError(w, http.StatusBadRequest, common.ErrorNotNullValue.Error())
			return
		}
		if errors.Is(err, common.ErrorInvalidBoardData) {
			api.RespondError(w, http.StatusBadRequest, common.ErrorInvalidBoardData.Error())
			return
		}
		if errors.Is(err, common.ErrorInvalidBoardReference) {
			api.RespondError(w, http.StatusBadRequest, common.ErrorInvalidBoardReference.Error())
			return
		}
		if errors.Is(err, common.ErrorUserAlreadyMember) {
			api.RespondError(w, http.StatusConflict, common.ErrorUserAlreadyMember.Error())
			return
		}

		logger.Error().Err(ErrCannotCreateBoard).Msg("board service CreateBoard")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotCreateBoard.Error())
		return
	}

	api.HandleError(api.RespondCreated(w, dto.BoardInfoResponseFromInfo(board)))
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
func (h *BoardHandler) DeleteBoard(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	rawUserLink := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := rawUserLink.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, ErrUserUnauthorized.Error())
		return
	}

	vars := mux.Vars(r)

	rawBoardLink, ok := vars[boardLinkKey]
	if !ok {
		api.RespondError(w, http.StatusBadRequest, ErrBoardLinkMissing.Error())
		return
	}

	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidBoardLink.Error())
		return
	}

	err = h.srv.DeleteBoard(r.Context(), boardLink, userLink)
	if err != nil {
		if errors.Is(err, common.ErrActionDenied) {
			api.RespondError(w, http.StatusForbidden, common.ErrActionDenied.Error())
			return
		}

		if errors.Is(err, common.ErrBoardNotFound) {
			api.RespondError(w, http.StatusNotFound, common.ErrBoardNotFound.Error())
			return
		}

		logger.Error().Err(ErrCannotDeleteBoard).Msg("board service DeleteBoard")
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
func (h *BoardHandler) UpdateBoard(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	rawUserLink := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := rawUserLink.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, ErrUserUnauthorized.Error())
		return
	}

	vars := mux.Vars(r)

	rawBoardLink, ok := vars[boardLinkKey]
	if !ok {
		api.RespondError(w, http.StatusBadRequest, ErrBoardLinkMissing.Error())
		return
	}

	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidBoardLink.Error())
		return
	}

	var updateRequest dto.UpdateBoardRequest
	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		logger.Error().Err(ErrInvalidRequestSchema).Msg("decode update board request")
		api.RespondError(w, http.StatusBadRequest, ErrInvalidRequestSchema.Error())
		return
	}
	updateRequest.Sanitize()

	err = h.srv.UpdateBoard(r.Context(), dto.ToUpdateBoardInfo(updateRequest, boardLink), userLink)
	if err != nil {
		if errors.Is(err, common.ErrActionDenied) {
			api.RespondError(w, http.StatusForbidden, common.ErrActionDenied.Error())
			return
		}

		if errors.Is(err, common.ErrBoardNotFound) {
			api.RespondError(w, http.StatusNotFound, common.ErrBoardNotFound.Error())
			return
		}

		if errors.Is(err, common.ErrorNotNullValue) {
			api.RespondError(w, http.StatusBadRequest, common.ErrorNotNullValue.Error())
			return
		}

		if errors.Is(err, common.ErrorInvalidBoardData) {
			api.RespondError(w, http.StatusBadRequest, common.ErrorInvalidBoardData.Error())
			return
		}

		logger.Error().Err(ErrCannotUpdateBoard).Msg("board service UpdateBoard")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotUpdateBoard.Error())
		return
	}

	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}

var detectContentTypeBufferPool = sync.Pool{
	New: func() any {
		return make([]byte, 512)
	},
}

// @Summary		Загрузить фон для доски
// @Description	Загружает изображение (multipart/form-data) и устанавливает его как фон доски
// @Tags			boards
// @Accept			multipart/form-data
// @Produce		json
// @Param			link		path		string	true	"UUID доски"	Format(uuid)
// @Param			background	formData	file	true	"Файл изображения (например, PNG/JPEG)"
// @Success		200			{object}	api.OkResponse[dto.BackgroundUpdateResponse]
// @Failure		400			{object}	api.ErrorResponse	"invalid board link / invalid content type / cannot find background key"
// @Failure		401			{object}	api.ErrorResponse	"unauthorized"
// @Failure		404			{object}	api.ErrorResponse	"board not found"
// @Failure		500			{object}	api.ErrorResponse	"cannot update background / cannot read file"
// @Router			/boards/{link}/background [put]
func (h *BoardHandler) UploadBackground(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	rawUserLink := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := rawUserLink.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, ErrUserUnauthorized.Error())
		return
	}

	vars := mux.Vars(r)

	rawBoardLink, ok := vars[boardLinkKey]
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
		logger.Error().Err(err).Msg("cannot find 'background' key")
		api.RespondError(w, http.StatusBadRequest, ErrCannotFindBackground.Error())
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Error().Err(err).Msg("BoardsHandler.UploadBackground close user sended file")
		}
	}()

	buf := detectContentTypeBufferPool.Get().([]byte)
	defer func() {
		clear(buf)
		detectContentTypeBufferPool.Put(buf)
	}()

	_, err = file.Read(buf)
	if err != nil && err != io.EOF {
		api.RespondError(w, http.StatusInternalServerError, ErrCannotReadFile.Error())
		return
	}

	contentType := http.DetectContentType(buf)
	if !strings.HasPrefix(contentType, "image/") {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidContentType.Error())
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		api.RespondError(w, http.StatusInternalServerError, ErrCannotOperateWithFile.Error())
		return
	}

	extension := filepath.Ext(header.Filename)

	backgroundURL, err := h.srv.UpdateBackground(r.Context(), file, contentType, extension, boardLink, userLink)
	if err != nil {
		if errors.Is(err, common.ErrBoardNotFound) {
			api.RespondError(w, http.StatusNotFound, common.ErrBoardNotFound.Error())
			return
		}

		logger.Error().Err(err).Msg("board service UpdateBackground")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotUpdateBackground.Error())
		return
	}

	api.HandleError(api.RespondOk(w, dto.BackgroundUpdateResponse{
		BackgroundURL: backgroundURL,
	}))

}

// @Summary		Получить пользователей доски
// @Description	Возвращает массив UUID всех пользователей, имеющих доступ к доске
// @Tags			boards
// @Produce		json
// @Param			link	path		string						true	"UUID доски"	Format(uuid)
// @Success		200		{object}	api.OkResponse[[]string]	"Список UUID пользователей"
// @Failure		400		{object}	api.ErrorResponse			"invalid board link / board link missing"
// @Failure		404		{object}	api.ErrorResponse			"board not found"
// @Failure		500		{object}	api.ErrorResponse			"cannot get users of board"
// @Router			/boards/{link}/users [get]
func (h *BoardHandler) GetUsersOfBoard(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	vars := mux.Vars(r)

	rawBoardLink, ok := vars[boardLinkKey]
	if !ok {
		api.RespondError(w, http.StatusBadRequest, ErrBoardLinkMissing.Error())
		return
	}

	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidBoardLink.Error())
		return
	}

	usersLinks, err := h.srv.GetUsersOfBoard(r.Context(), boardLink)
	if err != nil {
		if errors.Is(err, common.ErrBoardNotFound) {
			api.RespondError(w, http.StatusNotFound, common.ErrBoardNotFound.Error())
			return
		}

		logger.Error().Err(err).Msg("BoardService.GetUsersOfBoard")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotGetUsersOfBoard.Error())
		return
	}

	api.HandleError(api.RespondOk(w, usersLinks))
}
