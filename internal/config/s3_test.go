package config_test

import (
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestS3Config(t *testing.T) {
	t.Run("test reading from env", func(t *testing.T) {
		want := config.S3{
			Region:                  "ru-msk",
			Endpoint:                "vk.ru",
			AccessKey:               "my-access-key",
			SecretKey:               "my-secret-key",
			AvatarsBucket:           "avatars-bucket",
			AvatarsPrefix:           "avatars/",
			BoardsBackgroundsBucket: "backgrounds-bucket",
			BoardsBackgroundsPrefix: "backgrounds/",
			ConnectTimeout:          "60",
		}

		t.Setenv("S3_REGION", want.Region)
		t.Setenv("S3_ENDPOINT", want.Endpoint)
		t.Setenv("S3_ACCESS_KEY", want.AccessKey)
		t.Setenv("S3_SECRET_KEY", want.SecretKey)
		t.Setenv("S3_AVATARS_BUCKET", want.AvatarsBucket)
		t.Setenv("S3_AVATARS_PREFIX", want.AvatarsPrefix)
		t.Setenv("S3_BOARDS_BACKGROUNDS_BUCKET", want.BoardsBackgroundsBucket)
		t.Setenv("S3_BOARDS_BACKGROUNDS_PREFIX", want.BoardsBackgroundsPrefix)
		t.Setenv("S3_CONNECT_TIMEOUT", want.ConnectTimeout)

		v := viper.New()

		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()

		config.SetupEnvS3(v)

		var conf struct {
			S3 config.S3 `mapstructure:"s3"`
		}

		err := v.Unmarshal(&conf)

		require.NoError(t, err, "viper must not return error")
		assert.Equal(t, want, conf.S3)
	})

	t.Run("test default values", func(t *testing.T) {
		v := viper.New()
		config.SetupEnvS3(v)

		var conf struct {
			S3 config.S3 `mapstructure:"s3"`
		}

		err := v.Unmarshal(&conf)
		require.NoError(t, err)

		assert.Equal(t, "", conf.S3.Region)
		assert.Equal(t, "", conf.S3.ConnectTimeout)
	})
}
