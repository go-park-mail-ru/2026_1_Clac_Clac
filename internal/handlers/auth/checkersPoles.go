package handlers

import (
	"errors"
	"net/mail"
)

var (
	ErrorLenPassword     = errors.New("password must contain minimum 6")
	ErrorIncorrectEmail  = errors.New("invalid email format")
	ErrorIncorrectSymbol = errors.New("allowed only a-z, A-Z, 0-9, and /?!@")
)

func CheckAsciiSymbol(strings ...string) bool {
	for _, str := range strings {
		for _, symbol := range str {
			if symbol > 127 {
				return false
			}
		}
	}

	return true
}

func CheckEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
