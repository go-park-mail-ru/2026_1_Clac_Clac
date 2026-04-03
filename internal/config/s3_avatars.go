package config

import (
	"github.com/spf13/viper"
)

const (
	defaultS3AvatarsValue = ""
)

var defaultVaildExtensions = map[string]struct{}{
	"image/jpg":  {},
	"image/png":  {},
	"image/webp": {},
}

type S3Avatars struct {
	ConnectTimeout string `mapstructure:"connect_timeout"`
	Bucket         string `mapstructure:"bucket"`
	Prefix         string `mapstructure:"prefix"`
	Region         string `mapstructure:"region"`
	Endpoint       string `mapstructure:"endpoint"`
	AccessKey      string `mapstructure:"access_key"`
	SecretKey      string `mapstructure:"secret_key"`
	CDNBaseURL     string `mapstructure:"cdn_base_url"`

	ValidExtensions map[string]struct{} `mapstructure:"valid_extensions"`
}

func DefaultS3AvatarsConfig() S3Avatars {
	return S3Avatars{
		ConnectTimeout: defaultS3AvatarsValue,
		Bucket:         defaultS3AvatarsValue,
		Prefix:         defaultS3AvatarsValue,
		Region:         defaultS3AvatarsValue,
		Endpoint:       defaultS3AvatarsValue,
		AccessKey:      defaultS3AvatarsValue,
		SecretKey:      defaultS3AvatarsValue,

		ValidExtensions: defaultVaildExtensions,
	}
}

func SetupEnvS3Avatars(v *viper.Viper) {
	v.SetDefault("s3_avatars.connect_timeout", defaultS3AvatarsValue)
	v.SetDefault("s3_avatars.bucket", defaultS3AvatarsValue)
	v.SetDefault("s3_avatars.prefix", defaultS3AvatarsValue)
	v.SetDefault("s3_avatars.region", defaultS3AvatarsValue)
	v.SetDefault("s3_avatars.endpoint", defaultS3AvatarsValue)
	v.SetDefault("s3_avatars.access_key", defaultS3AvatarsValue)
	v.SetDefault("s3_avatars.secret_key", defaultS3AvatarsValue)

	v.SetDefault("s3_avatars.connect_timeout", "s3_avatars_connect_timeout")
	v.RegisterAlias("s3_avatars.bucket", "s3_avatars_bucket")
	v.RegisterAlias("s3_avatars.prefix", "s3_avatars_prefix")
	v.RegisterAlias("s3_avatars.region", "s3_avatars_region")
	v.RegisterAlias("s3_avatars.endpoint", "s3_avatars_endpoint")
	v.RegisterAlias("s3_avatars.access_key", "s3_avatars_access_key")
	v.RegisterAlias("s3_avatars.secret_key", "s3_avatars_secret_key")
}
