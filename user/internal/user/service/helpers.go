package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrorCreateHash    = errors.New("failed to cresate hash")
	ErrorWrongPassword = errors.New("write wrong password")
)

func HashPassword(password string) (string, error) {
	sha256Hash := sha256.Sum256([]byte(password))
	hashString := hex.EncodeToString(sha256Hash[:])

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(hashString), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrorCreateHash, err)
	}

	return string(hashedPassword), nil
}

func CheckPassword(inputPassword, hashPassword string) error {
	sha256Hash := sha256.Sum256([]byte(inputPassword))
	inputHashString := hex.EncodeToString(sha256Hash[:])

	if err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(inputHashString)); err != nil {
		return ErrorWrongPassword
	}

	return nil
}

func GenerateAvatarKey() (string, error) {
	return uuid.New().String(), nil
}
