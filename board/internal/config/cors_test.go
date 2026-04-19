package config_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCORSConfig(t *testing.T) {
	var conf struct {
		CORS config.CORS `mapstructure:"cors"`
	}

	t.Run("reading", func(t *testing.T) {
		want := config.CORS{
			Credentials: "true",
			Origin:      "lsdhflsdhf",
			Methods:     "lskdkfdsjf",
			Headers:     "lksdfklje",
			MaxAge:      "739397",
		}

		var yamlTest = []byte(`
cors:
  credentials: "true"
  methods: "lskdkfdsjf"
  headers: "lksdfklje"
  max_age: 739397
`)

		t.Setenv("CORS_ORIGIN", want.Origin)

		v := viper.New()

		v.SetConfigType("yaml")
		err := v.ReadConfig(bytes.NewBuffer(yamlTest))
		require.NoError(t, err, "viper must not return error when reading yaml")

		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()

		config.SetupEnvCORS(v)

		err = v.Unmarshal(&conf)
		require.NoError(t, err, "viper must not return error when marshal")
		assert.Equal(t, want, conf.CORS)
	})
}
