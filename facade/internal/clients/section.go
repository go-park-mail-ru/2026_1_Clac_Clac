package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/section/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type Section struct {
	client pb.SectionServiceClient
}

func NewSectionClient(connection *grpc.ClientConn) *Section {
	return &Section{
		client: pb.NewSectionServiceClient(connection),
	}
}

func (s *Section) GetSections(ctx context.Context, sectionRequest domain.GetSectionsRequest) ([]domain.SectionInfo, error) {
	req := &pb.GetSectionsRequest{
		UserLink:  sectionRequest.UserLink.String(),
		BoardLink: sectionRequest.BoardLink.String(),
	}

	res, err := s.client.GetSections(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("client.GetSections: %w", convertSectionGRPCError(err))
	}

	sections := make([]domain.SectionInfo, 0, len(res.SectionsInfo))
	for _, si := range res.SectionsInfo {
		link, err := uuid.Parse(si.Link)
		if err != nil {
			return nil, common.ErrorParseLink
		}
		sections = append(sections, domain.SectionInfo{
			Link:        link,
			Name:        si.Name,
			Position:    si.Position,
			IsMandatory: si.IsMandatory,
			Color:       si.Color,
			MaxTasks:    si.MaxTasks,
		})
	}

	return sections, nil
}

func (s *Section) GetSection(ctx context.Context, sectionRequest domain.GetSectionRequest) (domain.SectionInfo, error) {
	req := &pb.GetSectionRequest{
		UserLink:    sectionRequest.UserLink.String(),
		SectionLink: sectionRequest.SectionLink.String(),
	}

	res, err := s.client.GetSection(ctx, req)
	if err != nil {
		return domain.SectionInfo{}, fmt.Errorf("client.GetSection: %w", convertSectionGRPCError(err))
	}

	link, err := uuid.Parse(res.SectionInfo.Link)
	if err != nil {
		return domain.SectionInfo{}, common.ErrorParseLink
	}

	return domain.SectionInfo{
		Link:        link,
		Name:        res.SectionInfo.Name,
		Position:    res.SectionInfo.Position,
		IsMandatory: res.SectionInfo.IsMandatory,
		Color:       res.SectionInfo.Color,
		MaxTasks:    res.SectionInfo.MaxTasks,
	}, nil
}

func (s *Section) GetCards(ctx context.Context, cardRequest domain.GetCardsRequest) ([]domain.CardInfo, error) {
	req := &pb.GetCardsRequest{
		UserLink:    cardRequest.UserLink.String(),
		SectionLink: cardRequest.SectionLink.String(),
	}

	res, err := s.client.GetCards(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("client.GetCards: %w", convertSectionGRPCError(err))
	}

	cards := make([]domain.CardInfo, 0, len(res.CardsInfo))
	for _, ci := range res.CardsInfo {
		link, err := uuid.Parse(ci.Link)
		if err != nil {
			return nil, common.ErrorParseLink
		}

		var deadline *time.Time
		if ci.Deadline != nil {
			t := ci.Deadline.AsTime()
			deadline = &t
		}

		subtasks := make([]domain.SubtaskInfo, 0, len(ci.Subtasks))
		for _, st := range ci.Subtasks {
			subtaskLink, err := uuid.Parse(st.SubtaskLink)
			if err != nil {
				return nil, common.ErrorParseLink
			}
			subtasks = append(subtasks, domain.SubtaskInfo{
				SubtaskLink: subtaskLink,
				Description: st.Description,
				IsDone:      st.IsDone,
				Position:    int(st.Position),
			})
		}

		var executorLink *uuid.UUID
		if ci.ExecutorLink != nil {
			el, err := uuid.Parse(*ci.ExecutorLink)
			if err != nil {
				return nil, common.ErrorParseLink
			}
			executorLink = &el
		}

		cards = append(cards, domain.CardInfo{
			CardLink:     link,
			ExecutorLink: executorLink,
			Title:        ci.Title,
			Deadline:     deadline,
			Subtasks:     subtasks,
		})
	}

	return cards, nil
}

func (s *Section) CreateSection(ctx context.Context, sectionInfo domain.CreateSectionRequest) (domain.SectionInfo, error) {
	req := &pb.CreateSectionRequest{
		UserLink:    sectionInfo.UserLink.String(),
		BoardLink:   sectionInfo.BoardLink.String(),
		Name:        sectionInfo.Name,
		IsMandatory: sectionInfo.IsMandatory,
		Color:       sectionInfo.Color,
		MaxTasks:    sectionInfo.MaxTasks,
	}

	res, err := s.client.CreateSection(ctx, req)
	if err != nil {
		return domain.SectionInfo{}, fmt.Errorf("client.CreateSection: %w", convertSectionGRPCError(err))
	}

	link, err := uuid.Parse(res.SectionInfo.Link)
	if err != nil {
		return domain.SectionInfo{}, common.ErrorParseLink
	}

	return domain.SectionInfo{
		Link:        link,
		Name:        res.SectionInfo.Name,
		Position:    res.SectionInfo.Position,
		IsMandatory: res.SectionInfo.IsMandatory,
		Color:       res.SectionInfo.Color,
		MaxTasks:    res.SectionInfo.MaxTasks,
	}, nil
}

func (s *Section) DeleteSection(ctx context.Context, sectionInfo domain.DeleteSectionRequest) error {
	req := &pb.DeleteSectionRequest{
		UserLink:    sectionInfo.UserLink.String(),
		SectionLink: sectionInfo.SectionLink.String(),
	}

	_, err := s.client.DeleteSection(ctx, req)
	if err != nil {
		return fmt.Errorf("client.DeleteSection: %w", convertSectionGRPCError(err))
	}

	return nil
}

func (s *Section) ReorderSection(ctx context.Context, sectionInfo domain.ReorderSectionRequest) error {
	linksList := make([]string, 0, len(sectionInfo.LinksList))
	for _, l := range sectionInfo.LinksList {
		linksList = append(linksList, l.String())
	}

	req := &pb.ReorderSectionRequest{
		UserLink:  sectionInfo.UserLink.String(),
		BoardLink: sectionInfo.BoardLink.String(),
		LinksList: linksList,
	}

	_, err := s.client.ReorderSection(ctx, req)
	if err != nil {
		return fmt.Errorf("client.ReorderSection: %w", convertSectionGRPCError(err))
	}

	return nil
}

func (s *Section) UpdateSection(ctx context.Context, sectionInfo domain.UpdateSectionRequest) error {
	req := &pb.UpdateSectionRequest{
		UserLink:    sectionInfo.UserLink.String(),
		SectionLink: sectionInfo.SectionLink.String(),
		Name:        sectionInfo.Name,
		IsMandatory: sectionInfo.IsMandatory,
		Color:       sectionInfo.Color,
		MaxTasks:    sectionInfo.MaxTasks,
	}

	_, err := s.client.UpdateSection(ctx, req)
	if err != nil {
		return fmt.Errorf("client.UpdateSection: %w", convertSectionGRPCError(err))
	}

	return nil
}
