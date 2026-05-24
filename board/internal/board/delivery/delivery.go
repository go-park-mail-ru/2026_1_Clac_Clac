package delivery

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/delivery/dto"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/service/dto"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/boardRbac"
	sentryLogger "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/logger"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/board/v1"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

var (
	ErrUserLinkRequired       = errors.New("user link is required")
	ErrBoardLinkRequired      = errors.New("board link is required")
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

	ErrInviteLinkRequired      = errors.New("invite link is required")
	ErrInvalidInviteLink       = errors.New("invalid invite link")
	ErrCannotCreateInvite      = errors.New("cannot create invite")
	ErrCannotAcceptInvite      = errors.New("cannot accept invite")
	ErrCannotCloseInvite       = errors.New("cannot close invite")
	ErrRoleRequired            = errors.New("role is required")
	ErrInvalidRole             = errors.New("invalid role")
)

//go:generate mockery --name=BoardService --output mock_board_srv
type BoardService interface {
	GetBoards(ctx context.Context, userLink uuid.UUID) ([]serviceDto.BoardInfo, error)
	GetBoard(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID) (serviceDto.BoardInfo, error)
	CreateBoard(ctx context.Context, boardInfo serviceDto.NewBoardInfo, authorLink uuid.UUID) (serviceDto.BoardInfo, error)
	DeleteBoard(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID) error
	UpdateBoard(ctx context.Context, boardInfo serviceDto.UpdateBoardInfo, userLink uuid.UUID) error
	UpdateBackground(ctx context.Context, file io.Reader, contentType string, extension string, boardLink uuid.UUID, userLink uuid.UUID) (string, error)
	GetUsersOfBoard(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID) ([]serviceDto.MemberInfo, error)

	CreateInvite(ctx context.Context, inviteInfo serviceDto.NewInviteInfo, creatorLink uuid.UUID) (serviceDto.InviteInfo, error)
	AcceptInvite(ctx context.Context, inviteLink uuid.UUID, userLink uuid.UUID) (serviceDto.InviteInfo, error)
	CloseInvite(ctx context.Context, inviteLink uuid.UUID, userLink uuid.UUID) error

	UpdateMemberRole(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID, newRole rbac.Role, callerLink uuid.UUID) error
	RemoveMemberFromBoard(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID, callerLink uuid.UUID) error
	GetActiveInvites(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID) ([]serviceDto.InviteInfo, error)
	CanView(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID) error

	CreatePoll(ctx context.Context, boardLink, adminLink uuid.UUID, cards []uuid.UUID, invitees []uuid.UUID) error
	DeletePoll(ctx context.Context, boardLink, userLink uuid.UUID) error
	NextPollCard(ctx context.Context, boardLink, userLink uuid.UUID) error
	VotePoll(ctx context.Context, boardLink, userLink uuid.UUID, points int) error
}

type Config struct {
	BaseBackgroundURL          string
	MultipartBackgroundFileKey string
	MaxBackgroundSize          int64
}

func NewHandler(srv BoardService, conf Config) *BoardHandler {
	return &BoardHandler{
		srv:  srv,
		conf: conf,
	}
}

type BoardHandler struct {
	pb.UnimplementedBoardServiceServer

	conf Config
	srv  BoardService
}

func (h *BoardHandler) GetBoards(ctx context.Context, req *pb.GetBoardsRequest) (*pb.GetBoardsResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrUserLinkRequired.Error())
	}

	boards, err := h.srv.GetBoards(ctx, userLink)
	if err != nil {
		errLog := fmt.Errorf("srv.GetBoards: %w", err)
		logger.Error().Err(errLog).Msg("board service GetBoards")
		sentryLogger.CaptureFromContext(ctx, errLog, "GetBoards", map[string]interface{}{
			"user_link": rawUserLink,
			"action":    "get_boards",
		})
		return nil, status.Error(codes.Internal, ErrCannotGetBoards.Error())
	}

	boardsResponse := make([]*pb.BoardInfo, 0)
	for _, board := range boards {
		info := &pb.BoardInfo{
			Link:        board.Link.String(),
			Name:        board.Name,
			Description: board.Description,
			Background:  board.Background,
			CreatedAt:   timestamppb.New(board.CreatedAt),
		}

		if strings.HasPrefix(info.Background, "backgrounds") {
			info.Background = fmt.Sprintf("%s/%s", h.conf.BaseBackgroundURL, info.Background)
		}

		boardsResponse = append(boardsResponse, info)
	}

	return &pb.GetBoardsResponse{
		BoardsInfo: boardsResponse,
	}, nil
}

