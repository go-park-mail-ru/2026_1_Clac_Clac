package clients

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/contracts/mail_sender"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MailSender struct {
	client pb.MailSenderServiceClient
}

func NewMailSenderClient(conn *grpc.ClientConn) *MailSender {
	return &MailSender{
		client: pb.NewMailSenderServiceClient(conn),
	}
}

func convertMailSenderGRPCError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	if st.Code() == codes.NotFound {
		return ErrResetTokenNotFound
	}
	return err
}

func (m *MailSender) SendRecoveryCode(ctx context.Context, recoveryInfo domain.RecoveryCode) error {
	req := &pb.SendRecoveryCodeRequest{
		Email:    recoveryInfo.Email,
		UserLink: recoveryInfo.UserLink.String(),
	}

	_, err := m.client.SendRecoveryCode(ctx, req)
	if err != nil {
		return fmt.Errorf("client.SendRecoveryCode: %w", convertMailSenderGRPCError(err))
	}

	return nil
}

func (m *MailSender) CheckRecoveryCode(ctx context.Context, check domain.RecoveryCodeCheck) error {
	req := &pb.CheckRecoveryCodeRequest{
		Code: check.Code,
	}

	_, err := m.client.CheckRecoveryCode(ctx, req)
	if err != nil {
		return fmt.Errorf("client.CheckRecoveryCode: %w", convertMailSenderGRPCError(err))
	}

	return nil
}

func (m *MailSender) ExchangeTokenForUser(ctx context.Context, resetToken domain.ResetToken) (uuid.UUID, error) {
	req := &pb.ExchangeTokenRequest{
		ResetToken: resetToken.Token,
	}

	resp, err := m.client.ExchangeTokenForUser(ctx, req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("client.ExchangeTokenForUser: %w", convertMailSenderGRPCError(err))
	}

	userLink, err := uuid.Parse(resp.UserLink)
	if err != nil {
		return uuid.Nil, common.ErrorParseLink
	}

	return userLink, nil
}
