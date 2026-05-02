package config

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultCardConfig_Table(t *testing.T) {
	conf := DefaultCardConfig()

	tests := []struct {
		name     string
		validate func(t *testing.T)
	}{
		{
			name: "max len title equals default constant",
			validate: func(t *testing.T) {
				assert.Equal(t, cardConfigDefaultMaxLenTitle, conf.Handler.MaxLenTitle)
				assert.Equal(t, 128, conf.Handler.MaxLenTitle)
			},
		},
		{
			name: "max len description equals default constant",
			validate: func(t *testing.T) {
				assert.Equal(t, cardConfigDefaultMaxLenDescription, conf.Handler.MaxLenDescription)
				assert.Equal(t, 500, conf.Handler.MaxLenDescription)
			},
		},
		{
			name: "default client equals DefaultClientConfig",
			validate: func(t *testing.T) {
				assert.Equal(t, DefaultClientConfig(), conf.Client.ClientConfig)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t)
		})
	}
}

func TestSetupEnvCard_Table(t *testing.T) {
	tests := []struct {
		name          string
		envVars       map[string]string
		expectedTitle int
		expectedDesc  int
	}{
		{
			name: "card max len title from env",
			envVars: map[string]string{
				"SERVICES_CARD_HANDLER_MAX_LEN_TITLE": "256",
			},
			expectedTitle: 256,
			expectedDesc:  0,
		},
		{
			name: "card max len description from env",
			envVars: map[string]string{
				"SERVICES_CARD_HANDLER_MAX_LEN_DESCRIPTION": "1024",
			},
			expectedTitle: 0,
			expectedDesc:  1024,
		},
		{
			name: "both title and description from env",
			envVars: map[string]string{
				"SERVICES_CARD_HANDLER_MAX_LEN_TITLE":       "64",
				"SERVICES_CARD_HANDLER_MAX_LEN_DESCRIPTION": "2048",
			},
			expectedTitle: 64,
			expectedDesc:  2048,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.envVars {
				t.Setenv(key, val)
			}

			v := viper.New()
			v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
			v.AutomaticEnv()

			SetupEnvCard(v)

			var conf struct {
				Services struct {
					Card Card `mapstructure:"card"`
				} `mapstructure:"services"`
			}

			err := v.Unmarshal(&conf)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedTitle, conf.Services.Card.Handler.MaxLenTitle)
			assert.Equal(t, tt.expectedDesc, conf.Services.Card.Handler.MaxLenDescription)
		})
	}
}
