package usecase

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func GenerateRandomCSRFToken() (string, error) {
	const tokenLength = 32

	b := make([]byte, tokenLength)

	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("rand.Read: %w", err)
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
