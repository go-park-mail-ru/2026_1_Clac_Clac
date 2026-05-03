package delivery

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/delivery/dto"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/service/dto"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/boardRbac"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/section/v1"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

var (
	ErrInvalidBoardLink           = errors.New("invalid board link")
	ErrInvalidSectionLink         = errors.New("invalid section link")
	ErrCannotGetSections          = errors.New("cannot get sections")
	ErrCannotGetSection           = errors.New("cannot get section")
	ErrCannotGetCards             = errors.New("cannot get cards")
	ErrIncorrectMaxTasksValue     = errors.New("incorrect max tasks value")
	ErrCannotCreateSection        = errors.New("cannot create section")
	ErrCannotDeleteSection        = errors.New("cannot delete section")
	ErrCannotReorderSections      = errors.New("cannot reorder section")
	ErrIncorrectSectionNameLength = errors.New("incorrect section name length")
	ErrIncorrectColor             = errors.New("incorrect color value")
	ErrCannotUpdateSection        = errors.New("cannot update section")
	ErrInvalidUserLink            = errors.New("invalid user link")
)

//go:generate mockery --name=SectionService --output=mock_section_service --outpkg=mockSectionService
type SectionService interface {
	GetSections(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID) ([]serviceDto.FullSectionInfo, error)
	GetSection(ctx context.Context, linkSection uuid.UUID, userLink uuid.UUID) (serviceDto.FullSectionInfo, error)
	GetCards(ctx context.Context, linkSection uuid.UUID, userLink uuid.UUID) ([]serviceDto.Card, error)
	CreateSection(ctx context.Context, newSection serviceDto.CreatingSection, userLink uuid.UUID) (serviceDto.EntitySection, error)
	DeleteSection(ctx context.Context, linkSection uuid.UUID, userLink uuid.UUID) error
	UpdateSection(ctx context.Context, updatingSection serviceDto.FullSectionInfo, userLink uuid.UUID) error
	ReorderSection(ctx context.Context, linkBoard uuid.UUID, listLinkSection []uuid.UUID, userLink uuid.UUID) error
}

type Config struct {
	MaxQuantityTasks  int
	MinQuantityTasks  int
	MaxLenNameSection int
}

type SectionHandler struct {
	pb.UnimplementedSectionServiceServer

	srv SectionService
	cnf Config
}

func NewHandler(srv SectionService, cnf Config) *SectionHandler {
	return &SectionHandler{
		srv: srv,
		cnf: cnf,
	}
}

func (h *SectionHandler) GetSections(ctx context.Context, req *pb.GetSectionsRequest) (*pb.GetSectionsResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawBoardLink := req.GetBoardLink()
	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidBoardLink.Error())
	}

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidUserLink.Error())
	}

	sections, err := h.srv.GetSections(ctx, boardLink, userLink)
	if err != nil {
		switch {
		case errors.Is(err, rbac.ErrActionDenied):
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		}

		logger.Error().Err(err).Msg("SectionService.GetAllSections")
		return nil, status.Error(codes.Internal, ErrCannotGetSections.Error())
	}

	sectionsResponse := []*pb.SectionInfo{}
	for _, section := range sections {
		var maxTasks int64

		if section.MaxTasks != nil {
			maxTasks = int64(*section.MaxTasks)
		}

		sectionsResponse = append(sectionsResponse, &pb.SectionInfo{
			Link:        section.SectionLink.String(),
			Name:        section.SectionName,
			Position:    int64(section.Position),
			IsMandatory: section.IsMandatory,
			Color:       section.Color,
			MaxTasks:    &maxTasks,
		})
	}

	return &pb.GetSectionsResponse{
		SectionsInfo: sectionsResponse,
	}, nil
}

func (h *SectionHandler) GetSection(ctx context.Context, req *pb.GetSectionRequest) (*pb.GetSectionResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawSectionLink := req.GetSectionLink()
	sectionLink, err := uuid.Parse(rawSectionLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidSectionLink.Error())
	}

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidUserLink.Error())
	}

	section, err := h.srv.GetSection(ctx, sectionLink, userLink)
	if err != nil {
		switch {
		case errors.Is(err, rbac.ErrActionDenied):
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		case errors.Is(err, common.ErrSectionNotFound):
			return nil, status.Error(codes.NotFound, common.ErrSectionNotFound.Error())
		}

		logger.Error().Err(err).Msg("SectionService.GetSectionInfo")
		return nil, status.Error(codes.Internal, ErrCannotGetSection.Error())
	}

	var maxTasks int64

	if section.MaxTasks != nil {
		maxTasks = int64(*section.MaxTasks)
	}

	sectionInfo := &pb.SectionInfo{
		Link:        section.SectionLink.String(),
		Name:        section.SectionName,
		Position:    int64(section.Position),
		IsMandatory: section.IsMandatory,
		Color:       section.Color,
		MaxTasks:    &maxTasks,
	}

	return &pb.GetSectionResponse{
		SectionInfo: sectionInfo,
	}, nil
}