func (h *BoardHandler) GetBoard(ctx context.Context, req *pb.GetBoardRequest) (*pb.GetBoardResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrUserLinkRequired.Error())
	}

	rawBoardLink := req.GetBoardLink()
	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrBoardLinkRequired.Error())
	}

	board, err := h.srv.GetBoard(ctx, boardLink, userLink)
	if err != nil {
		switch {
		case errors.Is(err, rbac.ErrActionDenied):
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		case errors.Is(err, common.ErrBoardNotFound):
			return nil, status.Error(codes.NotFound, common.ErrBoardNotFound.Error())
		}

		errLog := fmt.Errorf("srv.GetBoard: %w", err)
		logger.Error().Err(errLog).Msg("board service GetBoard")
		sentryLogger.CaptureFromContext(ctx, errLog, "GetBoard", map[string]interface{}{
			"user_link":  rawUserLink,
			"board_link": rawBoardLink,
			"action":     "get_board",
		})
		return nil, status.Error(codes.Internal, ErrCannotGetBoards.Error())
	}

	info := &pb.BoardInfo{
		Link:        board.Link.String(),
		Name:        board.Name,
		Description: board.Description,
		CreatedAt:   timestamppb.New(board.CreatedAt),
	}

	if strings.HasPrefix(info.Background, "backgrounds") {
		info.Background = fmt.Sprintf("%s/%s", h.conf.BaseBackgroundURL, info.Background)
	}

	return &pb.GetBoardResponse{
		BoardInfo: info,
	}, nil
}

func (h *BoardHandler) CreateBoard(ctx context.Context, req *pb.CreateBoardRequest) (*pb.CreateBoardResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrUserLinkRequired.Error())
	}

	var createBoardRequest = dto.CreateBoardRequest{
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Background:  req.GetBackground(),
	}
	createBoardRequest.Sanitize()

	board, err := h.srv.CreateBoard(ctx, dto.ToNewBoardInfo(createBoardRequest), userLink)
	if err != nil {
		switch {
		case errors.Is(err, common.ErrNotNullValue):
			return nil, status.Error(codes.InvalidArgument, common.ErrNotNullValue.Error())
		case errors.Is(err, common.ErrInvalidBoardData):
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidBoardData.Error())
		case errors.Is(err, common.ErrInvalidBoardReference):
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidBoardReference.Error())
		case errors.Is(err, common.ErrUserAlreadyMember):
			return nil, status.Error(codes.InvalidArgument, common.ErrUserAlreadyMember.Error())
		}

		errLog := fmt.Errorf("srv.CreateBoard: %w", err)
		logger.Error().Err(errLog).Msg("board service CreateBoard")
		sentryLogger.CaptureFromContext(ctx, errLog, "CreateBoard", map[string]interface{}{
			"user_link": rawUserLink,
			"action":    "create_board",
		})
		return nil, status.Error(codes.Internal, ErrCannotCreateBoard.Error())
	}

	return &pb.CreateBoardResponse{
		BoardInfo: &pb.BoardInfo{
			Link:        board.Link.String(),
			Name:        board.Name,
			Description: board.Description,
			Background:  board.Background,
			CreatedAt:   timestamppb.New(board.CreatedAt),
		},
	}, nil
}

func (h *BoardHandler) DeleteBoard(ctx context.Context, req *pb.DeleteBoardRequest) (*pb.DeleteBoardResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrUserLinkRequired.Error())
	}

	rawBoardLink := req.GetBoardLink()
	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrBoardLinkRequired.Error())
	}

	err = h.srv.DeleteBoard(ctx, boardLink, userLink)
	if err != nil {
		switch {
		case errors.Is(err, rbac.ErrActionDenied):
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		case errors.Is(err, common.ErrBoardNotFound):
			return nil, status.Error(codes.NotFound, common.ErrBoardNotFound.Error())
		}

		errLog := fmt.Errorf("srv.DeleteBoard: %w", err)
		logger.Error().Err(errLog).Msg("board service DeleteBoard")
		sentryLogger.CaptureFromContext(ctx, errLog, "DeleteBoard", map[string]interface{}{
			"user_link":  rawUserLink,
			"board_link": rawBoardLink,
			"action":     "delete_board",
		})
		return nil, status.Error(codes.Internal, ErrCannotDeleteBoard.Error())
	}

	return &pb.DeleteBoardResponse{}, nil
}

