package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
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
	"github.com/mailru/easyjson"
	"github.com/rs/zerolog"
)

var (
	ErrCannotGetBoards        = errors.New("cannot get boards")
	ErrCannotCreateBoard      = errors.New("cannot create board")
	ErrCannotDeleteBoard      = errors.New("cannot delete board")
	ErrCannotUpdateBoard      = errors.New("cannot update board")
	ErrBoardLinkMissing       = errors.New("board link missing")
	ErrInvalidBoardLink       = errors.New("invalid board link")
	ErrCannotFindBackground   = errors.New("cannot find 'background' key")
	ErrCannotUpdateBackground = errors.New("cannot update background")
	ErrCannotGetMembers       = errors.New("cannot get members")
	ErrCannotGetProfiles      = errors.New("cannot get profiles")

	ErrCannotCreateInvite = errors.New("cannot create invite")
	ErrCannotAcceptInvite = errors.New("cannot accept invite")
	ErrCannotCloseInvite  = errors.New("cannot close invite")
	ErrRoleRequired       = errors.New("role is required")
	ErrInvalidRole        = errors.New("invalid role")
	ErrInviteLinkMissing  = errors.New("invite link missing")
	ErrInvalidInviteLink  = errors.New("invalid invite link")
)

const (
	boardLinkKey  = "link"
	inviteLinkKey = "invite_link"
)

//go:generate mockery --name=BoardUsecase --output=mock_board_use_case
type BoardUsecase interface {
	GetBoards(ctx context.Context, userLink uuid.UUID) ([]domain.BoardInfo, error)
	GetBoard(ctx context.Context, boardInfo domain.GetBoardRequest) (domain.BoardInfo, error)
	CreateBoard(ctx context.Context, boardInfo domain.CreateBoardRequest) (domain.BoardInfo, error)
	DeleteBoard(ctx context.Context, boardInfo domain.GetBoardRequest) error
	UpdateBoard(ctx context.Context, boardInfo domain.UpdateBoardRequest) error
	UploadBackground(ctx context.Context, backgroundInfo domain.UploadBackgroundRequest, image io.Reader) (domain.UploadBackgroundResponse, error)
	GetMembers(ctx context.Context, membersInfo domain.GetMembersRequest) (domain.GetMembersResponse, error)

	CreateInvite(ctx context.Context, inviteInfo domain.CreateInviteRequest) (domain.CreateInviteResponse, error)
	AcceptInvite(ctx context.Context, inviteInfo domain.AcceptInviteRequest) (string, string, error)
	CloseInvite(ctx context.Context, inviteInfo domain.CloseInviteRequest) error

	UpdateMemberRole(ctx context.Context, req domain.UpdateMemberRoleRequest) error
	RemoveMemberFromBoard(ctx context.Context, req domain.RemoveMemberRequest) error
	GetActiveInvites(ctx context.Context, userLink, boardLink uuid.UUID) ([]domain.InviteInfo, error)
}

type ProfileUsecase interface {
	GetProfiles(ctx context.Context, links []uuid.UUID) ([]domain.FullInfoUser, error)
}

type BoardConfig struct {
	MultipartBackgroundFileKey string
	MaxBackgroundSize          int64
	MaxDisplayName             int
}

type Board struct {
	boardSrv   BoardUsecase
	profileSrv ProfileUsecase
	conf       BoardConfig
}

