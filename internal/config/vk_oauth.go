package config

import "github.com/spf13/viper"

const (
	vkOAuthDefaultValue = ""
)

type VkOAuth struct {
	AppID       string `mapstructure:"app_id"`
	AppKey      string `mapstructure:"app_key"`
	AppSecret   string `mapstructure:"app_secret"`
	RedirectURL string `mapstructure:"redirect_url"`
}

func DefaultVkOAuthConfig() VkOAuth {
	return VkOAuth{
		AppID:       vkOAuthDefaultValue,
		AppKey:      vkOAuthDefaultValue,
		AppSecret:   vkOAuthDefaultValue,
		RedirectURL: vkOAuthDefaultValue,
	}
}

func SetDefaultEnvVkOAuth(v *viper.Viper) {
	v.SetDefault("vk_oauth.app_id", vkOAuthDefaultValue)
	v.SetDefault("vk_oauth.app_key", vkOAuthDefaultValue)
	v.SetDefault("vk_oauth.app_secret", vkOAuthDefaultValue)
	v.SetDefault("vk_oauth.redirect_url", vkOAuthDefaultValue)
}
