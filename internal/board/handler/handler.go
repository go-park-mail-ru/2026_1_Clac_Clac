package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/handler/dto"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/service/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
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
)

//go:generate mockery --name=BoardService --output mock_board_srv
type BoardService interface {
	GetBoards(ctx context.Context, userLink uuid.UUID) ([]serviceDto.BoardInfo, error)
	GetBoard(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID) (serviceDto.BoardInfo, error)
	CreateBoard(ctx context.Context, boardInfo serviceDto.NewBoardInfo, authorLink uuid.UUID) (serviceDto.BoardInfo, error)
	DeleteBoard(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID) error
	UpdateBoard(ctx context.Context, boardInfo serviceDto.UpdateBoardInfo, userLink uuid.UUID) error
	UpdateBackground(ctx context.Context, file io.Reader, contentType string, extension string, boardLink uuid.UUID, userLink uuid.UUID) (string, error)
}

func NewHandler(srv BoardService) *BoardHandler {
	return &BoardHandler{
		srv: srv,
	}
}

type BoardHandler struct {
	srv BoardService
}

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
		boardsResponse = append(boardsResponse, dto.BoardInfoResponseFromInfo(board))
	}

	api.HandleError(api.RespondOk(w, boardsResponse))
}

func (h *BoardHandler) GetBoard(w http.ResponseWriter, r *http.Request) {
	const boardLinkKey = "link"

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
		}

		if errors.Is(err, common.ErrBoardNotFound) {
			api.RespondError(w, http.StatusNotFound, common.ErrBoardNotFound.Error())
			return
		}

		logger.Error().Err(err).Msg("board service GetBoard")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotGetBoards.Error())
		return
	}

	api.HandleError(api.RespondOk(w, dto.BoardInfoResponseFromInfo(board)))
}

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
		logger.Error().Err(ErrCannotCreateBoard).Msg("board service CreateBoard")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotCreateBoard.Error())
		return
	}

	api.HandleError(api.RespondCreated(w, dto.BoardInfoResponseFromInfo(board)))
}

func (h *BoardHandler) DeleteBoard(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	rawUserLink := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := rawUserLink.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, ErrUserUnauthorized.Error())
		return
	}

	var deleteRequest dto.DeleteBoardRequest
	if err := json.NewDecoder(r.Body).Decode(&deleteRequest); err != nil {
		logger.Error().Err(ErrInvalidRequestSchema).Msg("decode delete board request")
		api.RespondError(w, http.StatusBadRequest, ErrInvalidRequestSchema.Error())
		return
	}

	err := h.srv.DeleteBoard(r.Context(), deleteRequest.Link, userLink)
	if err != nil {
		if errors.Is(err, common.ErrActionDenied) {
			api.RespondError(w, http.StatusForbidden, common.ErrActionDenied.Error())
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

func (h *BoardHandler) UpdateBoard(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	rawUserLink := r.Context().Value(middleware.UserContextLink{})

	userLink, ok := rawUserLink.(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, ErrUserUnauthorized.Error())
		return
	}

	var updateRequest dto.UpdateBoardRequest
	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		logger.Error().Err(ErrInvalidRequestSchema).Msg("decode update board request")
		api.RespondError(w, http.StatusBadRequest, ErrInvalidRequestSchema.Error())
		return
	}
	updateRequest.Sanitize()

	err := h.srv.UpdateBoard(r.Context(), dto.ToUpdateBoardInfo(updateRequest), userLink)
	if err != nil {
		if errors.Is(err, common.ErrActionDenied) {
			api.RespondError(w, http.StatusForbidden, common.ErrActionDenied.Error())
		}

		if errors.Is(err, common.ErrBoardNotFound) {
			api.RespondError(w, http.StatusNotFound, common.ErrBoardNotFound.Error())
			return
		}

		logger.Error().Err(ErrCannotUpdateBoard).Msg("board service UpdateBoard")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotUpdateBoard.Error())
		return
	}

	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}

var buffersPool = sync.Pool{
	New: func() any {
		const sizeForDetectContentType = 512
		return make([]byte, sizeForDetectContentType)
	},
}

func (h *BoardHandler) UploadBackground(w http.ResponseWriter, r *http.Request) {
	const boardLinkKey = "link"
	const backgroundFileKey = "background"
	const maxUploadSize = 2 << 26 // 8 МБ

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
	}

	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidBoardLink.Error())
		return
	}

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		logger.Error().Err(err).Msg("parse multipart form")
		api.RespondError(w, http.StatusBadRequest, ErrParseMultipartForm.Error())
		return
	}

	file, header, err := r.FormFile(backgroundFileKey)
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

	buf := buffersPool.Get().([]byte)
	defer func() {
		clear(buf)
		buffersPool.Put(buf)
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
