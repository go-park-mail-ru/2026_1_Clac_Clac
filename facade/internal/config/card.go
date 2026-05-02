package config

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
	Handler CardHandler `mapstructure:"handler"`
	Client  CardClient  `mapstructure:"client"`
}

func DefaultCardConfig() Card {
	return Card{
		Handler: CardHandler{
			MaxLenTitle:       cardConfigDefaultMaxLenTitle,
			MaxLenDescription: cardConfigDefaultMaxLenDescription,
		},
		Client: CardClient{
			ClientConfig: DefaultClientConfig(),
		},
	}
}
