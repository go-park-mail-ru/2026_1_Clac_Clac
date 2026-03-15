package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	service "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service"
)

func TestHashPassword(t *testing.T) {
	t.Run("sucess hash password", func(t *testing.T) {
		password := "my_secret_password"

		hash1, err := service.HashPassword(password)
		assert.NoError(t, err, "expected no error while hashing")
		assert.NotEmpty(t, hash1, "hash should not be empty")

		hash2, err := service.HashPassword(password)
		assert.NoError(t, err)
		assert.NotEmpty(t, hash2)
		assert.NotEqual(t, hash1, hash2, "bcrypt should generate unique hashes for the same password")
	})
}

func TestGenerateSessionID(t *testing.T) {
	t.Run("sucess generate session id", func(t *testing.T) {
		id1, err := service.GenerateSessionID()
		assert.NoError(t, err, "expected no error while generating session ID")
		assert.Equal(t, 64, len(id1), "hex encoded array should be 64 characters long")

		id2, err := service.GenerateSessionID()
		assert.NoError(t, err)
		assert.NotEqual(t, id1, id2, "generated sessionID should be unique")
	})
}

func TestCheckPassword(t *testing.T) {
	t.Run("sucess check passsword", func(t *testing.T) {
		newPassword := "newPassword"
		hashNewPassword, err := service.HashPassword(newPassword)
		assert.NoError(t, err, "expected no error while creating password hash")

		inputPassword := "newPassword"
		err = service.CheckPassword(inputPassword, hashNewPassword)

		assert.Nil(t, err, "expected passwords must be same")
	})
}

func TestCheckPasswordError(t *testing.T) {
	t.Run("check passsword error", func(t *testing.T) {
		newPassword := "newPassword"
		hashNewPassword, err := service.HashPassword(newPassword)
		assert.NoError(t, err, "expected no error while creating password hash")

		inputPassword := "inputPassword"
		err = service.CheckPassword(inputPassword, hashNewPassword)
		assert.EqualError(t, err, service.ErrorWrongPassword.Error(), "expected error for wrong password")
	})
}
func TestGeneratorCode(t *testing.T) {
	t.Run("generator reset token", func(t *testing.T) {
		code1, err := service.GenerateSessionID()
		assert.NoError(t, err, "expected no error while generating session ID")
		assert.Equal(t, 64, len(code1), "hex encoded array should be 64 characters long")

		code2, err := service.GenerateSessionID()
		assert.NoError(t, err)
		assert.NotEqual(t, code1, code2, "generated sessionID should be unique")
	})
}