func NewBoard(boardSrv BoardUsecase, profileSrv ProfileUsecase, conf BoardConfig) *Board {
	return &Board{
		boardSrv:   boardSrv,
		profileSrv: profileSrv,
		conf:       conf,
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
// @Tags			Boards
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

	boards, err := h.boardSrv.GetBoards(r.Context(), userLink)
	if err != nil {
		errLog := fmt.Errorf("srv.GetBoards: %w", err)
		logger.Error().Err(errLog).Msg("board usecase GetBoards")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "GetBoards", map[string]interface{}{
			"user_link": userLink,
			"action":    "get_boards",
		})
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
// @Tags			Boards
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

	board, err := h.boardSrv.GetBoard(r.Context(), domain.GetBoardRequest{
		UserLink:  userLink,
		BoardLink: boardLink,
	})
	if err != nil {
		if errors.Is(err, common.ErrorBoardNotFound) {
			api.RespondError(w, http.StatusNotFound, common.ErrorBoardNotFound.Error())
			return
		}
		if errors.Is(err, common.ErrorBoardPermissionDenied) {
			api.RespondError(w, http.StatusForbidden, common.ErrorBoardPermissionDenied.Error())
			return
		}
		errLog := fmt.Errorf("srv.GetBoard: %w", err)
		logger.Error().Err(errLog).Msg("board usecase GetBoard")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "GetBoard", map[string]interface{}{
			"user_link":  userLink,
			"board_link": boardLink,
			"action":     "get_board",
		})
		api.RespondError(w, http.StatusInternalServerError, ErrCannotGetBoards.Error())
		return
	}

	api.HandleError(api.RespondOk(w, boardInfoToDTO(board)))
}

// @Summary		Создать новую доску
// @Description	Создает новую доску на основе переданных данных
// @Tags			Boards
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
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	if req.Name == "" {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	if len([]rune(req.Name)) > h.conf.MaxDisplayName {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	board, err := h.boardSrv.CreateBoard(r.Context(), domain.CreateBoardRequest{
		UserLink:    userLink,
		Name:        req.Name,
		Description: req.Description,
		Background:  req.Background,
	})

	if err != nil {
		errLog := fmt.Errorf("srv.CreateBoard: %w", err)
		logger.Error().Err(errLog).Msg("board usecase CreateBoard")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "CreateBoard", map[string]interface{}{
			"user_link": userLink,
			"action":    "create_board",
		})
		api.RespondError(w, http.StatusInternalServerError, ErrCannotCreateBoard.Error())
		return
	}

	api.HandleError(api.RespondCreated(w, boardInfoToDTO(board)))
}

// @Summary		Удалить доску
// @Description	Удаляет доску по её UUID ссылке
// @Tags			Boards
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

	err = h.boardSrv.DeleteBoard(r.Context(), domain.GetBoardRequest{
		UserLink:  userLink,
		BoardLink: boardLink,
	})
	if err != nil {
		if errors.Is(err, common.ErrorBoardNotFound) {
			api.RespondError(w, http.StatusNotFound, common.ErrorBoardNotFound.Error())
			return
		}
		if errors.Is(err, common.ErrorBoardPermissionDenied) {
			api.RespondError(w, http.StatusForbidden, common.ErrorBoardPermissionDenied.Error())
			return
		}
		errLog := fmt.Errorf("srv.DeleteBoard: %w", err)
		logger.Error().Err(errLog).Msg("board usecase DeleteBoard")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "DeleteBoard", map[string]interface{}{
			"user_link":  userLink,
			"board_link": boardLink,
			"action":     "delete_board",
		})
		api.RespondError(w, http.StatusInternalServerError, ErrCannotDeleteBoard.Error())
		return
	}

	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}

// @Summary		Обновить информацию о доске
// @Description	Обновляет метаданные доски (имя, описание, фон)
// @Tags			Boards
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
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	if len([]rune(req.Name)) > h.conf.MaxDisplayName {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidSizeDisplayName.Error())
		return
	}

	err = h.boardSrv.UpdateBoard(r.Context(), domain.UpdateBoardRequest{
		UserLink:    userLink,
		BoardLink:   boardLink,
		Name:        req.Name,
		Description: req.Description,
		Background:  req.Background,
	})
	if err != nil {
		if errors.Is(err, common.ErrorBoardNotFound) {
			api.RespondError(w, http.StatusNotFound, common.ErrorBoardNotFound.Error())
			return
		}
		if errors.Is(err, common.ErrorBoardPermissionDenied) {
			api.RespondError(w, http.StatusForbidden, common.ErrorBoardPermissionDenied.Error())
			return
		}
		errLog := fmt.Errorf("srv.UpdateBoard: %w", err)
		logger.Error().Err(errLog).Msg("board usecase UpdateBoard")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "UpdateBoard", map[string]interface{}{
			"user_link":  userLink,
			"board_link": boardLink,
			"action":     "update_board",
		})
		api.RespondError(w, http.StatusInternalServerError, ErrCannotUpdateBoard.Error())
		return
	}

	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}

