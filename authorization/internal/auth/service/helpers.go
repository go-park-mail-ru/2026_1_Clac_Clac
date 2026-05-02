package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GenerateSessionKey() (string, error) {
	buffer := make([]byte, 32)

	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("cannot generate sessionId: %w", err)
	}

	return hex.EncodeToString(buffer), nil
}

func CreateSessionKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}
