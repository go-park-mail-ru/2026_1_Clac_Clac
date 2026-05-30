package app

import (
	"net/http"
	"time"

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
		ClientID:     conf.VkOAuth.AppID,
		ClientSecret: conf.VkOAuth.AppKey,
		RedirectURI:  conf.VkOAuth.RedirectURL,
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	return &Delivery{
		User: user.NewHandler(m.User, userConfig, httpClient),
	}
}

func (d *Delivery) Register(grpcServer *grpc.Server) {
	pb.RegisterUserServiceServer(grpcServer, d.User)
}
