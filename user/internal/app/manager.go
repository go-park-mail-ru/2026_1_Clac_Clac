package app

import (
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/s3"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/config"
	userService "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/user/service"
)

type Manager struct {
	User *userService.Service
}

func NewManager(s *Store, conf config.Config) *Manager {
	serviceCfg := userService.Config{
		BaseURLAvatar: s3.GetURL(conf.S3.Endpoint, conf.S3.AvatarsBucket),
	}

	tools := userService.Tools{
		Hasher:            userService.HashPassword,
		Checker:           userService.CheckPassword,
		GenerateAvatarKey: userService.GenerateAvatarKey,
	}

	return &Manager{
		User: userService.NewService(s.User, serviceCfg, tools),
	}
}
