package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultAppealConfig(t *testing.T) {
	t.Run("handler multipart key is empty string", func(t *testing.T) {
		conf := DefaultAppealConfig()
		assert.Equal(t, "", conf.Handler.MultipartAttachmentFileKey)
	})

	t.Run("client config equals DefaultClientConfig", func(t *testing.T) {
		conf := DefaultAppealConfig()
		assert.Equal(t, DefaultClientConfig(), conf.Client.ClientConfig)
	})
}
