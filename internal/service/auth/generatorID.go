package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GenerateSessionID() (string, error) {
	buffer := make([]byte, 32)

	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("cannot generate sessinId: %w", err)
	}

	return hex.EncodeToString(buffer), nil
}
