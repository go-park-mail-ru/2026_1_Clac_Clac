package config

const (
	defaultMaxLenTitle       = 128
	defaultMaxLenDescription = 500
)

type CardHandler struct {
	MaxLenTitle       int `mapstructure:"max_len_title"`
	MaxLenDescription int `mapstructure:"max_len_description"`
}

type Card struct {
	Handler CardHandler `mapstructure:"handler"`
}

func DefaultCardConfig() Card {
	return Card{
		Handler: CardHandler{
			MaxLenTitle:       defaultMaxLenTitle,
			MaxLenDescription: defaultMaxLenDescription,
		},
	}
}
