package auth

import "fmt"

func spyHasherError(password string) (string, error) {
	return "", fmt.Errorf("%w: %q", ErrorCreateHash, "error bcrypt")
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

	return ErrorWrongPassword
}
