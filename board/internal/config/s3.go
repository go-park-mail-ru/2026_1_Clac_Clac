package config

type S3 struct {
	Region                  string `mapstructure:"region"`
	Endpoint                string `mapstructure:"endpoint"`
	AccessKey               string `mapstructure:"access_key"`
	SecretKey               string `mapstructure:"secret_key"`
	BoardsBackgroundsBucket string `mapstructure:"boards_backgrounds_bucket"`
	BoardsBackgroundsPrefix string `mapstructure:"boards_backgrounds_prefix"`
	ConnectTimeout          string `mapstructure:"connect_timeout"`
}
