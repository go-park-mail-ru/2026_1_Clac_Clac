package config

type AppealHandler struct {
	MultipartAttachmentFileKey string `json:"multipart_attachment_file_key"`
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
		},
		Client: ClientAppeal{
			ClientConfig: DefaultClientConfig(),
		},
	}
}
