package service

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GeneratorCode() (string, error) {
	max := big.NewInt(1000000)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", n.Int64()), nil
}

func CreatorResetKey(tokenID string) string {
	return fmt.Sprintf("reset_token:%s", tokenID)
}