// @Summary		Загрузить фон для доски
// @Description	Загружает изображение (multipart/form-data) и устанавливает его как фон доски
// @Tags			Boards
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
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			api.RespondError(w, http.StatusRequestEntityTooLarge, ErrParseMultipartForm.Error())
		} else {
			api.RespondError(w, http.StatusBadRequest, ErrParseMultipartForm.Error())
		}
		return
	}

	file, header, err := r.FormFile(h.conf.MultipartBackgroundFileKey)
	if err != nil {
		logger.Error().Err(err).Str("expected key", h.conf.MultipartBackgroundFileKey).Msg("cannot find background key")
		api.RespondError(w, http.StatusBadRequest, ErrCannotFindBackground.Error())
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Error().Err(err).Msg("BoardHandler.UploadBackground close file")
		}
	}()

	resp, err := h.boardSrv.UploadBackground(r.Context(), domain.UploadBackgroundRequest{
		UserLink:  userLink,
		BoardLink: boardLink,
		Filename:  header.Filename,
	}, file)
	if err != nil {
		switch {
		case errors.Is(err, common.ErrorNonexistentUser):
			api.RespondError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, common.ErrorBoardNotFound):
			api.RespondError(w, http.StatusNotFound, common.ErrorBoardNotFound.Error())
		case errors.Is(err, common.ErrorBoardPermissionDenied):
			api.RespondError(w, http.StatusForbidden, common.ErrorBoardPermissionDenied.Error())
		case errors.Is(err, common.ErrorInvalidInput):
			api.RespondError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, common.ErrorInvalidContentType):
			api.RespondError(w, http.StatusUnsupportedMediaType, err.Error())
		default:
			errLog := fmt.Errorf("srv.UploadBackground: %w", err)
			logger.Error().Err(errLog).Msg("board usecase UploadBackground")
			sentryLogger.CaptureFromContext(r.Context(), errLog, "UploadBackground", map[string]interface{}{
				"user_link":  userLink,
				"board_link": boardLink,
				"action":     "upload_background",
			})
			api.RespondError(w, http.StatusInternalServerError, ErrCannotUpdateBackground.Error())
		}
		return
	}

	api.HandleError(api.RespondOk(w, dto.UploadBackgroundResponse{
		BackgroundKey: resp.BackgroundKey,
	}))
}

// @Summary		Получить участников доски
// @Description	Возвращает список участников доски с ролями, аватарами, именами, описаниями и email
// @Tags		Boards
// @Produce		json
// @Param		link	path	string	true	"UUID доски"	Format(uuid)
// @Success		200	{object}	api.OkResponse[dto.GetMembersResponse]
// @Failure		400	{object}	api.ErrorResponse	"invalid board link / board link missing / invalid input"
// @Failure		401	{object}	api.ErrorResponse	"unauthorized"
// @Failure		403	{object}	api.ErrorResponse	"action denied"
// @Failure		404	{object}	api.ErrorResponse	"board not found / user not found"
// @Failure		500	{object}	api.ErrorResponse	"cannot get members / cannot get profiles"
// @Router		/boards/{link}/users [get]
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

	members, err := h.boardSrv.GetMembers(r.Context(), domain.GetMembersRequest{
		UserLink:  userLink,
		BoardLink: boardLink,
	})
	if err != nil {
		if errors.Is(err, common.ErrorBoardNotFound) {
			api.RespondError(w, http.StatusNotFound, common.ErrorBoardNotFound.Error())
			return
		}
		if errors.Is(err, common.ErrorBoardPermissionDenied) {
			api.RespondError(w, http.StatusForbidden, common.ErrorBoardPermissionDenied.Error())
			return
		}
		errLog := fmt.Errorf("srv.GetMembers: %w", err)
		logger.Error().Err(errLog).Msg("board usecase GetMembers")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "GetMembers", map[string]interface{}{
			"user_link":  userLink,
			"board_link": boardLink,
			"action":     "get_members",
		})
		api.RespondError(w, http.StatusInternalServerError, ErrCannotGetMembers.Error())
		return
	}

	links := make([]uuid.UUID, 0, len(members.Members))
	for _, member := range members.Members {
		links = append(links, member.Link)
	}

	profiles, err := h.profileSrv.GetProfiles(r.Context(), links)
	if err != nil {
		if errors.Is(err, common.ErrorInvalidInput) {
			api.RespondError(w, http.StatusBadRequest, common.ErrorInvalidInput.Error())
			return
		}
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, common.ErrorNonexistentUser.Error())
			return
		}
		errLog := fmt.Errorf("srv.GetProfiles: %w", err)
		logger.Error().Err(errLog).Msg("profile usecase GetProfiles")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "GetMembers", map[string]interface{}{
			"user_link":  userLink,
			"board_link": boardLink,
			"action":     "get_profiles",
		})
		api.RespondError(w, http.StatusInternalServerError, ErrCannotGetProfiles.Error())
		return
	}

	result := make([]dto.MemberInfo, 0, len(members.Members))
	for index, p := range profiles {
		result = append(result, dto.MemberInfo{
			Link:        members.Members[index].Link,
			Role:        members.Members[index].Role,
			AvatarUrl:   p.AvatarURL,
			Description: p.Description,
			DisplayName: p.DisplayName,
			Email:       p.Email,
		})
	}

	api.HandleError(api.RespondOk(w, dto.GetMembersResponse{
		Members: result,
	}))
}

