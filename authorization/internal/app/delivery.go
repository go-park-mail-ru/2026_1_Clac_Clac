package app

import (
	auth "github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/auth/handler"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/config"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/auth/v1"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
)

type Delivery struct {
	Auth *auth.Handler
}

func NewDelivery(m *Manager, conf *config.Config, vkOAuth *oauth2.Config) *Delivery {
	return &Delivery{
		Auth: auth.NewHandler(m.Auth, vkOAuth),
	}
}

func (d *Delivery) Register(grpcServer *grpc.Server) {
	pb.RegisterAuthServiceServer(grpcServer, d.Auth)
}