func (h *BoardHandler) UpdateBoard(ctx context.Context, req *pb.UpdateBoardRequest) (*pb.UpdateBoardResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrUserLinkRequired.Error())
	}

	rawBoardLink := req.GetBoardLink()
	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrBoardLinkRequired.Error())
	}

	var updateBoardRequest = dto.UpdateBoardRequest{
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Background:  req.GetBackground(),
	}
	updateBoardRequest.Sanitize()

	err = h.srv.UpdateBoard(ctx, dto.ToUpdateBoardInfo(updateBoardRequest, boardLink), userLink)
	if err != nil {
		switch {
		case errors.Is(err, rbac.ErrActionDenied):
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		case errors.Is(err, common.ErrBoardNotFound):
			return nil, status.Error(codes.NotFound, common.ErrBoardNotFound.Error())
		case errors.Is(err, common.ErrInvalidBoardData):
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidBoardData.Error())
		case errors.Is(err, common.ErrInvalidBoardReference):
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidBoardReference.Error())
		}

		errLog := fmt.Errorf("srv.UpdateBoard: %w", err)
		logger.Error().Err(errLog).Msg("board service UpdateBoard")
		sentryLogger.CaptureFromContext(ctx, errLog, "UpdateBoard", map[string]interface{}{
			"user_link":  rawUserLink,
			"board_link": rawBoardLink,
			"action":     "update_board",
		})
		return nil, status.Error(codes.Internal, ErrCannotUpdateBoard.Error())
	}

	return &pb.UpdateBoardResponse{}, nil
}

func (h *BoardHandler) UploadBackground(ctx context.Context, req *pb.UploadBackgroundRequest) (*pb.UploadBackgroundResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrUserLinkRequired.Error())
	}

	rawBoardLink := req.GetBoardLink()
	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrBoardLinkRequired.Error())
	}

	image := req.GetImage()

	contentType := http.DetectContentType(image)
	if !strings.HasPrefix(contentType, "image/") {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidContentType.Error())
	}

	filename := req.GetFilename()
	extension := filepath.Ext(filename)

	backgroundKey, err := h.srv.UpdateBackground(ctx, bytes.NewReader(image), contentType, extension, boardLink, userLink)
	if err != nil {
		if errors.Is(err, common.ErrBoardNotFound) {
			return nil, status.Error(codes.NotFound, common.ErrBoardNotFound.Error())
		}

		errLog := fmt.Errorf("srv.UpdateBackground: %w", err)
		logger.Error().Err(errLog).Msg("board service UpdateBackground")
		sentryLogger.CaptureFromContext(ctx, errLog, "UploadBackground", map[string]interface{}{
			"user_link":  rawUserLink,
			"board_link": rawBoardLink,
			"action":     "upload_background",
		})
		return nil, status.Error(codes.Internal, ErrCannotUpdateBackground.Error())
	}

	return &pb.UploadBackgroundResponse{
		BackgroundKey: fmt.Sprintf("%s/%s", h.conf.BaseBackgroundURL, backgroundKey),
	}, nil
}

func (h *BoardHandler) GetMembers(ctx context.Context, req *pb.GetMembersRequest) (*pb.GetMembersResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawBoardLink := req.GetBoardLink()
	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrBoardLinkRequired.Error())
	}

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrUserLinkRequired.Error())
	}

	members, err := h.srv.GetUsersOfBoard(ctx, boardLink, userLink)
	if err != nil {
		switch {
		case errors.Is(err, rbac.ErrActionDenied):
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		case errors.Is(err, common.ErrBoardNotFound):
			return nil, status.Error(codes.NotFound, common.ErrBoardNotFound.Error())
		}

		errLog := fmt.Errorf("srv.GetUsersOfBoard: %w", err)
		logger.Error().Err(errLog).Msg("BoardService.GetUsersOfBoard")
		sentryLogger.CaptureFromContext(ctx, errLog, "GetMembers", map[string]interface{}{
			"user_link":  rawUserLink,
			"board_link": rawBoardLink,
			"action":     "get_members",
		})
		return nil, status.Error(codes.Internal, ErrCannotGetUsersOfBoard.Error())
	}

	pbMembers := make([]*pb.MemberInfo, 0, len(members))
	for _, m := range members {
		pbMembers = append(pbMembers, &pb.MemberInfo{
			Link: m.Link.String(),
			Role: m.Role.String(),
		})
	}

	return &pb.GetMembersResponse{
		Members: pbMembers,
	}, nil
}