// @Summary		Создать приглашение на доску
// @Description	Создает ссылку-приглашение для добавления пользователя на доску
// @Tags		Boards
// @Accept		json
// @Produce		json
// @Param		link	path	string	true	"UUID доски"	Format(uuid)
// @Param		request	body	dto.CreateInviteRequest	true	"Данные для создания приглашения"
// @Success		201	{object}	api.OkResponse[dto.CreateInviteResponse]
// @Failure	400	{object}	api.ErrorResponse	"invalid request schema"
// @Failure	401	{object}	api.ErrorResponse	"unauthorized"
// @Failure	403	{object}	api.ErrorResponse	"action denied"
// @Failure	404	{object}	api.ErrorResponse	"board not found"
// @Failure	500	{object}	api.ErrorResponse	"cannot create invite"
// @Router		/boards/{link}/invites [post]
func (h *Board) CreateInvite(w http.ResponseWriter, r *http.Request) {
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

	var req dto.CreateInviteRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	if req.DefaultRole == "" {
		api.RespondError(w, http.StatusBadRequest, ErrRoleRequired.Error())
		return
	}

	var targetUserLink *uuid.UUID
	if req.UserLink != "" {
		parsed, err := uuid.Parse(req.UserLink)
		if err != nil {
			api.RespondError(w, http.StatusBadRequest, ErrInvalidBoardLink.Error())
			return
		}
		targetUserLink = &parsed
	}

	invite, err := h.boardSrv.CreateInvite(r.Context(), domain.CreateInviteRequest{
		UserLink:       userLink,
		BoardLink:      boardLink,
		TargetUserLink: targetUserLink,
		DefaultRole:    req.DefaultRole,
		ExpireSeconds:  req.ExpireSeconds,
	})
	if err != nil {
		if errors.Is(err, common.ErrorBoardNotFound) {
			api.RespondError(w, http.StatusNotFound, common.ErrorBoardNotFound.Error())
			return
		}
		if errors.Is(err, common.ErrorBoardPermissionDenied) {
			api.RespondError(w, http.StatusForbidden, common.ErrorBoardPermissionDenied.Error())
			return
		}

		errLog := fmt.Errorf("srv.CreateInvite: %w", err)
		logger.Error().Err(errLog).Msg("board usecase CreateInvite")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "CreateInvite", map[string]interface{}{
			"user_link":  userLink,
			"board_link": boardLink,
			"action":     "create_invite",
		})
		api.RespondError(w, http.StatusInternalServerError, ErrCannotCreateInvite.Error())
		return
	}

	api.HandleError(api.RespondCreated(w, dto.CreateInviteResponse{
		InviteLink:     invite.InviteLink,
		BoardLink:      invite.BoardLink,
		TargetUserLink: invite.TargetUserLink,
		DefaultRole:    invite.DefaultRole,
		Status:         invite.Status,
		ExpireAt:       invite.ExpireAt,
		CreatedAt:      invite.CreatedAt,
	}))
}

