package config

import (
	"bytes"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationConfig(t *testing.T) {
	t.Run("test unmarshal", func(t *testing.T) {
		want := DefaultApplicationConfig()

		var yamlTest = []byte(`
			log_level: debug
			max_text_request_size: 10240
			max_upload_image_size: 10485760
		`)

		v := viper.New()
		v.SetConfigType("yaml")
		_ = v.ReadConfig(bytes.NewBuffer(yamlTest))

		conf := DefaultApplicationConfig()
		err := v.Unmarshal(&conf)

		require.NoError(t, err, "viper.Unmarshal should not return error")
		assert.Equal(t, want, conf)

	})
}

func TestIsDebug(t *testing.T) {
	assert.True(t, IsDebug("debug"), "this is debug level")
	assert.False(t, IsDebug("info"), "this is not debug level")
}