func (h *BoardHandler) CreateInvite(ctx context.Context, req *pb.CreateInviteRequest) (*pb.CreateInviteResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrUserLinkRequired.Error())
	}

	rawBoardLink := req.GetBoardLink()
	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrBoardLinkRequired.Error())
	}

	defaultRole := req.GetDefaultRole()
	if defaultRole == "" {
		return nil, status.Error(codes.InvalidArgument, ErrRoleRequired.Error())
	}

	role, err := rbac.ParseRole(defaultRole)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidRole.Error())
	}

	if !common.InviteAssignableRoles[role] {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidRole.Error())
	}

	var targetUserLink *uuid.UUID
	if req.TargetUserLink != nil {
		parsed, err := uuid.Parse(req.GetTargetUserLink())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, ErrInvalidBoardLink.Error())
		}
		targetUserLink = &parsed
	}

	var expireTime *time.Time
	if req.ExpireSeconds > 0 {
		t := time.Now().Add(time.Duration(req.ExpireSeconds) * time.Second)
		expireTime = &t
	}

	inviteInfo := serviceDto.NewInviteInfo{
		BoardLink:   boardLink,
		UserLink:    targetUserLink,
		DefaultRole: role,
		ExpireTime:  expireTime,
	}

	invite, err := h.srv.CreateInvite(ctx, inviteInfo, userLink)
	if err != nil {
		switch {
		case errors.Is(err, rbac.ErrActionDenied):
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		case errors.Is(err, common.ErrInvalidBoardReference):
			return nil, status.Error(codes.NotFound, common.ErrBoardNotFound.Error())
		case errors.Is(err, common.ErrNotNullValue):
			return nil, status.Error(codes.InvalidArgument, common.ErrNotNullValue.Error())
		}

		errLog := fmt.Errorf("srv.CreateInvite: %w", err)
		logger.Error().Err(errLog).Msg("BoardService.CreateInvite")
		sentryLogger.CaptureFromContext(ctx, errLog, "CreateInvite", map[string]interface{}{
			"user_link":   rawUserLink,
			"board_link":  rawBoardLink,
			"action":      "create_invite",
		})
		return nil, status.Error(codes.Internal, ErrCannotCreateInvite.Error())
	}

	resp := &pb.CreateInviteResponse{
		InviteLink:   invite.InviteLink.String(),
		BoardLink:    invite.BoardLink.String(),
		DefaultRole:  invite.DefaultRole.String(),
		Status:       invite.Status.String(),
		CreatedAt:    invite.CreatedAt.Unix(),
	}

	if invite.TargetUser != nil {
		target := invite.TargetUser.String()
		resp.TargetUserLink = &target
	}

	if invite.ExpireAt != nil {
		resp.ExpireAt = invite.ExpireAt.Unix()
	}

	return resp, nil
}

func (h *BoardHandler) AcceptInvite(ctx context.Context, req *pb.AcceptInviteRequest) (*pb.AcceptInviteResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawInviteLink := req.GetInviteLink()
	inviteLink, err := uuid.Parse(rawInviteLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInviteLinkRequired.Error())
	}

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrUserLinkRequired.Error())
	}

	invite, err := h.srv.AcceptInvite(ctx, inviteLink, userLink)
	if err != nil {
		switch {
		case errors.Is(err, common.ErrInviteNotFound):
			return nil, status.Error(codes.NotFound, common.ErrInviteNotFound.Error())
		case errors.Is(err, common.ErrInviteClosed):
			return nil, status.Error(codes.FailedPrecondition, common.ErrInviteClosed.Error())
		case errors.Is(err, common.ErrInviteExpired):
			return nil, status.Error(codes.FailedPrecondition, common.ErrInviteExpired.Error())
		case errors.Is(err, common.ErrInviteNotForUser):
			return nil, status.Error(codes.PermissionDenied, common.ErrInviteNotForUser.Error())
		case errors.Is(err, common.ErrUserAlreadyMember):
			return nil, status.Error(codes.AlreadyExists, common.ErrUserAlreadyMember.Error())
		}

		errLog := fmt.Errorf("srv.AcceptInvite: %w", err)
		logger.Error().Err(errLog).Msg("BoardService.AcceptInvite")
		sentryLogger.CaptureFromContext(ctx, errLog, "AcceptInvite", map[string]interface{}{
			"user_link":   rawUserLink,
			"invite_link": rawInviteLink,
			"action":      "accept_invite",
		})
		return nil, status.Error(codes.Internal, ErrCannotAcceptInvite.Error())
	}

	return &pb.AcceptInviteResponse{
		BoardLink: invite.BoardLink.String(),
		UserLink:  userLink.String(),
		Role:      invite.DefaultRole.String(),
	}, nil
}

