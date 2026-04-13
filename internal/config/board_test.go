package config_test

import (
	"bytes"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBoardConfig(t *testing.T) {
	t.Run("test unmarshal board config", func(t *testing.T) {
		want := config.Board{
			Handler: config.BoardHandler{
				MultipartBackgroundFileKey: "custom_bg_key",
				MaxBackgroundSize:          20 * 1024 * 1024, // 20 МБайт для теста
			},
			Repository: config.BoardRepository{
				CreateBoardDefaultUserRole: "admin",
			},
		}

		var yamlTest = []byte(`
handler:
  multipart_background_file_key: custom_bg_key
  max_background_size: 20971520
repository:
  create_board_default_user_role: admin
`)

		viper.Reset()
		viper.SetConfigType("yaml")
		err := viper.ReadConfig(bytes.NewBuffer(yamlTest))
		require.NoError(t, err, "viper.ReadConfig should not return error")

		conf := config.DefaultBoardConfig()
		err = viper.Unmarshal(&conf)

		require.NoError(t, err, "viper.Unmarshal should not return error")
		assert.Equal(t, want, conf)
	})
}
