package config

import (
	"github.com/spf13/viper"
)

const (
	defaultS3BoardsValue = ""
)

type S3Boards struct {
	ConnectTimeout string `mapstructure:"connect_timeout"`
	Bucket         string `mapstructure:"bucket"`
	Prefix         string `mapstructure:"prefix"`
	Region         string `mapstructure:"region"`
	Endpoint       string `mapstructure:"endpoint"`
	AccessKey      string `mapstructure:"access_key"`
	SecretKey      string `mapstructure:"secret_key"`
}

func DefaultS3BoardsConfig() S3Boards {
	return S3Boards{
		ConnectTimeout: defaultS3BoardsValue,
		Bucket:         defaultS3BoardsValue,
		Prefix:         defaultS3BoardsValue,
		Region:         defaultS3BoardsValue,
		Endpoint:       defaultS3BoardsValue,
		AccessKey:      defaultS3BoardsValue,
		SecretKey:      defaultS3BoardsValue,
	}
}

func SetupEnvS3Boards(v *viper.Viper) {
	v.SetDefault("s3_boards.connect_timeout", defaultS3BoardsValue)
	v.SetDefault("s3_boards.bucket", defaultS3BoardsValue)
	v.SetDefault("s3_boards.prefix", defaultS3BoardsValue)
	v.SetDefault("s3_boards.region", defaultS3BoardsValue)
	v.SetDefault("s3_boards.endpoint", defaultS3BoardsValue)
	v.SetDefault("s3_boards.access_key", defaultS3BoardsValue)
	v.SetDefault("s3_boards.secret_key", defaultS3BoardsValue)

	v.SetDefault("s3_boards.connect_timeout", "s3_boards_connect_timeout")
	v.RegisterAlias("s3_boards.bucket", "s3_boards_bucket")
	v.RegisterAlias("s3_boards.prefix", "s3_boards_prefix")
	v.RegisterAlias("s3_boards.region", "s3_boards_region")
	v.RegisterAlias("s3_boards.endpoint", "s3_boards_endpoint")
	v.RegisterAlias("s3_boards.access_key", "s3_boards_access_key")
	v.RegisterAlias("s3_boards.secret_key", "s3_boards_secret_key")
}
