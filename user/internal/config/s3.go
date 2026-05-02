package config

import (
	"github.com/spf13/viper"
)

const (
	defaultS3Value = ""
)

type S3 struct {
	Region         string `mapstructure:"region"`
	Endpoint       string `mapstructure:"endpoint"`
	AccessKey      string `mapstructure:"access_key"`
	SecretKey      string `mapstructure:"secret_key"`
	AvatarsBucket  string `mapstructure:"avatars_bucket"`
	AvatarsPrefix  string `mapstructure:"avatars_prefix"`
	ConnectTimeout string `mapstructure:"connect_timeout"`
}

func DefaultS3Config() S3 {
	return S3{}
}

func SetupEnvS3(v *viper.Viper) {
	v.SetDefault("s3.region", defaultS3Value)
	v.SetDefault("s3.endpoint", defaultS3Value)
	v.SetDefault("s3.access_key", defaultS3Value)
	v.SetDefault("s3.secret_key", defaultS3Value)
	v.SetDefault("s3.avatars_bucket", defaultS3Value)
	v.SetDefault("s3.avatars_prefix", defaultS3Value)
	v.SetDefault("s3.connect_timeout", defaultS3Value)

	v.BindEnv("s3.region", "S3_REGION")
	v.BindEnv("s3.endpoint", "S3_ENDPOINT")
	v.BindEnv("s3.access_key", "S3_ACCESS_KEY")
	v.BindEnv("s3.secret_key", "S3_SECRET_KEY")
	v.BindEnv("s3.avatars_bucket", "S3_AVATARS_BUCKET")
	v.BindEnv("s3.avatars_prefix", "S3_AVATARS_PREFIX")
	v.BindEnv("s3.connect_timeout", "S3_CONNECT_TIMEOUT")
}
