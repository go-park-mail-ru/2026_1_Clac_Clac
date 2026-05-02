package config

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultCSRFConfig(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		want := CSRF{
			TTL:                            defaultTTL,
			Secret:                         defaultCSRFSecret,
			ExpireTimeConvertationBase:     defaultCSRFTokenExpireTimeConvertationBase,
			ExpireTimeConvertationTypeSize: defaultCSRFTokenExpireTimeConvertationTypeSize,
			PartsCount:                     defaultPartsCount,
		}

		actual := DefaultCSRFConfig()
		assert.Equal(t, want, actual)
	})

	t.Run("default TTL is 24h", func(t *testing.T) {
		conf := DefaultCSRFConfig()
		assert.Equal(t, 24*time.Hour, conf.TTL)
	})

	t.Run("default conversion base is 10", func(t *testing.T) {
		conf := DefaultCSRFConfig()
		assert.Equal(t, 10, conf.ExpireTimeConvertationBase)
	})

	t.Run("default parts count is 2", func(t *testing.T) {
		conf := DefaultCSRFConfig()
		assert.Equal(t, 2, conf.PartsCount)
	})
}

func TestSetupEnvCSRFConfig(t *testing.T) {
	t.Run("secret from env", func(t *testing.T) {
		secret := "super-secret-key-32-bytes-long!!"
		t.Setenv("CSRF_SECRET", secret)

		v := viper.New()

		SetupEnvCSRFConfig(v)

		var conf struct {
			CSRF CSRF `mapstructure:"csrf"`
		}

		err := v.Unmarshal(&conf)
		require.NoError(t, err, "viper.Unmarshal must not return error")
		assert.Equal(t, secret, conf.CSRF.Secret)
	})

	t.Run("default secret is empty without env", func(t *testing.T) {
		v := viper.New()
		SetupEnvCSRFConfig(v)

		assert.Equal(t, defaultCSRFSecret, v.GetString("csrf.secret"))
	})
}
