package config

const (
	defaultMaxLenDescription = 500
)

type AppealHandler struct {
	MultipartAttachmentFileKey string `mapstructure:"multipart_attachment_file_key"`
	MaxLenDisplayName          int    `mapstructure:"max_len_display_name"`
	MaxLenDescription          int    `mapstructure:"max_len_description"`
}

type ClientAppeal struct {
	ClientConfig `mapstructure:",squash"`
}

type Appeal struct {
	Handler AppealHandler `mapstructure:"handler"`
	Client  ClientAppeal  `mapstructure:"client"`
}

func DefaultAppealConfig() Appeal {
	return Appeal{
		Handler: AppealHandler{
			MultipartAttachmentFileKey: "",
			MaxLenDisplayName:          defaultMaxDisplayName,
			MaxLenDescription:          defaultMaxLenDescription,
		},
		Client: ClientAppeal{
			ClientConfig: DefaultClientConfig(),
		},
	}
}
