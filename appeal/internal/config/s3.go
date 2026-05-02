package config

import "github.com/spf13/viper"

type S3 struct {
	Region                 string `mapstructure:"region"`
	Endpoint               string `mapstructure:"endpoint"`
	AccessKey              string `mapstructure:"access_key"`
	SecretKey              string `mapstructure:"secret_key"`
	AppealAttachmentBucket string `mapstructure:"appeal_attachment_bucket"`
	AppealAttachmentPrefix string `mapstructure:"appeal_attachment_prefix"`
	ConnectTimeout         string `mapstructure:"connect_timeout"`
}

func SetupEnvS3(v *viper.Viper) {
	v.SetDefault("s3.region", "")
	v.SetDefault("s3.endpoint", "")
	v.SetDefault("s3.access_key", "")
	v.SetDefault("s3.secret_key", "")
	v.SetDefault("s3.appeal_attachment_bucket", "")
	v.SetDefault("s3.appeal_attachment_prefix", "")
	v.SetDefault("s3.connect_timeout", "")

	v.RegisterAlias("s3.region", "s3_region")
	v.RegisterAlias("s3.endpoint", "s3_endpoint")
	v.RegisterAlias("s3.access_key", "s3_access_key")
	v.RegisterAlias("s3.secret_key", "s3_secret_key")
	v.RegisterAlias("s3.appeal_attachment_bucket", "s3_appeal_attachment_bucket")
	v.RegisterAlias("s3.appeal_attachment_prefix", "s3_appeal_attachment_prefix")
	v.RegisterAlias("s3.connect_timeout", "s3_connect_timeout")
}