func (h *BoardHandler) CloseInvite(ctx context.Context, req *pb.CloseInviteRequest) (*pb.CloseInviteResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrUserLinkRequired.Error())
	}

	rawInviteLink := req.GetInviteLink()
	inviteLink, err := uuid.Parse(rawInviteLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInviteLinkRequired.Error())
	}

	err = h.srv.CloseInvite(ctx, inviteLink, userLink)
	if err != nil {
		switch {
		case errors.Is(err, rbac.ErrActionDenied):
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		case errors.Is(err, common.ErrInviteNotFound):
			return nil, status.Error(codes.NotFound, common.ErrInviteNotFound.Error())
		}

		errLog := fmt.Errorf("srv.CloseInvite: %w", err)
		logger.Error().Err(errLog).Msg("BoardService.CloseInvite")
		sentryLogger.CaptureFromContext(ctx, errLog, "CloseInvite", map[string]interface{}{
			"user_link":   rawUserLink,
			"invite_link": rawInviteLink,
			"action":      "close_invite",
		})
		return nil, status.Error(codes.Internal, ErrCannotCloseInvite.Error())
	}

	return &pb.CloseInviteResponse{}, nil
}

func (h *BoardHandler) UpdateMemberRole(ctx context.Context, req *pb.UpdateMemberRoleRequest) (*pb.UpdateMemberRoleResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrUserLinkRequired.Error())
	}

	rawBoardLink := req.GetBoardLink()
	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrBoardLinkRequired.Error())
	}

	rawTargetLink := req.GetTargetUserLink()
	targetLink, err := uuid.Parse(rawTargetLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrUserLinkRequired.Error())
	}

	role, err := rbac.ParseRole(req.GetNewRole())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidRole.Error())
	}

	err = h.srv.UpdateMemberRole(ctx, boardLink, targetLink, role, userLink)
	if err != nil {
		switch {
		case errors.Is(err, rbac.ErrActionDenied):
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		case errors.Is(err, common.ErrBoardNotFound):
			return nil, status.Error(codes.NotFound, common.ErrBoardNotFound.Error())
		case errors.Is(err, common.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, common.ErrUserNotFound.Error())
		case errors.Is(err, common.ErrSelfRoleChange):
			return nil, status.Error(codes.InvalidArgument, common.ErrSelfRoleChange.Error())
		}
		errLog := fmt.Errorf("srv.UpdateMemberRole: %w", err)
		logger.Error().Err(errLog).Msg("BoardService.UpdateMemberRole")
		sentryLogger.CaptureFromContext(ctx, errLog, "UpdateMemberRole", map[string]interface{}{
			"user_link":    rawUserLink,
			"board_link":   rawBoardLink,
			"target_link":  rawTargetLink,
			"action":       "update_member_role",
		})
		return nil, status.Error(codes.Internal, ErrCannotUpdateBoard.Error())
	}

	return &pb.UpdateMemberRoleResponse{}, nil
}

