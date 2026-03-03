package service

import "fmt"

func SpyHasherError(password string) (string, error) {
	return "", fmt.Errorf("%w: %q", ErrorCreateHash, "error bcrypt")
}

func SpyHasher(password string) (string, error) {
	return "hash_" + password, nil
}

func SpyGenerator() (string, error) {
	return "sessionCLAC", nil
}

func SpyChecker(inputPassword, hashPassword string) error {
	if inputPassword == hashPassword {
		return nil
	}

	return ErrorWrongPassword
}
