package config_test

import (
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestS3BoardsConfig(t *testing.T) {
	t.Run("test reading from env", func(t *testing.T) {
		want := config.S3Boards{
			ConnectTimeout: "90",
			Bucket:         "klsdfskljf",
			Prefix:         "slkdfnmls",
			Region:         "lksdjfklsdj",
			Endpoint:       "https://localhost",
			AccessKey:      "lksdksfj",
			SecretKey:      "lksjdnleun",
		}

		t.Setenv("S3_BOARDS_CONNECT_TIMEOUT", want.ConnectTimeout)
		t.Setenv("S3_BOARDS_BUCKET", want.Bucket)
		t.Setenv("S3_BOARDS_PREFIX", want.Prefix)
		t.Setenv("S3_BOARDS_REGION", want.Region)
		t.Setenv("S3_BOARDS_ENDPOINT", want.Endpoint)
		t.Setenv("S3_BOARDS_ACCESS_KEY", want.AccessKey)
		t.Setenv("S3_BOARDS_SECRET_KEY", want.SecretKey)

		v := viper.New()

		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()

		config.SetupEnvS3Boards(v)

		var conf struct {
			S3Boards config.S3Boards `mapstructure:"s3_boards"`
		}

		err := v.Unmarshal(&conf)
		require.NoError(t, err, "viper must not return error")
		assert.Equal(t, want, conf.S3Boards)
	})
}