func (h *BoardHandler) RemoveMemberFromBoard(ctx context.Context, req *pb.RemoveMemberFromBoardRequest) (*pb.RemoveMemberFromBoardResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrUserLinkRequired.Error())
	}

	rawBoardLink := req.GetBoardLink()
	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrBoardLinkRequired.Error())
	}

	rawTargetLink := req.GetTargetUserLink()
	targetLink, err := uuid.Parse(rawTargetLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrUserLinkRequired.Error())
	}

	err = h.srv.RemoveMemberFromBoard(ctx, boardLink, targetLink, userLink)
	if err != nil {
		switch {
		case errors.Is(err, rbac.ErrActionDenied):
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		case errors.Is(err, common.ErrBoardNotFound):
			return nil, status.Error(codes.NotFound, common.ErrBoardNotFound.Error())
		case errors.Is(err, common.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, common.ErrUserNotFound.Error())
		case errors.Is(err, common.ErrCreatorCannotLeave):
			return nil, status.Error(codes.PermissionDenied, common.ErrCreatorCannotLeave.Error())
		}
		errLog := fmt.Errorf("srv.RemoveMemberFromBoard: %w", err)
		logger.Error().Err(errLog).Msg("BoardService.RemoveMemberFromBoard")
		sentryLogger.CaptureFromContext(ctx, errLog, "RemoveMemberFromBoard", map[string]interface{}{
			"user_link":    rawUserLink,
			"board_link":   rawBoardLink,
			"target_link":  rawTargetLink,
			"action":       "remove_member",
		})
		return nil, status.Error(codes.Internal, ErrCannotUpdateBoard.Error())
	}

	return &pb.RemoveMemberFromBoardResponse{}, nil
}

func (h *BoardHandler) GetActiveInvites(ctx context.Context, req *pb.GetActiveInvitesRequest) (*pb.GetActiveInvitesResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrUserLinkRequired.Error())
	}

	rawBoardLink := req.GetBoardLink()
	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrBoardLinkRequired.Error())
	}

	invites, err := h.srv.GetActiveInvites(ctx, boardLink, userLink)
	if err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		}
		errLog := fmt.Errorf("srv.GetActiveInvites: %w", err)
		logger.Error().Err(errLog).Msg("BoardService.GetActiveInvites")
		sentryLogger.CaptureFromContext(ctx, errLog, "GetActiveInvites", map[string]interface{}{
			"user_link":  rawUserLink,
			"board_link": rawBoardLink,
			"action":     "get_active_invites",
		})
		return nil, status.Error(codes.Internal, ErrCannotGetBoards.Error())
	}

	pbInvites := make([]*pb.InviteInfo, 0, len(invites))
	for _, inv := range invites {
		info := &pb.InviteInfo{
			InviteLink:  inv.InviteLink.String(),
			BoardLink:   inv.BoardLink.String(),
			DefaultRole: inv.DefaultRole.String(),
			Status:      inv.Status.String(),
			CreatedAt:   inv.CreatedAt.Unix(),
		}

		if inv.TargetUser != nil {
			target := inv.TargetUser.String()
			info.TargetUserLink = &target
		}

		if inv.ExpireAt != nil {
			info.ExpireAt = inv.ExpireAt.Unix()
		}

		pbInvites = append(pbInvites, info)
	}

	return &pb.GetActiveInvitesResponse{
		Invites: pbInvites,
	}, nil
}

func (h *BoardHandler) CanView(ctx context.Context, req *pb.CanViewRequest) (*pb.CanViewResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrUserLinkRequired.Error())
	}

	rawBoardLink := req.GetBoardLink()
	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrBoardLinkRequired.Error())
	}

	if err := h.srv.CanView(ctx, boardLink, userLink); err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		}

		errLog := fmt.Errorf("srv.CanView: %w", err)
		logger.Error().Err(errLog).Msg("board service CanView")
		sentryLogger.CaptureFromContext(ctx, errLog, "CanView", map[string]interface{}{
			"user_link":  rawUserLink,
			"board_link": rawBoardLink,
			"action":     "can_view",
		})
		return nil, status.Error(codes.Internal, ErrCannotGetBoards.Error())
	}

	return &pb.CanViewResponse{}, nil
}

func (h *BoardHandler) CreatePoll(ctx context.Context, req *pb.CreatePollRequest) (*pb.CreatePollResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrUserLinkRequired.Error())
	}

	rawBoardLink := req.GetBoardLink()
	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrBoardLinkRequired.Error())
	}

	cards := make([]uuid.UUID, 0, len(req.GetCardLinks()))
	for _, raw := range req.GetCardLinks() {
		cardLink, err := uuid.Parse(raw)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid card link in poll")
		}
		cards = append(cards, cardLink)
	}

	invitees := make([]uuid.UUID, 0, len(req.GetInvitees()))
	for _, raw := range req.GetInvitees() {
		uid, err := uuid.Parse(raw)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid invitee link")
		}
		invitees = append(invitees, uid)
	}

	if err := h.srv.CreatePoll(ctx, boardLink, userLink, cards, invitees); err != nil {
		if errors.Is(err, rbac.ErrActionDenied) {
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		}
		if errors.Is(err, common.ErrPollAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, common.ErrPollAlreadyExists.Error())
		}
		errLog := fmt.Errorf("srv.CreatePoll: %w", err)
		logger.Error().Err(errLog).Msg("CreatePoll")
		sentryLogger.CaptureFromContext(ctx, errLog, "CreatePoll", map[string]interface{}{
			"user_link":  rawUserLink,
			"board_link": rawBoardLink,
			"action":     "create_poll",
		})
		return nil, status.Error(codes.Internal, "cannot create poll")
	}

	return &pb.CreatePollResponse{}, nil
}

