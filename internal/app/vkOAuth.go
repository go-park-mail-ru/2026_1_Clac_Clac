package app

import (
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/vk"
)

func NewVKOAuth(conf *config.VkOAuth) *oauth2.Config {
	const emailKey = "email"
	var vkOAuthScopes = []string{emailKey}

	return &oauth2.Config{
		ClientID:     conf.AppID,
		ClientSecret: conf.AppKey,
		RedirectURL:  conf.RedirectURL,
		Scopes:       vkOAuthScopes,
		Endpoint:     vk.Endpoint,
	}
}
