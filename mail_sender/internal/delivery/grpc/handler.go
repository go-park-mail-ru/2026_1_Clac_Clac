package grpc

import (
	"context"

	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/mail"
)

type Sender interface {
	SendLetter(ctx context.Context, to, subject, body string) error
}

type MailHandler struct {
	pb.UnimplementedMailServiceServer
	manager Sender
}

func NewMailHandler(manager Sender) *MailHandler {
	return &MailHandler{
		manager: manager,
	}
}

func (m *MailHandler) SendEmail(ctx context.Context, req *pb.SendMailRequest) (*pb.SendMmailResponse, error) {
	to := req.To
	subject := req.Subject
	body := req.Body

	err := m.manager.SendLetter(ctx, to, subject, body)
	if err != nil {
		return &pb.SendMmailResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.SendMmailResponse{
		Success: true,
	}, nil
}
