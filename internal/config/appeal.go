package config

const (
	defaultMultipartAttachmentFileKey = "attachment"
	defaultMaxAttachmentSize          = 10 * 1024 * 1024 // 10 МБайт
)

type AppealHandler struct {
	MultipartAttachmentFileKey string `mapstructure:"multipart_attachment_file_key"`
	MaxAttachmentSize          int64  `mapstructure:"max_attachment_size"`
}

type Appeal struct {
	Handler AppealHandler `mapstructure:"handler"`
}

func DefaultAppealConfig() Appeal {
	return Appeal{
		Handler: AppealHandler{
			MultipartAttachmentFileKey: defaultMultipartAttachmentFileKey,
			MaxAttachmentSize:          defaultMaxAttachmentSize,
		},
	}
}
