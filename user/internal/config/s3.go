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

	v.RegisterAlias("s3.region", "s3_region")
	v.RegisterAlias("s3.endpoint", "s3_endpoint")
	v.RegisterAlias("s3.access_key", "s3_access_key")
	v.RegisterAlias("s3.secret_key", "s3_secret_key")
	v.RegisterAlias("s3.avatars_bucket", "s3_avatars_bucket")
	v.RegisterAlias("s3.avatars_prefix", "s3_avatars_prefix")
	v.RegisterAlias("s3.connect_timeout", "s3_connect_timeout")
}
