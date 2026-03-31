package config_test

import (
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestS3AvatarsConfig(t *testing.T) {
	t.Run("test reading from env", func(t *testing.T) {
		want := config.S3Avatars{
			Bucket:    "klsdfskljf",
			Region:    "lksdjfklsdj",
			Endpoint:  "https://localhost",
			AccessKey: "lksdksfj",
			SecretKey: "lksjdnleun",
		}

		t.Setenv("S3_AVATARS_BUCKET", want.Bucket)
		t.Setenv("S3_AVATARS_REGION", want.Region)
		t.Setenv("S3_AVATARS_ENDPOINT", want.Endpoint)
		t.Setenv("S3_AVATARS_ACCESS_KEY", want.AccessKey)
		t.Setenv("S3_AVATARS_SECRET_KEY", want.SecretKey)

		v := viper.New()

		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()

		config.SetupEnvS3Avatars(v)

		var conf struct {
			S3Avatars config.S3Avatars `mapstructure:"s3_avatars"`
		}

		err := v.Unmarshal(&conf)
		require.NoError(t, err, "viper must not return error")
		assert.Equal(t, want, conf.S3Avatars)
	})
}
