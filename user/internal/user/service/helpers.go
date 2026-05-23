package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const myCost = 8

var (
	ErrorCreateHash    = errors.New("failed to cresate hash")
	ErrorWrongPassword = errors.New("write wrong password")
)

func HashPassword(password string) (string, error) {
	sha256Hash := sha256.Sum256([]byte(password))
	hashString := hex.EncodeToString(sha256Hash[:])

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(hashString), myCost)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrorCreateHash, err)
	}

	return string(hashedPassword), nil
}

func CheckPassword(inputPassword, hashPassword string) (string, error) {
	sha256Hash := sha256.Sum256([]byte(inputPassword))
	inputHashString := hex.EncodeToString(sha256Hash[:])

	if err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(inputHashString)); err != nil {
		return "", ErrorWrongPassword
	}

	cost, err := bcrypt.Cost([]byte(hashPassword))
	if err != nil || cost > myCost {
		newHash, err := HashPassword(inputPassword)
		if err != nil {
			return "", fmt.Errorf("re-hash: %w", err)
		}
		return newHash, nil
	}

	return "", nil
}

func GenerateAvatarKey() (string, error) {
	return uuid.New().String(), nil
}