// @Summary		Принять приглашение на доску
// @Description	Принимает приглашение по ссылке и добавляет пользователя в участники доски
// @Tags		Boards
// @Produce		json
// @Param		invite_link	path	string	true	"UUID ссылки-приглашения"	Format(uuid)
// @Success		200	{object}	api.OkResponse[dto.AcceptInviteResponse]
// @Failure	400	{object}	api.ErrorResponse	"invalid invite link"
// @Failure	401	{object}	api.ErrorResponse	"unauthorized"
// @Failure	403	{object}	api.ErrorResponse	"invite not for user"
// @Failure	404	{object}	api.ErrorResponse	"invite not found"
// @Failure	409	{object}	api.ErrorResponse	"user is already a member"
// @Failure	412	{object}	api.ErrorResponse	"invite is expired"
// @Failure	412	{object}	api.ErrorResponse	"invite is closed"
// @Failure	500	{object}	api.ErrorResponse	"cannot accept invite"
// @Router		/invites/{invite_link} [post]
func (h *Board) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	userLink, ok := r.Context().Value(middleware.UserContextLink{}).(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}

	rawInviteLink, ok := mux.Vars(r)[inviteLinkKey]
	if !ok {
		api.RespondError(w, http.StatusBadRequest, ErrInviteLinkMissing.Error())
		return
	}

	boardLink, role, err := h.boardSrv.AcceptInvite(r.Context(), domain.AcceptInviteRequest{
		InviteLink: rawInviteLink,
		UserLink:   userLink,
	})
	if err != nil {
		switch {
		case errors.Is(err, common.ErrorInviteNotFound):
			api.RespondError(w, http.StatusNotFound, common.ErrorInviteNotFound.Error())
			return
		case errors.Is(err, common.ErrorInviteClosed):
			api.RespondError(w, http.StatusPreconditionFailed, common.ErrorInviteClosed.Error())
			return
		case errors.Is(err, common.ErrorInviteExpired):
			api.RespondError(w, http.StatusPreconditionFailed, common.ErrorInviteExpired.Error())
			return
		case errors.Is(err, common.ErrorInviteNotForUser):
			api.RespondError(w, http.StatusForbidden, common.ErrorInviteNotForUser.Error())
			return
		case errors.Is(err, common.ErrorUserAlreadyMember):
			api.RespondError(w, http.StatusConflict, common.ErrorUserAlreadyMember.Error())
			return
		}

		errLog := fmt.Errorf("srv.AcceptInvite: %w", err)
		logger.Error().Err(errLog).Msg("board usecase AcceptInvite")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "AcceptInvite", map[string]interface{}{
			"user_link":   userLink,
			"invite_link": rawInviteLink,
			"action":      "accept_invite",
		})
		api.RespondError(w, http.StatusInternalServerError, ErrCannotAcceptInvite.Error())
		return
	}

	api.HandleError(api.RespondOk(w, dto.AcceptInviteResponse{
		BoardLink: boardLink,
		Role:      role,
	}))
}

