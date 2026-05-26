package clients

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/appeal/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

const (
	appealChunkSize = 1024 * 1024
)

var appealBufferPool = sync.Pool{
	New: func() any {
		buffer := make([]byte, appealChunkSize)
		return &buffer
	},
}

type ConfigAppeal struct {
	MaxAppealAttachmentBytesSize int
}

type Appeal struct {
	client pb.AppealServiceClient
	cfg    ConfigAppeal
}

func NewAppealClient(connection *grpc.ClientConn, cfg ConfigAppeal) *Appeal {
	return &Appeal{
		client: pb.NewAppealServiceClient(connection),
		cfg:    cfg,
	}
}

func (a *Appeal) categoryToProto(c string) pb.Category {
	switch c {
	case "bug":
		return pb.Category_CATEGORY_BUG
	case "proposal":
		return pb.Category_CATEGORY_PROPOSAL
	case "complaint":
		return pb.Category_CATEGORY_COMPLAINT
	}

	return pb.Category_CATEGORY_UNSPECIFIED
}

func (a *Appeal) statusToProto(s string) pb.Status {
	switch s {
	case "open", "new":
		return pb.Status_STATUS_OPEN
	case "in_work", "in_progress":
		return pb.Status_STATUS_IN_WORK
	case "close", "closed":
		return pb.Status_STATUS_CLOSE
	}

	return pb.Status_STATUS_UNSPECIFIED
}

func (a *Appeal) parseProtoRole(role pb.Role) string {
	switch role {
	case pb.Role_ROLE_USER:
		return "user"
	case pb.Role_ROLE_SUPPORT:
		return "support"
	case pb.Role_ROLE_ADMIN:
		return "admin"
	}

	return "unknown"
}

func (a *Appeal) parseProtoCategory(c pb.Category) string {
	switch c {
	case pb.Category_CATEGORY_BUG:
		return "bug"
	case pb.Category_CATEGORY_PROPOSAL:
		return "proposal"
	case pb.Category_CATEGORY_COMPLAINT:
		return "complaint"
	}

	return "unknown"
}

func (a *Appeal) parseProtoStatus(s pb.Status) string {
	switch s {
	case pb.Status_STATUS_OPEN:
		return "new"
	case pb.Status_STATUS_IN_WORK:
		return "in_progress"
	case pb.Status_STATUS_CLOSE:
		return "closed"
	}

	return "unknown"
}

func (a *Appeal) CreateAppeal(ctx context.Context, newAppeal domain.CreateAppealInfo) (uuid.UUID, error) {
	req := &pb.CreateAppealRequest{
		UserLink:    newAppeal.UserLink.String(),
		Email:       newAppeal.Email,
		Category:    a.categoryToProto(newAppeal.Category),
		Description: newAppeal.Description,
		DisplayName: newAppeal.DisplayName,
	}

	res, err := a.client.CreateAppeal(ctx, req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("AppealClient.CreateAppeal: %w", convertAppealGRPCError(err))
	}

	rawAppealLink := res.GetAppealLink()
	appealLink, err := uuid.Parse(rawAppealLink)
	if err != nil {
		return uuid.Nil, fmt.Errorf("uuid.Parse: %w", err)
	}

	return appealLink, nil
}

func (a *Appeal) GetAppeal(ctx context.Context, userLink uuid.UUID) (string, []domain.AppealInfo, error) {
	req := &pb.GetAppealsRequest{
		UserLink: userLink.String(),
	}

	res, err := a.client.GetAppeals(ctx, req)
	if err != nil {
		return "", []domain.AppealInfo{}, fmt.Errorf("AppealClient.GetAppeals: %w", convertAppealGRPCError(err))
	}

	appeals := make([]domain.AppealInfo, 0, len(res.AppealsInfo))
	for _, appeal := range res.AppealsInfo {
		rawAppealLink := appeal.GetAppealLink()
		appealLink, err := uuid.Parse(rawAppealLink)
		if err != nil {
			return "", []domain.AppealInfo{}, fmt.Errorf("uuid.Parse: %w", err)
		}

		appeals = append(appeals, domain.AppealInfo{
			AppealID:      appeal.GetAppealId(),
			AppealLink:    appealLink,
			Email:         appeal.GetEmail(),
			Category:      a.parseProtoCategory(appeal.GetCategory()),
			Status:        a.parseProtoStatus(appeal.GetStatus()),
			DisplayName:   appeal.GetDisplayName(),
			Description:   appeal.GetDescription(),
			AttachmentURL: appeal.GetAttachmentUrl(),
			CreatedAt:     appeal.GetCreatedAt().AsTime(),
		})
	}

	return a.parseProtoRole(res.Role), appeals, nil
}

