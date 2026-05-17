package config

const (
	cardConfigDefaultMaxLenTitle       = 128
	cardConfigDefaultMaxLenDescription = 500
	cardConfigDefaultAttachmentFileKey = "attachment"
)

type CardHandler struct {
	MaxLenTitle       int `mapstructure:"max_len_title"`
	MaxLenDescription int `mapstructure:"max_len_description"`

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
			MultipartAttachmentFileKey: cardConfigDefaultAttachmentFileKey,
		},
	}
}
