package config

import "github.com/spf13/viper"

const (
	cardConfigDefaultMaxLenTitle       = 128
	cardConfigDefaultMaxLenDescription = 500
)

type CardHandler struct {
	MaxLenTitle       int `mapstructure:"max_len_title"`
	MaxLenDescription int `mapstructure:"max_len_description"`
}

type CardClient struct {
	ClientConfig `mapstructure:",squash"`
}

type Card struct {
	Client  CardClient  `mapstructure:"client"`
	Handler CardHandler `mapstructure:"handler"`
}

func DefaultCardConfig() Card {
	return Card{
		Client: CardClient{
			ClientConfig: DefaultClientConfig(),
		},
		Handler: CardHandler{
			MaxLenTitle:       cardConfigDefaultMaxLenTitle,
			MaxLenDescription: cardConfigDefaultMaxLenDescription,
		},
	}
}

func SetupEnvCard(v *viper.Viper) {
	v.SetDefault("services.card.handler.max_len_title", cardConfigDefaultMaxLenTitle)
	v.SetDefault("services.card.handler.max_len_description", cardConfigDefaultMaxLenDescription)
	v.RegisterAlias("services.card.handler.max_len_title", "services_card_handler_max_len_title")
	v.RegisterAlias("services.card.handler.max_len_description", "services_card_handler_max_len_description")
}
