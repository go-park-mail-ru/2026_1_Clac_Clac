package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
