package config_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

<<<<<<<< HEAD:board/internal/config/config_test.go
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/grpc_engine"
========
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/config"
>>>>>>>> feat/add-facade:monolith/internal/config/config_test.go
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	DebugLevel = "debug"
	InfoLevel  = "info"
)

func TestConfigReading(t *testing.T) {
	expectedConfig := config.Config{
		App: config.DefaultApplicationConfig(),
		Engine: grpc_engine.Config{
			Addr:                    ":8080",
			WriteTimeout:            30,
			ReadTimeout:             30,
			IdleTimeout:             90,
			GracefulShutdownTimeout: 25,
		},
		S3:    config.S3{},
		Board: config.DefaultBoardConfig(),
		Section: config.Section{
			Handler: config.SectionHandler{
				MaxQuantityTasks:  100,
				MinQuantityTasks:  0,
				MaxLenNameSection: 128,
			},
		},

		Card: config.Card{
			Handler: config.CardHandler{
				MaxLenTitle:       128,
				MaxLenDescription: 500,
			},
		},
	}

	var yamlTest = []byte(`
app:
  log_level: debug
  max_text_request_size: 10240
  max_upload_image_size: 10485760


engine:
  addr: ":8080"
  write_timeout: 30
  read_timeout: 30
  idle_timeout: 90
  graceful_shutdown_timeout: 25
`)

	v := viper.New()
	v.SetConfigType("yaml")
	err := v.ReadConfig(bytes.NewBuffer(yamlTest))

	require.NoError(t, err, "reading should not returt error")

	conf := config.DefaultConfig()
	err = v.Unmarshal(&conf)

	require.NoError(t, err, "error while reading config")
	assert.Equal(t, expectedConfig, conf)
}

func TestSetupViper(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		const configFilename = "config.yaml"
		const envFilename = ".env"

		var yamlTest = []byte(`
app:
  log_level: debug

engine:
  addr: ":8080"
  write_timeout: 30
  read_timeout: 30
  idle_timeout: 90
  graceful_shutdown_timeout: 25
`)
		var envTest = []byte(`MAIL_SENDER_HOST=test.ru`)

		tempDir := t.TempDir()
		os.WriteFile(filepath.Join(tempDir, configFilename), yamlTest, 0644)
		os.WriteFile(filepath.Join(tempDir, envFilename), envTest, 0644)

		v, err := config.SetupViper(tempDir)

		require.NoError(t, err, "must not return error")

		conf := config.DefaultConfig()
		err = v.Unmarshal(&conf)

		require.NoError(t, err, "error while reading config")
	})

	t.Run("error", func(t *testing.T) {
		const currentDir = "."
		_, err := config.SetupViper(currentDir)

		require.Error(t, err, "must return error")
	})
}