func (a *Appeal) UploadAttachment(ctx context.Context, attachmentInfo domain.UploadAttachmentInfo, attachment io.Reader) (string, error) {
	stream, err := a.client.UploadAttachment(ctx)
	if err != nil {
		return "", fmt.Errorf("can not open stream client.UploadAttachment: %w", err)
	}

	req := &pb.UploadAttachmentRequest{
		Request: &pb.UploadAttachmentRequest_Metadata{
			Metadata: &pb.MetadataUploadAttachment{
				UserLink:   attachmentInfo.UserLink.String(),
				AppealLink: attachmentInfo.AppealLink.String(),
				Filename:   attachmentInfo.Filename,
			},
		},
	}

	if err := stream.Send(req); err != nil {
		return "", fmt.Errorf("metadata appeal stream.Send: %w", err)
	}

	bufPtr := appealBufferPool.Get().(*[]byte)
	buffer := *bufPtr

	defer appealBufferPool.Put(bufPtr)

	for {
		n, err := attachment.Read(buffer)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("attachment.Read: %w", err)
		}

		if n > 0 {
			chunkReq := &pb.UploadAttachmentRequest{
				Request: &pb.UploadAttachmentRequest_Image{
					Image: buffer[:n],
				},
			}

			if err := stream.Send(chunkReq); err != nil {
				return "", fmt.Errorf("stream.Send: %w", err)
			}
		}

		if err == io.EOF {
			break
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		return "", fmt.Errorf("AppealService.UploadAttachment: %w", convertAppealGRPCError(err))
	}

	return res.GetAttachmentUrl(), nil
}

func (a *Appeal) DeleteAppeal(ctx context.Context, deleteInfo domain.DeleteInfo) error {
	req := &pb.DeleteAppealRequest{
		UserLink:   deleteInfo.UserLink.String(),
		AppealLink: deleteInfo.AppealLink.String(),
	}

	_, err := a.client.DeleteAppeal(ctx, req)
	if err != nil {
		return fmt.Errorf("AppealClient.DeleteAppeal: %w", convertAppealGRPCError(err))
	}

	return nil
}

func (a *Appeal) GetStats(ctx context.Context, userLink uuid.UUID) (domain.AppealsStats, error) {
	req := &pb.GetStatsRequest{
		UserLink: userLink.String(),
	}

	res, err := a.client.GetStats(ctx, req)
	if err != nil {
		return domain.AppealsStats{}, fmt.Errorf("AppealClient.GetStats: %w", convertAppealGRPCError(err))
	}

	return domain.AppealsStats{
		OpenAppeals:   res.GetOpenAppeals(),
		InWorkAppeals: res.GetInWorkAppeals(),
		CloseAppeals:  res.GetCloseAppeals(),
	}, nil
}

func (a *Appeal) ChangeAppealStatus(ctx context.Context, changeStatusInfo domain.ChangeAppealStatusInfo) error {
	req := &pb.ChangeAppealStatusRequest{
		UserLink:   changeStatusInfo.UserLink.String(),
		AppealLink: changeStatusInfo.AppealLink.String(),
		NewStatus:  a.statusToProto(changeStatusInfo.NewStatus),
	}

	_, err := a.client.ChangeAppealStatus(ctx, req)
	if err != nil {
		return fmt.Errorf("AppealClient.ChangeAppealStatus: %w", convertAppealGRPCError(err))
	}

	return nil
}
