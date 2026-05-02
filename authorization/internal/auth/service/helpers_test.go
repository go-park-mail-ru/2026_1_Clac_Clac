package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSessionKey(t *testing.T) {
	t.Run("success generate session key", func(t *testing.T) {
		key1, err := GenerateSessionKey()
		assert.NoError(t, err)
		assert.Equal(t, 64, len(key1))

		key2, err := GenerateSessionKey()
		assert.NoError(t, err)
		assert.NotEqual(t, key1, key2)
	})
}

func TestCreateSessionKey(t *testing.T) {
	t.Run("success create session key", func(t *testing.T) {
		actual := CreateSessionKey("abc123")
		assert.Equal(t, "session:abc123", actual)
	})

	t.Run("empty session id", func(t *testing.T) {
		actual := CreateSessionKey("")
		assert.Equal(t, "session:", actual)
	})
}
