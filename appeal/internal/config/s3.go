package config

type S3 struct {
	Region                 string `mapstructure:"region"`
	Endpoint               string `mapstructure:"endpoint"`
	AccessKey              string `mapstructure:"access_key"`
	SecretKey              string `mapstructure:"secret_key"`
	AppealAttachmentBucket string `mapstructure:"appeal_attachment_bucket"`
	AppealAttachmentPrefix string `mapstructure:"appeal_attachment_prefix"`
	ConnectTimeout         string `mapstructure:"connect_timeout"`
}
