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
	APIMethod   string `mapstructure:"api_method"`
}

func DefaultVkOAuthConfig() VkOAuth {
	return VkOAuth{
		AppID:       vkOAuthDefaultValue,
		AppKey:      vkOAuthDefaultValue,
		AppSecret:   vkOAuthDefaultValue,
		RedirectURL: vkOAuthDefaultValue,
		APIMethod:   vkOAuthDefaultValue,
	}
}

func SetupEnvVkOAuth(v *viper.Viper) {
	// Надо для того, чтобы viper мог считать переменные из окружения
	v.SetDefault("vk_oauth.app_id", vkOAuthDefaultValue)
	v.SetDefault("vk_oauth.app_key", vkOAuthDefaultValue)
	v.SetDefault("vk_oauth.app_secret", vkOAuthDefaultValue)
	v.SetDefault("vk_oauth.redirect_url", vkOAuthDefaultValue)
	v.SetDefault("vk_oauth.api_method", vkOAuthDefaultValue)
	// Надо, чтобы viper мог считать .env файл
	v.RegisterAlias("vk_oauth.app_id", "vk_oauth_app_id")
	v.RegisterAlias("vk_oauth.app_key", "vk_oauth_app_key")
	v.RegisterAlias("vk_oauth.app_secret", "vk_oauth_app_secret")
	v.RegisterAlias("vk_oauth.redirect_url", "vk_oauth_redirect_url")
	v.RegisterAlias("vk_oauth.api_method", "vk_oauth_api_method")
}
