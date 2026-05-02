package app

import (
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/vk"
)

func NewVKOAuth(conf *config.VkOAuth) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     conf.AppID,
		ClientSecret: conf.AppSecret,
		RedirectURL:  conf.RedirectURL,
		Scopes:       []string{"email"},
		Endpoint:     vk.Endpoint,
	}
}