func (h *SectionHandler) GetCards(ctx context.Context, req *pb.GetCardsRequest) (*pb.GetCardsResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawSectionLink := req.GetSectionLink()
	sectionLink, err := uuid.Parse(rawSectionLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidSectionLink.Error())
	}

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidUserLink.Error())
	}

	cards, err := h.srv.GetCards(ctx, sectionLink, userLink)
	if err != nil {
		switch {
		case errors.Is(err, rbac.ErrActionDenied):
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		case errors.Is(err, common.ErrSectionNotFound):
			return nil, status.Error(codes.NotFound, common.ErrSectionNotFound.Error())
		}

		logger.Error().Err(err).Msg("SectionService.GetCards")
		return nil, status.Error(codes.Internal, ErrCannotGetCards.Error())
	}

	cardsResponse := make([]*pb.CardInfo, 0, len(cards))
	for _, card := range cards {
		var deadline *timestamppb.Timestamp
		if card.DeadLine != nil {
			deadline = timestamppb.New(*card.DeadLine)
		}

		var subtasks []*pb.SubtaskInfo

		for _, sub := range card.Subtasks {
			subtasks = append(subtasks, &pb.SubtaskInfo{
				SubtaskLink: sub.SubtaskLink.String(),
				Description: sub.Description,
				IsDone:      sub.IsDone,
				Position:    int64(sub.Position),
			})
		}

		var executorLink *string
		if card.ExecutorLink != nil {
			s := card.ExecutorLink.String()
			executorLink = &s
		}

		cardsResponse = append(cardsResponse, &pb.CardInfo{
			Link:         card.CardLink.String(),
			ExecutorLink: executorLink,
			Title:        card.Title,
			Deadline:     deadline,
			Subtasks:    subtasks,
			Position:   int64(card.Position),
		})
	}

	return &pb.GetCardsResponse{
		CardsInfo: cardsResponse,
	}, nil
}

func (h *SectionHandler) CreateSection(ctx context.Context, req *pb.CreateSectionRequest) (*pb.CreateSectionResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawBoardLink := req.GetBoardLink()
	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidBoardLink.Error())
	}

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidUserLink.Error())
	}

	var maxTasks int
	if req.MaxTasks != nil {
		maxTasks = int(req.GetMaxTasks())
	}

	var newSection = dto.CreatingSection{
		BoardLink:   boardLink,
		SectionName: req.GetName(),
		IsMandatory: req.GetIsMandatory(),
		Color:       req.GetColor(),
		MaxTasks:    &maxTasks,
	}

	if newSection.MaxTasks != nil && (*newSection.MaxTasks > h.cnf.MaxQuantityTasks ||
		*newSection.MaxTasks < h.cnf.MinQuantityTasks) {
		return nil, status.Error(codes.InvalidArgument, ErrIncorrectMaxTasksValue.Error())
	}

	section, err := h.srv.CreateSection(ctx, serviceDto.CreatingSection{
		BoardLink:   newSection.BoardLink,
		SectionName: newSection.SectionName,
		IsMandatory: newSection.IsMandatory,
		Color:       newSection.Color,
		MaxTasks:    newSection.MaxTasks,
	}, userLink)

	if err != nil {
		switch {
		case errors.Is(err, rbac.ErrActionDenied):
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		case errors.Is(err, common.ErrSectionAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, common.ErrSectionAlreadyExists.Error())
		case errors.Is(err, common.ErrInvalidReferenceSectionData):
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidReferenceSectionData.Error())
		case errors.Is(err, common.ErrInvalidSectionData):
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidSectionData.Error())
		case errors.Is(err, common.ErrMissingRequiredField):
			return nil, status.Error(codes.InvalidArgument, common.ErrMissingRequiredField.Error())
		}

		logger.Error().Err(err).Msg("SectionService.CreateSection")
		return nil, status.Error(codes.Internal, ErrCannotCreateSection.Error())
	}

	var maxTasksResponse int64
	if section.MaxTasks != nil {
		maxTasksResponse = int64(*section.MaxTasks)
	}

	return &pb.CreateSectionResponse{
		SectionInfo: &pb.SectionInfo{
			Link:        section.SectionLink.String(),
			Name:        section.SectionName,
			IsMandatory: section.IsMandatory,
			Position:    int64(section.Position),
			Color:       section.Color,
			MaxTasks:    &maxTasksResponse,
		},
	}, nil
}

func (h *SectionHandler) DeleteSection(ctx context.Context, req *pb.DeleteSectionRequest) (*pb.DeleteSectionResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawSectionLink := req.GetSectionLink()
	sectionLink, err := uuid.Parse(rawSectionLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidSectionLink.Error())
	}

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidUserLink.Error())
	}

	err = h.srv.DeleteSection(ctx, sectionLink, userLink)
	if err != nil {
		switch {
		case errors.Is(err, rbac.ErrActionDenied):
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		case errors.Is(err, common.ErrSectionNotFound):
			return nil, status.Error(codes.NotFound, common.ErrSectionNotFound.Error())
		case errors.Is(err, common.ErrCannotDeleteBacklog):
			return nil, status.Error(codes.PermissionDenied, common.ErrCannotDeleteBacklog.Error())
		case errors.Is(err, common.ErrInvalidReferenceSectionData):
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidReferenceSectionData.Error())
		case errors.Is(err, common.ErrInvalidSectionData):
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidSectionData.Error())
		case errors.Is(err, common.ErrMissingRequiredField):
			return nil, status.Error(codes.InvalidArgument, common.ErrMissingRequiredField.Error())
		}

		logger.Error().Err(err).Msg("SectionService.DeleteSection")
		return nil, status.Error(codes.Internal, ErrCannotDeleteSection.Error())
	}

	return &pb.DeleteSectionResponse{}, nil
}

