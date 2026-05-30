package config

const (
	cardConfigDefaultMaxLenTitle              = 128
	cardConfigDefaultMaxLenDescription        = 500
	cardConfigDefaultAttachmentFileKey        = "attachment"
	cardConfigDefaultMaxLenComment            = 2000
	cardConfigDefaultMaxLenSubtaskDescription = 500
	cardConfigDefaultMinPoints                = 1
	cardConfigDefaultMaxPoints                = 21
)

type CardHandler struct {
	MaxLenTitle              int `mapstructure:"max_len_title"`
	MaxLenDescription        int `mapstructure:"max_len_description"`
	MaxLenComment            int `mapstructure:"max_len_comment"`
	MaxLenSubtaskDescription int `mapstructure:"max_len_subtask_description"`
	MinPoints                int `mapstructure:"min_points"`
	MaxPoints                int `mapstructure:"max_points"`

	MultipartAttachmentFileKey string `mapstructure:"multipart_attachment_file_key"`
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
			MaxLenTitle:                cardConfigDefaultMaxLenTitle,
			MaxLenDescription:          cardConfigDefaultMaxLenDescription,
			MaxLenComment:              cardConfigDefaultMaxLenComment,
			MaxLenSubtaskDescription:   cardConfigDefaultMaxLenSubtaskDescription,
			MinPoints:                  cardConfigDefaultMinPoints,
			MaxPoints:                  cardConfigDefaultMaxPoints,
			MultipartAttachmentFileKey: cardConfigDefaultAttachmentFileKey,
		},
	}
}
