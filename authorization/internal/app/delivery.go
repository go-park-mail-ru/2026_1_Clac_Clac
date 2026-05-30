package app

import (
	auth "github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/auth/handler"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/auth/v1"
	"google.golang.org/grpc"
)

type Delivery struct {
	Auth *auth.Handler
}

func NewDelivery(m *Manager) *Delivery {
	return &Delivery{
		Auth: auth.NewHandler(m.Auth),
	}
}

func (d *Delivery) Register(grpcServer *grpc.Server) {
	pb.RegisterAuthServiceServer(grpcServer, d.Auth)
}
