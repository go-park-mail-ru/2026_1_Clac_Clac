package config

const (
	defaultMaxLenTitle       = 128
	defaultMaxLenDescription = 500
	defaultMaxAttachments    = 100
	defaultMaxNestingDepth   = 100
)

type CardHandler struct {
	MaxLenTitle       int `mapstructure:"max_len_title"`
	MaxLenDescription int `mapstructure:"max_len_description"`
}

type CardRepository struct {
	MaxAttachments  int `mapstructure:"max_attachments"`
	MaxNestingDepth int `mapstructure:"max_nesting_depth"`
}

type Card struct {
	Handler    CardHandler    `mapstructure:"handler"`
	Repository CardRepository `mapstructure:"repository"`
}

func DefaultCardConfig() Card {
	return Card{
		Handler: CardHandler{
			MaxLenTitle:       defaultMaxLenTitle,
			MaxLenDescription: defaultMaxLenDescription,
		},
		Repository: CardRepository{
			MaxAttachments:  defaultMaxAttachments,
			MaxNestingDepth: defaultMaxNestingDepth,
		},
	}
}
