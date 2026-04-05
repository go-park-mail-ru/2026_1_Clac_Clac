package handler

import (
	"errors"
	"net/mail"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
)

var (
	ErrorLenPassword         = errors.New("password must contain minimum 8 and maximum 128 symbols")
	ErrorIncorrectEmail      = errors.New("invalid email format")
	ErrorIncorrectSymbol     = errors.New("allowed only a-z, A-Z, 0-9, and /?!@")
	ErrorDifferencePasswords = errors.New("passwords don't match")
)

func ValidatorWithCheckPassword(email, password, repeatedPassword string) error {
	if password != repeatedPassword {
		return ErrorDifferencePasswords
	}

	return ValidatorRequestAuth(email, password)
}

func ValidatorRequestAuth(email, password string) error {
	correctSymbols := common.CheckAsciiSymbol(email, password)
	if !correctSymbols {
		return ErrorIncorrectSymbol
	}

	if len(password) < 8 || len(password) > 128 {
		return ErrorLenPassword
	}

	correctEmail := ValidateEmail(email)
	if !correctEmail {
		return ErrorIncorrectEmail
	}

	return nil
}

func ValidatorRequestNewPassword(password, repeatedPassword string) error {
	if password != repeatedPassword {
		return ErrorDifferencePasswords
	}

	correctSymbols := common.CheckAsciiSymbol(password)
	if !correctSymbols {
		return ErrorIncorrectSymbol
	}

	if len(password) < 8 || len(password) > 128 {
		return ErrorLenPassword
	}

	return nil
}

func ValidateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
