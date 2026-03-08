package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	password := "my_secret_password"

	hash1, err := HashPassword(password)
	assert.NoError(t, err, "expected no error while hashing")
	assert.NotEmpty(t, hash1, "hash should not be empty")

	hash2, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash2)
	assert.NotEqual(t, hash1, hash2, "bcrypt should generate unique hashes for the same password")
}

func TestGenerateSessionID(t *testing.T) {
	id1, err := GenerateSessionID()
	assert.NoError(t, err, "expected no error while generating session ID")
	assert.Equal(t, 64, len(id1), "hex encoded array should be 64 characters long")

	id2, err := GenerateSessionID()
	assert.NoError(t, err)
	assert.NotEqual(t, id1, id2, "generated sessionID should be unique")
}

func TestCheckPassword(t *testing.T) {
	newPassword := "newPassword"
	hashNewPassword, err := HashPassword(newPassword)
	assert.NoError(t, err, "expected no error while creating password hash")

	inputPassword := "newPassword"
	err = CheckPassword(inputPassword, hashNewPassword)

	assert.Nil(t, err, "expected passwords must be same")
}

func TestCheckPasswordError(t *testing.T) {
	newPassword := "newPassword"
	hashNewPassword, err := HashPassword(newPassword)
	assert.NoError(t, err, "expected no error while creating password hash")

	inputPassword := "inputPassword"
	err = CheckPassword(inputPassword, hashNewPassword)
	assert.EqualError(t, err, ErrorWrongPassword.Error(), "expected error for wrong password")
}
