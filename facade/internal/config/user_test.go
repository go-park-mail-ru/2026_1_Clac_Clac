package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultUserConfig(t *testing.T) {
	t.Run("default handler values", func(t *testing.T) {
		conf := DefaultUserConfig()

		assert.Equal(t, profileConfigDefaultSignatureBytes, conf.Handler.SignatureTypeBytes)
		assert.Equal(t, int64(profileConfigDefaultMaxReadBytes), conf.Handler.MaxReadBytes)
		assert.Equal(t, profileConfigDefaultMaxLenNameUser, conf.Handler.MaxLenNameUser)
		assert.Equal(t, profileConfigDefaultMaxLenDescriptionUser, conf.Handler.MaxLenDescriptionUser)
		assert.Equal(t, profileConfigDefaultMaxLenPassword, conf.Handler.MaxLenPassword)
		assert.Equal(t, profileConfigDefaultMinLenPassword, conf.Handler.MinLenPassword)
	})

	t.Run("default client equals DefaultClientConfig", func(t *testing.T) {
		conf := DefaultUserConfig()
		assert.Equal(t, DefaultClientConfig(), conf.Client.ClientConfig)
	})

	t.Run("valid extensions are populated", func(t *testing.T) {
		conf := DefaultUserConfig()
		require.NotEmpty(t, conf.Handler.ValidExtensions)
		assert.Contains(t, conf.Handler.ValidExtensions, "image/png")
		assert.Contains(t, conf.Handler.ValidExtensions, "image/jpeg")
		assert.Contains(t, conf.Handler.ValidExtensions, "image/jpg")
		assert.Contains(t, conf.Handler.ValidExtensions, "image/webp")
	})
}

func TestDefaultValidExtensions(t *testing.T) {
	t.Run("contains all allowed types", func(t *testing.T) {
		exts := DefaultValidExtensions()

		assert.Contains(t, exts, "image/png")
		assert.Contains(t, exts, "image/jpeg")
		assert.Contains(t, exts, "image/jpg")
		assert.Contains(t, exts, "image/webp")
	})

	t.Run("does not contain disallowed types", func(t *testing.T) {
		exts := DefaultValidExtensions()

		assert.NotContains(t, exts, "image/gif")
		assert.NotContains(t, exts, "image/bmp")
		assert.NotContains(t, exts, "application/pdf")
	})

	t.Run("struct values are empty (set membership)", func(t *testing.T) {
		exts := DefaultValidExtensions()
		for _, v := range exts {
			assert.Equal(t, struct{}{}, v, "map value must be empty struct")
		}
	})
}
