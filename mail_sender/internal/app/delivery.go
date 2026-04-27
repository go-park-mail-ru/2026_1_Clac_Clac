package app

import (
	sender "github.com/go-park-mail-ru/2026_1_Clac_Clac/mail_sender/internal/sender/handler"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/contracts/mail_sender"
	"google.golang.org/grpc"
)

type Delivery struct {
	Sender *sender.Handler
}

func NewDelivery(m *Manager) *Delivery {
	return &Delivery{
		Sender: sender.NewHandler(m.Sender),
	}
}

func (d *Delivery) Register(grpcServer *grpc.Server) {
	pb.RegisterMailSenderServiceServer(grpcServer, d.Sender)
}
