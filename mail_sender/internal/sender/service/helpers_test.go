package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratorCode(t *testing.T) {
	t.Run("success generate code", func(t *testing.T) {
		code1, err := GeneratorCode()
		assert.NoError(t, err)
		assert.Equal(t, 6, len(code1), "code must be 6 digits")

		code2, err := GeneratorCode()
		assert.NoError(t, err)
		assert.NotEqual(t, code1, code2, "codes should differ")
	})
}

func TestCreatorResetKey(t *testing.T) {
	t.Run("success create reset key", func(t *testing.T) {
		actual := CreatorResetKey("123456")
		assert.Equal(t, "reset_token:123456", actual)
	})

	t.Run("empty token id", func(t *testing.T) {
		actual := CreatorResetKey("")
		assert.Equal(t, "reset_token:", actual)
	})
}
