package config

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestS3AvatarsConfig(t *testing.T) {
	t.Run("test reading from env", func(t *testing.T) {
		want := S3Avatars{
			ConnectTimeout:  "60",
			Bucket:          "avatars-bucket",
			Prefix:          "avatars/",
			Region:          "ru-msk",
			Endpoint:        "vk.ru",
			AccessKey:       "my-access-key",
			SecretKey:       "my-secret-key",
			ValidExtensions: defaultVaildExtensions,
		}

		t.Setenv("S3_AVATARS_CONNECT_TIMEOUT", want.ConnectTimeout)
		t.Setenv("S3_AVATARS_BUCKET", want.Bucket)
		t.Setenv("S3_AVATARS_PREFIX", want.Prefix)
		t.Setenv("S3_AVATARS_REGION", want.Region)
		t.Setenv("S3_AVATARS_ENDPOINT", want.Endpoint)
		t.Setenv("S3_AVATARS_ACCESS_KEY", want.AccessKey)
		t.Setenv("S3_AVATARS_SECRET_KEY", want.SecretKey)

		v := viper.New()
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()

		SetupEnvS3Avatars(v)

		var conf struct {
			S3Avatars S3Avatars `mapstructure:"s3_avatars"`
		}
		conf.S3Avatars = DefaultS3AvatarsConfig()

		err := v.Unmarshal(&conf)

		require.NoError(t, err, "viper must not return error")
		assert.Equal(t, want, conf.S3Avatars)
	})

	t.Run("default config", func(t *testing.T) {
		actual := DefaultS3AvatarsConfig()
		assert.Equal(t, defaultVaildExtensions, actual.ValidExtensions)
	})
}