// @Summary		Закрыть приглашение
// @Description	Закрывает активное приглашение (только для Admin/Creator)
// @Tags		Boards
// @Produce		json
// @Param		invite_link	path	string	true	"UUID ссылки-приглашения"	Format(uuid)
// @Success		200	{object}	api.Response	"status ok"
// @Failure	400	{object}	api.ErrorResponse	"invalid invite link"
// @Failure	401	{object}	api.ErrorResponse	"unauthorized"
// @Failure	403	{object}	api.ErrorResponse	"action denied"
// @Failure	404	{object}	api.ErrorResponse	"invite not found"
// @Failure	500	{object}	api.ErrorResponse	"cannot close invite"
// @Router		/invites/{invite_link} [delete]
func (h *Board) CloseInvite(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	userLink, ok := r.Context().Value(middleware.UserContextLink{}).(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusUnauthorized, handlerCommon.ErrUserNotAuthorized.Error())
		return
	}

	rawInviteLink, ok := mux.Vars(r)[inviteLinkKey]
	if !ok {
		api.RespondError(w, http.StatusBadRequest, ErrInviteLinkMissing.Error())
		return
	}

	err := h.boardSrv.CloseInvite(r.Context(), domain.CloseInviteRequest{
		UserLink:   userLink,
		InviteLink: rawInviteLink,
	})
	if err != nil {
		switch {
		case errors.Is(err, common.ErrorInviteNotFound):
			api.RespondError(w, http.StatusNotFound, common.ErrorInviteNotFound.Error())
			return
		case errors.Is(err, common.ErrorBoardPermissionDenied):
			api.RespondError(w, http.StatusForbidden, common.ErrorBoardPermissionDenied.Error())
			return
		}

		errLog := fmt.Errorf("srv.CloseInvite: %w", err)
		logger.Error().Err(errLog).Msg("board usecase CloseInvite")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "CloseInvite", map[string]interface{}{
			"user_link":   userLink,
			"invite_link": rawInviteLink,
			"action":      "close_invite",
		})
		api.RespondError(w, http.StatusInternalServerError, ErrCannotCloseInvite.Error())
		return
	}

	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}

// @Summary		Получить активные приглашения доски
// @Description	Возвращает все активные приглашения для доски
// @Tags		Boards
// @Produce		json
// @Param		link	path	string	true	"UUID доски"	Format(uuid)
// @Success		200	{object}	api.OkResponse[[]dto.InviteInfo]
// @Failure	401	{object}	api.ErrorResponse	"unauthorized"
// @Failure	403	{object}	api.ErrorResponse	"action denied"
// @Failure	500	{object}	api.ErrorResponse	"cannot get boards"
// @Router		/boards/{link}/invites [get]
func (h *Board) GetActiveInvites(w http.ResponseWriter, r *http.Request) {
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

	invites, err := h.boardSrv.GetActiveInvites(r.Context(), userLink, boardLink)
	if err != nil {
		if errors.Is(err, common.ErrorBoardPermissionDenied) {
			api.RespondError(w, http.StatusForbidden, common.ErrorBoardPermissionDenied.Error())
			return
		}
		errLog := fmt.Errorf("srv.GetActiveInvites: %w", err)
		logger.Error().Err(errLog).Msg("board usecase GetActiveInvites")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotGetBoards.Error())
		return
	}

	result := make([]dto.InviteInfo, 0, len(invites))
	for _, inv := range invites {
		info := dto.InviteInfo{
			InviteLink:  inv.InviteLink,
			BoardLink:   inv.BoardLink,
			DefaultRole: inv.DefaultRole,
			Status:      inv.Status,
			CreatedAt:   inv.CreatedAt,
		}
		if inv.TargetUserLink != nil {
			info.TargetUserLink = inv.TargetUserLink
		}
		if inv.ExpireAt != nil {
			info.ExpireAt = inv.ExpireAt
		}
		result = append(result, info)
	}

	api.HandleError(api.RespondOk(w, result))
}

