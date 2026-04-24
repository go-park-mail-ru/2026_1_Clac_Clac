package app

import (
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/contracts/user"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/config"
	user "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/user/handler"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
)

type Delivery struct {
	User *user.Handler
}

func NewDelivery(m *Manager, conf *config.Config, vkOAuth *oauth2.Config) *Delivery {
	userConfig := user.Config{
		MaxLenPassword: conf.User.Handler.MaxLenPassword,
		MinLenPassword: conf.User.Handler.MinLenPassword,

		ValidExtensions:       conf.S3Avatars.ValidExtensions,
		SignatureTypeBytes:    conf.User.Handler.SiganatureTypeBytes,
		MaxLenNameUser:        conf.User.Handler.MaxLenNameUser,
		MaxLenDescriptionUser: conf.User.Handler.MaxLenDescriptionUser,
		MaxReadBytes:          conf.User.Handler.MaxReadBytes,

		APIMethod: conf.VkOAuth.APIMethod,
	}

	return &Delivery{
		User: user.NewHandler(m.User, userConfig, vkOAuth),
	}
}

func (d *Delivery) Register(grpcServer *grpc.Server) {
	pb.RegisterUserServiceServer(grpcServer, d.User)
}