func (h *SectionHandler) UpdateSection(ctx context.Context, req *pb.UpdateSectionRequest) (*pb.UpdateSectionResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawSectionLink := req.GetSectionLink()
	sectionLink, err := uuid.Parse(rawSectionLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidSectionLink.Error())
	}

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidUserLink.Error())
	}

	var maxTasks int
	if req.MaxTasks != nil {
		maxTasks = int(req.GetMaxTasks())
	}

	sectionInfo := dto.FullSectionInfo{
		SectionLink: sectionLink,
		SectionName: req.GetName(),
		IsMandatory: req.GetIsMandatory(),
		Color:       req.GetColor(),
		MaxTasks:    &maxTasks,
	}

	if !common.CheckSectionNameLength(sectionInfo.SectionName, h.cnf.MaxLenNameSection) {
		return nil, status.Error(codes.InvalidArgument, ErrIncorrectSectionNameLength.Error())
	}

	if !common.CheckColor(sectionInfo.Color) {
		return nil, status.Error(codes.InvalidArgument, ErrIncorrectColor.Error())
	}

	if sectionInfo.MaxTasks != nil {
		if !common.CheckMaxTasks(*sectionInfo.MaxTasks, h.cnf.MaxQuantityTasks, h.cnf.MinQuantityTasks) {
			return nil, status.Error(codes.InvalidArgument, ErrIncorrectMaxTasksValue.Error())
		}
	}

	err = h.srv.UpdateSection(ctx, serviceDto.FullSectionInfo{
		SectionLink: sectionLink,
		SectionName: sectionInfo.SectionName,
		Position:    sectionInfo.Position,
		IsMandatory: sectionInfo.IsMandatory,
		Color:       sectionInfo.Color,
		MaxTasks:    sectionInfo.MaxTasks,
	}, userLink)
	if err != nil {
		switch {
		case errors.Is(err, rbac.ErrActionDenied):
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		case errors.Is(err, common.ErrSectionNotFound):
			return nil, status.Error(codes.NotFound, common.ErrSectionNotFound.Error())
		case errors.Is(err, common.ErrCannotUpdateBacklog):
			return nil, status.Error(codes.PermissionDenied, common.ErrCannotUpdateBacklog.Error())
		case errors.Is(err, common.ErrInvalidReferenceSectionData):
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidReferenceSectionData.Error())
		case errors.Is(err, common.ErrInvalidSectionData):
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidSectionData.Error())
		case errors.Is(err, common.ErrMissingRequiredField):
			return nil, status.Error(codes.InvalidArgument, common.ErrMissingRequiredField.Error())
		}

		logger.Error().Err(err).Msg("SectionService.UpdateSection")
		return nil, status.Error(codes.Internal, ErrCannotUpdateSection.Error())
	}

	return &pb.UpdateSectionResponse{}, nil
}

func (h *SectionHandler) ReorderSection(ctx context.Context, req *pb.ReorderSectionRequest) (*pb.ReorderSectionResponse, error) {
	logger := zerolog.Ctx(ctx)

	rawBoardLink := req.GetBoardLink()
	boardLink, err := uuid.Parse(rawBoardLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidBoardLink.Error())
	}

	rawUserLink := req.GetUserLink()
	userLink, err := uuid.Parse(rawUserLink)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidUserLink.Error())
	}

	sectionsLinks := make([]uuid.UUID, 0)
	for _, rawLink := range req.GetLinksList() {
		link, err := uuid.Parse(rawLink)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, ErrInvalidSectionLink.Error())
		}

		sectionsLinks = append(sectionsLinks, link)
	}

	err = h.srv.ReorderSection(ctx, boardLink, sectionsLinks, userLink)
	if err != nil {
		switch {
		case errors.Is(err, rbac.ErrActionDenied):
			return nil, status.Error(codes.PermissionDenied, rbac.ErrActionDenied.Error())
		case errors.Is(err, common.ErrNotFindAllLinks):
			return nil, status.Error(codes.NotFound, common.ErrNotFindAllLinks.Error())
		case errors.Is(err, common.ErrInvalidReferenceSectionData):
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidReferenceSectionData.Error())
		case errors.Is(err, common.ErrInvalidSectionData):
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidSectionData.Error())
		case errors.Is(err, common.ErrMissingRequiredField):
			return nil, status.Error(codes.InvalidArgument, common.ErrMissingRequiredField.Error())
		}

		logger.Error().Err(err).Msg("SectionService.ReorderSection")
		return nil, status.Error(codes.Internal, ErrCannotReorderSections.Error())
	}

	return &pb.ReorderSectionResponse{}, nil
}
