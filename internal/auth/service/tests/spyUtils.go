package tests

import (
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service"
)

func spyHasherError(password string) (string, error) {
	return "", fmt.Errorf("%w: %q", service.ErrorCreateHash, "error bcrypt")
}

func spyHasher(password string) (string, error) {
	return "hash_" + password, nil
}

func spyGenerator() (string, error) {
	return "sessionCLAC", nil
}

func spyChecker(inputPassword, hashPassword string) error {
	if inputPassword == hashPassword {
		return nil
	}

	return service.ErrorWrongPassword
}
