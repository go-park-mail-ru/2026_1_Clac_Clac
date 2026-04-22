package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateAvatarKey(t *testing.T) {
	t.Run("Create avatar key", func(t *testing.T) {
		_, err := GenerateAvatarKey()
		assert.NoError(t, err, "not wait error")
	})
}