// @Summary		Изменить роль пользователя на доске
// @Description	Изменяет роль пользователя на доске. Доступно Admin и Creator. Нельзя изменить свою собственную роль.
// @Tags		Boards
// @Accept		json
// @Produce		json
// @Param		link		path	string					true	"UUID доски"					Format(uuid)
// @Param		user_link	path	string					true	"UUID пользователя"				Format(uuid)
// @Param		request		body	dto.UpdateMemberRoleRequest	true	"Новая роль"
// @Success		200			{object}	api.Response			"status ok"
// @Failure	400	{object}	api.ErrorResponse	"cannot change your own role"
// @Failure	401	{object}	api.ErrorResponse	"unauthorized"
// @Failure	403	{object}	api.ErrorResponse	"action denied"
// @Failure	404	{object}	api.ErrorResponse	"board not found / user not found"
// @Failure	500	{object}	api.ErrorResponse	"cannot update board"
// @Router		/boards/{link}/members/{user_link}/role [put]
func (h *Board) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
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

	rawTargetLink, ok := mux.Vars(r)["user_link"]
	if !ok {
		api.RespondError(w, http.StatusBadRequest, ErrBoardLinkMissing.Error())
		return
	}

	targetLink, err := uuid.Parse(rawTargetLink)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidBoardLink.Error())
		return
	}

	var req dto.UpdateMemberRoleRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	err = h.boardSrv.UpdateMemberRole(r.Context(), domain.UpdateMemberRoleRequest{
		UserLink:       userLink,
		BoardLink:      boardLink,
		TargetUserLink: targetLink,
		NewRole:        req.NewRole,
	})
	if err != nil {
		if errors.Is(err, common.ErrorBoardNotFound) {
			api.RespondError(w, http.StatusNotFound, common.ErrorBoardNotFound.Error())
			return
		}
		if errors.Is(err, common.ErrorBoardPermissionDenied) {
			api.RespondError(w, http.StatusForbidden, common.ErrorBoardPermissionDenied.Error())
			return
		}
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, common.ErrorNonexistentUser.Error())
			return
		}
		if errors.Is(err, common.ErrorSelfRoleChange) {
			api.RespondError(w, http.StatusBadRequest, common.ErrorSelfRoleChange.Error())
			return
		}
		errLog := fmt.Errorf("srv.UpdateMemberRole: %w", err)
		logger.Error().Err(errLog).Msg("board usecase UpdateMemberRole")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotUpdateBoard.Error())
		return
	}

	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}

// @Summary		Удалить пользователя с доски
// @Description	Удаляет пользователя из участников доски. Пользователь может выйти из доски самостоятельно, но создатель не может покинуть доску. Admin и Creator могут удалять других участников.
// @Tags		Boards
// @Produce		json
// @Param		link		path	string		true	"UUID доски"				Format(uuid)
// @Param		user_link	path	string		true	"UUID пользователя"			Format(uuid)
// @Success		200			{object}	api.Response	"status ok"
// @Failure	400	{object}	api.ErrorResponse	"invalid input"
// @Failure	401	{object}	api.ErrorResponse	"unauthorized"
// @Failure	403	{object}	api.ErrorResponse	"action denied / creator cannot leave the board"
// @Failure	404	{object}	api.ErrorResponse	"board not found / user not found"
// @Failure	500	{object}	api.ErrorResponse	"cannot update board"
// @Router		/boards/{link}/members/{user_link} [delete]
func (h *Board) RemoveMemberFromBoard(w http.ResponseWriter, r *http.Request) {
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

	rawTargetLink, ok := mux.Vars(r)["user_link"]
	if !ok {
		api.RespondError(w, http.StatusBadRequest, ErrBoardLinkMissing.Error())
		return
	}

	targetLink, err := uuid.Parse(rawTargetLink)
	if err != nil {
		api.RespondError(w, http.StatusBadRequest, ErrInvalidBoardLink.Error())
		return
	}

	err = h.boardSrv.RemoveMemberFromBoard(r.Context(), domain.RemoveMemberRequest{
		UserLink:       userLink,
		BoardLink:      boardLink,
		TargetUserLink: targetLink,
	})
	if err != nil {
		if errors.Is(err, common.ErrorBoardNotFound) {
			api.RespondError(w, http.StatusNotFound, common.ErrorBoardNotFound.Error())
			return
		}
		if errors.Is(err, common.ErrorBoardPermissionDenied) {
			api.RespondError(w, http.StatusForbidden, common.ErrorBoardPermissionDenied.Error())
			return
		}
		if errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, common.ErrorNonexistentUser.Error())
			return
		}
		if errors.Is(err, common.ErrorCreatorCannotLeave) {
			api.RespondError(w, http.StatusForbidden, common.ErrorCreatorCannotLeave.Error())
			return
		}
		errLog := fmt.Errorf("srv.RemoveMemberFromBoard: %w", err)
		logger.Error().Err(errLog).Msg("board usecase RemoveMemberFromBoard")
		api.RespondError(w, http.StatusInternalServerError, ErrCannotUpdateBoard.Error())
		return
	}

	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}
