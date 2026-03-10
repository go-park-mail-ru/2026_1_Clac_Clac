package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

func GenerateSessionID() (string, error) {
	buffer := make([]byte, 32)

	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("cannot generate sessinId: %w", err)
	}

	return hex.EncodeToString(buffer), nil
}

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

func GeneratorCode() (string, error) {
	max := big.NewInt(1000000)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", n.Int64()), nil
}
