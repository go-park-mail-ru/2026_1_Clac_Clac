package config_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/config"
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
		Engine: config.Engine{
			Addr:                    ":8080",
			WriteTimeout:            30,
			ReadTimeout:             30,
			IdleTimeout:             90,
			GracefulShutdownTimeout: 25,
		},
		DBRateLimiters: config.DataBaseRateLimiters{
			DBActions: map[string]config.ActionRateLimiters{
				"login": {
					Limit:  5,
					Action: "login",
					Window: 1 * time.Minute,
				},
				"register": {
					Limit:  5,
					Action: "register",
					Window: 1 * time.Hour,
				},
			},
		},
		S3:    config.DefaultS3Config(),
		Board: config.DefaultBoardConfig(),
		S3Avatars: config.S3Avatars{
			ValidExtensions: map[string]struct{}{
				"image/jpg":  {},
				"image/jpeg": {},
				"image/png":  {},
				"image/webp": {},
			},
		},
		CORS: config.CORS{
			Credentials: "false",
			Origin:      "localhost",
			Methods:     "GET,POST,OPTIONS",
			Headers:     "",
			MaxAge:      "60",
		},

		Auth: config.Auth{
			Handler: config.AuthHandler{
				MaxLenPassword:  128,
				MinLenPassword:  8,
				SessionLifetime: 24 * time.Hour,
			},
			Service: config.AuthService{
				SessionLifetime: 24 * time.Hour,
				CountRetries:    3,
			},
		},

		Profile: config.Profile{
			Handler: config.ProfileHandler{
				SiganatureTypeBytes:   512,
				MaxReadBytes:          5 << 20,
				MaxLenNameUser:        128,
				MaxLenDescriptionUser: 500,
			},
		},

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
