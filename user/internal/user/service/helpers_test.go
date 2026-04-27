package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	t.Run("success hash password", func(t *testing.T) {
		password := "my_secret_password"

		hash, err := HashPassword(password)
		assert.NoError(t, err, "expected no error while hashing")
		assert.NotEmpty(t, hash, "hash should not be empty")
	})

	t.Run("two hashes for same password differ", func(t *testing.T) {
		password := "same_password"

		hash1, err := HashPassword(password)
		assert.NoError(t, err)
		hash2, err := HashPassword(password)
		assert.NoError(t, err)

		assert.NotEqual(t, hash1, hash2, "bcrypt should generate unique salted hashes")
	})
}

func TestCheckPassword(t *testing.T) {
	t.Run("success check password", func(t *testing.T) {
		password := "correct_password"
		hash, err := HashPassword(password)
		assert.NoError(t, err)

		err = CheckPassword(password, hash)
		assert.NoError(t, err, "expected no error for correct password")
	})

	t.Run("error wrong password", func(t *testing.T) {
		password := "correct_password"
		hash, err := HashPassword(password)
		assert.NoError(t, err)

		err = CheckPassword("wrong_password", hash)
		assert.EqualError(t, err, ErrorWrongPassword.Error(), "expected wrong password error")
	})
}

func TestGenerateAvatarKey(t *testing.T) {
	t.Run("success generate avatar key", func(t *testing.T) {
		key, err := GenerateAvatarKey()
		assert.NoError(t, err, "expected no error generating key")
		assert.NotEmpty(t, key, "key should not be empty")
	})

	t.Run("generated keys are unique", func(t *testing.T) {
		key1, err := GenerateAvatarKey()
		assert.NoError(t, err)
		key2, err := GenerateAvatarKey()
		assert.NoError(t, err)

		assert.NotEqual(t, key1, key2, "avatar keys must be unique")
	})
}
