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

//go:generate mockery --name=SectionService --output=mock_section_service --outpkg=mockSectionService
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

	sections, err := h.srv.GetAllSections(ctx, boardLink)
	if err != nil {
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

	section, err := h.srv.GetSectionInfo(ctx, sectionLink)
	if err != nil {
		if errors.Is(err, common.ErrSectionNotFound) {
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

	cards, err := h.srv.GetCards(ctx, sectionLink)
	if err != nil {
		if errors.Is(err, common.ErrSectionNotFound) {
			return nil, status.Error(codes.NotFound, common.ErrSectionNotFound.Error())
		}

		logger.Error().Err(err).Msg("SectionService.GetCards")
		return nil, status.Error(codes.Internal, ErrCannotGetCards.Error())
	}

	cardsResponse := make([]*pb.CardInfo, 0)
	for _, card := range cards {
		cardsResponse = append(cardsResponse, &pb.CardInfo{
			Link:         card.CardLink.String(),
			ExecuterName: card.ExecuterName,
			Title:        card.Title,
			Deadline:     timestamppb.New(*card.DeadLine),
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
	})

	if err != nil {
		if errors.Is(err, common.ErrSectionAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, common.ErrSectionAlreadyExists.Error())
		}

		if errors.Is(err, common.ErrInvalidReferenceSectionData) {
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidReferenceSectionData.Error())
		}

		if errors.Is(err, common.ErrInvalidSectionData) {
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidSectionData.Error())
		}

		if errors.Is(err, common.ErrMissingRequiredField) {
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

	err = h.srv.DeleteSection(ctx, sectionLink)
	if err != nil {
		if errors.Is(err, common.ErrSectionNotFound) {
			return nil, status.Error(codes.NotFound, common.ErrSectionNotFound.Error())
		}

		if errors.Is(err, common.ErrCannotDeleteBacklog) {
			return nil, status.Error(codes.PermissionDenied, common.ErrCannotDeleteBacklog.Error())
		}

		if errors.Is(err, common.ErrInvalidReferenceSectionData) {
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidReferenceSectionData.Error())
		}

		if errors.Is(err, common.ErrInvalidSectionData) {
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidSectionData.Error())
		}

		if errors.Is(err, common.ErrMissingRequiredField) {
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
		return nil, status.Error(codes.InvalidArgument, ErrIncorrectMaxTasksValue.Error())
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
	})
	if err != nil {
		if errors.Is(err, common.ErrSectionNotFound) {
			return nil, status.Error(codes.NotFound, common.ErrSectionNotFound.Error())
		}

		if errors.Is(err, common.ErrCannotUpdateBacklog) {
			return nil, status.Error(codes.PermissionDenied, common.ErrCannotUpdateBacklog.Error())
		}

		if errors.Is(err, common.ErrInvalidReferenceSectionData) {
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidReferenceSectionData.Error())
		}

		if errors.Is(err, common.ErrInvalidSectionData) {
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidSectionData.Error())
		}

		if errors.Is(err, common.ErrMissingRequiredField) {
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

	sectionsLinks := make([]uuid.UUID, 0)
	for _, rawLink := range req.GetLinksList() {
		link, err := uuid.Parse(rawLink)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, ErrInvalidSectionLink.Error())
		}

		sectionsLinks = append(sectionsLinks, link)
	}

	err = h.srv.ReorderSection(ctx, boardLink, sectionsLinks)
	if err != nil {
		if errors.Is(err, common.ErrNotFindAllLinks) {
			return nil, status.Error(codes.NotFound, common.ErrNotFindAllLinks.Error())
		}

		if errors.Is(err, common.ErrInvalidReferenceSectionData) {
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidReferenceSectionData.Error())
		}

		if errors.Is(err, common.ErrInvalidSectionData) {
			return nil, status.Error(codes.InvalidArgument, common.ErrInvalidSectionData.Error())
		}

		if errors.Is(err, common.ErrMissingRequiredField) {
			return nil, status.Error(codes.InvalidArgument, common.ErrMissingRequiredField.Error())
		}

		logger.Error().Err(err).Msg("SectionService.ReorderSection")
		return nil, status.Error(codes.Internal, ErrCannotReorderSections.Error())
	}

	return &pb.ReorderSectionResponse{}, nil
}
