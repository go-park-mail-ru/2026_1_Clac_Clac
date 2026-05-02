package config

import (
	"github.com/spf13/viper"
)

const (
	defaultS3AvatarsValue = ""
)

var defaultVaildExtensions = map[string]struct{}{
	"image/jpg":  {},
	"image/jpeg": {},
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

	v.BindEnv("s3_avatars.connect_timeout", "S3_AVATARS_CONNECT_TIMEOUT")
	v.BindEnv("s3_avatars.bucket", "S3_AVATARS_BUCKET")
	v.BindEnv("s3_avatars.prefix", "S3_AVATARS_PREFIX")
	v.BindEnv("s3_avatars.region", "S3_AVATARS_REGION")
	v.BindEnv("s3_avatars.endpoint", "S3_AVATARS_ENDPOINT")
	v.BindEnv("s3_avatars.access_key", "S3_AVATARS_ACCESS_KEY")
	v.BindEnv("s3_avatars.secret_key", "S3_AVATARS_SECRET_KEY")
}