func (h *BoardHandler) DeletePoll(ctx context.Context, req *pb.DeletePollRequest) (*pb.DeletePollResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrUserLinkRequired.Error())
	}

	rawBoardLink := req.GetBoardLink()
	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrBoardLinkRequired.Error())
	}

	if err := h.srv.DeletePoll(ctx, boardLink, userLink); err != nil {
		if errors.Is(err, common.ErrPollNotFound) {
			return nil, status.Error(codes.NotFound, common.ErrPollNotFound.Error())
		}
		if errors.Is(err, common.ErrNotPollAdmin) {
			return nil, status.Error(codes.PermissionDenied, common.ErrNotPollAdmin.Error())
		}
		errLog := fmt.Errorf("srv.DeletePoll: %w", err)
		logger.Error().Err(errLog).Msg("DeletePoll")
		sentryLogger.CaptureFromContext(ctx, errLog, "DeletePoll", map[string]interface{}{
			"user_link":  rawUserLink,
			"board_link": rawBoardLink,
			"action":     "delete_poll",
		})
		return nil, status.Error(codes.Internal, "cannot delete poll")
	}

	return &pb.DeletePollResponse{}, nil
}

func (h *BoardHandler) NextPollCard(ctx context.Context, req *pb.NextPollCardRequest) (*pb.NextPollCardResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrUserLinkRequired.Error())
	}

	rawBoardLink := req.GetBoardLink()
	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrBoardLinkRequired.Error())
	}

	if err := h.srv.NextPollCard(ctx, boardLink, userLink); err != nil {
		if errors.Is(err, common.ErrPollNotFound) {
			return nil, status.Error(codes.NotFound, common.ErrPollNotFound.Error())
		}
		if errors.Is(err, common.ErrNotPollAdmin) {
			return nil, status.Error(codes.PermissionDenied, common.ErrNotPollAdmin.Error())
		}
		if errors.Is(err, common.ErrPollNoMoreCards) {
			return &pb.NextPollCardResponse{}, nil
		}
		errLog := fmt.Errorf("srv.NextPollCard: %w", err)
		logger.Error().Err(errLog).Msg("NextPollCard")
		sentryLogger.CaptureFromContext(ctx, errLog, "NextPollCard", map[string]interface{}{
			"user_link":  rawUserLink,
			"board_link": rawBoardLink,
			"action":     "next_poll_card",
		})
		return nil, status.Error(codes.Internal, "cannot advance poll")
	}

	return &pb.NextPollCardResponse{}, nil
}

func (h *BoardHandler) VotePoll(ctx context.Context, req *pb.VotePollRequest) (*pb.VotePollResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrUserLinkRequired.Error())
	}

	rawBoardLink := req.GetBoardLink()
	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrBoardLinkRequired.Error())
	}

	if err := h.srv.VotePoll(ctx, boardLink, userLink, int(req.GetPoints())); err != nil {
		if errors.Is(err, common.ErrPollNotFound) {
			return nil, status.Error(codes.NotFound, common.ErrPollNotFound.Error())
		}
		if errors.Is(err, common.ErrUserNotInvited) {
			return nil, status.Error(codes.PermissionDenied, common.ErrUserNotInvited.Error())
		}
		errLog := fmt.Errorf("srv.VotePoll: %w", err)
		logger.Error().Err(errLog).Msg("VotePoll")
		sentryLogger.CaptureFromContext(ctx, errLog, "VotePoll", map[string]interface{}{
			"user_link":  rawUserLink,
			"board_link": rawBoardLink,
			"action":     "vote_poll",
		})
		return nil, status.Error(codes.Internal, "cannot vote")
	}

	return &pb.VotePollResponse{}, nil
}
