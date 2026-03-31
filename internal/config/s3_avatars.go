package config

import (
	"github.com/spf13/viper"
)

const (
	defaultS3AvatarsValue = ""
)

type S3Avatars struct {
	Bucket    string `mapstructure:"bucket"`
	Region    string `mapstructure:"region"`
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
}

func DefaultS3AvatarsConfig() S3Avatars {
	return S3Avatars{
		Bucket:    defaultS3AvatarsValue,
		Region:    defaultS3AvatarsValue,
		Endpoint:  defaultS3AvatarsValue,
		AccessKey: defaultS3AvatarsValue,
		SecretKey: defaultS3AvatarsValue,
	}
}

func SetupEnvS3Avatars(v *viper.Viper) {
	v.SetDefault("s3_avatars.bucket", defaultS3AvatarsValue)
	v.SetDefault("s3_avatars.region", defaultS3AvatarsValue)
	v.SetDefault("s3_avatars.endpoint", defaultS3AvatarsValue)
	v.SetDefault("s3_avatars.access_key", defaultS3AvatarsValue)
	v.SetDefault("s3_avatars.secret_key", defaultS3AvatarsValue)

	v.RegisterAlias("s3_avatars.bucket", "s3_avatars_bucket")
	v.RegisterAlias("s3_avatars.region", "s3_avatars_region")
	v.RegisterAlias("s3_avatars.endpoint", "s3_avatars_endpoint")
	v.RegisterAlias("s3_avatars.access_key", "s3_avatars_access_key")
	v.RegisterAlias("s3_avatars.secret_key", "s3_avatars_secret_key")
}
