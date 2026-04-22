package app

import (
	pbAuth "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/auth"
	pbProfile "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/profile"
	auth "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/auth/handler"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/config"
	profile "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/profile/handler"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
)

type Delivery struct {
	Auth    *auth.Handler
	Profile *profile.Handler
}

func NewDelivery(m *Manager, conf *config.Config, vkOAuth *oauth2.Config) *Delivery {
	authConfig := auth.Config{
		MaxLenPassword:  conf.Auth.Handler.MaxLenPassword,
		MinLenPassword:  conf.Auth.Handler.MinLenPassword,
		SessionLifetime: conf.Auth.Handler.SessionLifetime,

		APIMethod: conf.VkOAuth.APIMethod,
	}

	profileConfig := profile.Config{
		ValidExtensions: conf.S3Avatars.ValidExtensions,

		SiganatureTypeBytes:   conf.Profile.Handler.SiganatureTypeBytes,
		MaxLenNameUser:        conf.Profile.Handler.MaxLenNameUser,
		MaxLenDescriptionUser: conf.Profile.Handler.MaxLenDescriptionUser,
		MaxReadBytes:          conf.Profile.Handler.MaxReadBytes,
	}

	return &Delivery{
		Auth:    auth.NewHandler(m.Auth, authConfig, vkOAuth),
		Profile: profile.NewHandler(m.Profile, profileConfig),
	}
}

func (d *Delivery) Register(grpcServer *grpc.Server) {
	pbAuth.RegisterAuthServiceServer(grpcServer, d.Auth)
	pbProfile.RegisterProfileServiceServer(grpcServer, d.Profile)
}
