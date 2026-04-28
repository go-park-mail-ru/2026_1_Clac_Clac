package app

import (
	"net/http"

	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/user/v1"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/config"
	user "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/user/handler"
	"google.golang.org/grpc"
)

type Delivery struct {
	User *user.Handler
}

func NewDelivery(m *Manager, conf *config.Config) *Delivery {
	userConfig := user.Config{
		APIMethod: conf.VkOAuth.APIMethod,
	}

	return &Delivery{
		User: user.NewHandler(m.User, userConfig, http.DefaultClient),
	}
}

func (d *Delivery) Register(grpcServer *grpc.Server) {
	pb.RegisterUserServiceServer(grpcServer, d.User)
}
