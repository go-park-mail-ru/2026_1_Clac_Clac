package auth

import (
	"errors"
	"net/mail"
)

var (
	ErrorLenPassword         = errors.New("password must contain minimum 6")
	ErrorIncorrectEmail      = errors.New("invalid email format")
	ErrorIncorrectSymbol     = errors.New("allowed only a-z, A-Z, 0-9, and /?!@")
	ErrorDifferencePasswords = errors.New("passwords don't match")
)

func ValidatorRegistraionRequest(email, password, repeatedPassword string) error {
	if password != repeatedPassword {
		return ErrorDifferencePasswords
	}

	return ValidatorRequestAuth(email, password)
}

func ValidatorRequestAuth(email, password string) error {
	correctSymbols := checkAsciiSymbol(email, password)
	if !correctSymbols {
		return ErrorIncorrectSymbol
	}

	if len(password) < 6 {
		return ErrorLenPassword
	}

	correctEmail := checkEmail(email)
	if !correctEmail {
		return ErrorIncorrectEmail
	}

	return nil
}

func checkAsciiSymbol(strings ...string) bool {
	for _, str := range strings {
		for _, symbol := range str {
			if symbol > 127 {
				return false
			}
		}
	}

	return true
}

func checkEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
